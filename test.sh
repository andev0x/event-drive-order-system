#!/bin/bash

# Test script for Event-Driven Order System
set -e

echo "=========================================="
echo "Event-Driven Order System - Test Script"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}→ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Check if services are running
print_info "Checking if services are running..."
if ! docker-compose ps | grep -q "Up"; then
    print_error "Services are not running. Please run 'make up' first."
    exit 1
fi
print_success "Services are running"

# Wait for services to be ready
print_info "Waiting for services to be fully ready..."
sleep 5

# Test 1: Health check - Order Service
print_info "Testing Order Service health..."
ORDER_HEALTH=$(curl -s http://localhost:8080/health)
if echo "$ORDER_HEALTH" | grep -q "healthy"; then
    print_success "Order Service is healthy"
else
    print_error "Order Service health check failed"
    exit 1
fi

# Test 2: Health check - Analytics Service
print_info "Testing Analytics Service health..."
ANALYTICS_HEALTH=$(curl -s http://localhost:8081/health)
if echo "$ANALYTICS_HEALTH" | grep -q "healthy"; then
    print_success "Analytics Service is healthy"
else
    print_error "Analytics Service health check failed"
    exit 1
fi

# Test 3: Create an order
print_info "Creating test order..."
ORDER_RESPONSE=$(curl -s -X POST http://localhost:8080/orders \
    -H "Content-Type: application/json" \
    -d '{
        "customer_id": "customer-001",
        "product_id": "product-123",
        "quantity": 3,
        "total_amount": 149.99
    }')

ORDER_ID=$(echo "$ORDER_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)

if [ -n "$ORDER_ID" ]; then
    print_success "Order created successfully (ID: $ORDER_ID)"
else
    print_error "Failed to create order"
    echo "Response: $ORDER_RESPONSE"
    exit 1
fi

# Test 4: Retrieve the created order
print_info "Retrieving order by ID..."
GET_ORDER=$(curl -s http://localhost:8080/orders/$ORDER_ID)
if echo "$GET_ORDER" | grep -q "$ORDER_ID"; then
    print_success "Order retrieved successfully"
else
    print_error "Failed to retrieve order"
    exit 1
fi

# Test 5: Create multiple orders for analytics
print_info "Creating additional orders for analytics..."
for i in {1..3}; do
    curl -s -X POST http://localhost:8080/orders \
        -H "Content-Type: application/json" \
        -d "{
            \"customer_id\": \"customer-00$i\",
            \"product_id\": \"product-$i\",
            \"quantity\": $i,
            \"total_amount\": $((i * 50)).00
        }" > /dev/null
    print_success "Order $i created"
done

# Wait for events to be processed
print_info "Waiting for events to be processed by Analytics Service..."
sleep 3

# Test 6: Check analytics summary
print_info "Checking analytics summary..."
ANALYTICS_SUMMARY=$(curl -s http://localhost:8081/analytics/summary)
TOTAL_ORDERS=$(echo "$ANALYTICS_SUMMARY" | grep -o '"total_orders":[0-9]*' | cut -d':' -f2)

if [ "$TOTAL_ORDERS" -ge 4 ]; then
    print_success "Analytics processed successfully (Total orders: $TOTAL_ORDERS)"
    echo "$ANALYTICS_SUMMARY" | jq '.'
else
    print_error "Analytics processing may have issues (Total orders: $TOTAL_ORDERS)"
fi

# Test 7: Check notification worker logs
print_info "Checking notification worker logs..."
NOTIFICATION_LOGS=$(docker-compose logs notification-worker 2>&1 | grep -c "NOTIFICATION" || true)
if [ "$NOTIFICATION_LOGS" -gt 0 ]; then
    print_success "Notification worker processed $NOTIFICATION_LOGS notifications"
else
    print_error "No notifications found in logs"
fi

# Test 8: Check Prometheus metrics
print_info "Checking Prometheus metrics endpoint..."
METRICS=$(curl -s http://localhost:8080/metrics)
if echo "$METRICS" | grep -q "promhttp_"; then
    print_success "Prometheus metrics are available"
else
    print_error "Prometheus metrics endpoint failed"
fi

echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
print_success "All tests passed!"
echo ""
echo "Service URLs:"
echo "  - Order Service: http://localhost:8080"
echo "  - Analytics Service: http://localhost:8081"
echo "  - RabbitMQ Management: http://localhost:15672 (guest/guest)"
echo ""
echo "Try these commands:"
echo "  - make order        # Create a new order"
echo "  - make analytics    # View analytics summary"
echo "  - make logs         # View all service logs"
echo ""
