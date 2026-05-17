# /perf-review — Performance review for this Go + MongoDB project

Analyze the current branch changes (or `$ARGUMENTS` path) for performance issues.
Produce findings with measured impact estimates.

## What to analyze

If `$ARGUMENTS` is empty: `git diff main...HEAD -- '*.go'`
If `$ARGUMENTS` is a path: read those files.
Also read the referenced source files to understand full context of each change.

---

## Performance checks

### [1] MongoDB query efficiency

Scan `internal/adapters/outbound/mongodb/`:

**N+1 queries**
- Any loop that calls `repo.FindByID` or `collection.FindOne` inside it?
- Flag: fetch in bulk with `$in` query instead.

**Missing index usage**
- `bson.D` / `bson.M` filter keys — cross-check against `migrations/` JSON files.
- Filter on `name`, `role`, `active`, `email` must have backing index.
- Flag any filter field not found in migration indexes.

**Unbounded queries**
- `FindAll` without `Limit` set on cursor options?
- `CountDocuments` called separately from `Find` — can sometimes be replaced with `estimatedDocumentCount` for non-filtered counts.

**Regex performance**
- Case-insensitive regex search (`/pattern/i`) without a text index = full collection scan.
- Flag if Search filter uses `$regex` on a large collection without a text index.

**Projection missing**
- `Find` fetching entire document when only a few fields are needed?
- Look for cases where only `id` or `email` is used from a full user doc.

**Write amplification**
- `$set` updates — are all modified fields included, or is the whole document being replaced?

### [2] Context timeout discipline

Scan all adapter files for:
```bash
grep -n "context\.Background()\|context\.TODO()" internal/adapters/ 2>/dev/null
```
Flag: adapter code using Background() instead of propagating the request context — this defeats per-request deadline enforcement.

```bash
grep -n "time\.Second\*[0-9]\|time\.Minute\*[0-9]" internal/adapters/outbound/mongodb/ 2>/dev/null
```
Hardcoded timeouts that ignore config — flag for config-driven values.

### [3] HTTP server / Echo performance

**Response body size**
- List endpoints — any returning full user list without pagination enforced?
- `Limit` clamped to [1, 100] in ListFilter — verify enforcement in service.

**Middleware overhead**
- Middleware on health/ready endpoints? These should skip all middleware except Recover.
- Rate limiter applied to health check? Unnecessary overhead.

**Body reading**
- Any handler reading `c.Request().Body` more than once? (Body is a stream — second read = empty)

**JSON marshaling**
- Large response structs with many unused fields? Consider projection or response-specific DTOs.

### [4] Connection pool and resty client

Scan `internal/adapters/outbound/`:

**MongoDB pool**
- `MinPool` / `MaxPool` values from config — are they used in `client.go`?
- Missing `SetMinPoolSize` or `SetMaxPoolSize` in options?

**Resty client**
- Single shared `*resty.Client` instance (correct) vs new client per request (expensive)?
- `SetRetryCount` + `SetRetryWaitTime` configured? Unbounded retries = cascading failure.
- `SetTimeout` set? Missing = goroutine leak on hanging external calls.

### [5] Goroutine management

```bash
grep -n "go func\|go [a-z]" internal/ -r 2>/dev/null | grep -v "_test\.go"
```

For each goroutine spawn:
- Is there a done channel, WaitGroup, or context cancellation to bound its lifetime?
- Fire-and-forget for welcome email is intentional — but it should still log on completion.
- Any goroutine that could panic without a recover?

### [6] Memory allocation hotspots

**Slice growth**
- `append` in a loop without pre-allocated capacity?
  ```bash
  grep -n "append(" internal/adapters/outbound/mongodb/ 2>/dev/null
  ```
  Flag if inside a cursor loop without `make([]T, 0, estimatedSize)`.

**String building**
- `+` string concatenation in loops? Use `strings.Builder`.

**Interface boxing**
- Passing concrete types as `interface{}` in hot paths (e.g. inside cursor loops)?

### [7] Service layer efficiency

- `Register`: `ExistsByEmail` + `Insert` = 2 round trips. Could use upsert with unique index and catch 11000 in one round trip. Flag as optimization opportunity.
- `Update`: `FindByID` + `Update` = 2 round trips. Could use `findOneAndUpdate` for atomic read-modify. Flag.
- `Delete`: `FindByID` + `Delete` = 2 round trips. Same optimization applies. Flag.

### [8] JWT verification overhead

Scan `internal/adapters/outbound/jwtissuer/`:
- `Verify` called on EVERY authenticated request. Is result cached? (in-memory LRU with TTL could eliminate crypto overhead for repeated calls within JWT validity window)
- Flag as potential optimization if traffic is high.

---

## Output format

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  PERFORMANCE REVIEW  [branch: <current branch>]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

Group by impact:

**[HIGH IMPACT]** — measurable latency / throughput effect at production scale
**[MEDIUM IMPACT]** — noticeable under load, worth fixing before scale
**[LOW IMPACT]** — micro-optimization or architectural suggestion
**[INFO]** — observation, not a problem, but worth knowing

For each finding:
```
[IMPACT] category — file.go:line
  Issue:    <what is slow and why>
  Evidence: <the specific code pattern>
  Fix:      <concrete code or approach>
  Estimate: <rough impact — e.g. "saves 1 RTT per write, ~2-5ms p99 improvement">
```

End with:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  N high-impact, N medium, N low, N info items
  Current bottleneck estimate: <DB layer | HTTP | CPU | N/A>
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```
