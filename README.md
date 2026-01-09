# Event-Driven Order System

A production-inspired **event-driven order management system** built with **Go**, focusing on scalable microservices architecture, asynchronous event processing, distributed system design, and production-ready operational practices.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Services](#services)
- [Project Structure](#project-structure)
- [Quick Start](#quick-start)
- [API Documentation](#api-documentation)
- [Development Guide](#development-guide)
- [Testing](#testing)
- [Deployment](#deployment)
- [Monitoring & Observability](#monitoring--observability)
- [Design Decisions](#design-decisions)
- [Future Improvements](#future-improvements)
- [Contributing](#contributing)
- [License](#license)

## Overview

### Key Features

- **Microservices Architecture**: Independent services with clear responsibilities
- **Event-Driven Communication**: Asynchronous messaging via RabbitMQ for service decoupling
- **Distributed Persistence**: Each service maintains its own database (database-per-service pattern)
- **Caching Strategy**: Redis cache-aside pattern for performance optimization
- **Clean Architecture**: Layered design with clear separation of concerns (handlers, services, repositories)
- **Production-Ready Patterns**: Proper error handling, structured logging, and metrics

### Technology Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Language | Go 1.21+ | High-performance backend |
| Message Broker | RabbitMQ 3 | Asynchronous event communication |
| Caching | Redis 7 | In-memory data cache |
| Primary Database | MySQL 8 | Persistent storage |
| Containerization | Docker | Consistent deployment environment |
| Orchestration | Docker Compose | Local development orchestration |

## Architecture

### High-Level System Design

```
┌─────────────────────────────────────────────────────────────────┐
│                          External Client                        │
└────────────────┬─────────────────────────────────────────────────┘
                 │ REST API
                 ▼
         ┌──────────────────┐
         │  Order Service   │─── MySQL (order_db)
         │  :8080           │─── Redis (cache)
         └────────┬─────────┘
                  │ Publish Events
                  │ OrderCreated
                  │ OrderProcessed
                  ▼
         ┌──────────────────┐
         │    RabbitMQ      │
         │    :5672         │
         └────────┬─────────┘
                  │
         ┌────────┴──────────┐
         │                   │
         ▼                   ▼
    ┌──────────────┐   ┌──────────────────┐
    │ Analytics    │   │  Notification    │
    │ Service      │   │  Worker          │
    │ :8081        │   │  (background)    │
    └──────────────┘   └──────────────────┘
    │ MySQL        │   (Event consumer)
    │ Redis        │
```

### Architectural Principles

1. **Service Decoupling**: Services communicate through events, not direct API calls
2. **Database Isolation**: Each service owns its data; no shared databases
3. **Asynchronous-First**: Event-driven design for scalability and resilience
4. **Failure Isolation**: Service failures don't cascade to dependent services
5. **Clear Contracts**: Well-defined event schemas for inter-service communication

## Services

### 1. Order Service

**Responsibilities:**
- Accept and validate order creation requests
- Persist orders to database
- Cache frequently accessed orders
- Publish domain events for order state changes

**Technology:**
- REST API (Go)
- MySQL Database (order_db)
- Redis Cache
- RabbitMQ Producer

**Key Features:**
- Order creation with validation
- Order retrieval with cache-aside pattern
- Event publishing for downstream services

**Environment Variables:**
```
DB_HOST=localhost
DB_PORT=3306
DB_USER=orderuser
DB_PASSWORD=orderpass
DB_NAME=order_db
REDIS_HOST=localhost
REDIS_PORT=6379
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
SERVICE_PORT=8080
```

---

### 2. Analytics Service

**Responsibilities:**
- Consume order events from message queue
- Aggregate business metrics (total orders, revenue, etc.)
- Provide analytics query endpoints
- Cache aggregated data

**Technology:**
- RabbitMQ Consumer
- MySQL Database (analytics_db)
- Redis Cache
- REST API (Go)

**Key Features:**
- Event consumption from RabbitMQ
- Real-time metrics aggregation
- Analytics API endpoints

**Environment Variables:**
```
DB_HOST=localhost
DB_PORT=3306
DB_USER=analyticsuser
DB_PASSWORD=analyticspass
DB_NAME=analytics_db
REDIS_HOST=localhost
REDIS_PORT=6379
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
SERVICE_PORT=8081
```

---

### 3. Notification Worker

**Responsibilities:**
- Consume order events from message queue
- Simulate notification sending (email, SMS, etc.)
- Log notification events
- Demonstrate event fan-out pattern

**Technology:**
- RabbitMQ Consumer
- Event processing (in-memory)

**Key Features:**
- Background event processing
- Multiple consumer pattern demonstration
- Extensible notification system

**Environment Variables:**
```
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

## Project Structure

```
event-drive-order-system/
├── services/
│   ├── order-service/
│   │   ├── cmd/order-api/
│   │   │   └── main.go                 # Application entry point
│   │   ├── internal/
│   │   │   ├── handler/                # HTTP request handlers
│   │   │   │   └── order_handler.go
│   │   │   ├── service/                # Business logic layer
│   │   │   │   └── order_service.go
│   │   │   ├── repository/             # Data access layer
│   │   │   │   └── order_repository.go
│   │   │   ├── mq/                     # Message queue (RabbitMQ)
│   │   │   │   └── publisher.go
│   │   │   ├── cache/                  # Redis caching
│   │   │   │   └── order_cache.go
│   │   │   └── model/                  # Domain models
│   │   │       └── order.go
│   │   ├── migrations/                 # Database schema
│   │   ├── tests/
│   │   │   └── order_service_test.go
│   │   ├── go.mod
│   │   ├── go.sum
│   │   └── Dockerfile
│   │
│   ├── analytics-service/
│   │   ├── cmd/analytics-api/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   ├── mq/
│   │   │   ├── cache/
│   │   │   └── model/
│   │   ├── migrations/
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   └── notification-worker/
│       ├── cmd/notification-worker/
│       │   └── main.go
│       ├── internal/
│       │   ├── mq/
│       │   └── model/
│       ├── go.mod
│       └── Dockerfile
│
├── docs/                              # Architecture and design docs
├── scripts/                           # Utility scripts
├── docker-compose.yml                 # Local development environment
├── Makefile                           # Build and development targets
├── test.sh                            # Test runner script
├── LICENSE
└── README.md
```

## Quick Start

### Prerequisites

- Docker & Docker Compose (v2.10+)
- Go 1.21+ (optional, for local development)
- Make (optional, for convenience commands)

### Running with Docker Compose

```bash
# Start all services and dependencies
docker compose up --build

# Wait for services to be healthy (~30 seconds)
# Services are ready when all containers show "healthy" status
```

**Service Endpoints:**
- Order Service: `http://localhost:8080`
- Analytics Service: `http://localhost:8081`
- RabbitMQ Management: `http://localhost:15672` (guest/guest)
- MySQL (Order DB): `localhost:3306` (user: orderuser, pass: orderpass)
- MySQL (Analytics DB): `localhost:3307` (user: analyticsuser, pass: analyticspass)
- Redis: `localhost:6379`

### Quick Test

```bash
# Create an order (using Makefile)
make order

# Get analytics summary
make analytics

# View notification worker logs
make notification
```

### Stopping Services

```bash
# Stop all services
docker compose down

# Remove volumes (clean state)
docker compose down -v
```

## API Documentation

### Order Service

#### Create Order

Create a new order in the system. This triggers an async event that gets consumed by analytics and notification services.

**Request:**
```http
POST /orders
Content-Type: application/json

{
  "customer_id": "customer-123",
  "product_id": "product-456",
  "quantity": 2,
  "total_amount": 99.99
}
```

**Response (201 Created):**
```json
{
  "id": "order-uuid-xxxx",
  "customer_id": "customer-123",
  "product_id": "product-456",
  "quantity": 2,
  "total_amount": 99.99,
  "status": "created",
  "created_at": "2026-01-09T12:34:56Z",
  "updated_at": "2026-01-09T12:34:56Z"
}
```

**Error Response (400 Bad Request):**
```json
{
  "error": "invalid_request",
  "message": "total_amount must be greater than 0",
  "details": {}
}
```

**Using curl:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "customer-123",
    "product_id": "product-456",
    "quantity": 2,
    "total_amount": 99.99
  }'
```

---

#### Get Order

Retrieve a previously created order. Results are cached in Redis for subsequent requests.

**Request:**
```http
GET /orders/{order_id}
```

**Response (200 OK):**
```json
{
  "id": "order-uuid-xxxx",
  "customer_id": "customer-123",
  "product_id": "product-456",
  "quantity": 2,
  "total_amount": 99.99,
  "status": "created",
  "created_at": "2026-01-09T12:34:56Z",
  "updated_at": "2026-01-09T12:34:56Z"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "not_found",
  "message": "order not found",
  "details": {}
}
```

**Using curl:**
```bash
curl http://localhost:8080/orders/order-uuid-xxxx
```

---

### Analytics Service

#### Get Summary

Retrieve aggregated analytics metrics. Results are cached and updated as new orders arrive via events.

**Request:**
```http
GET /analytics/summary
```

**Response (200 OK):**
```json
{
  "total_orders": 42,
  "total_revenue": 12450.50,
  "average_order_value": 296.44,
  "last_updated": "2026-01-09T12:45:30Z"
}
```

**Using curl:**
```bash
curl http://localhost:8081/analytics/summary
```

---

## Development Guide

### Local Development Setup

```bash
# Clone the repository
git clone https://github.com/anvndev/event-drive-order-system.git
cd event-drive-order-system

# Start dependencies only (MySQL, Redis, RabbitMQ)
docker compose up -d order-db analytics-db redis rabbitmq

# In separate terminals, run each service locally:

# Terminal 1: Order Service
cd services/order-service
go run cmd/order-api/main.go

# Terminal 2: Analytics Service
cd services/analytics-service
go run cmd/analytics-api/main.go

# Terminal 3: Notification Worker
cd services/notification-worker
go run cmd/notification-worker/main.go
```

### Available Make Commands

```bash
make help              # Show all available commands
make tidy              # Tidy go modules for all services
make test              # Run all tests
make build             # Build Docker images
make build-go          # Build Go binaries locally
make up                # Start all services
make down              # Stop all services
make logs              # Stream logs from all services
make restart           # Restart all services
make clean             # Remove containers and volumes
make order             # Create a test order
make analytics         # Fetch analytics summary
make notification      # View notification worker logs
```

### Code Style & Standards

This project follows Go best practices:

- **Package Structure**: Domain-driven design with layered architecture
- **Error Handling**: Explicit error handling without panics in production code
- **Naming**: Clear, descriptive names following Go conventions
- **Comments**: Package-level documentation and exported function comments
- **Testing**: Unit tests for business logic with mocked dependencies

### Adding a New Service

1. Create directory structure in `services/new-service/`
2. Initialize Go module: `go mod init github.com/anvndev/new-service`
3. Implement layers: handler → service → repository
4. Add Dockerfile to service root
5. Update docker-compose.yml with service definition
6. Update Makefile with new service

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests for specific service
cd services/order-service && go test ./...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestOrderServiceCreate ./...
```

### Testing Strategy

**Unit Tests:**
- Business logic in service layer
- Repository interfaces with mocked dependencies
- Edge cases and error handling

**Integration Tests:**
- Database operations
- Cache behavior
- Message queue integration

**Test Example:**
```bash
cd services/order-service && go test -v ./tests/
```

## Deployment

### Docker Compose (Development)

```bash
docker compose up --build
```

### Docker Build Manually

```bash
# Build individual services
docker build -t order-service:latest services/order-service/
docker build -t analytics-service:latest services/analytics-service/
docker build -t notification-worker:latest services/notification-worker/

# Run services with docker run
docker run -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e RABBITMQ_URL=amqp://guest:guest@host.docker.internal:5672/ \
  order-service:latest
```

### Environment Configuration

Create a `.env` file in the project root:

```env
# Database Configuration
DB_ROOT_PASSWORD=rootpassword

# Logging
LOG_LEVEL=info

# Application
ENVIRONMENT=development
```

## Monitoring & Observability

### Structured Logging

All services implement structured logging:

```
{"level":"info","service":"order-service","event":"order_created","order_id":"123","timestamp":"2026-01-09T12:34:56Z"}
```

### Metrics Endpoint

Metrics are exposed on `/metrics` endpoint (Prometheus format):

```bash
curl http://localhost:8080/metrics
```

**Available Metrics:**
- `orders_created_total` - Total orders created (counter)
- `orders_created_duration_seconds` - Order creation latency (histogram)
- `http_request_duration_seconds` - HTTP request latency (histogram)
- `http_requests_total` - Total HTTP requests (counter)

### RabbitMQ Management UI

Monitor queue health and message throughput:

```
http://localhost:15672
Username: guest
Password: guest
```

### Health Checks

Each service provides basic health checks via logs. Monitor service health:

```bash
# Check service logs
docker compose logs order-service
docker compose logs analytics-service
docker compose logs notification-worker
```

## Design Decisions

### Why Event-Driven Architecture?

**Benefits:**
- **Decoupling**: Services don't know about each other
- **Scalability**: Handle traffic spikes independently
- **Resilience**: Failure in one service doesn't affect others
- **Auditability**: Complete event history

**Trade-offs:**
- Increased complexity in distributed tracing
- Potential message duplication handling
- Eventual consistency instead of strong consistency

### Why RabbitMQ?

- **Reliability**: Message persistence and acknowledgments
- **Flexibility**: Multiple exchange types and routing patterns
- **Management**: Built-in UI for monitoring
- **Proven**: Battle-tested in production systems
- **Ecosystem**: Excellent Go client libraries

### Why Database-Per-Service?

**Advantages:**
- Services can choose optimal database technology
- Independent scaling and tuning
- No bottleneck shared database
- Clear service boundaries

**Considerations:**
- No ACID transactions across services
- Data consistency eventual (not immediate)
- Join operations across services require application logic

### Why Redis Cache?

- **Performance**: In-memory data structure store
- **Simplicity**: Cache-aside pattern is straightforward
- **Ubiquity**: Industry standard for caching

**Cache Invalidation Strategy:**
- TTL-based expiration
- Manual invalidation on data updates
- Cache warming during service startup

## Event Schema

### OrderCreated Event

Emitted when an order is successfully created.

```json
{
  "event_type": "OrderCreated",
  "event_id": "event-uuid-xxxx",
  "timestamp": "2026-01-09T12:34:56Z",
  "data": {
    "order_id": "order-uuid-xxxx",
    "customer_id": "customer-123",
    "product_id": "product-456",
    "quantity": 2,
    "total_amount": 99.99
  }
}
```

### OrderProcessed Event

Emitted when an order is processed by the analytics service.

```json
{
  "event_type": "OrderProcessed",
  "event_id": "event-uuid-yyyy",
  "timestamp": "2026-01-09T12:35:00Z",
  "data": {
    "order_id": "order-uuid-xxxx",
    "processed_by": "analytics-service",
    "metrics_updated": true
  }
}
```

## Future Improvements

- [ ] **gRPC Communication**: High-performance inter-service communication
- [ ] **Dead-Letter Queue (DLQ)**: Handle failed message processing
- [ ] **Distributed Tracing**: OpenTelemetry integration with Jaeger
- [ ] **Circuit Breaker Pattern**: Prevent cascading failures
- [ ] **Rate Limiting**: Protect services from overload
- [ ] **Authentication & Authorization**: JWT tokens and RBAC
- [ ] **API Versioning**: Support multiple API versions
- [ ] **GraphQL Gateway**: Alternative query interface
- [ ] **Kubernetes Deployment**: Production container orchestration
- [ ] **Blue-Green Deployment**: Zero-downtime deployments
- [ ] **Kafka Migration**: Consider for higher throughput requirements
- [ ] **CQRS Pattern**: Separate read and write models
- [ ] **Saga Pattern**: Distributed transactions across services

## Troubleshooting

### Service Won't Start

**Symptom**: `Connection refused` errors

**Solutions:**
```bash
# Ensure all dependencies are healthy
docker compose ps

# Check service logs
docker compose logs [service-name]

# Verify database connectivity
docker compose exec order-db mysql -u orderuser -p -e "SELECT 1;"

# Restart services
docker compose down -v && docker compose up --build
```

### High Memory Usage

**Symptom**: Services consuming excessive memory

**Solutions:**
- Check for memory leaks in logs
- Verify Redis key expiration is working: `docker exec redis redis-cli INFO memory`
- Reduce cache TTL values
- Monitor with: `docker stats`

### Messages Not Being Processed

**Symptom**: Orders created but analytics not updating

**Solutions:**
```bash
# Check RabbitMQ queue status
curl -u guest:guest http://localhost:15672/api/queues

# Verify consumer is running
docker compose logs analytics-service | grep "consumer"

# Check for dead-letter messages
docker compose logs -f analytics-service
```

### Database Connection Issues

**Symptom**: `Access denied` or connection timeouts

**Solutions:**
```bash
# Test connection manually
docker compose exec order-db mysql -h order-db -u orderuser -porderpass -e "SELECT 1;"

# Check environment variables
docker compose config | grep DB_

# Verify network connectivity
docker compose exec order-service ping order-db
```

### Redis Cache Not Working

**Symptom**: Cache hits not happening

**Solutions:**
```bash
# Connect to Redis CLI
docker exec -it redis redis-cli

# List all keys
> KEYS *

# Check key TTL
> TTL [key-name]

# Flush cache
> FLUSHALL
```

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following Go best practices
4. Write or update tests
5. Commit with clear messages (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Author

**andev0x** - [GitHub](https://github.com/andev0x)

> Backend-focused project showcasing Go microservices, event-driven architecture, and production-ready engineering practices.

### Key Learning Objectives

- Microservices design and communication patterns
- Event-driven architecture fundamentals
- Distributed system resilience patterns
- Go backend best practices
- System design trade-offs and decisions

---

**Last Updated**: January 2026

**Project Status**: Active Development

**Go Version**: 1.21+

**License**: [MIT](LICENSE)

