# Go Starter

Production-ready REST API starter built with **Go 1.25**, **Echo v4**, and **MongoDB**.

## Architecture

Hexagonal (Ports & Adapters) with strict dependency direction:

```
HTTP Handler → Service → Domain ← Repository (MongoDB)
```

Domain core has zero framework dependencies. Swapping MongoDB or Echo touches only the adapter layer.

## Tech Stack

| Concern | Library |
|---|---|
| HTTP | Echo v4 |
| Database | MongoDB v2 driver |
| Auth | JWT (HS256, golang-jwt/v5) |
| Config | Cobra + Viper |
| Logging | Uber Zap (structured JSON) |
| Validation | go-playground/validator/v10 |
| Migrations | golang-migrate/v4 (MongoDB JSON) |
| Outbound HTTP | go-resty/v2 |
| API Docs | swaggo/swag + echo-swagger |

## Quick Start

```bash
# 1. Copy env
cp .env.example .env

# 2. Start dependencies
docker compose up -d mongo

# 3. Run with hot reload
make run

# API available at http://localhost:8080
# Swagger UI at http://localhost:8080/swagger/index.html
```

## Docker

```bash
# Full stack (app + mongo)
make docker-up

# Logs
make docker-logs

# Teardown
make docker-down
```

## API Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/auth/register` | — | Register a new user |
| POST | `/api/v1/auth/login` | — | Login, returns access + refresh token |
| POST | `/api/v1/auth/refresh` | — | Refresh access token |
| GET | `/api/v1/users` | JWT | List users (paginated, filterable) |
| GET | `/api/v1/users/:id` | JWT | Get user by ID |
| PUT | `/api/v1/users/:id` | JWT | Update user |
| DELETE | `/api/v1/users/:id` | JWT | Delete user |
| GET | `/health` | — | Liveness probe |
| GET | `/ready` | — | Readiness probe (pings Mongo) |
| GET | `/version` | — | Build info |
| GET | `/swagger/*` | — | Swagger UI (disabled in production by default) |

## Roles

| Role | Description |
|---|---|
| `user` | Default role assigned on register |
| `admin` | Administrator with elevated privileges |

## Development

```bash
make test          # unit tests with race detector
make test-cover    # coverage report
make lint          # golangci-lint
make swagger       # regenerate Swagger docs
make mocks         # regenerate testify mocks
make migrate-up    # run pending migrations
make migrate-status # migration state
```

## Project Structure

```
internal/
├── core/
│   ├── domain/        # Entities (User), value objects (Email, Role), sentinel errors
│   ├── ports/
│   │   ├── inbound/   # AuthService, UserService interfaces + DTOs
│   │   └── outbound/  # UserRepository, Notifier, Clock, IDGenerator, PasswordHasher, TokenIssuer
│   └── services/      # AuthService, UserService implementations
├── adapters/
│   ├── inbound/http/
│   │   ├── handler/   # AuthHandler, UserHandler, HealthHandler
│   │   ├── middleware/ # RequestID, Logger, Recover, Auth, RateLimit, CORS, SecurityHeaders, BodyLimit
│   │   ├── dto/       # Request/response structs + mappers
│   │   ├── validator.go
│   │   ├── error_handler.go
│   │   └── router.go
│   └── outbound/
│       ├── mongodb/    # UserRepository implementation + migration runner
│       ├── jwtissuer/  # HS256 JWT (access + refresh tokens)
│       ├── crypto/     # bcrypt PasswordHasher
│       ├── clock/      # System clock adapter
│       ├── idgen/      # UUID v7 generator
│       └── httpclient/ # Resty-based Notifier
├── config/            # Viper-based config loading + validation
└── mocks/             # mockery-generated testify mocks
```

## Environment Variables

See `.env.example` for all configurable values. Key variables:

```env
APP_MONGO_URI=mongodb://localhost:27017
APP_JWT_SECRET=<32+ char random string>
APP_SERVER_PORT=8080
APP_APP_ENV=development   # development | staging | production
AUTO_MIGRATE=true         # run migrations on startup
```