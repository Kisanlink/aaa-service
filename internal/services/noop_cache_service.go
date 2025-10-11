package services

import (
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
)

// NoOpCacheService is a no-operation cache service that does nothing
// Used when CACHE_DISABLED=true to avoid Redis connection errors
type NoOpCacheService struct {
	logger interfaces.Logger
}

// NewNoOpCacheService creates a new no-op cache service
func NewNoOpCacheService(logger interfaces.Logger) interfaces.CacheService {
	return &NoOpCacheService{
		logger: logger,
	}
}

// Get always returns cache miss
func (c *NoOpCacheService) Get(key string) (interface{}, bool) {
	return nil, false
}

// Set does nothing and returns success
func (c *NoOpCacheService) Set(key string, value interface{}, ttl int) error {
	return nil
}

// Delete does nothing and returns success
func (c *NoOpCacheService) Delete(key string) error {
	return nil
}

// Exists always returns false
func (c *NoOpCacheService) Exists(key string) bool {
	return false
}

// Clear does nothing and returns success
func (c *NoOpCacheService) Clear() error {
	return nil
}

// Keys returns empty slice
func (c *NoOpCacheService) Keys(pattern string) ([]string, error) {
	return []string{}, nil
}

// Expire does nothing and returns success
func (c *NoOpCacheService) Expire(key string, ttl int) error {
	return nil
}

// TTL returns 0 (no TTL)
func (c *NoOpCacheService) TTL(key string) (int, error) {
	return 0, nil
}

// Close does nothing and returns success
func (c *NoOpCacheService) Close() error {
	return nil
}
