// Package cache provides caching implementations for analytics data.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/andev0x/analytics-service/internal/model"
	"github.com/redis/go-redis/v9"
)

const (
	summaryKey = "analytics:summary"
	summaryTTL = 5 * time.Minute
)

// AnalyticsCache interface defines methods for caching analytics data
type AnalyticsCache interface {
	GetSummary(ctx context.Context) (*model.AnalyticsSummary, error)
	SetSummary(ctx context.Context, summary *model.AnalyticsSummary) error
	InvalidateSummary(ctx context.Context) error
}

// RedisAnalyticsCache implements AnalyticsCache using Redis
type RedisAnalyticsCache struct {
	client *redis.Client
}

// NewRedisAnalyticsCache creates a new Redis analytics cache
func NewRedisAnalyticsCache(client *redis.Client) *RedisAnalyticsCache {
	return &RedisAnalyticsCache{client: client}
}

// GetSummary retrieves analytics summary from cache
func (c *RedisAnalyticsCache) GetSummary(ctx context.Context) (*model.AnalyticsSummary, error) {
	data, err := c.client.Get(ctx, summaryKey).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("summary not found in cache")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get summary from cache: %w", err)
	}

	var summary model.AnalyticsSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, fmt.Errorf("failed to unmarshal summary: %w", err)
	}

	return &summary, nil
}

// SetSummary stores analytics summary in cache
func (c *RedisAnalyticsCache) SetSummary(ctx context.Context, summary *model.AnalyticsSummary) error {
	data, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	if err := c.client.Set(ctx, summaryKey, data, summaryTTL).Err(); err != nil {
		return fmt.Errorf("failed to set summary in cache: %w", err)
	}

	return nil
}

// InvalidateSummary removes analytics summary from cache
func (c *RedisAnalyticsCache) InvalidateSummary(ctx context.Context) error {
	if err := c.client.Del(ctx, summaryKey).Err(); err != nil {
		return fmt.Errorf("failed to invalidate summary cache: %w", err)
	}
	return nil
}

// InitRedis initializes Redis client
func InitRedis(host, port string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		Password:     "", // no password
		DB:           0,  // default DB
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}
