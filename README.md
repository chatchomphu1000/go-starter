# Go Starter — Dormitory Rental Backend

Production-ready REST API for managing dormitory/apartment rentals, built with **Go 1.25**, **Echo v4**, and **MongoDB**.

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
| Background jobs | Custom goroutine pool + ticker scheduler |

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

## Roles

| Role | Description |
|---|---|
| `admin` | Full access to everything |
| `owner` | Manages rooms, bookings, invoices, reports |
| `tenant` | Books rooms, pays invoices, submits maintenance |
| `user` | Basic authenticated user (default on register) |

## Background Jobs

| Job | Schedule | Description |
|---|---|---|
| `invoice.overdue` | Every 1h | Marks sent invoices past due date as overdue |
| `rent.reminder` | Every 24h | Sends notification for invoices due within 3 days |

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
│   ├── domain/        # Entities, value objects, sentinel errors
│   ├── ports/         # Inbound (service interfaces + DTOs), Outbound (repo/infra interfaces)
│   └── services/      # Business logic
├── adapters/
│   ├── inbound/http/  # Echo handlers, middleware, router, DTOs
│   └── outbound/      # MongoDB repos, JWT issuer, bcrypt, clock, UUID gen, HTTP notifier
└── worker/            # Goroutine pool + scheduled jobs
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
