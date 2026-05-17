# /health — Full project quality gate

Run the complete quality pipeline and report a structured summary.
No arguments needed.

## Steps (run in this exact order)

### 1. Build check
```bash
go build ./...
```
Report: PASS / FAIL with error output

### 2. Static analysis
```bash
go vet ./...
```
Report: PASS / FAIL

### 3. Unit tests (short mode — no integration)
```bash
go test ./... -race -count=1 -short -coverprofile=/tmp/go-starter-cov.out 2>&1
```
Report: PASS / FAIL, number of packages tested, test count

### 4. Coverage threshold
```bash
go tool cover -func=/tmp/go-starter-cov.out | tail -1
```
PASS if total >= 80%, WARN if 60–79%, FAIL if < 60%

### 5. Lint (if golangci-lint is installed)
```bash
golangci-lint run ./... --timeout=120s 2>&1 | head -50
```
Skip gracefully if not installed (note it in report).

### 6. Architecture guard (quick scan)
Check for forbidden imports in core layer:
```bash
grep -r "\"github.com/labstack/echo" internal/core/ 2>/dev/null
grep -r "\"go.mongodb.org/mongo-driver" internal/core/ 2>/dev/null
grep -r "\"go.uber.org/zap" internal/core/ 2>/dev/null
grep -r "time\.Now()" internal/core/services/ 2>/dev/null
grep -r "uuid\.New()" internal/core/services/ 2>/dev/null
```
Report any hits as ARCH VIOLATION.

### 7. Docker status (non-fatal)
```bash
docker compose ps 2>/dev/null
```
Report which services are running/stopped.

---

## Output format

Print a clean summary table:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  GO-STARTER HEALTH REPORT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Build         [ PASS | FAIL ]
  Vet           [ PASS | FAIL ]
  Tests         [ PASS | FAIL ]  (N pkgs, N tests)
  Coverage      [ XX.X% ]        (threshold: 80%)
  Lint          [ PASS | FAIL | SKIPPED ]
  Arch Guard    [ PASS | N violations ]
  Docker        [ app=running mongo=running | ... ]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Overall: [ ALL GREEN | ISSUES FOUND ]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

After the table, list any failures with actionable fix hints.
If all pass, say "Ready to ship." and nothing else.
