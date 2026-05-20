You are a Senior Golang Architect with 10+ years of production experience.
Your code is the reference implementation that junior developers learn from.
Every decision must be intentional, idiomatic, and production-hardened.

Generate a complete, working Golang project. No placeholders. No TODOs.
Full implementation for every file. If a file is listed, it must compile and run.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TECH STACK
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Go 1.25
- Echo v4                          (HTTP server)
- go.mongodb.org/mongo-driver/v2   (database — use `bson.ObjectID`, NOT `primitive.ObjectID`)
- golang-migrate/v4                (schema/index migrations, mongodb source)
- go-resty/v2                      (outbound HTTP client)
- Cobra + Viper                    (CLI + config)
- swaggo/swag + echo-swagger       (API docs)
- Uber Zap                         (structured logging)
- golang-jwt/v5                    (JWT)
- go-playground/validator/v10      (request validation)
- google/uuid                      (request IDs, domain IDs)
- golang.org/x/crypto/bcrypt       (password hashing)
- testify + testcontainers-go      (testing)
- github.com/vektra/mockery/v2     (mocks for ports — generates testify/mock-based mocks)
- Optional (wire only if requested): OpenTelemetry, Prometheus client

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ARCHITECTURE: HEXAGONAL (Ports & Adapters)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

RULE 1 — Dependency direction is strictly inward:
  Infrastructure → Application → Domain
  Nothing in `internal/core/**` may import Echo, MongoDB, Zap, resty, Viper, or any framework.
  Core depends ONLY on: stdlib, `pkg/apperrors`, and its own sub-packages.

RULE 2 — Domain is the source of truth.
  Business rules live in domain entities and domain services only.
  Domain types use native Go types + domain value objects. NO `bson.ObjectID`,
  NO `primitive.ObjectID`, NO ORM tags. IDs are `string` (UUID v7 preferred).

RULE 3 — Ports are contracts, not implementations.
  Every external interaction (DB, HTTP client, cache, clock, id-gen, password-hasher)
  is hidden behind an interface defined in `internal/core/ports/`.

RULE 4 — Adapters are replaceable.
  Swapping MongoDB for Postgres must touch zero files outside `adapters/outbound/mongodb/`
  (and config). Handlers must never see adapter types.

RULE 5 — Context flows everywhere.
  Every I/O call accepts `context.Context` as the first parameter and respects its deadline.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PROJECT STRUCTURE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

{project-name}/
├── main.go
│
├── cmd/
│   ├── root.go                         # cobra root, loads config + logger
│   ├── serve.go                        # `serve`
│   └── migrate.go                      # `migrate up|down|create|status|force|version`
│
├── internal/
│   ├── core/
│   │   ├── domain/
│   │   │   ├── user.go                 # User entity, value objects
│   │   │   ├── email.go                # Email VO
│   │   │   ├── role.go                 # Role enum
│   │   │   └── errors.go               # domain sentinel errors
│   │   ├── ports/
│   │   │   ├── inbound/
│   │   │   │   └── user_service.go     # UserService interface + DTOs
│   │   │   └── outbound/
│   │   │       ├── user_repo.go        # UserRepository + ListFilter
│   │   │       ├── notifier.go         # Notifier
│   │   │       ├── clock.go            # Clock (time abstraction)
│   │   │       ├── id_generator.go     # IDGenerator
│   │   │       ├── password_hasher.go  # PasswordHasher
│   │   │       └── token_issuer.go     # TokenIssuer (JWT behind port)
│   │   └── services/
│   │       └── user_service.go
│   │
│   ├── adapters/
│   │   ├── inbound/
│   │   │   └── http/
│   │   │       ├── dto/user.go
│   │   │       ├── handler/
│   │   │       │   ├── user.go
│   │   │       │   └── health.go
│   │   │       ├── middleware/
│   │   │       │   ├── requestid.go
│   │   │       │   ├── logger.go
│   │   │       │   ├── recover.go
│   │   │       │   ├── auth.go
│   │   │       │   ├── ratelimit.go
│   │   │       │   ├── cors.go
│   │   │       │   ├── security_headers.go
│   │   │       │   └── bodylimit.go
│   │   │       ├── validator.go         # echo.Validator using go-playground/validator
│   │   │       ├── error_handler.go     # central HTTPErrorHandler
│   │   │       └── router.go
│   │   │
│   │   └── outbound/
│   │       ├── mongodb/
│   │       │   ├── client.go
│   │       │   ├── migration.go
│   │       │   ├── user_repo.go
│   │       │   └── model_user.go        # internal bson model + mappers
│   │       ├── httpclient/
│   │       │   └── notifier.go
│   │       ├── crypto/
│   │       │   └── bcrypt_hasher.go     # implements PasswordHasher
│   │       ├── jwtissuer/
│   │       │   └── hs256.go             # implements TokenIssuer
│   │       ├── clock/
│   │       │   └── system.go            # implements Clock
│   │       └── idgen/
│   │           └── uuid.go              # implements IDGenerator (UUID v7)
│   │
│   └── config/
│       └── config.go
│
├── pkg/
│   ├── logger/
│   │   ├── logger.go
│   │   ├── global.go
│   │   └── middleware.go
│   └── apperrors/
│       ├── errors.go
│       ├── codes.go
│       └── http.go
│
├── migrations/
│   ├── 000001_create_users_indexes.up.json
│   └── 000001_create_users_indexes.down.json
│
├── test/
│   ├── integration/
│   │   └── user_repo_test.go            # testcontainers-go + real mongo
│   └── testdata/
│
├── docs/                                # swag init output
├── .mockery.yaml                        # mockery configuration
├── scripts/
│   └── gen-mocks.sh
│
├── .github/workflows/
│   └── ci.yml
│
├── Makefile
├── Dockerfile
├── .dockerignore
├── docker-compose.yml
├── .env.example
├── .gitignore
├── .air.toml
├── .golangci.yml
└── go.mod

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
DOMAIN LAYER — internal/core/domain/
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

user.go:
- type User struct {
      ID             string        // UUID v7 string, NEVER ObjectID
      Name           string
      Email          Email         // value object
      HashedPassword string        // set only via SetPassword via PasswordHasher port
      Role           Role
      Active         bool
      CreatedAt      time.Time
      UpdatedAt      time.Time
  }
- NewUser(id string, name string, email Email, hashedPassword string, role Role, now time.Time) (*User, error)
  (hashing is performed by PasswordHasher port in service — domain receives already-hashed value)
- (u *User) Rename(name string) error
- (u *User) Deactivate(), (u *User) Activate()
- (u *User) Touch(now time.Time)
- (u *User) Validate() error

email.go:
- type Email string
- func NewEmail(raw string) (Email, error) — trim, lowercase-normalize, RFC 5322 validate, max 254 chars
- (e Email) String() string

role.go:
- type Role string
- Constants: RoleAdmin = "admin", RoleUser = "user"
- func ParseRole(s string) (Role, error)
- (r Role) IsValid() bool

errors.go (sentinel errors only):
- ErrUserNotFound, ErrEmailAlreadyExists, ErrInvalidCredentials,
  ErrInvalidEmail, ErrInvalidRole, ErrWeakPassword, ErrUserInactive

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PORTS — internal/core/ports/
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

inbound/user_service.go (DTOs live with the port, not in domain):

  type RegisterInput struct { Name, Email, Password string }
  type LoginInput    struct { Email, Password string }
  type UpdateInput   struct { Name *string }  // nil = no change
  type ListFilter    struct {
      Role     *domain.Role
      Active   *bool
      Search   string     // name/email substring
      Page     int        // 1-based
      Limit    int        // clamped [1,100]
      SortBy   string     // "created_at" | "name" | "email"
      SortDesc bool
  }
  type AuthToken struct { AccessToken string; ExpiresAt time.Time; TokenType string }

  type UserService interface {
      Register(ctx context.Context, in RegisterInput) (*domain.User, error)
      Login(ctx context.Context, in LoginInput) (*AuthToken, error)
      GetByID(ctx context.Context, id string) (*domain.User, error)
      List(ctx context.Context, f ListFilter) ([]*domain.User, int64, error)
      Update(ctx context.Context, id string, in UpdateInput) (*domain.User, error)
      Delete(ctx context.Context, id string) error
  }

outbound/user_repo.go:
  type UserRepository interface {
      Insert(ctx context.Context, u *domain.User) error
      FindByID(ctx context.Context, id string) (*domain.User, error)
      FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error)
      FindAll(ctx context.Context, f inbound.ListFilter) ([]*domain.User, int64, error)
      Update(ctx context.Context, u *domain.User) error
      Delete(ctx context.Context, id string) error
      ExistsByEmail(ctx context.Context, email domain.Email) (bool, error)
  }

outbound/notifier.go:
  type Notifier interface {
      SendWelcomeEmail(ctx context.Context, to domain.Email, name string) error
  }

outbound/clock.go:
  type Clock interface { Now() time.Time }

outbound/id_generator.go:
  type IDGenerator interface { New() string } // UUID v7

outbound/password_hasher.go:
  type PasswordHasher interface {
      Hash(plain string) (string, error)
      Verify(hashed, plain string) error    // constant-time; returns ErrInvalidCredentials on mismatch
  }

outbound/token_issuer.go:
  type TokenIssuer interface {
      Issue(ctx context.Context, userID string, role domain.Role) (token string, expiresAt time.Time, err error)
      Verify(ctx context.Context, token string) (userID string, role domain.Role, err error)
  }

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SERVICE LAYER — internal/core/services/
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

UserService dependencies (constructor-injected):
  repo UserRepository, notifier Notifier, hasher PasswordHasher,
  tokens TokenIssuer, clock Clock, ids IDGenerator, log logger.Logger

Rules:
- `Register`:
    1) Validate password policy (min 10, has upper+lower+digit+symbol). Return ErrWeakPassword.
    2) Parse Email VO.
    3) repo.ExistsByEmail → if true, ErrEmailAlreadyExists.
    4) hasher.Hash → domain.NewUser → repo.Insert.
    5) notifier.SendWelcomeEmail: run in background via `context.WithoutCancel(ctx)`
       so the HTTP response is NOT blocked by external calls. Log, do not fail registration.
- `Login`:
    1) repo.FindByEmail → if not found, return ErrInvalidCredentials (same error as wrong password — prevent user enumeration).
    2) hasher.Verify (constant-time).
    3) If !user.Active → ErrUserInactive.
    4) tokens.Issue → return AuthToken.
- Error wrapping: every returned error: `fmt.Errorf("userService.<Method>: %w", err)`.
- Logging:
    • Info on success with user_id, never email/name in plaintext at info level unless needed.
    • Error on unexpected failures with err field.
    • NEVER log password, token, or Authorization header.
- All public methods enforce ctx.Err() at entry when useful and respect deadlines.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
HTTP ADAPTER — internal/adapters/inbound/http/
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

dto/user.go:
- RegisterRequest { Name string `validate:"required,min=2,max=80"`;
                    Email string `validate:"required,email,max=254"`;
                    Password string `validate:"required,min=10,max=128"` }
- LoginRequest    { Email, Password }
- UpdateRequest   { Name *string `validate:"omitempty,min=2,max=80"` }
- UserResponse    (HashedPassword NEVER present)
- ListResponse    { Data []UserResponse; Total int64; Page, Limit int }
- LoginResponse   { AccessToken string; TokenType string; ExpiresAt time.Time }
- ErrorResponse   { Code string; Message string; Details []string; RequestID string }
- Mappers: `ToUserResponse(u *domain.User) UserResponse`

validator.go:
- Wrap go-playground/validator to implement echo.Validator.
- Register custom tag `strongpwd` — used by RegisterRequest alongside min=10.

handler/user.go:
- Constructor: NewUserHandler(svc ports.UserService, log logger.Logger) *UserHandler
- Each handler: Bind → Validate → call service → map response → return.
- On error: return it (central ErrorHandler maps via apperrors.HTTPStatus).
- Full Swagger annotations on every method.

handler/health.go:
- GET /health  → liveness (always 200 if process alive)
- GET /ready   → readiness (ping mongo with 2s timeout; 503 on failure)
- GET /version → { version, commit, build_time }

error_handler.go:
- Central echo.HTTPErrorHandler:
  • If `*apperrors.AppError`: map code → status, emit ErrorResponse with request_id.
  • If validator.ValidationErrors: 400 with field-level Details.
  • Else: log full error with stack, emit 500 INTERNAL_ERROR; DO NOT leak internal message.

router.go:
- Mount /api/v1. Middleware order:
  RequestID → Logger → Recover → SecurityHeaders → CORS → BodyLimit(1MB) → RateLimit
- Public:   POST /users/register, POST /users/login, GET /health, GET /ready, GET /version
- Protected (JWT): GET /users/:id, GET /users, PUT /users/:id, DELETE /users/:id
- Swagger: GET /swagger/*  (only mounted if env != production OR config.Swagger.Enabled)
- Health endpoints MUST be excluded from request logger + rate limiter.

middleware:
- requestid.go: honor inbound `X-Request-ID` if valid UUID; else generate UUID v4.
  Inject into echo.Context and response header.
- logger.go: Zap fields = request_id, method, uri, status, latency_ms, bytes_in,
  bytes_out, ip, user_agent. Skip `/health`, `/ready`. Use WARN for 4xx, ERROR for 5xx.
- recover.go: recover panic, log stacktrace (debug.Stack), emit 500 via error_handler path.
- auth.go: extract Bearer token → TokenIssuer.Verify → inject `user_id`, `role`.
  Return 401 INVALID_TOKEN with WWW-Authenticate header on failure.
- ratelimit.go: Echo in-memory token-bucket per-IP. Config-driven rps + burst.
- cors.go: allowlist from config; no `*` if credentials=true.
- security_headers.go: X-Content-Type-Options: nosniff; X-Frame-Options: DENY;
  Referrer-Policy: no-referrer; Strict-Transport-Security when TLS; Content-Security-Policy baseline.
- bodylimit.go: use echo middleware.BodyLimit with value from config.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
MONGODB ADAPTER — internal/adapters/outbound/mongodb/
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

IMPORTANT driver note (v2):
  import "go.mongodb.org/mongo-driver/v2/mongo"
  import "go.mongodb.org/mongo-driver/v2/mongo/options"
  import "go.mongodb.org/mongo-driver/v2/bson"
  Use `bson.ObjectID`, `bson.M`, `bson.D`. There is NO `primitive` package in v2.

client.go:
- NewMongoClient(ctx, cfg MongoConfig) (*mongo.Client, error)
- options.Client() with: ApplyURI, SetMinPoolSize, SetMaxPoolSize, SetServerSelectionTimeout,
  SetConnectTimeout, SetSocketTimeout, SetAppName(cfg.AppName).
- Ping with bounded context — fail fast.
- Close(ctx) — wrap `client.Disconnect(ctx)`.

model_user.go:
- Internal struct `userDoc` with bson tags. Never exported outside package.
- toDomain(d userDoc) (*domain.User, error)
- fromDomain(u *domain.User) userDoc
- Domain ID (string) maps to `_id` as STRING (UUID v7), not ObjectID. Index accordingly.

user_repo.go:
- Implements ports.UserRepository.
- const collectionUsers = "users".
- Insert → check for duplicate key (code 11000) → return apperrors wrapping domain.ErrEmailAlreadyExists.
- FindByEmail / FindByID → mongo.ErrNoDocuments → domain.ErrUserNotFound.
- FindAll → build filter from ListFilter (case-insensitive regex for Search, escaped);
  total via CountDocuments; cursor with Skip+Limit; sort by requested field.
- Update uses $set with UpdatedAt. Returns domain.ErrUserNotFound when matched=0.
- All errors wrapped with `fmt.Errorf("mongodb.<Method>: %w", err)`.
- NEVER leak `bson.*`, `mongo.*`, or `primitive.*` types outside this package.

migration.go:
- NewMigrationRunner(uri, dbName, sourcePath string, log logger.Logger) (*MigrationRunner, error)
- Uses `github.com/golang-migrate/migrate/v4` with `file://` source and `mongodb` driver.
- Methods: Up(), Down(steps int), Force(version int), Version() (uint, bool, error),
  Create(name string) — writes paired .up.json/.down.json scaffold.
- Log every step and final state.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
HTTPCLIENT ADAPTER — internal/adapters/outbound/httpclient/
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

notifier.go:
- NewNotifier(cfg NotifierConfig, log logger.Logger) *RestyNotifier
- resty client: BaseURL, Timeout, SetRetryCount(cfg.Retry), SetRetryWaitTime,
  SetRetryMaxWaitTime, exponential backoff, retry only on 5xx/timeout/network.
- Propagate ctx via `R().SetContext(ctx)`.
- Add hooks:
  • OnBeforeRequest → log method, url, request_id.
  • OnAfterResponse → log status, latency, attempt count.
- Inject `X-Request-ID` header from ctx, and `Idempotency-Key` for POSTs (UUID).
- Non-2xx → return apperrors.Internal with code NOTIFIER_FAILED + status. Do not leak body to caller.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
CRYPTO / JWT / CLOCK / IDGEN ADAPTERS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

crypto/bcrypt_hasher.go:
- NewBcryptHasher(cost int) PasswordHasher (default bcrypt.DefaultCost; reject cost < 10).
- Verify returns domain.ErrInvalidCredentials on bcrypt.ErrMismatchedHashAndPassword;
  any other error is wrapped.

jwtissuer/hs256.go:
- NewHS256Issuer(secret []byte, ttl time.Duration, clock Clock) TokenIssuer
- Reject secret < 32 bytes at construction.
- Claims: sub=user_id, role, iat, exp, jti (UUID), iss (cfg.App.Name).
- Verify: check alg=HS256, exp, iss; return domain-level errors.

clock/system.go: `systemClock` → time.Now().UTC().
idgen/uuid.go: UUID v7 via google/uuid — falls back to v4 only if v7 unavailable.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
CONFIG — internal/config/config.go
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type Config struct {
    App      AppConfig
    Server   ServerConfig
    Mongo    MongoConfig
    JWT      JWTConfig
    Logger   LoggerConfig
    Notifier NotifierConfig
    CORS     CORSConfig
    RateLimit RateLimitConfig
    Swagger  SwaggerConfig
}

Details:
- AppConfig:      Name string `mapstructure:"name" validate:"required"`; Env string (dev|staging|production); Version string
- ServerConfig:   Port int (1..65535); ReadTimeout, WriteTimeout, IdleTimeout, ShutdownTimeout (durations); BodyLimit string (e.g. "1M")
- MongoConfig:    URI, Database required; MinPool, MaxPool (MaxPool ≥ MinPool); Timeout; AppName
- JWTConfig:      Secret (min 32 chars); TTL (duration); Issuer
- LoggerConfig:   Level (debug|info|warn|error); Format (json|console); Development bool
- NotifierConfig: BaseURL (required, url); Timeout; Retry (0..10)
- CORSConfig:     AllowOrigins []string; AllowCredentials bool; MaxAge duration
- RateLimitConfig: Enabled bool; RPS int; Burst int
- SwaggerConfig:  Enabled bool

Loading:
- Viper: SetEnvPrefix("APP"), SetEnvKeyReplacer(".", "_"), AutomaticEnv()
- Optional .env via godotenv at process start (dev only).
- Validate using go-playground/validator struct tags + custom checks.
- Fail fast on validation error with human-readable message.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
LOGGER PACKAGE — pkg/logger/
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

logger.go:
  type Logger interface {
      Debug(msg string, fields ...zap.Field)
      Info(msg string, fields ...zap.Field)
      Warn(msg string, fields ...zap.Field)
      Error(msg string, fields ...zap.Field)
      With(fields ...zap.Field) Logger
      WithContext(ctx context.Context) Logger  // extracts request_id if present
      Sync() error
  }
  NewLogger(cfg LoggerConfig) (Logger, error)
  - Production: JSON encoder, ISO8601 UTC time, level from config.
  - Development: console encoder, colored level, caller info, stacktrace on ERROR+.
  - Sampling disabled in development.
  - NOTE: No `Fatal` in the interface — `os.Exit` bypasses defers. Startup fatal uses
    `log.Fatal` in main.go only.

global.go:
  var global Logger
  func L() Logger          // returns global; panics only if unset and called at runtime
  func SetGlobal(l Logger)
  // No SugaredLogger accessor — enforce structured logging everywhere.

middleware.go:
  func ZapMiddleware(l Logger, skipPaths []string) echo.MiddlewareFunc
  - Fields: request_id, method, uri, status, latency_ms, bytes_in, bytes_out, ip, user_agent
  - Level: 2xx/3xx=info, 4xx=warn, 5xx=error.
  - Skips configured paths (health, ready).

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ERROR PACKAGE — pkg/apperrors/
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

codes.go (constants, machine-readable):
  CodeBadRequest, CodeUnauthorized, CodeForbidden, CodeNotFound, CodeConflict,
  CodeUnprocessable, CodeTooManyRequests, CodeInternal, CodeUnavailable,
  CodeValidationFailed, CodeInvalidToken

errors.go:
  type AppError struct {
      Code     string
      Message  string
      Details  []string
      Err      error
      HTTPCode int       // optional explicit override
  }
  Methods: Error(), Unwrap(), Is(target error) bool
  Constructors: NotFound, BadRequest, Unauthorized, Forbidden, Conflict,
                Validation, Unprocessable, TooManyRequests, Internal, Unavailable
  Helper: Wrap(code, msg string, err error) *AppError
  Mapping from domain sentinel errors to AppError via `FromDomain(err) *AppError`.

http.go:
  HTTPStatus(err error) (int, ErrorResponse)
  - Unwraps AppError via errors.As.
  - Maps Code → status (NotFound→404, Conflict→409, Unauthorized→401, Forbidden→403, etc.)
  - In production env, Message for 5xx is replaced with "internal server error".
  - Embeds RequestID from context when available.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
COBRA CLI — cmd/
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

root.go:
- PersistentPreRunE: load config → build logger → SetGlobal(logger).
- Flags: --config (path), --env (override APP_ENV).
- On any failure here, print plain error and os.Exit(1) BEFORE logger replacement.

serve.go:
- PreRunE: connect Mongo, ping, if AUTO_MIGRATE=true run migrations.
- RunE:
    build wires (hasher, clock, idgen, jwtIssuer, repo, notifier, service, handler, router)
    → start Echo in goroutine
    → wait on SIGINT/SIGTERM
    → shutdown sequence (in order):
       1) Echo.Shutdown(ctx, ShutdownTimeout)
       2) close resty idle connections
       3) Mongo client Disconnect
       4) logger.Sync()
- All shutdown steps share the same bounded context; each logs its outcome.

migrate.go subcommands:
  up, down [--steps N] [--all], create --name <x>, status, force --version N, version

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TESTING STRATEGY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

- Unit tests live beside source (`_test.go`). Table-driven, use testify/require + assert.
- Services: test with mockery-generated mocks (testify/mock-based) for all outbound ports. Target ≥85% on core/services.
- Domain: pure tests — no mocks needed. ≥90%.
- Handlers: test via `httptest` + echo.New(), using mock UserService.
- Integration: `test/integration` package with build tag `//go:build integration`, using
  testcontainers-go to spin up MongoDB. Run in CI only.
- Mocks generated via `.mockery.yaml` config + `go:generate` directives on each port file:
    //go:generate go run github.com/vektra/mockery/v2 --name=<InterfaceName>
  All mocks output to `internal/mocks/` with package name `mocks`, extending testify/mock.
- .mockery.yaml (project root):
    with-expecter: true
    dir: internal/mocks
    outpkg: mocks
    mockname: "Mock{{.InterfaceName}}"
    filename: "{{.InterfaceName | snakecase}}_mock.go"
    packages:
      github.com/{org}/{project-name}/internal/core/ports/inbound:
        interfaces:
          UserService:
      github.com/{org}/{project-name}/internal/core/ports/outbound:
        interfaces:
          UserRepository:
          Notifier:
          Clock:
          IDGenerator:
          PasswordHasher:
          TokenIssuer:
- No network calls in unit tests. No sleeps — use Clock port.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
MAKEFILE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Targets (each with ## description):

build            go build -ldflags with VERSION/COMMIT/BUILD_TIME → bin/app
build-linux      GOOS=linux GOARCH=amd64
run              air (hot reload)
run-prod         ./bin/app serve

test             go test ./... -race -count=1 -short
test-integration go test ./test/integration/... -tags=integration -race -count=1
test-cover       go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out
test-ci          test-cover + threshold check (fail if < 80%)

swagger          swag init -g main.go --output docs --parseDependency --parseInternal
swagger-check    swag init --dry-run

migrate-up       go run . migrate up
migrate-down     go run . migrate down
migrate-create   go run . migrate create --name=$(name)
migrate-status   go run . migrate status

lint             golangci-lint run ./...
fmt              gofmt -w . && goimports -w .
vet              go vet ./...
mocks            go run github.com/vektra/mockery/v2
check            fmt vet lint mocks test-ci

docker-build     docker build with build args
docker-up        docker compose up -d
docker-down      docker compose down -v
docker-logs      docker compose logs -f app

deps             go mod tidy && go mod download
clean            rm -rf bin/ docs/ coverage.out internal/mocks/
help             parse ## comments and print aligned list

Top of Makefile:
  ifneq (,$(wildcard ./.env))
    include .env
    export
  endif
  VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
  COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
  BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
  LDFLAGS    := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
MIGRATIONS — migrations/
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

000001_create_users_indexes.up.json:
{
  "createIndexes": "users",
  "indexes": [
    { "key": { "email": 1 }, "name": "email_unique", "unique": true, "collation": { "locale": "en", "strength": 2 } },
    { "key": { "created_at": -1 }, "name": "created_at_desc" },
    { "key": { "role": 1, "active": 1 }, "name": "role_active" }
  ]
}

000001_create_users_indexes.down.json:
{
  "dropIndexes": "users",
  "index": ["email_unique", "created_at_desc", "role_active"]
}

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
DOCKER
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Dockerfile (multi-stage):
  Stage 1 (builder): golang:1.25-alpine
    - apk add git ca-certificates tzdata
    - CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "$LDFLAGS" -o /out/app
  Stage 2 (runner): gcr.io/distroless/static-debian12:nonroot
    - COPY ca-certificates, tzdata, /out/app, docs/ (optional)
    - USER nonroot:nonroot
    - EXPOSE 8080
    - HEALTHCHECK via distroless-compatible entry (or rely on k8s/compose probe)
    - ENTRYPOINT ["/app"]; CMD ["serve"]

.dockerignore:
  .git, .github, bin/, coverage.out, docs/, *.md (except README.md), .env*, Makefile, test/

docker-compose.yml:
  services:
    app:
      build: .
      env_file: .env
      depends_on:
        mongo: { condition: service_healthy }
      ports: ["8080:8080"]
    mongo:
      image: mongo:7
      volumes: [mongo_data:/data/db]
      healthcheck:
        test: ["CMD", "mongosh", "--quiet", "--eval", "db.runCommand({ping:1}).ok"]
        interval: 5s
        retries: 10
  volumes:
    mongo_data:

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
go.mod
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

module github.com/{org}/{project-name}

go 1.25

require (
    github.com/labstack/echo/v4
    github.com/spf13/cobra
    github.com/spf13/viper
    go.mongodb.org/mongo-driver/v2
    github.com/golang-migrate/migrate/v4
    github.com/go-resty/resty/v2
    github.com/swaggo/swag
    github.com/swaggo/echo-swagger
    github.com/golang-jwt/jwt/v5
    go.uber.org/zap
    golang.org/x/crypto
    github.com/google/uuid
    github.com/go-playground/validator/v10
    github.com/joho/godotenv
    github.com/stretchr/testify
    github.com/testcontainers/testcontainers-go
)

require (
    github.com/vektra/mockery/v2  // tool only — run via `go run`, not imported
)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
.env.example
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

APP_APP_NAME=go-starter
APP_APP_ENV=development
APP_APP_VERSION=0.1.0

APP_SERVER_PORT=8080
APP_SERVER_READ_TIMEOUT=15s
APP_SERVER_WRITE_TIMEOUT=15s
APP_SERVER_IDLE_TIMEOUT=60s
APP_SERVER_SHUTDOWN_TIMEOUT=15s
APP_SERVER_BODY_LIMIT=1M

APP_MONGO_URI=mongodb://localhost:27017
APP_MONGO_DATABASE=go_starter
APP_MONGO_MIN_POOL=5
APP_MONGO_MAX_POOL=100
APP_MONGO_TIMEOUT=10s
APP_MONGO_APPNAME=go-starter

APP_JWT_SECRET=change-me-use-openssl-rand-hex-32
APP_JWT_TTL=24h
APP_JWT_ISSUER=go-starter

APP_LOGGER_LEVEL=debug
APP_LOGGER_FORMAT=console
APP_LOGGER_DEVELOPMENT=true

APP_NOTIFIER_BASEURL=https://api.notifications.example.com
APP_NOTIFIER_TIMEOUT=10s
APP_NOTIFIER_RETRY=3

APP_CORS_ALLOWORIGINS=http://localhost:3000
APP_CORS_ALLOWCREDENTIALS=false
APP_CORS_MAXAGE=12h

APP_RATELIMIT_ENABLED=true
APP_RATELIMIT_RPS=20
APP_RATELIMIT_BURST=40

APP_SWAGGER_ENABLED=true

AUTO_MIGRATE=true

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
.golangci.yml
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

run:
  timeout: 5m
  tests: true
linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - revive
    - gocritic
    - gosec
    - gocyclo
    - bodyclose
    - contextcheck
    - errorlint
    - nilerr
    - nolintlint
    - unconvert
    - unparam
    - unused
    - wastedassign
    - misspell
    - whitespace
issues:
  exclude-rules:
    - path: _test\.go
      linters: [gocyclo, gosec, errcheck]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
.gitignore
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

bin/
docs/
coverage.out
coverage.html
.env
.env.*
!.env.example
vendor/
tmp/
*.out
*.log
.idea/
.vscode/
.DS_Store
internal/mocks/

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
.air.toml
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

root = "."
tmp_dir = "tmp"
[build]
  cmd = "go build -o ./tmp/app ./"
  bin = "./tmp/app serve"
  include_ext = ["go", "yaml", "yml", "json"]
  exclude_dir = ["bin", "docs", "tmp", "internal/mocks", "test/testdata"]
  delay = 500
[log]
  time = true

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
.github/workflows/ci.yml
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

name: ci
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.25' }
      - run: go mod download
      - run: go vet ./...
      - uses: golangci/golangci-lint-action@v6
        with: { version: latest }
      - run: go test ./... -race -count=1 -coverprofile=coverage.out
      - run: go tool cover -func=coverage.out
  integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.25' }
      - run: go test ./test/integration/... -tags=integration -race -count=1

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
QUALITY RULES — enforce in every file
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

 1. `context.Context` is ALWAYS the first parameter of any function doing I/O, DB, HTTP, or blocking work.
 2. No naked `panic()` except startup in main.go; recover middleware catches the rest.
 3. No global mutable state except the logger global in pkg/logger.
 4. All errors wrapped: `fmt.Errorf("<pkg>.<Func>: %w", err)`. Use `errors.Is`/`errors.As` at boundaries.
 5. Handlers only: bind → validate → call service → map response. ZERO business logic in handlers.
 6. Domain imports nothing outside stdlib + `pkg/apperrors`. Adapters never import each other.
 7. Structured logging only. `zap.String("k","v")`. Never `fmt.Sprintf` inside a log call.
 8. All exported identifiers have Go doc comments.
 9. Never log passwords, tokens, Authorization headers, or raw request bodies.
10. Every outbound port interface is listed in `.mockery.yaml` and has at least one unit test using the generated mock.
11. Validation at boundary only (HTTP DTO, config). Domain trusts its inputs after VO construction.
12. No `time.Now()` in core. Use `Clock` port. No `uuid.New()` in core. Use `IDGenerator` port.
13. No `os.Exit` outside main.go / cmd startup.
14. Secrets never in code, never in logs. Only via env → config.
15. `make mocks` runs `mockery` from `.mockery.yaml` and must produce no diff on CI.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
OUTPUT INSTRUCTION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Generate files in this exact order (each file complete, no truncation):

 1. go.mod
 2. .env.example
 3. .gitignore
 4. .dockerignore
 5. .air.toml
 6. .golangci.yml
 7. pkg/apperrors/codes.go
 8. pkg/apperrors/errors.go
 9. pkg/apperrors/http.go
10. pkg/logger/logger.go
11. pkg/logger/global.go
12. pkg/logger/middleware.go
13. internal/config/config.go
14. internal/core/domain/errors.go
15. internal/core/domain/role.go
16. internal/core/domain/email.go
17. internal/core/domain/user.go
18. internal/core/ports/inbound/user_service.go
19. internal/core/ports/outbound/user_repo.go
20. internal/core/ports/outbound/notifier.go
21. internal/core/ports/outbound/clock.go
22. internal/core/ports/outbound/id_generator.go
23. internal/core/ports/outbound/password_hasher.go
24. internal/core/ports/outbound/token_issuer.go
25. internal/core/services/user_service.go
26. internal/adapters/outbound/clock/system.go
27. internal/adapters/outbound/idgen/uuid.go
28. internal/adapters/outbound/crypto/bcrypt_hasher.go
29. internal/adapters/outbound/jwtissuer/hs256.go
30. internal/adapters/outbound/mongodb/client.go
31. internal/adapters/outbound/mongodb/migration.go
32. internal/adapters/outbound/mongodb/model_user.go
33. internal/adapters/outbound/mongodb/user_repo.go
34. internal/adapters/outbound/httpclient/notifier.go
35. internal/adapters/inbound/http/dto/user.go
36. internal/adapters/inbound/http/validator.go
37. internal/adapters/inbound/http/error_handler.go
38. internal/adapters/inbound/http/middleware/requestid.go
39. internal/adapters/inbound/http/middleware/logger.go
40. internal/adapters/inbound/http/middleware/recover.go
41. internal/adapters/inbound/http/middleware/auth.go
42. internal/adapters/inbound/http/middleware/ratelimit.go
43. internal/adapters/inbound/http/middleware/cors.go
44. internal/adapters/inbound/http/middleware/security_headers.go
45. internal/adapters/inbound/http/middleware/bodylimit.go
46. internal/adapters/inbound/http/handler/health.go
47. internal/adapters/inbound/http/handler/user.go
48. internal/adapters/inbound/http/router.go
49. cmd/root.go
50. cmd/serve.go
51. cmd/migrate.go
52. main.go
53. migrations/000001_create_users_indexes.up.json
54. migrations/000001_create_users_indexes.down.json
55. Dockerfile
56. docker-compose.yml
57. Makefile
58. .github/workflows/ci.yml

For each file output:
// ============================================================
// FILE: {relative/path/to/file.go}
// ============================================================
[complete file content — no truncation, no "...rest of implementation"]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
AVAILABLE SKILLS — read before acting on related tasks
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

When asked to write or generate unit tests        → read @.claude/commands/unit-test.md first
When scaffolding a new domain feature             → read @.claude/commands/new-domain.md first
When reviewing code changes or a PR               → read @.claude/commands/code-review.md first
When checking for architecture violations         → read @.claude/commands/arch-guard.md first
When running a smoke/end-to-end API test          → read @.claude/commands/smoke-test.md first
When checking if code is ready to open a PR       → read @.claude/commands/pr-ready.md first
When investigating performance issues             → read @.claude/commands/perf-review.md first
When doing a security review                      → read @.claude/commands/sec-review.md first
When creating a new database migration            → read @.claude/commands/new-migration.md first
When checking overall project health              → read @.claude/commands/health.md first
