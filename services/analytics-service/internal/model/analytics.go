package model

import (
	"time"
)

// OrderCreatedEvent represents the event consumed from RabbitMQ
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

// OrderMetric represents aggregated order metrics
type OrderMetric struct {
	ID          int       `json:"id"`
	OrderID     string    `json:"order_id"`
	CustomerID  string    `json:"customer_id"`
	ProductID   string    `json:"product_id"`
	Quantity    int       `json:"quantity"`
	TotalAmount float64   `json:"total_amount"`
	ProcessedAt time.Time `json:"processed_at"`
}

// AnalyticsSummary represents aggregated analytics data
type AnalyticsSummary struct {
	TotalOrders      int       `json:"total_orders"`
	TotalRevenue     float64   `json:"total_revenue"`
	AverageOrderSize float64   `json:"average_order_size"`
	LastUpdated      time.Time `json:"last_updated"`
}
