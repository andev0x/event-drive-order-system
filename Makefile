.PHONY: help tidy test build up down logs clean restart order analytics notification

help: ## Show this help message
	@echo "Available targets:"
	@echo "  make tidy          - Tidy go modules for all services"
	@echo "  make test          - Run tests for all services"
	@echo "  make build         - Build all services"
	@echo "  make up            - Start all services with docker-compose"
	@echo "  make down          - Stop all services"
	@echo "  make logs          - Show logs from all services"
	@echo "  make clean         - Clean up containers and volumes"
	@echo "  make restart       - Restart all services"
	@echo "  make order         - Create a test order"
	@echo "  make analytics     - Get analytics summary"
	@echo "  make notification  - Check notification worker logs"

tidy: ## Tidy go modules for all services
	@echo "Tidying go modules..."
	cd services/order-service && go mod tidy
	cd services/analytics-service && go mod tidy
	cd services/notification-worker && go mod tidy
	@echo "Done!"

test: ## Run tests for all services
	@echo "Running tests..."
	cd services/order-service && go test -v ./tests/...
	@echo "Tests completed!"

build: ## Build Docker images for all services
	@echo "Building services..."
	docker-compose build
	@echo "Build completed!"

up: ## Start all services with docker-compose
	@echo "Starting services..."
	docker-compose up -d
	@echo "Services started! Waiting for services to be ready..."
	@sleep 10
	@echo "Services should be ready:"
	@echo "  - Order Service: http://localhost:8080"
	@echo "  - Analytics Service: http://localhost:8081"
	@echo "  - RabbitMQ Management: http://localhost:15672 (guest/guest)"

down: ## Stop all services
	@echo "Stopping services..."
	docker-compose down
	@echo "Services stopped!"

logs: ## Show logs from all services
	docker-compose logs -f

clean: ## Clean up containers, volumes, and images
	@echo "Cleaning up..."
	docker-compose down -v
	@echo "Cleanup completed!"

restart: down up ## Restart all services

order: ## Create a test order
	@echo "Creating test order..."
	@curl -X POST http://localhost:8080/orders \
		-H "Content-Type: application/json" \
		-d '{"customer_id": "customer-123", "product_id": "product-456", "quantity": 2, "total_amount": 99.99}' \
		| jq '.'
	@echo ""

analytics: ## Get analytics summary
	@echo "Fetching analytics summary..."
	@curl -X GET http://localhost:8081/analytics/summary | jq '.'
	@echo ""

notification: ## Check notification worker logs
	@echo "Notification worker logs (last 20 lines):"
	@docker-compose logs --tail=20 notification-worker
