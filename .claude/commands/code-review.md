# /code-review — Go-specific code review for this project

Review the code changes on the current branch (or `$ARGUMENTS` if a file/path is given).
Produce a structured review graded by severity.

## What to review

If `$ARGUMENTS` is empty: `git diff main...HEAD -- '*.go'`
If `$ARGUMENTS` is a path: read those files directly.

---

## Review checklist (check ALL — collect findings, report at end)

### [A] Hexagonal Architecture
- [ ] `internal/core/**` imports nothing from Echo, MongoDB, Zap, Viper, Cobra, resty, swaggo
- [ ] Domain structs have no bson/json framework tags
- [ ] No `time.Now()` or `uuid.New()` called directly in `internal/core/services/`
- [ ] Handlers do zero business logic — only: bind → validate → call service → map response
- [ ] Adapters don't import each other (mongodb ↔ httpclient ↔ jwtissuer)
- [ ] No `bson.*`, `mongo.*` types leak outside `internal/adapters/outbound/mongodb/`

### [B] Error handling
- [ ] Every error in `internal/core/services/` is wrapped: `fmt.Errorf("<service>.<Method>: %w", err)`
- [ ] Domain sentinel errors used correctly (`domain.ErrUserNotFound` etc.) — not raw `errors.New` at call sites
- [ ] `errors.Is` / `errors.As` used at error boundaries (not string comparison)
- [ ] No swallowed errors (`_ = someFunc()` or blank `if err != nil {}`)
- [ ] HTTP handler errors returned to Echo's central error handler, not manually written to response

### [C] Context discipline
- [ ] Every I/O function has `ctx context.Context` as first parameter
- [ ] `ctx.Err()` checked at entry of long operations
- [ ] `context.WithoutCancel(ctx)` used for background fire-and-forget (e.g. welcome email)
- [ ] No `context.Background()` used inside handlers (only acceptable at app startup)

### [D] Idiomatic Go
- [ ] No naked returns in functions longer than 3 lines
- [ ] Interfaces defined where consumed (ports), not where implemented
- [ ] Exported identifiers have Go doc comments
- [ ] No unexported struct fields with capital letters
- [ ] `defer` used correctly — no `defer` inside loops
- [ ] Goroutines have explicit lifetime management and cancellation
- [ ] No `init()` functions (except in `main.go` if absolutely necessary)
- [ ] Slice pre-allocation with `make([]T, 0, n)` where n is known
- [ ] Named return values only when they meaningfully document the return

### [E] Logging discipline
- [ ] Zap structured fields used: `zap.String("k","v")` — no `fmt.Sprintf` inside log calls
- [ ] No password, token, Authorization header, or raw request body logged anywhere
- [ ] Info level for success paths, Error for unexpected failures, Warn for 4xx
- [ ] `logger.With(fields...)` used to avoid repeating fields per-method

### [F] Validation boundary
- [ ] Input validation happens only at HTTP DTO layer (go-playground/validator tags)
- [ ] Domain trusts its inputs after value object construction (Email VO, Role.ParseRole)
- [ ] No double-validation (both DTO tag and service layer checking the same field)

### [G] Test quality (if test files changed)
- [ ] Tests are table-driven with `testify/require` for fatal assertions, `assert` for non-fatal
- [ ] Every method under test has a `context_already_cancelled` case with zero mock expectations
- [ ] Mock setup uses `.EXPECT().Method(...).Return(...)` (with-expecter style) and `AssertExpectations(t)` at the end of every test
- [ ] No `time.Sleep` in tests — use `mocks.MockClock` with `EXPECT().Now().Return(fixedNow)`
- [ ] No `time.Now()` in test files — use a fixed `var fixedNow = time.Date(...)` constant
- [ ] No network calls, no real DB, no file I/O in unit tests
- [ ] Error assertions use `errors.Is(err, target)`, not string comparison
- [ ] Service tests are `package services` (same-package) to enable testing unexported helpers
- [ ] Integration tests have `//go:build integration` build tag
- [ ] Test names follow `TestXxxService_MethodName_ScenarioDescription` pattern
- [ ] Coverage ≥ 85% for `internal/core/services/` files — use `/unit-test` command to scaffold missing tests

### [H] MongoDB adapter specifics
- [ ] `_id` stored as string (UUID v7), never `bson.ObjectID`
- [ ] Duplicate key error (11000) mapped to `domain.ErrXxxAlreadyExists`
- [ ] `mongo.ErrNoDocuments` mapped to `domain.ErrXxxNotFound`
- [ ] FindAll uses `CountDocuments` + cursor (not two full scans)
- [ ] Search filter uses escaped regex (`regexp.QuoteMeta`) — no regex injection

### [I] Code smells
- [ ] No magic numbers/strings — use named constants
- [ ] Functions longer than 50 lines (flag for review)
- [ ] Cyclomatic complexity > 10 (flag for review)
- [ ] Copy-paste code that should be extracted (3+ identical blocks)
- [ ] Unnecessary pointer indirection or unnecessary dereferencing

---

## Output format

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  CODE REVIEW  [branch: <current branch>]
  Files reviewed: N  |  Lines changed: +X -Y
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

Group findings by severity:

**[CRITICAL]** — must fix before merge (architecture violations, security bugs, data loss risk)
**[MAJOR]** — should fix before merge (incorrect behavior, missing error handling)
**[MINOR]** — fix soon (code smell, style, missing doc comment)
**[NIT]** — optional (cosmetic, personal preference)

For each finding:
```
[SEVERITY] file.go:line
  Issue: <what is wrong>
  Why:   <why it matters>
  Fix:   <concrete code suggestion>
```

End with:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Summary: N critical, N major, N minor, N nits
  Verdict: [ APPROVE | REQUEST CHANGES | DISCUSS ]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

If zero findings: "LGTM. Clean, idiomatic, architecture-compliant."
