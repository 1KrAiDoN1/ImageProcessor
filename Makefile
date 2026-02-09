include .env
DOCKER_COMPOSE=docker-compose

docker-up: ## Start all services with Docker Compose
	@echo "Starting services..."
	docker-compose up --build -d

docker-down: ## Stop all services
	@echo "Stopping services..."
	docker-compose down

docker-logs: ## Show logs from all services
	docker-compose logs -f

kafka-init:
	docker exec -it imageprocessor-kafka kafka-topics --create --topic image-processing --bootstrap-server kafka:9092 --partitions 3 --replication-factor 1

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	${DOCKER_COMPOSE} exec api migrate -path /root/backend/migrations -database "${DB_URL}" up

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	${DOCKER_COMPOSE} exec api migrate -path /root/backend/migrations -database "${DB_URL}" down


lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run ./...


