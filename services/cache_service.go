package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/go-redis/redis/v8"
)

// CacheService implements the CacheService interface using Redis
type CacheService struct {
	client *redis.Client
	logger interfaces.Logger
}

// NewCacheService creates a new CacheService instance
func NewCacheService(redisAddr, redisPassword string, redisDB int, logger interfaces.Logger) (interfaces.CacheService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
		PoolSize: 10,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Cache service initialized successfully", "redis_addr", redisAddr)

	return &CacheService{
		client: client,
		logger: logger,
	}, nil
}

// Get retrieves a value from cache
func (c *CacheService) Get(key string) (interface{}, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false
		}
		c.logger.Error("Failed to get from cache", "key", key, "error", err)
		return nil, false
	}

	// Try to unmarshal as JSON first
	var data interface{}
	if err := json.Unmarshal([]byte(result), &data); err == nil {
		return data, true
	}

	// Return as string if not JSON
	return result, true
}

// Set stores a value in cache with TTL
func (c *CacheService) Set(key string, value interface{}, ttl int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Marshal to JSON for complex types
	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Error("Failed to marshal cache value", "key", key, "error", err)
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	// Set with TTL
	duration := time.Duration(ttl) * time.Second
	if err := c.client.Set(ctx, key, data, duration).Err(); err != nil {
		c.logger.Error("Failed to set cache value", "key", key, "error", err)
		return fmt.Errorf("failed to set cache value: %w", err)
	}

	return nil
}

// Delete removes a value from cache
func (c *CacheService) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := c.client.Del(ctx, key).Err(); err != nil {
		c.logger.Error("Failed to delete from cache", "key", key, "error", err)
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}

// Clear removes all values from cache
func (c *CacheService) Clear() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := c.client.FlushDB(ctx).Err(); err != nil {
		c.logger.Error("Failed to clear cache", "error", err)
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	return nil
}

// Exists checks if a key exists in cache
func (c *CacheService) Exists(key string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		c.logger.Error("Failed to check cache existence", "key", key, "error", err)
		return false
	}

	return result > 0
}

// Close closes the Redis connection
func (c *CacheService) Close() error {
	return c.client.Close()
}

// GetWithTTL retrieves a value and its TTL from cache
func (c *CacheService) GetWithTTL(key string) (interface{}, time.Duration, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Get value
	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, 0, false
		}
		c.logger.Error("Failed to get from cache", "key", key, "error", err)
		return nil, 0, false
	}

	// Get TTL
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		c.logger.Error("Failed to get TTL from cache", "key", key, "error", err)
		ttl = 0
	}

	// Try to unmarshal as JSON first
	var data interface{}
	if err := json.Unmarshal([]byte(result), &data); err == nil {
		return data, ttl, true
	}

	// Return as string if not JSON
	return result, ttl, true
}

// SetNX sets a value only if it doesn't exist
func (c *CacheService) SetNX(key string, value interface{}, ttl int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Marshal to JSON for complex types
	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Error("Failed to marshal cache value", "key", key, "error", err)
		return false, fmt.Errorf("failed to marshal cache value: %w", err)
	}

	// Set with TTL only if not exists
	duration := time.Duration(ttl) * time.Second
	result, err := c.client.SetNX(ctx, key, data, duration).Result()
	if err != nil {
		c.logger.Error("Failed to set cache value", "key", key, "error", err)
		return false, fmt.Errorf("failed to set cache value: %w", err)
	}

	return result, nil
}

// Increment increments a numeric value in cache
func (c *CacheService) Increment(key string, value int64) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := c.client.IncrBy(ctx, key, value).Result()
	if err != nil {
		c.logger.Error("Failed to increment cache value", "key", key, "error", err)
		return 0, fmt.Errorf("failed to increment cache value: %w", err)
	}

	return result, nil
}

// DeletePattern deletes all keys matching a pattern
func (c *CacheService) DeletePattern(pattern string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			c.logger.Error("Failed to delete pattern from cache", "pattern", pattern, "key", iter.Val(), "error", err)
		}
	}

	if err := iter.Err(); err != nil {
		c.logger.Error("Failed to iterate over cache keys", "pattern", pattern, "error", err)
		return fmt.Errorf("failed to iterate over cache keys: %w", err)
	}

	return nil
}
