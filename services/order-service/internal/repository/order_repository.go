// Package repository provides data access implementations for the order service.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/andev0x/order-service/internal/model"
	// MySQL driver for database/sql
	_ "github.com/go-sql-driver/mysql"
)

// OrderRepository interface defines methods for order persistence
type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id string) (*model.Order, error)
	List(ctx context.Context, limit, offset int) ([]*model.Order, error)
}

// MySQLOrderRepository implements OrderRepository using MySQL
type MySQLOrderRepository struct {
	db *sql.DB
}

// NewMySQLOrderRepository creates a new MySQL order repository
func NewMySQLOrderRepository(db *sql.DB) *MySQLOrderRepository {
	return &MySQLOrderRepository{db: db}
}

// Create inserts a new order into the database
func (r *MySQLOrderRepository) Create(ctx context.Context, order *model.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, product_id, quantity, total_amount, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		order.ID,
		order.CustomerID,
		order.ProductID,
		order.Quantity,
		order.TotalAmount,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// GetByID retrieves an order by its ID
func (r *MySQLOrderRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	query := `
		SELECT id, customer_id, product_id, quantity, total_amount, status, created_at, updated_at
		FROM orders
		WHERE id = ?
	`

	order := &model.Order{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.ProductID,
		&order.Quantity,
		&order.TotalAmount,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

// List retrieves a list of orders with pagination
func (r *MySQLOrderRepository) List(ctx context.Context, limit, offset int) ([]*model.Order, error) {
	query := `
		SELECT id, customer_id, product_id, quantity, total_amount, status, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var orders []*model.Order
	for rows.Next() {
		order := &model.Order{}
		err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.ProductID,
			&order.Quantity,
			&order.TotalAmount,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// InitDB initializes the database connection
func InitDB(host, port, user, password, dbname string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		user, password, host, port, dbname)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
