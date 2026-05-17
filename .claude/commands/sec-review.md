# /sec-review — Security review tailored for this Go + Echo + MongoDB + JWT project

Deep security audit of the current branch changes (or `$ARGUMENTS` path).
This is a project-specific review — not generic OWASP. Every check is grounded in the actual tech stack.

## Scope

If `$ARGUMENTS` is empty: `git diff main...HEAD -- '*.go'` + read full source of changed files for context.
If `$ARGUMENTS` is a path: audit those files.

---

## Security checks

### [AUTH-1] JWT configuration
Read `internal/adapters/outbound/jwtissuer/hs256.go` and `internal/config/config.go`:
- Secret minimum length enforced at construction (must be ≥ 32 bytes)?
- Algorithm explicitly pinned to `HS256` — reject tokens with `alg: none` or RS256?
- `exp` claim verified on every `Verify()` call?
- `iss` claim verified against config value?
- `jti` (JWT ID) present — enables token revocation if blacklist is ever added?
- Token TTL from config (not hardcoded)?

### [AUTH-2] Auth middleware gaps
Read `internal/adapters/inbound/http/middleware/auth.go` and `router.go`:
- All sensitive routes protected by the auth middleware?
- `DELETE /users/:id` protected?
- `PUT /users/:id` protected?
- Admin-only operations further restricted by role check?
- `WWW-Authenticate` header returned on 401?
- Error message same for expired vs invalid token (prevent enumeration)?

### [AUTH-3] Password security
Read `internal/adapters/outbound/crypto/bcrypt_hasher.go` and `internal/core/services/user_service.go`:
- Bcrypt cost ≥ 12 in production? (cost < 10 is weak)
- Constant-time comparison used — `bcrypt.CompareHashAndPassword` (not `==`)?
- Weak password policy validated BEFORE hashing (saves compute on invalid input)?
- Password policy: min 10 chars + uppercase + lowercase + digit + symbol?
- `ErrInvalidCredentials` returned for BOTH wrong email AND wrong password (prevent user enumeration)?
- Plain password NEVER appears in any log call?

### [INJECT-1] MongoDB injection
Read `internal/adapters/outbound/mongodb/user_repo.go` and any other repo files:
- Search filter uses `regexp.QuoteMeta()` before building `$regex`? (without it: user input controls regex = ReDoS risk)
- Filter fields use typed values (string, bool, int) — not raw user strings inserted into `bson.D` operators?
- No raw MongoDB operator keys (`$where`, `$expr`) constructed from user input?

### [INJECT-2] HTTP request handling
Read handlers and middleware:
- `c.QueryParam()` / `c.Param()` values validated before use?
- Pagination `page` and `limit` params clamped (negative page, limit=999999)?
- `Content-Type` header enforced for POST/PUT — Echo's `Bind()` respects it?
- Request body size limited via `BodyLimit` middleware?

### [CRYPTO-1] Sensitive data exposure
Scan all `*.go` files in the diff:
```bash
# Passwords in logs
grep -n "log.*[Pp]assword\|log.*[Tt]oken\|log.*[Ss]ecret\|log.*[Aa]uthorization" .
# HashedPassword in response DTOs
grep -n "HashedPassword\|hashed_password" internal/adapters/inbound/http/dto/
# Sensitive fields in Swagger annotations
grep -n "HashedPassword\|password.*example" internal/adapters/inbound/http/handler/
```

### [CRYPTO-2] Secret management
- `.env` file in `.gitignore`?
- Secrets read from environment via Viper (not hardcoded)?
- `.env.example` contains only placeholder values (no real secrets)?
```bash
grep -n "jwt\|secret\|password\|key" .env.example | grep -v "change-me\|example\|placeholder\|your-"
```

### [RATE-LIMIT] Rate limiting configuration
Read `internal/adapters/inbound/http/middleware/ratelimit.go` and `router.go`:
- Rate limiter applied to auth endpoints (`/register`, `/login`)?
- Health/ready endpoints excluded from rate limiter?
- Rate limiter per-IP (not global) — global limiter can be weaponized for DoS?
- `RateLimitConfig.Enabled` checked — disabled in production config is a finding?
- Burst size reasonable (not 10x RPS)?

### [HEADERS] Security headers
Read `internal/adapters/inbound/http/middleware/security_headers.go`:
- `X-Content-Type-Options: nosniff` present?
- `X-Frame-Options: DENY` present?
- `Referrer-Policy: no-referrer` present?
- `Strict-Transport-Security` set when TLS enabled?
- `Content-Security-Policy` baseline set?
- `Server` header removed or obscured (don't advertise Echo version)?

### [CORS] CORS configuration
Read `internal/adapters/inbound/http/middleware/cors.go` and config:
- `AllowOrigins` is not `["*"]` when `AllowCredentials: true`?
- Origins validated against allowlist (not just prefix match)?
- Wildcard `*` origin rejected when credentials are used?

### [TIMING] Timing attacks
- Login flow: same error (`ErrInvalidCredentials`) returned for wrong email AND wrong password?
- `bcrypt.CompareHashAndPassword` called even when user not found? (dummy hash to equalize timing)
- Token verify errors normalized to same 401 response?

### [SWAGGER] API documentation security exposure
Read `internal/adapters/inbound/http/router.go`:
- Swagger UI disabled in production (`Swagger.Enabled = false` or `env == production`)?
- Swagger endpoint not protected by auth (by design) but guarded by env check?

### [DEPS] Dependency audit
```bash
go list -m -json all 2>/dev/null | grep -E '"Path"|"Version"' | head -60
```
Check if any direct dependencies have known CVEs. Flag outdated major versions.

---

## Output format

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  SECURITY REVIEW  [branch: <current branch>]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

Group by severity:

**[CRITICAL]** — exploitable vulnerability, block merge immediately
**[HIGH]** — significant risk, fix before production deploy
**[MEDIUM]** — risk present but requires specific conditions to exploit
**[LOW]** — defense-in-depth improvement, fix in next sprint
**[INFO]** — observation, hardening suggestion, not a vulnerability

For each finding:
```
[SEVERITY] category — file.go:line
  Vulnerability: <what is wrong>
  Impact:        <what an attacker can do>
  Exploit:       <how to trigger it — be specific enough to fix, not a recipe>
  Fix:           <concrete code change or config>
  Reference:     <CWE-XXX or OWASP ASVS section if applicable>
```

End with:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  N critical, N high, N medium, N low, N info
  Overall risk: [ CRITICAL | HIGH | MEDIUM | LOW | MINIMAL ]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

If zero findings above MEDIUM: "No significant vulnerabilities found in reviewed scope."
Note: this review covers code patterns — not runtime pentesting or infrastructure.
