// Package model defines the data structures used throughout the order service.
package model

import (
	"time"
)

// Order represents an order in the system
type Order struct {
	ID          string    `json:"id"`
	CustomerID  string    `json:"customer_id"`
	ProductID   string    `json:"product_id"`
	Quantity    int       `json:"quantity"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	CustomerID  string  `json:"customer_id" validate:"required"`
	ProductID   string  `json:"product_id" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	TotalAmount float64 `json:"total_amount" validate:"required,gt=0"`
}

// OrderCreatedEvent represents the event published when an order is created
type OrderCreatedEvent struct {
	OrderID     string    `json:"order_id"`
	CustomerID  string    `json:"customer_id"`
	ProductID   string    `json:"product_id"`
	Quantity    int       `json:"quantity"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	EventType   string    `json:"event_type"`
}

// OrderStatus constants
const (
	OrderStatusPending   = "pending"
	OrderStatusConfirmed = "confirmed"
	OrderStatusCancelled = "cancelled"
)
