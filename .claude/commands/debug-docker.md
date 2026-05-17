# /debug-docker — Docker environment diagnostics and log inspector

Diagnose the Docker Compose stack and surface actionable information fast.
Optional argument: `$ARGUMENTS` — service name to focus on (`app` or `mongo`), or empty for all.

## Step 1: Stack status
```bash
docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}" 2>&1
```

## Step 2: Recent logs (last 100 lines per service)
```bash
docker compose logs --tail=100 --no-log-prefix app 2>&1 | head -80
```
```bash
docker compose logs --tail=50 --no-log-prefix mongo 2>&1 | head -40
```

Highlight in output:
- Lines containing `ERROR`, `FATAL`, `panic`, `exception` → flag as CRITICAL
- Lines containing `WARN` → flag as WARNING
- Lines containing `connection refused`, `timeout`, `no such host` → flag as NETWORK ISSUE

## Step 3: MongoDB connectivity check
```bash
docker compose exec mongo mongosh --quiet --eval "db.runCommand({ping:1})" 2>&1
```
PASS: returns `{ ok: 1 }`
FAIL: report connection issue with hint

## Step 4: App health endpoint
```bash
APP_PORT=$(docker compose port app 8080 2>/dev/null | cut -d: -f2 || echo "8080")
curl -sf "http://localhost:${APP_PORT}/health" 2>&1
curl -sf "http://localhost:${APP_PORT}/ready" 2>&1
```

## Step 5: Container resource usage
```bash
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}" 2>&1
```

## Step 6: Environment variable audit (security check)
```bash
docker compose exec app env 2>/dev/null | grep -E "^APP_" | grep -v "SECRET\|PASSWORD\|JWT" | sort
```
Print non-sensitive env vars. Mask any that contain SECRET/PASSWORD/JWT (show key name only, value as `***`).

## Step 7: Port conflicts (if app fails to start)
```bash
lsof -i :8080 2>/dev/null | head -5
lsof -i :27017 2>/dev/null | head -5
```

---

## Output format

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  DOCKER DIAGNOSTICS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Services:
    app    [ running | stopped | restarting ]  :8080
    mongo  [ running | stopped ]               :27017

  Health:
    /health  [ OK | FAIL ]
    /ready   [ OK | FAIL ]
    MongoDB  [ OK | FAIL ]

  Resources:
    app    CPU: X%  MEM: X/XMiB (X%)
    mongo  CPU: X%  MEM: X/XMiB (X%)

  Issues detected: N
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

Then print:
- `CRITICAL:` lines from logs (with timestamp context)
- `WARNING:` lines from logs
- `NETWORK:` connection-related errors
- Recommended fix for each issue detected

Common fixes to suggest:
- App won't start → `make docker-down && make docker-up`
- Port conflict → `lsof -i :8080 | kill -9 <pid>`
- Mongo not healthy → `docker compose restart mongo`
- Migration pending → `make migrate-up`
- Build error → `make docker-build`
