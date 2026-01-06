package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/andev0x/order-service/internal/model"
	"github.com/andev0x/order-service/internal/service"
)

// MockOrderRepository is a mock implementation of OrderRepository
type MockOrderRepository struct {
	CreateFunc  func(ctx context.Context, order *model.Order) error
	GetByIDFunc func(ctx context.Context, id string) (*model.Order, error)
	ListFunc    func(ctx context.Context, limit, offset int) ([]*model.Order, error)
}

func (m *MockOrderRepository) Create(ctx context.Context, order *model.Order) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, order)
	}
	return nil
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockOrderRepository) List(ctx context.Context, limit, offset int) ([]*model.Order, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, limit, offset)
	}
	return nil, errors.New("not implemented")
}

// MockOrderCache is a mock implementation of OrderCache
type MockOrderCache struct {
	GetFunc    func(ctx context.Context, id string) (*model.Order, error)
	SetFunc    func(ctx context.Context, order *model.Order) error
	DeleteFunc func(ctx context.Context, id string) error
}

func (m *MockOrderCache) Get(ctx context.Context, id string) (*model.Order, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *MockOrderCache) Set(ctx context.Context, order *model.Order) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, order)
	}
	return nil
}

func (m *MockOrderCache) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// MockEventPublisher is a mock implementation of EventPublisher
type MockEventPublisher struct {
	PublishOrderCreatedFunc func(ctx context.Context, event *model.OrderCreatedEvent) error
}

func (m *MockEventPublisher) PublishOrderCreated(ctx context.Context, event *model.OrderCreatedEvent) error {
	if m.PublishOrderCreatedFunc != nil {
		return m.PublishOrderCreatedFunc(ctx, event)
	}
	return nil
}

func (m *MockEventPublisher) Close() error {
	return nil
}

// TestCreateOrder tests the CreateOrder method
func TestCreateOrder(t *testing.T) {
	tests := []struct {
		name        string
		request     *model.CreateOrderRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "valid order",
			request: &model.CreateOrderRequest{
				CustomerID:  "customer-123",
				ProductID:   "product-456",
				Quantity:    2,
				TotalAmount: 99.99,
			},
			wantErr: false,
		},
		{
			name: "missing customer_id",
			request: &model.CreateOrderRequest{
				CustomerID:  "",
				ProductID:   "product-456",
				Quantity:    2,
				TotalAmount: 99.99,
			},
			wantErr:     true,
			errContains: "customer_id",
		},
		{
			name: "missing product_id",
			request: &model.CreateOrderRequest{
				CustomerID:  "customer-123",
				ProductID:   "",
				Quantity:    2,
				TotalAmount: 99.99,
			},
			wantErr:     true,
			errContains: "product_id",
		},
		{
			name: "invalid quantity",
			request: &model.CreateOrderRequest{
				CustomerID:  "customer-123",
				ProductID:   "product-456",
				Quantity:    0,
				TotalAmount: 99.99,
			},
			wantErr:     true,
			errContains: "quantity",
		},
		{
			name: "invalid total_amount",
			request: &model.CreateOrderRequest{
				CustomerID:  "customer-123",
				ProductID:   "product-456",
				Quantity:    2,
				TotalAmount: 0,
			},
			wantErr:     true,
			errContains: "total_amount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockOrderRepository{
				CreateFunc: func(ctx context.Context, order *model.Order) error {
					return nil
				},
			}

			mockCache := &MockOrderCache{
				SetFunc: func(ctx context.Context, order *model.Order) error {
					return nil
				},
			}

			mockPublisher := &MockEventPublisher{
				PublishOrderCreatedFunc: func(ctx context.Context, event *model.OrderCreatedEvent) error {
					return nil
				},
			}

			svc := service.NewOrderService(mockRepo, mockCache, mockPublisher)

			order, err := svc.CreateOrder(context.Background(), tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateOrder() expected error but got none")
				}
				if tt.errContains != "" && err != nil {
					// Simple string contains check
					errStr := err.Error()
					found := false
					for i := 0; i < len(errStr)-len(tt.errContains)+1; i++ {
						if errStr[i:i+len(tt.errContains)] == tt.errContains {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("CreateOrder() error = %v, want error containing %v", err, tt.errContains)
					}
				}
			} else {
				if err != nil {
					t.Errorf("CreateOrder() unexpected error = %v", err)
				}
				if order == nil {
					t.Errorf("CreateOrder() returned nil order")
				}
				if order != nil && order.Status != model.OrderStatusPending {
					t.Errorf("CreateOrder() status = %v, want %v", order.Status, model.OrderStatusPending)
				}
			}
		})
	}
}

// TestGetOrderByID tests the GetOrderByID method
func TestGetOrderByID(t *testing.T) {
	testOrder := &model.Order{
		ID:          "order-123",
		CustomerID:  "customer-123",
		ProductID:   "product-456",
		Quantity:    2,
		TotalAmount: 99.99,
		Status:      model.OrderStatusPending,
	}

	t.Run("cache hit", func(t *testing.T) {
		mockRepo := &MockOrderRepository{}
		mockCache := &MockOrderCache{
			GetFunc: func(ctx context.Context, id string) (*model.Order, error) {
				return testOrder, nil
			},
		}
		mockPublisher := &MockEventPublisher{}

		svc := service.NewOrderService(mockRepo, mockCache, mockPublisher)

		order, err := svc.GetOrderByID(context.Background(), "order-123")
		if err != nil {
			t.Errorf("GetOrderByID() unexpected error = %v", err)
		}
		if order.ID != testOrder.ID {
			t.Errorf("GetOrderByID() returned wrong order")
		}
	})

	t.Run("cache miss, db hit", func(t *testing.T) {
		mockRepo := &MockOrderRepository{
			GetByIDFunc: func(ctx context.Context, id string) (*model.Order, error) {
				return testOrder, nil
			},
		}
		mockCache := &MockOrderCache{
			GetFunc: func(ctx context.Context, id string) (*model.Order, error) {
				return nil, errors.New("not found")
			},
			SetFunc: func(ctx context.Context, order *model.Order) error {
				return nil
			},
		}
		mockPublisher := &MockEventPublisher{}

		svc := service.NewOrderService(mockRepo, mockCache, mockPublisher)

		order, err := svc.GetOrderByID(context.Background(), "order-123")
		if err != nil {
			t.Errorf("GetOrderByID() unexpected error = %v", err)
		}
		if order.ID != testOrder.ID {
			t.Errorf("GetOrderByID() returned wrong order")
		}
	})
}
