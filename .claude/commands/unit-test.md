# /unit-test — Generate table-driven unit tests for a service

Generate a complete, compilable `{service}_service_test.go` file for the service given in `$ARGUMENTS`.

Parse `$ARGUMENTS` as the service name in lowercase (e.g., `auth`, `user`, `room`, `booking`,
`invoice`, `payment`, `maintenance`, `notice`, `message`, `report`).

Module path: `github.com/chatchomphu1000/go-starter`

---

## Step 1: Discover the service implementation

Read these files before generating anything:
```
internal/core/services/{name}_service.go      ← actual implementation
internal/core/ports/inbound/{name}_service.go ← interface + input/output DTOs
```

Identify:
- All struct fields (injected dependencies)
- All public methods with signatures
- Business rules enforced in each method (guard clauses, validations, error branches)
- Which outbound ports are called in each method

---

## Step 2: Identify required mocks

Check `internal/mocks/` for existing generated mocks.
If a required mock does not exist, remind the user to run `make gen-mock` first.

Common mocks used:
| Port                | Mock type                        |
|---------------------|----------------------------------|
| UserRepository      | `mocks.MockUserRepository`       |
| RoomRepository      | `mocks.MockRoomRepository`       |
| BookingRepository   | `mocks.MockBookingRepository`    |
| InvoiceRepository   | `mocks.MockInvoiceRepository`    |
| PaymentRepository   | `mocks.MockPaymentRepository`    |
| MaintenanceRepository | `mocks.MockMaintenanceRepository` |
| NoticeRepository    | `mocks.MockNoticeRepository`     |
| MessageRepository   | `mocks.MockMessageRepository`    |
| ActivityLogRepository | `mocks.MockActivityLogRepository` |
| Notifier            | `mocks.MockNotifier`             |
| Clock               | `mocks.MockClock`                |
| IDGenerator         | `mocks.MockIDGenerator`          |
| PasswordHasher      | `mocks.MockPasswordHasher`       |
| TokenIssuer         | `mocks.MockTokenIssuer`          |

---

## Step 3: Generate `internal/core/services/{name}_service_test.go`

### File conventions

```go
package services  // same package — allows testing unexported helpers if needed

import (
    "context"
    "errors"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
    "go.uber.org/zap"

    "github.com/chatchomphu1000/go-starter/internal/core/domain"
    "github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
    "github.com/chatchomphu1000/go-starter/internal/mocks"
)
```

### No-op logger helper (add once per file)

```go
// nopLogger adapts zap.NewNop() to the logger.Logger interface.
// Import "github.com/chatchomphu1000/go-starter/pkg/logger" and use:
//   pkg/logger.NewNopLogger() if it exists, otherwise:
func nopLogger() logger.Logger {
    // Use logger.NewLogger with a discarding config, OR wrap zap.NewNop() directly.
    // Check whether pkg/logger exposes a NewNopLogger() — if not, construct one inline.
    l, _ := logger.NewLogger(logger.LoggerConfig{Level: "error", Format: "json"})
    return l
}
```

### Fixed time constant (add once per file)

```go
var fixedNow = time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
```

### Service factory helper

Create one per service to avoid repetition:

```go
type {name}Mocks struct {
    repo    *mocks.Mock{Pascal}Repository
    // ... all dependencies
    clock   *mocks.MockClock
    ids     *mocks.MockIDGenerator
    log     logger.Logger
}

func newTest{Pascal}Service(t *testing.T) (*{name}Service, {name}Mocks) {
    t.Helper()
    m := {name}Mocks{
        repo:  new(mocks.Mock{Pascal}Repository),
        clock: new(mocks.MockClock),
        ids:   new(mocks.MockIDGenerator),
        log:   nopLogger(),
    }
    svc := New{Pascal}Service(m.repo, m.clock, m.ids, m.log).(*{name}Service)
    return svc, m
}
```

---

## Step 4: Write tests — method by method

### Naming pattern
`TestXxxService_MethodName_ScenarioDescription`  
e.g., `TestBookingService_Create_Success`, `TestBookingService_Create_RoomUnavailable`

### Universal first test case: context already cancelled

Every method that calls `ctx.Err()` at entry MUST have this test:

```go
{
    name: "context_already_cancelled",
    ctx:  cancelledCtx(),
    // no mock expectations — short-circuits before any outbound call
    wantErr: context.Canceled,
},
```

Helper:
```go
func cancelledCtx() context.Context {
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    return ctx
}
```

### Test case structure

```go
func TestXxxService_Method_Scenario(t *testing.T) {
    tests := []struct {
        name    string
        ctx     context.Context
        input   inbound.XxxInput     // or whatever the method accepts
        setup   func(m xxxMocks)
        want    *domain.Xxx          // or whatever the method returns
        wantErr error
    }{
        {
            name:  "success",
            ctx:   context.Background(),
            input: validInput,
            setup: func(m xxxMocks) {
                m.repo.EXPECT().FindByID(mock.Anything, "id-1").Return(someEntity, nil)
                // ... other expectations
            },
            want:    expectedResult,
            wantErr: nil,
        },
        {
            name:    "context_already_cancelled",
            ctx:     cancelledCtx(),
            input:   validInput,
            setup:   func(m xxxMocks) {},
            wantErr: context.Canceled,
        },
        // ... other cases
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            svc, m := newTestXxxService(t)
            tc.setup(m)

            got, err := svc.Method(tc.ctx, tc.input)

            if tc.wantErr != nil {
                require.Error(t, err)
                assert.True(t, errors.Is(err, tc.wantErr),
                    "expected error %v, got %v", tc.wantErr, err)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tc.want, got)
            }

            m.repo.AssertExpectations(t)
            // AssertExpectations on ALL mocks used
        })
    }
}
```

---

## Step 5: Required test scenarios per service

### **auth_service**

**Register:**
- `success` — ExistsByEmail=false, Hash OK, Insert OK, notifier called in background (do not assert notifier in sync test — it's async)
- `context_already_cancelled`
- `weak_password_too_short` — < 10 chars → ErrWeakPassword, zero DB calls
- `weak_password_no_upper` — missing uppercase → ErrWeakPassword
- `weak_password_no_symbol` — missing symbol → ErrWeakPassword
- `invalid_email_format` — malformed email → ErrInvalidEmail
- `email_already_exists` — ExistsByEmail=true → ErrEmailAlreadyExists
- `hash_error` — hasher.Hash fails → propagates error
- `repo_insert_error` — repo.Insert fails → propagates error

**Login:**
- `success` — FindByEmail OK, Verify OK, user active, Issue OK, IssueRefresh OK
- `context_already_cancelled`
- `invalid_email_format` → ErrInvalidCredentials (NOT ErrInvalidEmail — user enumeration prevention)
- `user_not_found` — FindByEmail returns ErrUserNotFound → ErrInvalidCredentials (not ErrUserNotFound)
- `wrong_password` — Verify returns error → ErrInvalidCredentials
- `user_inactive` — user.Active=false → ErrUserInactive
- `token_issue_error` — tokens.Issue fails → propagates error
- `refresh_issue_error` — tokens.IssueRefresh fails → propagates error

**RefreshToken:**
- `success` — VerifyRefresh OK, FindByID OK, Issue OK, IssueRefresh OK
- `context_already_cancelled`
- `invalid_refresh_token` — VerifyRefresh fails → propagates error
- `user_not_found` — FindByID returns ErrUserNotFound → wraps error
- `user_inactive` — user.Active=false → ErrUserInactive

### **user_service**

**GetByID:**
- `success`
- `context_already_cancelled`
- `not_found` — repo.FindByID returns ErrUserNotFound → wraps with `errors.Is`

**List:**
- `success_default_pagination` — Page=0 clamped to 1, Limit=0 clamped to 10
- `success_limit_clamped_to_100` — Limit=200 → clamped to 100
- `context_already_cancelled`
- `repo_error`

**Update:**
- `success_rename`
- `success_no_changes` — UpdateInput.Name=nil, only Touch is called
- `context_already_cancelled`
- `user_not_found` — FindByID returns ErrUserNotFound
- `empty_name` — Name=ptr("") → bad request error

**Delete:**
- `success`
- `context_already_cancelled`
- `user_not_found` — repo.Delete returns ErrUserNotFound

### **room_service**

**Create:**
- `success`
- `context_already_cancelled`
- `room_number_already_exists` — ExistsByNumber=true → ErrRoomNumberExists
- `domain_validation_error` — NewRoom returns error (e.g., negative price)
- `repo_insert_error`

**GetByID:**
- `success`
- `context_already_cancelled`
- `not_found`

**List:**
- `success_with_clamping`
- `context_already_cancelled`
- `repo_error`

**Update:**
- `success`
- `context_already_cancelled`
- `not_found`
- `wrong_owner` (if service enforces ownerID check)

**Delete:**
- `success`
- `context_already_cancelled`
- `not_found`
- `wrong_owner` (if service enforces ownerID check)

### **booking_service**

**Create:**
- `success`
- `context_already_cancelled`
- `room_not_found`
- `room_unavailable` — room.IsAvailable()=false → ErrRoomUnavailable
- `booking_conflict` — HasActiveBooking=true → ErrBookingConflict
- `domain_validation_error`
- `repo_insert_error`

**Approve / Reject / Cancel / Activate / Complete:**
- `success` — FindByID OK, valid status transition
- `context_already_cancelled`
- `not_found`
- `invalid_transition` — booking in wrong status → ErrInvalidBookingTransition
- `wrong_actor` — wrong ownerID/requesterID (if enforced)

### **invoice_service**

**Create:**
- `success` — items summed correctly
- `context_already_cancelled`
- `repo_insert_error`

**GetByID / List:** standard patterns

**MarkPaid / Cancel / Void:** status transition cases

### **payment_service**

**Create:**
- `success`
- `context_already_cancelled`
- `already_paid` — invoice/payment already completed → ErrPaymentAlreadyPaid

**HandleWebhook:**
- `success_completed`
- `success_failed`
- `payment_not_found`

### **maintenance_service**

**Create:** standard + owner notification (if any)

**StartWork / Resolve / Close:** status transition cases per TicketStatus

### **notice_service**

**Create / Update / Delete / List:** standard CRUD + expiry logic if any

### **message_service**

**Send:** upserts thread, inserts message
**MarkRead:** delegates to repo

### **report_service**

**IncomeExpense / OwnerDashboard / TenantStatement:** pure aggregation, assert math is correct

---

## Step 6: Post-generation checks

```bash
go build ./internal/core/services/...
```
Fix any compile errors.

```bash
go vet ./internal/core/services/...
```
Fix any vet issues.

```bash
go test ./internal/core/services/... -run Test{Pascal}Service -v -count=1 -race
```
All tests must pass.

```bash
go test ./internal/core/services/... -coverprofile=/tmp/svc.out && \
go tool cover -func=/tmp/svc.out | grep "{name}_service.go"
```
Report coverage. Target ≥ 85% per service file.

---

## Hard rules

- **Same package** — `package services`, not `package services_test`. Enables testing unexported helpers (`validatePassword` etc.)
- **No real I/O** — zero DB, zero HTTP, zero file system calls in any test
- **No `time.Sleep`** — use `mocks.MockClock.EXPECT().Now().Return(fixedNow)`
- **No `time.Now()`** — always use `fixedNow` constant
- **`errors.Is` for error assertions** — never `err.Error() == "some string"`
- **`mock.AssertExpectations(t)`** on every mock after the call under test
- **Table-driven** — even tests with a single case should use the table pattern for future extensibility
- **Async goroutines** — for methods that fire goroutines (e.g., welcome email), do NOT assert the goroutine behavior in the sync test. Add a brief comment: `// notifier is called asynchronously — not asserted here`
- **Background email test** — if you want to test the goroutine path, add a separate `TestXxx_Method_BackgroundNotification` that uses a `sync.WaitGroup` or channel to synchronize
