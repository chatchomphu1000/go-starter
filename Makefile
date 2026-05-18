ifneq (,$(wildcard ./.env))
	include .env
	export
endif

VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)

.PHONY: build build-linux run run-prod test test-integration test-cover test-ci \
        swagger swagger-check migrate-up migrate-down migrate-create migrate-status \
        lint fmt vet mocks check docker-build docker-up docker-down docker-logs \
        deps install clean help

build: ## Build the application binary
	go build -ldflags "$(LDFLAGS)" -o bin/app .

build-linux: ## Build for Linux amd64
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/app .

run: ## Run with hot reload (air)
	air

run-prod: ## Run the production binary
	./bin/app serve

test: ## Run unit tests
	go test ./... -race -count=1 -short

test-integration: ## Run integration tests (requires Docker)
	go test ./test/integration/... -tags=integration -race -count=1

test-cover: ## Run tests with coverage report
	go test ./... -race -count=1 -coverprofile=coverage.out
	go tool cover -func=coverage.out

test-ci: test-cover ## Run CI test suite with coverage threshold
	@echo "Checking coverage threshold..."
	@total=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | tr -d '%'); \
	if [ "$$(echo "$$total < 80" | bc)" -eq 1 ]; then \
		echo "Coverage $$total%% is below 80%% threshold"; \
		exit 1; \
	else \
		echo "Coverage $$total%% meets threshold"; \
	fi

swagger: ## Generate Swagger documentation
	swag init -g main.go --output docs

swagger-check: ## Dry-run Swagger generation
	swag init -g main.go --output docs --dryRun

migrate-up: ## Run database migrations up
	go run . migrate up

migrate-down: ## Run database migrations down
	go run . migrate down

migrate-create: ## Create a new migration (usage: make migrate-create name=migration_name)
	go run . migrate create --name=$(name)

migrate-status: ## Show migration status
	go run . migrate status

lint: ## Run linters
	golangci-lint run ./...

fmt: ## Format code
	gofmt -w .
	goimports -w .

vet: ## Run go vet
	go vet ./...

gen-mock: ## Generate mocks
	go run github.com/vektra/mockery/v2

check: fmt vet lint mocks test-ci ## Run all checks

docker-build: ## Build Docker image
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t go-starter:$(VERSION) .

docker-up: docker-down ## Start Docker Compose services (auto-creates .env from .env.example if missing)
	@if [ ! -f .env ]; then \
		echo "No .env file found — copying .env.example to .env"; \
		cp .env.example .env; \
	fi
	docker compose up --build -d
	@echo "✓ Services started successfully!"
	@echo ""
	@echo "📊 Available Services:"
	@echo "  • Swagger UI:			http://localhost:8080/swagger/index.html"

docker-down: ## Stop and remove Docker Compose services
	docker compose down -v

docker-logs: ## Follow Docker Compose logs
	docker compose logs -f app

install: ## Install required CLI tools (swag, air, golangci-lint, goimports, mockery)
	@echo "Installing tools..."
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/vektra/mockery/v2@latest
	@echo "All tools installed to $$(go env GOPATH)/bin"
	@echo "Make sure $$(go env GOPATH)/bin is in your PATH"

deps: ## Download and tidy dependencies
	go mod tidy
	go mod download

clean: ## Remove build artifacts
	rm -rf bin/ docs/ coverage.out internal/mocks/

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

run: docker-down docker-up
