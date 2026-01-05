package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/andev0x/order-service/internal/model"
	"github.com/redis/go-redis/v9"
)

const (
	orderKeyPrefix = "order:"
	orderTTL       = 15 * time.Minute
)

// OrderCache interface defines methods for caching orders
type OrderCache interface {
	Get(ctx context.Context, id string) (*model.Order, error)
	Set(ctx context.Context, order *model.Order) error
	Delete(ctx context.Context, id string) error
}

// RedisOrderCache implements OrderCache using Redis
type RedisOrderCache struct {
	client *redis.Client
}

// NewRedisOrderCache creates a new Redis order cache
func NewRedisOrderCache(client *redis.Client) *RedisOrderCache {
	return &RedisOrderCache{client: client}
}

// Get retrieves an order from cache
func (c *RedisOrderCache) Get(ctx context.Context, id string) (*model.Order, error) {
	key := orderKeyPrefix + id
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("order not found in cache")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order from cache: %w", err)
	}

	var order model.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return &order, nil
}

// Set stores an order in cache
func (c *RedisOrderCache) Set(ctx context.Context, order *model.Order) error {
	key := orderKeyPrefix + order.ID
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	if err := c.client.Set(ctx, key, data, orderTTL).Err(); err != nil {
		return fmt.Errorf("failed to set order in cache: %w", err)
	}

	return nil
}

// Delete removes an order from cache
func (c *RedisOrderCache) Delete(ctx context.Context, id string) error {
	key := orderKeyPrefix + id
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete order from cache: %w", err)
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
