APP_NAME := go-cicd
GIT_TAG := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-w -s -X main.version=$(GIT_TAG)"
DB_DSN := "postgres://postgres:postgres@localhost:5432/taskdb?sslmode=disable"

.PHONY: help run dev build test test-coverage lint fmt vet tidy \
        migrate-up migrate-down migrate-create \
        docker-up docker-down docker-build clean

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Run the application
	go run ./cmd/api

dev: ## Run with hot reload (requires Air: go install github.com/air-verse/air@latest)
	@command -v air >/dev/null 2>&1 || { echo "Air not found. Install with: go install github.com/air-verse/air@latest"; exit 1; }
	air

build: ## Build the application binary
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/server ./cmd/api

test: ## Run all tests
	go test -race -count=1 -v ./...

test-coverage: ## Run tests with coverage report
	go test -race -count=1 -coverprofile=coverage.out -v ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run golangci-lint
	golangci-lint run ./...

fmt: ## Format code
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	go vet ./...

tidy: ## Tidy go modules
	go mod tidy

migrate-up: ## Run database migrations up
	migrate -path migrations -database $(DB_DSN) up

migrate-down: ## Run database migrations down
	migrate -path migrations -database $(DB_DSN) down

migrate-create: ## Create a new migration (usage: make migrate-create name=create_users)
	migrate create -ext sql -dir migrations -seq $(name)

docker-up: ## Start Docker containers
	docker compose up -d

docker-down: ## Stop Docker containers
	docker compose down

docker-build: ## Build Docker image
	docker build -t $(APP_NAME) .

clean: ## Remove build artifacts
	rm -rf bin/ tmp/ coverage*
