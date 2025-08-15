package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// CacheService implements the CacheService interface using Redis
type CacheService struct {
	client *redis.Client
	logger interfaces.Logger
}

// NewCacheService creates a new CacheService instance
func NewCacheService(redisAddr, redisPassword string, redisDB int, logger interfaces.Logger) interfaces.CacheService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	logger.Info("Redis cache service initialized", zap.String("redis_addr", redisAddr))

	return &CacheService{
		client: rdb,
		logger: logger,
	}
}

// Get retrieves a value from cache
func (c *CacheService) Get(key string) (interface{}, bool) {
	ctx := context.Background()
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Key doesn't exist
			return nil, false
		}
		c.logger.Error("Failed to get from cache", zap.String("key", key), zap.Error(err))
		return nil, false
	}

	// Try to unmarshal JSON
	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		// If JSON unmarshal fails, return as string
		return val, true
	}

	return result, true
}

// Set stores a value in cache with TTL
func (c *CacheService) Set(key string, value interface{}, ttl int) error {
	ctx := context.Background()

	// Marshal value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Error("Failed to marshal value for cache", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	duration := time.Duration(ttl) * time.Second
	if err := c.client.Set(ctx, key, data, duration).Err(); err != nil {
		c.logger.Error("Failed to set cache", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Delete removes a key from cache
func (c *CacheService) Delete(key string) error {
	ctx := context.Background()
	if err := c.client.Del(ctx, key).Err(); err != nil {
		c.logger.Error("Failed to delete from cache", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to delete from cache: %w", err)
	}
	return nil
}

// Exists checks if a key exists in cache
func (c *CacheService) Exists(key string) bool {
	ctx := context.Background()
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		c.logger.Error("Failed to check existence in cache", zap.String("key", key), zap.Error(err))
		return false
	}
	return count > 0
}

// Clear removes all keys from cache
func (c *CacheService) Clear() error {
	ctx := context.Background()
	if err := c.client.FlushDB(ctx).Err(); err != nil {
		c.logger.Error("Failed to clear cache", zap.Error(err))
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	return nil
}

// Keys returns all keys matching a pattern
func (c *CacheService) Keys(pattern string) ([]string, error) {
	ctx := context.Background()
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		c.logger.Error("Failed to get keys from cache", zap.String("pattern", pattern), zap.Error(err))
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}
	return keys, nil
}

// Expire sets TTL for a key
func (c *CacheService) Expire(key string, ttl int) error {
	ctx := context.Background()
	duration := time.Duration(ttl) * time.Second
	if err := c.client.Expire(ctx, key, duration).Err(); err != nil {
		c.logger.Error("Failed to set expiry for cache key", zap.String("key", key), zap.Int("ttl", ttl), zap.Error(err))
		return fmt.Errorf("failed to set expiry: %w", err)
	}
	return nil
}

// TTL returns the TTL of a key
func (c *CacheService) TTL(key string) (int, error) {
	ctx := context.Background()
	duration, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		c.logger.Error("Failed to get TTL from cache", zap.String("key", key), zap.Error(err))
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}
	return int(duration.Seconds()), nil
}

// Close closes the cache connection
func (c *CacheService) Close() error {
	if err := c.client.Close(); err != nil {
		c.logger.Error("Failed to close cache connection", zap.Error(err))
		return fmt.Errorf("failed to close cache: %w", err)
	}
	return nil
}
