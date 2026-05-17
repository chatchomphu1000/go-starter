# /pr-ready — Pre-PR quality gate

Run every check required before opening a PR. Produce a go/no-go verdict.
No arguments needed.

## Steps (in order — do not skip on failure, collect all results)

### 1. Format check
```bash
gofmt -l . 2>/dev/null
```
PASS if no output. If files listed, run `gofmt -w .` to fix and note which files changed.

```bash
goimports -l . 2>/dev/null
```
PASS if no output. Fix automatically if goimports is installed.

### 2. Go vet
```bash
go vet ./...
```
PASS / FAIL with output.

### 3. Build
```bash
go build ./...
```
PASS / FAIL — build failures are blocking.

### 4. Mock freshness check
```bash
go run github.com/vektra/mockery/v2 --dry-run 2>/dev/null | head -20
```
If mockery is not installed, note it. If mocks are stale, run `make mocks` to regenerate.

### 5. Swagger doc freshness
```bash
swag init -g main.go --output docs --parseDependency --parseInternal --quiet 2>/dev/null
git diff --name-only docs/ 2>/dev/null
```
If docs/ has uncommitted changes after regen, note "Swagger docs are stale — commit the updated docs/".

### 6. Full test suite with coverage
```bash
go test ./... -race -count=1 -coverprofile=/tmp/pr-ready-cov.out 2>&1
go tool cover -func=/tmp/pr-ready-cov.out | tail -1
```
PASS if all tests pass and coverage >= 80%.
Show coverage per package for packages below 80%.

### 7. Architecture guard (abbreviated)
```bash
grep -rn "\"github.com/labstack/echo\|go.mongodb.org\|go.uber.org/zap" internal/core/ 2>/dev/null
grep -rn "time\.Now()\|uuid\.New()" internal/core/services/ 2>/dev/null
```
FAIL if any hits found.

### 8. Lint
```bash
golangci-lint run ./... --timeout=120s 2>&1
```
Skip gracefully if not installed. Otherwise PASS / FAIL.

### 9. Secret leak scan
```bash
git diff main...HEAD -- . | grep -iE "^\+.*(password|secret|api_key|apikey|private_key)\s*=\s*['\"][^'\"]{8,}" 2>/dev/null | grep -v "example\|placeholder\|change-me\|_test\.go"
```
FAIL if any matches — these indicate potential hardcoded secrets in the diff.

### 10. Diff summary
```bash
git diff --stat main...HEAD 2>/dev/null
git log --oneline main...HEAD 2>/dev/null
```
Print the list of changed files and commit history since main.

---

## Output format

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  PR READINESS REPORT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  1. Format (gofmt/goimports)   [ PASS | FIXED | FAIL ]
  2. go vet                     [ PASS | FAIL ]
  3. Build                      [ PASS | FAIL ]
  4. Mocks freshness            [ PASS | STALE | SKIPPED ]
  5. Swagger freshness          [ PASS | STALE | SKIPPED ]
  6. Tests + Coverage           [ PASS | FAIL ]  (XX.X%)
  7. Architecture guard         [ PASS | FAIL ]
  8. Lint                       [ PASS | FAIL | SKIPPED ]
  9. Secret leak scan           [ PASS | FAIL ]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Changed files: N  |  Commits ahead of main: N
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  VERDICT:  [ READY TO PR | BLOCKING ISSUES FOUND ]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

If BLOCKING ISSUES FOUND, list each issue with an actionable fix command.
If READY TO PR, suggest a PR title based on the commits and nothing else.
