# /new-domain — Scaffold a complete Hexagonal domain feature

Scaffold **all layers** for a new domain from a single argument: `$ARGUMENTS`

Parse `$ARGUMENTS` as the domain name in singular form (e.g. `product`, `order`, `invoice`).
Derive variants automatically:
- `singular` = as given (e.g. `product`)
- `plural`   = singular + "s" unless it already ends in s (e.g. `products`)
- `Pascal`   = PascalCase (e.g. `Product`)
- `Camel`    = camelCase (e.g. `product`)
- `snake`    = snake_case (e.g. `product`)

Module path is `github.com/chatchomphu1000/go-starter`.

---

## Files to generate — in order, NO placeholders, every file must compile

### 1. `internal/core/domain/{singular}.go`
Full domain entity mirroring `internal/core/domain/user.go` as reference:
- Struct `{Pascal}` with fields: `ID string`, `Name string`, `Active bool`, `CreatedAt time.Time`, `UpdatedAt time.Time`
- Constructor `New{Pascal}(id, name string, now time.Time) (*{Pascal}, error)`
- Methods: `Rename(name string) error`, `Activate()`, `Deactivate()`, `Touch(now time.Time)`, `Validate() error`
- No `bson.*`, no framework imports — stdlib + `pkg/apperrors` only

### 2. `internal/core/domain/errors.go` (append, do not overwrite)
Add sentinel errors: `Err{Pascal}NotFound`, `Err{Pascal}AlreadyExists`

### 3. `internal/core/ports/inbound/{singular}_service.go`
```go
type {Pascal}CreateInput struct { Name string }
type {Pascal}UpdateInput struct { Name *string }
type {Pascal}ListFilter struct {
    Active *bool
    Search string
    Page   int
    Limit  int
    SortBy string; SortDesc bool
}
type {Pascal}Service interface {
    Create(ctx context.Context, in {Pascal}CreateInput) (*domain.{Pascal}, error)
    GetByID(ctx context.Context, id string) (*domain.{Pascal}, error)
    List(ctx context.Context, f {Pascal}ListFilter) ([]*domain.{Pascal}, int64, error)
    Update(ctx context.Context, id string, in {Pascal}UpdateInput) (*domain.{Pascal}, error)
    Delete(ctx context.Context, id string) error
}
```

### 4. `internal/core/ports/outbound/{singular}_repo.go`
```go
type {Pascal}Repository interface {
    Insert(ctx context.Context, e *domain.{Pascal}) error
    FindByID(ctx context.Context, id string) (*domain.{Pascal}, error)
    FindAll(ctx context.Context, f inbound.{Pascal}ListFilter) ([]*domain.{Pascal}, int64, error)
    Update(ctx context.Context, e *domain.{Pascal}) error
    Delete(ctx context.Context, id string) error
}
```
Add `//go:generate go run github.com/vektra/mockery/v2 --name={Pascal}Repository` directive.

### 5. `internal/core/services/{singular}_service.go`
Full service implementing `{Pascal}Service` interface. Constructor-injected:
`repo {Pascal}Repository, clock Clock, ids IDGenerator, log logger.Logger`
- `Create`: validate name (non-empty), ids.New(), clock.Now(), domain.New{Pascal}(), repo.Insert()
- `GetByID`: repo.FindByID() → wrap ErrNotFound
- `List`: delegate to repo.FindAll()
- `Update`: FindByID → apply UpdateInput → Touch(clock.Now()) → repo.Update()
- `Delete`: FindByID (fail fast) → repo.Delete()
- Every method: `ctx.Err()` check at entry, `fmt.Errorf("{singular}Service.{Method}: %w", err)` wrapping, zap structured logging

### 6. `internal/adapters/outbound/mongodb/model_{singular}.go`
Internal bson model `{singular}Doc` with bson tags. `_id` is STRING (UUID v7).
Mapper functions `toDomain{Pascal}` and `from{Pascal}Domain` (unexported, package-private).

### 7. `internal/adapters/outbound/mongodb/{singular}_repo.go`
Full `mongo{Pascal}Repo` implementing `ports.{Pascal}Repository`:
- const `collection{Pascal}s = "{plural}"`
- Insert: duplicate key 11000 → wrap domain.Err{Pascal}AlreadyExists
- FindByID / FindAll / Update / Delete — mirror `user_repo.go` patterns exactly
- FindAll: build filter from ListFilter (case-insensitive regex for Search); CountDocuments for total; Skip+Limit cursor
- Update uses `$set` + UpdatedAt; matched=0 → domain.Err{Pascal}NotFound
- All errors: `fmt.Errorf("mongodb.{Pascal}Repo.{Method}: %w", err)`
- NEVER expose bson/mongo types outside the package

### 8. `internal/adapters/inbound/http/dto/{singular}.go`
```go
type {Pascal}CreateRequest struct {
    Name string `json:"name" validate:"required,min=2,max=100"`
}
type {Pascal}UpdateRequest struct {
    Name *string `json:"name" validate:"omitempty,min=2,max=100"`
}
type {Pascal}Response struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Active    bool      `json:"active"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
type {Pascal}ListResponse struct {
    Data  []*{Pascal}Response `json:"data"`
    Total int64               `json:"total"`
    Page  int                 `json:"page"`
    Limit int                 `json:"limit"`
}
func To{Pascal}Response(e *domain.{Pascal}) {Pascal}Response { ... }
```

### 9. `internal/adapters/inbound/http/handler/{singular}.go`
`{Pascal}Handler` with constructor `New{Pascal}Handler(svc ports.{Pascal}Service, log logger.Logger)`.
Handlers: `Create`, `GetByID`, `List`, `Update`, `Delete`
- Pattern: bind → validate → call service → map DTO → return
- Full Swagger annotations (`@Summary`, `@Tags`, `@Accept`, `@Produce`, `@Param`, `@Success`, `@Failure`, `@Router`)
- Zero business logic in handlers

### 10. Update `internal/adapters/inbound/http/router.go`
Add protected routes under `/api/v1/{plural}`:
```
POST   /api/v1/{plural}           → h.Create
GET    /api/v1/{plural}           → h.List
GET    /api/v1/{plural}/:id       → h.GetByID
PUT    /api/v1/{plural}/:id       → h.Update
DELETE /api/v1/{plural}/:id       → h.Delete
```

### 11. Update `.mockery.yaml`
Add under the `outbound` packages block:
```yaml
{Pascal}Repository:
```

### 14. `internal/core/services/{singular}_service_test.go`
Generate a complete table-driven unit test file following the patterns in `/unit-test`.

Required test coverage per method:
- `success` — happy path with all mocks returning valid data
- `context_already_cancelled` — `cancelledCtx()` helper, no mock expectations
- Error cases for every outbound call in the method (repo error, domain error, etc.)

```go
package services

import (
    "context"
    "errors"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"

    "github.com/chatchomphu1000/go-starter/internal/core/domain"
    "github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
    "github.com/chatchomphu1000/go-starter/internal/mocks"
    "github.com/chatchomphu1000/go-starter/pkg/logger"
)

var fixedNow = time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

func cancelledCtx() context.Context {
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    return ctx
}

func nopLogger() logger.Logger {
    l, _ := logger.NewLogger(logger.LoggerConfig{Level: "error", Format: "json"})
    return l
}

type {singular}Mocks struct {
    repo  *mocks.Mock{Pascal}Repository
    clock *mocks.MockClock
    ids   *mocks.MockIDGenerator
}

func newTest{Pascal}Service(t *testing.T) (*{singular}Service, {singular}Mocks) {
    t.Helper()
    m := {singular}Mocks{
        repo:  new(mocks.Mock{Pascal}Repository),
        clock: new(mocks.MockClock),
        ids:   new(mocks.MockIDGenerator),
    }
    svc := New{Pascal}Service(m.repo, m.clock, m.ids, nopLogger()).(*{singular}Service)
    return svc, m
}
```

### 12. `migrations/{NEXT_SEQ}_create_{plural}_indexes.up.json`
```json
{
  "createIndexes": "{plural}",
  "indexes": [
    { "key": { "name": 1 }, "name": "name_idx" },
    { "key": { "active": 1 }, "name": "active_idx" },
    { "key": { "created_at": -1 }, "name": "created_at_desc" }
  ]
}
```

### 13. `migrations/{NEXT_SEQ}_create_{plural}_indexes.down.json`
```json
{ "dropIndexes": "{plural}", "index": ["name_idx","active_idx","created_at_desc"] }
```

---

## After generating all files

1. Run `go build ./...` — fix any compile errors before reporting done
2. Run `go vet ./...` — fix any vet issues
3. Run `make gen-mock` — regenerate mocks to include the new `{Pascal}Repository`
4. Run `go test ./internal/core/services/... -run Test{Pascal}Service -v -count=1 -race` — all tests must pass
5. Tell the user:
   - List of files created/modified
   - Next steps: wire the service in `cmd/serve.go`, run `make swagger`
   - Migration sequence number used

## Hard rules (never violate)
- `internal/core/**` imports NOTHING outside stdlib + `pkg/apperrors` + its own sub-packages
- Domain struct has NO bson tags, NO framework tags
- Handlers have ZERO business logic
- Every I/O function takes `context.Context` as first param
- No `time.Now()` or `uuid.New()` in core — use ports
