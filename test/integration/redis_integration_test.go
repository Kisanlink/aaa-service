//go:build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestRedis_Connection(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6380"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}
}

func TestRedis_SetGet(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6380"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test Set
	key := "integration_test_key"
	value := "integration_test_value"
	if err := rdb.Set(ctx, key, value, time.Minute).Err(); err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	// Test Get
	result, err := rdb.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}

	if result != value {
		t.Errorf("Expected value '%s', got '%s'", value, result)
	}

	// Cleanup
	rdb.Del(ctx, key)
}
