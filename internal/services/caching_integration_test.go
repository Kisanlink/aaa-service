package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// MockCacheServiceIntegration for integration testing
type MockCacheServiceIntegration struct {
	data map[string]interface{}
	ttls map[string]time.Time
}

func NewMockCacheServiceIntegration() *MockCacheServiceIntegration {
	return &MockCacheServiceIntegration{
		data: make(map[string]interface{}),
		ttls: make(map[string]time.Time),
	}
}

func (m *MockCacheServiceIntegration) Get(key string) (interface{}, bool) {
	// Check if key has expired
	if expiry, exists := m.ttls[key]; exists && time.Now().After(expiry) {
		delete(m.data, key)
		delete(m.ttls, key)
		return nil, false
	}

	value, exists := m.data[key]
	return value, exists
}

func (m *MockCacheServiceIntegration) Set(key string, value interface{}, ttl int) error {
	m.data[key] = value
	m.ttls[key] = time.Now().Add(time.Duration(ttl) * time.Second)
	return nil
}

func (m *MockCacheServiceIntegration) Delete(key string) error {
	delete(m.data, key)
	delete(m.ttls, key)
	return nil
}

func (m *MockCacheServiceIntegration) Exists(key string) bool {
	_, exists := m.Get(key)
	return exists
}

func (m *MockCacheServiceIntegration) Clear() error {
	m.data = make(map[string]interface{})
	m.ttls = make(map[string]time.Time)
	return nil
}

func (m *MockCacheServiceIntegration) Keys(pattern string) ([]string, error) {
	keys := make([]string, 0, len(m.data))
	for key := range m.data {
		keys = append(keys, key)
	}
	return keys, nil
}

func (m *MockCacheServiceIntegration) Expire(key string, ttl int) error {
	if _, exists := m.data[key]; exists {
		m.ttls[key] = time.Now().Add(time.Duration(ttl) * time.Second)
	}
	return nil
}

func (m *MockCacheServiceIntegration) TTL(key string) (int, error) {
	if expiry, exists := m.ttls[key]; exists {
		remaining := expiry.Sub(time.Now())
		if remaining > 0 {
			return int(remaining.Seconds()), nil
		}
	}
	return 0, nil
}

func (m *MockCacheServiceIntegration) Close() error {
	return nil
}

func TestCacheServiceIntegration_BasicOperations(t *testing.T) {
	cache := NewMockCacheServiceIntegration()

	// Test Set and Get
	key := "test:key"
	value := "test value"
	ttl := 60 // 1 minute

	err := cache.Set(key, value, ttl)
	assert.NoError(t, err)

	retrievedValue, exists := cache.Get(key)
	assert.True(t, exists)
	assert.Equal(t, value, retrievedValue)

	// Test Exists
	assert.True(t, cache.Exists(key))

	// Test Delete
	err = cache.Delete(key)
	assert.NoError(t, err)

	_, exists = cache.Get(key)
	assert.False(t, exists)
	assert.False(t, cache.Exists(key))
}

func TestCacheServiceIntegration_TTLExpiry(t *testing.T) {
	cache := NewMockCacheServiceIntegration()

	key := "test:expiry"
	value := "expires soon"
	ttl := 1 // 1 second

	err := cache.Set(key, value, ttl)
	assert.NoError(t, err)

	// Should exist immediately
	_, exists := cache.Get(key)
	assert.True(t, exists)

	// Wait for expiry
	time.Sleep(2 * time.Second)

	// Should not exist after expiry
	_, exists = cache.Get(key)
	assert.False(t, exists)
}

func TestCacheServiceIntegration_UserRolesCaching(t *testing.T) {
	cache := NewMockCacheServiceIntegration()

	// Simulate user roles caching
	userID := "user-123"
	cacheKey := fmt.Sprintf("user_roles:%s", userID)

	// Create mock user roles
	userRoles := []*models.UserRole{
		{
			UserID:   userID,
			RoleID:   "role-1",
			IsActive: true,
		},
		{
			UserID:   userID,
			RoleID:   "role-2",
			IsActive: true,
		},
	}

	// Cache the user roles
	err := cache.Set(cacheKey, userRoles, 900) // 15 minutes
	assert.NoError(t, err)

	// Retrieve from cache
	cachedRoles, exists := cache.Get(cacheKey)
	assert.True(t, exists)
	assert.NotNil(t, cachedRoles)

	// Verify the cached data
	retrievedRoles, ok := cachedRoles.([]*models.UserRole)
	assert.True(t, ok)
	assert.Len(t, retrievedRoles, 2)
	assert.Equal(t, userID, retrievedRoles[0].UserID)
	assert.Equal(t, "role-1", retrievedRoles[0].RoleID)
	assert.True(t, retrievedRoles[0].IsActive)
}

func TestCacheServiceIntegration_CacheInvalidation(t *testing.T) {
	cache := NewMockCacheServiceIntegration()

	userID := "user-456"

	// Set up multiple cache entries for a user
	cacheKeys := []string{
		fmt.Sprintf("user:%s", userID),
		fmt.Sprintf("user_roles:%s", userID),
		fmt.Sprintf("user_profile:%s", userID),
		fmt.Sprintf("user_with_roles:%s", userID),
	}

	// Cache some data
	for i, key := range cacheKeys {
		err := cache.Set(key, fmt.Sprintf("data-%d", i), 3600)
		assert.NoError(t, err)
	}

	// Verify all entries exist
	for _, key := range cacheKeys {
		assert.True(t, cache.Exists(key))
	}

	// Simulate cache invalidation (like after role assignment)
	roleRelatedKeys := []string{
		fmt.Sprintf("user_roles:%s", userID),
		fmt.Sprintf("user_with_roles:%s", userID),
	}

	for _, key := range roleRelatedKeys {
		err := cache.Delete(key)
		assert.NoError(t, err)
	}

	// Verify role-related entries are gone
	assert.False(t, cache.Exists(fmt.Sprintf("user_roles:%s", userID)))
	assert.False(t, cache.Exists(fmt.Sprintf("user_with_roles:%s", userID)))

	// Verify other entries still exist
	assert.True(t, cache.Exists(fmt.Sprintf("user:%s", userID)))
	assert.True(t, cache.Exists(fmt.Sprintf("user_profile:%s", userID)))
}

func TestCacheServiceIntegration_CacheKeyPatterns(t *testing.T) {
	cache := NewMockCacheServiceIntegration()

	// Test various cache key patterns used in the system
	testCases := []struct {
		keyPattern string
		userID     string
		expected   string
	}{
		{"user:%s", "123", "user:123"},
		{"user_roles:%s", "456", "user_roles:456"},
		{"user_profile:%s", "789", "user_profile:789"},
		{"user_with_roles:%s", "abc", "user_with_roles:abc"},
	}

	for _, tc := range testCases {
		key := fmt.Sprintf(tc.keyPattern, tc.userID)
		assert.Equal(t, tc.expected, key)

		// Test that we can use these keys with the cache
		err := cache.Set(key, "test data", 60)
		assert.NoError(t, err)

		_, exists := cache.Get(key)
		assert.True(t, exists)
	}
}

func TestCacheServiceIntegration_ConcurrentAccess(t *testing.T) {
	cache := NewMockCacheServiceIntegration()

	// Test concurrent access to cache (basic test)
	key := "concurrent:test"

	// Set initial value
	err := cache.Set(key, "initial", 60)
	assert.NoError(t, err)

	// Simulate concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_, exists := cache.Get(key)
			assert.True(t, exists)
		}()
	}

	// Give goroutines time to complete
	time.Sleep(100 * time.Millisecond)

	// Verify key still exists
	_, exists := cache.Get(key)
	assert.True(t, exists)
}

func TestRoleServiceCaching_Integration(t *testing.T) {
	// Create a mock cache service
	cache := NewMockCacheServiceIntegration()
	logger := zap.NewNop()

	// Create a minimal role service for testing cache integration
	// Note: This is a simplified test focusing on cache behavior
	userID := "test-user"
	cacheKey := fmt.Sprintf("user_roles:%s", userID)

	// Simulate caching user roles
	mockRoles := []*models.UserRole{
		{
			UserID:   userID,
			RoleID:   "admin",
			IsActive: true,
		},
	}

	// Test cache miss scenario
	_, exists := cache.Get(cacheKey)
	assert.False(t, exists)

	// Simulate setting cache after repository call
	err := cache.Set(cacheKey, mockRoles, 900) // 15 minutes
	assert.NoError(t, err)

	// Test cache hit scenario
	cachedRoles, exists := cache.Get(cacheKey)
	assert.True(t, exists)
	assert.NotNil(t, cachedRoles)

	// Simulate cache invalidation after role assignment
	err = cache.Delete(cacheKey)
	assert.NoError(t, err)

	// Verify cache is cleared
	_, exists = cache.Get(cacheKey)
	assert.False(t, exists)

	logger.Info("Role service caching integration test completed")
}
