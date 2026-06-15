# /smoke-test — API smoke test against running server

Run a full end-to-end smoke test of the API to verify core flows work.
Optional argument: base URL (default: `http://localhost:8080`)

## Pre-flight

```bash
curl -sf http://localhost:8080/health > /dev/null 2>&1 || echo "SERVER_NOT_RUNNING"
```
If server not running, stop and tell the user to run `make run` first.

---

## Test sequence

Use a unique test email per run to avoid conflicts:
`TEST_EMAIL="smoketest+$(date +%s)@example.com"`
`TEST_PASSWORD="Smoke@Test123!"`

### [1/7] Health check
```bash
curl -s -w "\n%{http_code}" http://localhost:8080/health
curl -s -w "\n%{http_code}" http://localhost:8080/ready
```
PASS: both return 200

### [2/7] Readiness (mongo connected)
PASS: `/ready` returns `{"status":"ok"}` with 200

### [3/8] Register a new user
```bash
curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Smoke Test\",\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}"
```
PASS: 201, extract and store `id` from response

### [4/8] Register duplicate → expect 409 Conflict
```bash
curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Smoke Test\",\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}"
```
PASS: 409 with code `CONFLICT`

### [5/8] Login
```bash
curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}"
```
PASS: 200, extract `access_token` and `refresh_token` → store as `TOKEN` and `REFRESH_TOKEN`

### [6/8] Login with wrong password → expect 401
```bash
curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"WrongPassword1!\"}"
```
PASS: 401 with code `UNAUTHORIZED`

### [7/8] Refresh token
```bash
curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}"
```
PASS: 200, response contains new `access_token`

### [8/8] Get own profile (authenticated)
```bash
curl -s -w "\n%{http_code}" http://localhost:8080/api/v1/users/$USER_ID \
  -H "Authorization: Bearer $TOKEN"
```
PASS: 200, id matches, hashedPassword NOT present in response

### [BONUS] Unauthenticated access → expect 401
```bash
curl -s -w "\n%{http_code}" http://localhost:8080/api/v1/users/$USER_ID
```
PASS: 401

---

## Output format

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  SMOKE TEST REPORT  [http://localhost:8080]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  [1] GET  /health                  [ PASS | FAIL ] 200
  [2] GET  /ready                   [ PASS | FAIL ] 200
  [3] POST /api/v1/auth/register          [ PASS | FAIL ] 201
  [4] POST /auth/register (duplicate)     [ PASS | FAIL ] 409
  [5] POST /api/v1/auth/login             [ PASS | FAIL ] 200
  [6] POST /auth/login (wrong pwd)        [ PASS | FAIL ] 401
  [7] POST /api/v1/auth/refresh           [ PASS | FAIL ] 200
  [8] GET  /api/v1/users/:id              [ PASS | FAIL ] 200
  [B] GET  /api/v1/users/:id (no auth)    [ PASS | FAIL ] 401
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Result: N/9 passed
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

For each failure, print the actual response body to help debug.

**Security assertion**: verify the register/login/get-user responses do NOT contain the string `password` (raw or hashed) anywhere in the body. Flag it if found.
