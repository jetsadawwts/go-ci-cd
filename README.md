# Go CI/CD — Task Microservice

[![CI/CD](https://github.com/jetsadawwts/go-ci-cd/actions/workflows/ci.yml/badge.svg)](https://github.com/jetsadawwts/go-ci-cd/actions/workflows/ci.yml)
A production-ready Task microservice built with Go, following Clean Architecture principles. Features a full CI/CD pipeline with GitHub Actions, Docker containerization, and PostgreSQL persistence.

## Description

This project demonstrates a complete Go microservice with:

- **RESTful CRUD API** for task management
- **Clean Architecture** with clear separation of concerns
- **CI/CD Pipeline** using GitHub Actions (test → lint → build → Docker push)
- **Docker & Docker Compose** for containerized deployment
- **Database Migrations** with golang-migrate
- **Hot Reload** development with Air
- **Structured Logging** with `log/slog`
- **Graceful Shutdown** handling

## Architecture

The project follows **Clean Architecture** combined with layered patterns:

```
HTTP Request
    │
    ▼
┌──────────┐    ┌──────────┐    ┌────────────┐    ┌──────────────┐
│ Handler   │───▶│ Use Case │───▶│ Repository │───▶│  PostgreSQL  │
│ (Delivery)│    │ (Business│    │ (Data      │    │  (Database)  │
│           │◀───│  Logic)  │◀───│  Access)   │◀───│              │
└──────────┘    └──────────┘    └────────────┘    └──────────────┘
```

- **Delivery Layer** (`internal/delivery/http`): HTTP handlers, routing, request/response
- **Use Case Layer** (`internal/usecase`): Business logic and orchestration
- **Repository Layer** (`internal/repository`): Data access and persistence
- **Domain Layer** (`internal/domain`): Core entities and interfaces

Each layer depends only on the layer below it, with interfaces defining the contracts between them. This ensures testability, maintainability, and flexibility.

## Project Structure

```
go-ci-cd/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration loading
│   ├── database/                # Database connection & migrations
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/         # HTTP handlers
│   │       └── router.go        # Route definitions
│   ├── domain/                  # Entities & interfaces
│   ├── repository/              # Data access implementations
│   └── usecase/                 # Business logic
├── migrations/                  # SQL migration files
├── .github/
│   └── workflows/
│       └── ci.yml               # GitHub Actions CI/CD
├── .air.toml                    # Hot reload config
├── .golangci.yml                # Linter config
├── docker-compose.yml           # Docker Compose services
├── Dockerfile                   # Multi-stage Docker build
├── Makefile                     # Build automation
├── go.mod                       # Go module
├── go.sum                       # Dependency checksums
└── README.md                    # This file
```

## Prerequisites

| Tool | Version | Required |
|------|---------|----------|
| [Go](https://golang.org/dl/) | 1.25+ | ✅ |
| [Docker](https://docs.docker.com/get-docker/) | Latest | ✅ |
| [Docker Compose](https://docs.docker.com/compose/) | Latest | ✅ |
| [Air](https://github.com/air-verse/air) | Latest | Optional (hot reload) |
| [golangci-lint](https://golangci-lint.run/) | Latest | Optional (linting) |
| [golang-migrate](https://github.com/golang-migrate/migrate) | Latest | Optional (migrations CLI) |

## Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/jetsadawwts/go-cicd.git
cd go-cicd
```

### 2. Start PostgreSQL

```bash
make docker-up
```

This starts a PostgreSQL 16 container on port `5432`.

### 3. Run database migrations

```bash
make migrate-up
```

### 4. Start the application

**With hot reload (recommended for development):**

```bash
# Install Air first
go install github.com/air-verse/air@latest

make dev
```

**Without hot reload:**

```bash
make run
```

The server starts on `http://localhost:8080`.

### 5. Test the API

**Health check:**

```bash
curl http://localhost:8080/health
```

**Create a task:**

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Learn Go", "description": "Study Clean Architecture in Go"}'
```

**List all tasks:**

```bash
curl http://localhost:8080/api/v1/tasks
```

**Get a task by ID:**

```bash
curl http://localhost:8080/api/v1/tasks/{id}
```

**Update a task:**

```bash
curl -X PUT http://localhost:8080/api/v1/tasks/{id} \
  -H "Content-Type: application/json" \
  -d '{"title": "Learn Go CI/CD", "description": "Build a complete pipeline", "status": "in_progress"}'
```

**Delete a task:**

```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/{id}
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `POST` | `/api/v1/tasks` | Create a new task |
| `GET` | `/api/v1/tasks` | List all tasks |
| `GET` | `/api/v1/tasks/{id}` | Get a task by ID |
| `PUT` | `/api/v1/tasks/{id}` | Update a task |
| `DELETE` | `/api/v1/tasks/{id}` | Delete a task |

## CI/CD Pipeline

The project uses **GitHub Actions** for continuous integration and deployment:

```
Push/PR to main
      │
      ├──▶ Test ──────────┐
      │    (with Postgres) │
      │                    ├──▶ Build ──▶ Docker Build & Push
      ├──▶ Lint ──────────┘          (only on main push)
      │    (golangci-lint)            Pushes to GHCR
```

### Jobs

| Job | Trigger | Description |
|-----|---------|-------------|
| **Test** | Push & PR | Runs tests with PostgreSQL service, uploads coverage |
| **Lint** | Push & PR | Runs golangci-lint with project config |
| **Build** | After Test + Lint | Builds the Go binary, uploads as artifact |
| **Docker** | Push to main only | Builds & pushes Docker image to GitHub Container Registry |

### Docker Image

Images are pushed to `ghcr.io/<owner>/go-cicd` with tags:
- `latest` — latest build from main
- `sha-<commit>` — specific commit SHA

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make help` | Show all available commands |
| `make run` | Run the application |
| `make dev` | Run with hot reload (requires Air) |
| `make build` | Build production binary |
| `make test` | Run all tests with race detection |
| `make test-coverage` | Run tests and generate HTML coverage report |
| `make lint` | Run golangci-lint |
| `make fmt` | Format code with go fmt and goimports |
| `make vet` | Run go vet |
| `make tidy` | Tidy go modules |
| `make migrate-up` | Run database migrations |
| `make migrate-down` | Rollback database migrations |
| `make migrate-create name=xxx` | Create a new migration |
| `make docker-up` | Start Docker containers |
| `make docker-down` | Stop Docker containers |
| `make docker-build` | Build Docker image |
| `make clean` | Remove build artifacts |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `APP_ENV` | `development` | Application environment |
| `DB_HOST` | `localhost` | Database host |
| `DB_PORT` | `5432` | Database port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `taskdb` | Database name |
| `DB_SSLMODE` | `disable` | Database SSL mode |

## License

This project is licensed under the [MIT License](LICENSE).
