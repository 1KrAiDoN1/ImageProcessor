.PHONY: help build run test clean docker-up docker-down migrate-up migrate-down frontend

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build both API and Worker services
	@echo "Building services..."
	go build -o bin/api backend/cmd/api/main.go
	go build -o bin/worker backend/cmd/worker/main.go
	@echo "Build complete!"

build-api: ## Build API service
	@echo "Building API..."
	go build -o bin/api backend/cmd/api/main.go

build-worker: ## Build Worker service
	@echo "Building Worker..."
	go build -o bin/worker backend/cmd/worker/main.go

run-api: ## Run API service
	@echo "Starting API service..."
	go run backend/cmd/api/main.go

run-worker: ## Run Worker service
	@echo "Starting Worker service..."
	go run backend/cmd/worker/main.go

test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	docker-compose build

docker-up: ## Start all services with Docker Compose
	@echo "Starting services..."
	docker-compose up -d

docker-down: ## Stop all services
	@echo "Stopping services..."
	docker-compose down

docker-logs: ## Show logs from all services
	docker-compose logs -f

docker-logs-api: ## Show logs from API service
	docker-compose logs -f api

docker-logs-worker: ## Show logs from Worker service
	docker-compose logs -f worker

docker-restart: ## Restart all services
	@echo "Restarting services..."
	docker-compose restart

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	psql -h localhost -U postgres -d imageprocessor -f backend/migrations/001_init_schema.up.sql

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	psql -h localhost -U postgres -d imageprocessor -f backend/migrations/001_init_schema.down.sql

deps: ## Install dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

dev: ## Start development environment
	@echo "Starting development environment..."
	docker-compose up postgres kafka zookeeper minio -d
	@echo "Waiting for services to be ready..."
	sleep 10
	@echo "Running migrations..."
	make migrate-up
	@echo "Development environment ready!"

stop-dev: ## Stop development environment
	@echo "Stopping development environment..."
	docker-compose down

logs: ## Show application logs
	tail -f logs/*.log

frontend: ## Start frontend development server
	@echo "Starting frontend server on http://localhost:8000..."
	cd frontend && python3 -m http.server 8000

frontend-npm: ## Start frontend with npm http-server
	@echo "Starting frontend with npm..."
	cd frontend && npx http-server -p 8000 -c-1

docker-build-all: ## Build all Docker images
	@echo "Building all Docker images..."
	docker-compose build

docker-build-frontend: ## Build only frontend Docker image
	@echo "Building frontend Docker image..."
	docker-compose build frontend

docker-build-backend: ## Build only backend Docker images
	@echo "Building backend Docker images..."
	docker-compose build api worker

docker-up-all: ## Start all services with Docker Compose
	@echo "Starting all services..."
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 15
	@echo "Services are ready!"
	@echo "Frontend: http://localhost"
	@echo "API: http://localhost:8080"
	@echo "MinIO Console: http://localhost:9001"

docker-up-infra: ## Start only infrastructure services
	@echo "Starting infrastructure services..."
	docker-compose up -d postgres kafka zookeeper minio

docker-up-app: ## Start only application services
	@echo "Starting application services..."
	docker-compose up -d api worker frontend

docker-stop: ## Stop all Docker services
	@echo "Stopping all services..."
	docker-compose stop

docker-down-all: ## Stop and remove all Docker containers
	@echo "Stopping and removing all containers..."
	docker-compose down

docker-down-clean: ## Stop and remove all Docker containers with volumes
	@echo "WARNING: This will remove all data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose down -v; \
	fi

docker-logs-all: ## Show logs from all Docker services
	docker-compose logs -f

docker-logs-frontend: ## Show logs from frontend service
	docker-compose logs -f frontend

docker-status: ## Show status of all Docker services
	@echo "Docker services status:"
	@docker-compose ps

docker-restart-frontend: ## Restart frontend service
	docker-compose restart frontend

docker-restart-api: ## Restart API service
	docker-compose restart api

docker-restart-worker: ## Restart worker service
	docker-compose restart worker

docker-exec-frontend: ## Open shell in frontend container
	docker-compose exec frontend sh

docker-exec-api: ## Open shell in API container
	docker-compose exec api sh

docker-stats: ## Show Docker containers resource usage
	docker stats

.DEFAULT_GOAL := help

