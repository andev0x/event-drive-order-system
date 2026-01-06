package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/andev0x/analytics-service/internal/model"
	_ "github.com/go-sql-driver/mysql"
)

// AnalyticsRepository interface defines methods for analytics persistence
type AnalyticsRepository interface {
	SaveOrderMetric(ctx context.Context, metric *model.OrderMetric) error
	GetSummary(ctx context.Context) (*model.AnalyticsSummary, error)
}

// MySQLAnalyticsRepository implements AnalyticsRepository using MySQL
type MySQLAnalyticsRepository struct {
	db *sql.DB
}

// NewMySQLAnalyticsRepository creates a new MySQL analytics repository
func NewMySQLAnalyticsRepository(db *sql.DB) *MySQLAnalyticsRepository {
	return &MySQLAnalyticsRepository{db: db}
}

// SaveOrderMetric inserts a new order metric into the database
func (r *MySQLAnalyticsRepository) SaveOrderMetric(ctx context.Context, metric *model.OrderMetric) error {
	query := `
		INSERT INTO order_metrics (order_id, customer_id, product_id, quantity, total_amount, processed_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		metric.OrderID,
		metric.CustomerID,
		metric.ProductID,
		metric.Quantity,
		metric.TotalAmount,
		metric.ProcessedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save order metric: %w", err)
	}

	return nil
}

// GetSummary retrieves aggregated analytics summary
func (r *MySQLAnalyticsRepository) GetSummary(ctx context.Context) (*model.AnalyticsSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_orders,
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COALESCE(AVG(total_amount), 0) as average_order_size
		FROM order_metrics
	`

	summary := &model.AnalyticsSummary{
		LastUpdated: time.Now(),
	}

	err := r.db.QueryRowContext(ctx, query).Scan(
		&summary.TotalOrders,
		&summary.TotalRevenue,
		&summary.AverageOrderSize,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get summary: %w", err)
	}

	return summary, nil
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
