# Testing Guide

This guide explains how to test the Event-Driven Order System.

## Prerequisites

- Docker and Docker Compose installed
- `curl` and `jq` for API testing (optional but recommended)
- Make (optional, but makes commands easier)

## Quick Start

### 1. Start the System

```bash
# Using Make (recommended)
make up

# Or using docker-compose directly
docker-compose up --build -d
```

Wait about 10-15 seconds for all services to initialize.

### 2. Run Automated Tests

```bash
./test.sh
```

This script will:
- Check service health
- Create test orders
- Verify order retrieval
- Check analytics aggregation
- Verify notification worker
- Test metrics endpoints

## Manual Testing

### Test Order Creation

```bash
# Create an order
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "customer-123",
    "product_id": "product-456",
    "quantity": 2,
    "total_amount": 99.99
  }'

# Or use Make
make order
```

Expected response:
```json
{
  "id": "uuid-generated-id",
  "customer_id": "customer-123",
  "product_id": "product-456",
  "quantity": 2,
  "total_amount": 99.99,
  "status": "pending",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

### Test Order Retrieval

```bash
# Get order by ID
curl http://localhost:8080/orders/{order-id}

# List all orders
curl http://localhost:8080/orders
```

### Test Analytics

```bash
# Get analytics summary
curl http://localhost:8081/analytics/summary

# Or use Make
make analytics
```

Expected response:
```json
{
  "total_orders": 10,
  "total_revenue": 999.90,
  "average_order_size": 99.99,
  "last_updated": "2026-01-01T00:00:00Z"
}
```

### Check Service Health

```bash
# Order Service
curl http://localhost:8080/health

# Analytics Service
curl http://localhost:8081/health
```

### View Metrics

```bash
# Order Service metrics
curl http://localhost:8080/metrics

# Analytics Service metrics
curl http://localhost:8081/metrics
```

## Verify Event Flow

### 1. Create an Order

```bash
ORDER_ID=$(curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "test-customer",
    "product_id": "test-product",
    "quantity": 1,
    "total_amount": 50.00
  }' | jq -r '.id')

echo "Created order: $ORDER_ID"
```

### 2. Check Analytics Service Logs

```bash
docker-compose logs -f analytics-service
```

You should see logs like:
```
Received OrderCreated event: OrderID=xxx, CustomerID=test-customer, Amount=50.00
Successfully processed order event: OrderID=xxx, Amount=50.00
```

### 3. Check Notification Worker Logs

```bash
docker-compose logs -f notification-worker

# Or use Make
make notification
```

You should see logs like:
```
📧 [NOTIFICATION] Order xxx created for customer test-customer
   Product: test-product, Quantity: 1, Total: $50.00
```

### 4. Verify Analytics Update

```bash
curl http://localhost:8081/analytics/summary | jq '.'
```

The `total_orders` and `total_revenue` should reflect the new order.

## RabbitMQ Management UI

Access the RabbitMQ management interface at:
- URL: http://localhost:15672
- Username: guest
- Password: guest

Here you can:
- View exchanges and queues
- Monitor message rates
- See consumer connections
- Debug message routing

## Testing Cache Behavior

### Test Cache-Aside Pattern (Order Service)

```bash
# First request (cache miss)
time curl http://localhost:8080/orders/{order-id}

# Second request (cache hit - should be faster)
time curl http://localhost:8080/orders/{order-id}
```

Check logs to see cache hits/misses:
```bash
docker-compose logs order-service | grep -i cache
```

### Test Analytics Cache

```bash
# First request (cache miss)
time curl http://localhost:8081/analytics/summary

# Second request (cache hit - should be faster)
time curl http://localhost:8081/analytics/summary
```

The cache TTL is 5 minutes for analytics, so results are cached for better performance.

## Load Testing (Optional)

Create multiple orders to test system behavior:

```bash
for i in {1..10}; do
  curl -X POST http://localhost:8080/orders \
    -H "Content-Type: application/json" \
    -d "{
      \"customer_id\": \"customer-$i\",
      \"product_id\": \"product-$i\",
      \"quantity\": $i,
      \"total_amount\": $((i * 10)).00
    }"
  echo ""
done
```

Then check analytics:
```bash
curl http://localhost:8081/analytics/summary | jq '.'
```

## Troubleshooting

### Services Not Starting

```bash
# Check service status
docker-compose ps

# Check logs for errors
docker-compose logs

# Restart a specific service
docker-compose restart order-service
```

### Database Connection Issues

```bash
# Check MySQL logs
docker-compose logs order-db
docker-compose logs analytics-db

# Verify databases are healthy
docker-compose ps
```

### RabbitMQ Connection Issues

```bash
# Check RabbitMQ logs
docker-compose logs rabbitmq

# Verify RabbitMQ is healthy
curl http://localhost:15672/api/overview -u guest:guest
```

### Reset Everything

```bash
# Stop and remove all containers and volumes
make clean

# Start fresh
make up
```

## Running Unit Tests

```bash
# Run all tests
make test

# Run tests for specific service
cd services/order-service && go test -v ./tests/...
```

## Monitoring

### View All Logs

```bash
# Follow all service logs
make logs

# View specific service
docker-compose logs -f order-service
docker-compose logs -f analytics-service
docker-compose logs -f notification-worker
```

### Check Resource Usage

```bash
docker stats
```

## Expected Behavior

1. **Order Creation**: Order Service saves to MySQL, caches in Redis, publishes event to RabbitMQ
2. **Event Processing**: Analytics Service and Notification Worker consume the event
3. **Analytics Update**: Analytics Service saves metrics to its own MySQL database
4. **Cache Invalidation**: Analytics cache is invalidated to ensure fresh data
5. **Notifications**: Notification Worker logs the order details

## Success Criteria

- ✅ Order Service accepts and stores orders
- ✅ Orders can be retrieved by ID (with caching)
- ✅ Events are published to RabbitMQ
- ✅ Analytics Service consumes events and updates metrics
- ✅ Notification Worker consumes events and logs notifications
- ✅ Analytics summary reflects all processed orders
- ✅ Health checks return 200 OK
- ✅ Metrics endpoints are accessible
- ✅ Services are loosely coupled (can restart independently)

## Clean Up

```bash
# Stop services
make down

# Remove all data
make clean
```
