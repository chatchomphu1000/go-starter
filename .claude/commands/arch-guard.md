# /arch-guard — Hexagonal Architecture Violation Scanner

Deep-scan the entire codebase for architecture violations and produce a violation report.
No arguments needed.

## What to check (run ALL checks, collect results, report at end)

### RULE 1: Core must not import framework packages
Scan `internal/core/**/*.go` for any of these imports:
- `github.com/labstack/echo`
- `go.mongodb.org/mongo-driver`
- `go.uber.org/zap`
- `github.com/go-resty/resty`
- `github.com/spf13/viper`
- `github.com/spf13/cobra`
- `github.com/swaggo`

```bash
grep -rn "\"github.com/labstack/echo\|go.mongodb.org/mongo-driver\|go.uber.org/zap\|go-resty/resty\|spf13/viper\|spf13/cobra\|swaggo" internal/core/ 2>/dev/null
```

### RULE 2: No `time.Now()` in core services (must use Clock port)
```bash
grep -rn "time\.Now()" internal/core/ 2>/dev/null
```

### RULE 3: No direct UUID generation in core (must use IDGenerator port)
```bash
grep -rn "uuid\.New()\|uuid\.NewRandom()\|uuid\.NewString()" internal/core/ 2>/dev/null
```

### RULE 4: No bson/mongo types leaking outside mongodb adapter package
```bash
grep -rn "bson\.\|mongo\.\|primitive\." internal/core/ internal/adapters/inbound/ 2>/dev/null
grep -rn "\"go.mongodb.org" internal/adapters/inbound/ 2>/dev/null
```

### RULE 5: Adapters must not import each other
```bash
grep -rn "\"github.com/chatchomphu1000/go-starter/internal/adapters/outbound/mongodb" internal/adapters/outbound/httpclient/ internal/adapters/outbound/jwtissuer/ internal/adapters/outbound/crypto/ 2>/dev/null
grep -rn "\"github.com/chatchomphu1000/go-starter/internal/adapters/outbound/httpclient" internal/adapters/outbound/mongodb/ internal/adapters/outbound/jwtissuer/ internal/adapters/outbound/crypto/ 2>/dev/null
```

### RULE 6: Handlers must contain no business logic
Check handlers for patterns that indicate logic leaking in:
```bash
grep -n "if.*password\|bcrypt\.\|jwt\.\|mongo\.\|bson\.\|hash\." internal/adapters/inbound/http/handler/*.go 2>/dev/null
```
Also check for direct repo usage in handlers:
```bash
grep -n "Repository\|\.Insert(\|\.FindBy\|\.Delete(" internal/adapters/inbound/http/handler/*.go 2>/dev/null
```

### RULE 7: No `os.Exit` outside main.go and cmd/
```bash
grep -rn "os\.Exit" internal/ pkg/ 2>/dev/null
```

### RULE 8: No `panic()` outside main.go
```bash
grep -rn "panic(" internal/ pkg/ 2>/dev/null | grep -v "_test\.go"
```

### RULE 9: No naked error returns (check for `return err` without wrapping fmt.Errorf)
```bash
grep -rn "^\s*return err$\|^\s*return nil, err$" internal/core/services/ 2>/dev/null
```
Flag if found — these should be `fmt.Errorf("...: %w", err)`.

### RULE 10: No secrets or sensitive values hardcoded
```bash
grep -rni "password\s*=\s*\"[^\"]\+\"\|secret\s*=\s*\"[^\"]\+\"\|token\s*=\s*\"[^\"]\+" internal/ cmd/ main.go 2>/dev/null | grep -v "_test\.go\|example\|placeholder\|change-me"
```

### RULE 11: Context as first param in all I/O functions
Check for I/O functions (Insert, Find, Update, Delete, Issue, Verify, Send) missing ctx:
```bash
grep -n "^func.*Insert(\|^func.*Find\|^func.*Update(\|^func.*Delete(\|^func.*Issue(\|^func.*Verify(\|^func.*Send(" internal/ -r 2>/dev/null | grep -v "ctx context.Context"
```

### RULE 12: No domain errors defined outside domain package
```bash
grep -rn "ErrUser\|ErrEmail\|ErrInvalid\|ErrWeak" internal/core/services/ internal/adapters/ 2>/dev/null | grep "errors\.New\|fmt\.Errorf"
```

---

## Output format

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ARCHITECTURE GUARD REPORT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Rule 1: Core framework imports   [ PASS | N violations ]
  Rule 2: time.Now() in core       [ PASS | N violations ]
  Rule 3: Direct UUID in core      [ PASS | N violations ]
  Rule 4: bson/mongo leaks         [ PASS | N violations ]
  Rule 5: Adapter cross-imports    [ PASS | N violations ]
  Rule 6: Business logic in handlers [ PASS | N violations ]
  Rule 7: os.Exit in core/pkg      [ PASS | N violations ]
  Rule 8: naked panic              [ PASS | N violations ]
  Rule 9: Unwrapped errors in svc  [ PASS | N violations ]
  Rule 10: Hardcoded secrets       [ PASS | N violations ]
  Rule 11: Context as first param  [ PASS | N violations ]
  Rule 12: Domain errors outside domain [ PASS | N violations ]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Total violations: N
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

For each violation, print: `[RULE N] file:line — description and how to fix`

If zero violations: "Architecture is clean. Hexagonal boundaries intact."
