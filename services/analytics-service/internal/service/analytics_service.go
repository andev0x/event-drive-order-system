package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/andev0x/analytics-service/internal/cache"
	"github.com/andev0x/analytics-service/internal/model"
	"github.com/andev0x/analytics-service/internal/repository"
)

// AnalyticsService handles business logic for analytics
type AnalyticsService struct {
	repo  repository.AnalyticsRepository
	cache cache.AnalyticsCache
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(repo repository.AnalyticsRepository, cache cache.AnalyticsCache) *AnalyticsService {
	return &AnalyticsService{
		repo:  repo,
		cache: cache,
	}
}

// ProcessOrderEvent processes an order created event
func (s *AnalyticsService) ProcessOrderEvent(ctx context.Context, event *model.OrderCreatedEvent) error {
	// Create metric from event
	metric := &model.OrderMetric{
		OrderID:     event.OrderID,
		CustomerID:  event.CustomerID,
		ProductID:   event.ProductID,
		Quantity:    event.Quantity,
		TotalAmount: event.TotalAmount,
		ProcessedAt: time.Now(),
	}

	// Save to database
	if err := s.repo.SaveOrderMetric(ctx, metric); err != nil {
		return fmt.Errorf("failed to save order metric: %w", err)
	}

	// Invalidate cache to force fresh calculation on next request
	if err := s.cache.InvalidateSummary(ctx); err != nil {
		log.Printf("Warning: failed to invalidate cache: %v", err)
	}

	log.Printf("Successfully processed order event: OrderID=%s, Amount=%.2f", event.OrderID, event.TotalAmount)
	return nil
}

// GetSummary retrieves analytics summary (cache-aside pattern)
func (s *AnalyticsService) GetSummary(ctx context.Context) (*model.AnalyticsSummary, error) {
	// Try to get from cache first
	summary, err := s.cache.GetSummary(ctx)
	if err == nil {
		log.Println("Cache hit for analytics summary")
		return summary, nil
	}

	log.Println("Cache miss for analytics summary, fetching from database")

	// Cache miss, get from database
	summary, err = s.repo.GetSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get summary: %w", err)
	}

	// Update cache
	if err := s.cache.SetSummary(ctx, summary); err != nil {
		log.Printf("Warning: failed to cache summary: %v", err)
	}

	return summary, nil
}
