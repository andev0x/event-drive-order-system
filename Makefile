# Root Makefile for Event-driven Order System
# Author: anvndev

SHELL := /bin/bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c

.PHONY: help tidy test build build-go up down logs clean restart order analytics notification

SERVICES := order-service analytics-service notification-worker
SERVICES_DIR := services

help:
	@echo "Available targets:"
	@echo "  make tidy           - Tidy go modules for all services"
	@echo "  make test           - Run tests for all services"
	@echo "  make build          - Build Docker images"
	@echo "  make build-go       - Build Go binaries"
	@echo "  make up             - Start all services with docker compose"
	@echo "  make down           - Stop all services"
	@echo "  make logs           - Show logs from all services"
	@echo "  make clean          - Clean up containers and volumes"
	@echo "  make restart        - Restart all services"
	@echo "  make order          - Create a test order"
	@echo "  make analytics      - Get analytics summary"
	@echo "  make notification   - Check notification worker logs"

tidy:
	@echo "Tidying go modules..."
	for s in $(SERVICES); do \
		(cd $(SERVICES_DIR)/$$s && go mod tidy); \
	done
	@echo "Done."

test:
	@echo "Running tests..."
	for s in $(SERVICES); do \
		(cd $(SERVICES_DIR)/$$s && go test ./...); \
	done
	@echo "Tests completed."

build:
	@echo "Building Docker images..."
	docker compose build
	@echo "Build completed."

build-go:
	@echo "Building Go binaries..."
	(cd services/order-service && go build -o bin/order-api ./cmd/order-api)
	(cd services/analytics-service && go build -o bin/analytics-api ./cmd/analytics-api)
	(cd services/notification-worker && go build -o bin/notification-worker ./cmd/notification-worker)
	@echo "Binaries built."

up:
	@echo "Starting services..."
	docker compose up -d
	@echo "Services started."

down:
	@echo "Stopping services..."
	docker compose down

logs:
	docker compose logs -f

clean:
	@echo "Cleaning up..."
	docker compose down -v
	@echo "Cleanup completed."

restart: down up

order:
	@echo "Creating test order..."
	curl -s -X POST http://localhost:8080/orders \
		-H "Content-Type: application/json" \
		-d '{"customer_id":"customer-123","product_id":"product-456","quantity":2,"total_amount":99.99}' | jq .
	@echo ""

analytics:
	@echo "Fetching analytics summary..."
	curl -s http://localhost:8081/analytics/summary | jq .
	@echo ""

notification:
	@echo "Notification worker logs (last 20 lines):"
	docker compose logs --tail=20 notification-worker
