package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/andev0x/order-service/internal/cache"
	"github.com/andev0x/order-service/internal/model"
	"github.com/andev0x/order-service/internal/mq"
	"github.com/andev0x/order-service/internal/repository"
	"github.com/google/uuid"
)

// OrderService handles business logic for orders
type OrderService struct {
	repo      repository.OrderRepository
	cache     cache.OrderCache
	publisher mq.EventPublisher
}

// NewOrderService creates a new order service
func NewOrderService(repo repository.OrderRepository, cache cache.OrderCache, publisher mq.EventPublisher) *OrderService {
	return &OrderService{
		repo:      repo,
		cache:     cache,
		publisher: publisher,
	}
}

// CreateOrder creates a new order
func (s *OrderService) CreateOrder(ctx context.Context, req *model.CreateOrderRequest) (*model.Order, error) {
	// Validate request
	if req.CustomerID == "" || req.ProductID == "" {
		return nil, fmt.Errorf("customer_id and product_id are required")
	}
	if req.Quantity <= 0 {
		return nil, fmt.Errorf("quantity must be greater than 0")
	}
	if req.TotalAmount <= 0 {
		return nil, fmt.Errorf("total_amount must be greater than 0")
	}

	// Create order entity
	order := &model.Order{
		ID:          uuid.New().String(),
		CustomerID:  req.CustomerID,
		ProductID:   req.ProductID,
		Quantity:    req.Quantity,
		TotalAmount: req.TotalAmount,
		Status:      model.OrderStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Persist to database
	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Cache the order
	if err := s.cache.Set(ctx, order); err != nil {
		log.Printf("Warning: failed to cache order %s: %v", order.ID, err)
	}

	// Publish event asynchronously
	go func() {
		event := &model.OrderCreatedEvent{
			OrderID:     order.ID,
			CustomerID:  order.CustomerID,
			ProductID:   order.ProductID,
			Quantity:    order.Quantity,
			TotalAmount: order.TotalAmount,
			Status:      order.Status,
			CreatedAt:   order.CreatedAt,
		}

		if err := s.publisher.PublishOrderCreated(context.Background(), event); err != nil {
			log.Printf("Error: failed to publish order created event for order %s: %v", order.ID, err)
		}
	}()

	log.Printf("Order created successfully: %s", order.ID)
	return order, nil
}

// GetOrderByID retrieves an order by ID (cache-aside pattern)
func (s *OrderService) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	// Try to get from cache first
	order, err := s.cache.Get(ctx, id)
	if err == nil {
		log.Printf("Cache hit for order: %s", id)
		return order, nil
	}

	log.Printf("Cache miss for order: %s, fetching from database", id)

	// Cache miss, get from database
	order, err = s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Update cache
	if err := s.cache.Set(ctx, order); err != nil {
		log.Printf("Warning: failed to cache order %s: %v", id, err)
	}

	return order, nil
}

// ListOrders retrieves a list of orders
func (s *OrderService) ListOrders(ctx context.Context, limit, offset int) ([]*model.Order, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	orders, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	return orders, nil
}
