package services

import (
	"context"
	"sync"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/services/groups"
	"github.com/Kisanlink/aaa-service/v2/internal/services/organizations"
	"go.uber.org/zap"
)

// CacheWarmingService handles cache warming for frequently accessed data
type CacheWarmingService struct {
	orgService   interfaces.OrganizationService
	groupService interfaces.GroupService
	orgCache     *organizations.OrganizationCacheService
	groupCache   *groups.GroupCacheService
	cache        interfaces.CacheService
	logger       *zap.Logger

	// Configuration
	warmingInterval    time.Duration
	maxConcurrentWarms int

	// State
	isRunning bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
	mu        sync.RWMutex
}

// CacheWarmingConfig holds configuration for cache warming
type CacheWarmingConfig struct {
	WarmingInterval    time.Duration `json:"warming_interval"`
	MaxConcurrentWarms int           `json:"max_concurrent_warms"`

	// Frequently accessed entities
	FrequentOrganizations []string `json:"frequent_organizations"`
	FrequentGroups        []string `json:"frequent_groups"`
	FrequentUsers         []string `json:"frequent_users"`

	// Warming strategies
	WarmHierarchies    bool `json:"warm_hierarchies"`
	WarmGroupRoles     bool `json:"warm_group_roles"`
	WarmEffectiveRoles bool `json:"warm_effective_roles"`
	WarmStats          bool `json:"warm_stats"`
}

// NewCacheWarmingService creates a new cache warming service
func NewCacheWarmingService(
	orgService interfaces.OrganizationService,
	groupService interfaces.GroupService,
	cache interfaces.CacheService,
	logger *zap.Logger,
) *CacheWarmingService {
	orgCache := organizations.NewOrganizationCacheService(cache, logger)
	groupCache := groups.NewGroupCacheService(cache, logger)

	return &CacheWarmingService{
		orgService:         orgService,
		groupService:       groupService,
		orgCache:           orgCache,
		groupCache:         groupCache,
		cache:              cache,
		logger:             logger,
		warmingInterval:    30 * time.Minute, // Default 30 minutes
		maxConcurrentWarms: 5,                // Default 5 concurrent operations
		stopChan:           make(chan struct{}),
	}
}

// Start begins the cache warming process
func (s *CacheWarmingService) Start(ctx context.Context, config CacheWarmingConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		s.logger.Warn("Cache warming service is already running")
		return nil
	}

	s.warmingInterval = config.WarmingInterval
	s.maxConcurrentWarms = config.MaxConcurrentWarms
	s.isRunning = true

	s.logger.Info("Starting cache warming service",
		zap.Duration("interval", s.warmingInterval),
		zap.Int("max_concurrent", s.maxConcurrentWarms),
		zap.Int("frequent_orgs", len(config.FrequentOrganizations)),
		zap.Int("frequent_groups", len(config.FrequentGroups)),
		zap.Int("frequent_users", len(config.FrequentUsers)))

	// Start the warming goroutine
	s.wg.Add(1)
	go s.warmingLoop(ctx, config)

	return nil
}

// Stop stops the cache warming process
func (s *CacheWarmingService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		s.logger.Warn("Cache warming service is not running")
		return nil
	}

	s.logger.Info("Stopping cache warming service")
	s.isRunning = false
	close(s.stopChan)
	s.wg.Wait()

	s.logger.Info("Cache warming service stopped")
	return nil
}

// IsRunning returns whether the cache warming service is currently running
func (s *CacheWarmingService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// WarmNow performs immediate cache warming for specified entities
func (s *CacheWarmingService) WarmNow(ctx context.Context, config CacheWarmingConfig) error {
	s.logger.Info("Performing immediate cache warming")

	// Create a semaphore to limit concurrent operations
	semaphore := make(chan struct{}, s.maxConcurrentWarms)
	var wg sync.WaitGroup

	// Warm organizations
	if config.WarmHierarchies || config.WarmStats {
		for _, orgID := range config.FrequentOrganizations {
			wg.Add(1)
			go func(orgID string) {
				defer wg.Done()
				semaphore <- struct{}{}        // Acquire
				defer func() { <-semaphore }() // Release

				s.warmOrganization(ctx, orgID, config)
			}(orgID)
		}
	}

	// Warm groups
	if config.WarmGroupRoles {
		for _, groupID := range config.FrequentGroups {
			wg.Add(1)
			go func(groupID string) {
				defer wg.Done()
				semaphore <- struct{}{}        // Acquire
				defer func() { <-semaphore }() // Release

				s.warmGroup(ctx, groupID, config)
			}(groupID)
		}
	}

	// Warm user effective roles
	if config.WarmEffectiveRoles {
		for _, orgID := range config.FrequentOrganizations {
			for _, userID := range config.FrequentUsers {
				wg.Add(1)
				go func(orgID, userID string) {
					defer wg.Done()
					semaphore <- struct{}{}        // Acquire
					defer func() { <-semaphore }() // Release

					s.warmUserEffectiveRoles(ctx, orgID, userID)
				}(orgID, userID)
			}
		}
	}

	wg.Wait()
	s.logger.Info("Immediate cache warming completed")
	return nil
}

// warmingLoop runs the periodic cache warming
func (s *CacheWarmingService) warmingLoop(ctx context.Context, config CacheWarmingConfig) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.warmingInterval)
	defer ticker.Stop()

	// Perform initial warming
	s.WarmNow(ctx, config)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Cache warming loop stopped due to context cancellation")
			return
		case <-s.stopChan:
			s.logger.Info("Cache warming loop stopped")
			return
		case <-ticker.C:
			s.logger.Debug("Performing scheduled cache warming")
			if err := s.WarmNow(ctx, config); err != nil {
				s.logger.Error("Failed to perform scheduled cache warming", zap.Error(err))
			}
		}
	}
}

// warmOrganization warms cache for a specific organization
func (s *CacheWarmingService) warmOrganization(ctx context.Context, orgID string, config CacheWarmingConfig) {
	s.logger.Debug("Warming organization cache", zap.String("org_id", orgID))

	startTime := time.Now()
	defer func() {
		s.logger.Debug("Organization cache warming completed",
			zap.String("org_id", orgID),
			zap.Duration("duration", time.Since(startTime)))
	}()

	// Warm hierarchy if enabled
	if config.WarmHierarchies {
		if _, err := s.orgService.GetOrganizationHierarchy(ctx, orgID); err == nil {
			// The service method should automatically cache the result
			s.logger.Debug("Warmed organization hierarchy", zap.String("org_id", orgID))
		} else {
			s.logger.Warn("Failed to warm organization hierarchy",
				zap.String("org_id", orgID),
				zap.Error(err))
		}
	}

	// Warm stats if enabled
	if config.WarmStats {
		if _, err := s.orgService.GetOrganizationStats(ctx, orgID); err == nil {
			// The service method should automatically cache the result
			s.logger.Debug("Warmed organization stats", zap.String("org_id", orgID))
		} else {
			s.logger.Warn("Failed to warm organization stats",
				zap.String("org_id", orgID),
				zap.Error(err))
		}
	}

	// Warm organization groups
	if groups, err := s.orgService.GetOrganizationGroups(ctx, orgID, 100, 0, false); err == nil {
		s.orgCache.CacheOrganizationGroups(ctx, orgID, groups, false)
		s.logger.Debug("Warmed organization groups", zap.String("org_id", orgID))
	} else {
		s.logger.Warn("Failed to warm organization groups",
			zap.String("org_id", orgID),
			zap.Error(err))
	}

	// Warm active organization groups
	if activeGroups, err := s.orgService.GetOrganizationGroups(ctx, orgID, 100, 0, true); err == nil {
		s.orgCache.CacheOrganizationGroups(ctx, orgID, activeGroups, true)
		s.logger.Debug("Warmed active organization groups", zap.String("org_id", orgID))
	} else {
		s.logger.Warn("Failed to warm active organization groups",
			zap.String("org_id", orgID),
			zap.Error(err))
	}
}

// warmGroup warms cache for a specific group
func (s *CacheWarmingService) warmGroup(ctx context.Context, groupID string, config CacheWarmingConfig) {
	s.logger.Debug("Warming group cache", zap.String("group_id", groupID))

	startTime := time.Now()
	defer func() {
		s.logger.Debug("Group cache warming completed",
			zap.String("group_id", groupID),
			zap.Duration("duration", time.Since(startTime)))
	}()

	// Warm group roles if enabled
	if config.WarmGroupRoles {
		if _, err := s.groupService.GetGroupRoles(ctx, groupID); err == nil {
			// The service method should automatically cache the result
			s.logger.Debug("Warmed group roles", zap.String("group_id", groupID))
		} else {
			s.logger.Warn("Failed to warm group roles",
				zap.String("group_id", groupID),
				zap.Error(err))
		}
	}

	// Warm group members
	if _, err := s.groupService.GetGroupMembers(ctx, groupID, 100, 0); err == nil {
		// The service method should automatically cache the result
		s.logger.Debug("Warmed group members", zap.String("group_id", groupID))
	} else {
		s.logger.Warn("Failed to warm group members",
			zap.String("group_id", groupID),
			zap.Error(err))
	}
}

// warmUserEffectiveRoles warms cache for user effective roles
func (s *CacheWarmingService) warmUserEffectiveRoles(ctx context.Context, orgID, userID string) {
	s.logger.Debug("Warming user effective roles cache",
		zap.String("org_id", orgID),
		zap.String("user_id", userID))

	startTime := time.Now()
	defer func() {
		s.logger.Debug("User effective roles cache warming completed",
			zap.String("org_id", orgID),
			zap.String("user_id", userID),
			zap.Duration("duration", time.Since(startTime)))
	}()

	if _, err := s.groupService.GetUserEffectiveRoles(ctx, orgID, userID); err == nil {
		// The service method should automatically cache the result
		s.logger.Debug("Warmed user effective roles",
			zap.String("org_id", orgID),
			zap.String("user_id", userID))
	} else {
		s.logger.Warn("Failed to warm user effective roles",
			zap.String("org_id", orgID),
			zap.String("user_id", userID),
			zap.Error(err))
	}
}

// GetCacheStats returns statistics about cache warming operations
func (s *CacheWarmingService) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get cache service stats if available
	if keys, err := s.cache.Keys("org:*"); err == nil {
		stats["organization_cache_keys"] = len(keys)
	}

	if keys, err := s.cache.Keys("group:*"); err == nil {
		stats["group_cache_keys"] = len(keys)
	}

	if keys, err := s.cache.Keys("*:effective_roles*"); err == nil {
		stats["effective_roles_cache_keys"] = len(keys)
	}

	stats["is_running"] = s.IsRunning()
	stats["warming_interval"] = s.warmingInterval.String()
	stats["max_concurrent_warms"] = s.maxConcurrentWarms

	return stats, nil
}

// InvalidateFrequentlyAccessedCache invalidates cache for frequently accessed entities
func (s *CacheWarmingService) InvalidateFrequentlyAccessedCache(ctx context.Context, config CacheWarmingConfig) error {
	s.logger.Info("Invalidating frequently accessed cache")

	// Invalidate organization caches
	for _, orgID := range config.FrequentOrganizations {
		if err := s.orgCache.InvalidateOrganizationCache(ctx, orgID); err != nil {
			s.logger.Warn("Failed to invalidate organization cache",
				zap.String("org_id", orgID),
				zap.Error(err))
		}
	}

	// Invalidate group caches
	for _, groupID := range config.FrequentGroups {
		if err := s.groupCache.InvalidateGroupCache(ctx, groupID); err != nil {
			s.logger.Warn("Failed to invalidate group cache",
				zap.String("group_id", groupID),
				zap.Error(err))
		}
	}

	// Invalidate user effective roles caches
	for _, orgID := range config.FrequentOrganizations {
		for _, userID := range config.FrequentUsers {
			if err := s.groupCache.InvalidateUserEffectiveRolesCache(ctx, orgID, userID); err != nil {
				s.logger.Warn("Failed to invalidate user effective roles cache",
					zap.String("org_id", orgID),
					zap.String("user_id", userID),
					zap.Error(err))
			}
		}
	}

	s.logger.Info("Frequently accessed cache invalidation completed")
	return nil
}

// UpdateWarmingConfig updates the cache warming configuration
func (s *CacheWarmingService) UpdateWarmingConfig(config CacheWarmingConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.warmingInterval = config.WarmingInterval
	s.maxConcurrentWarms = config.MaxConcurrentWarms

	s.logger.Info("Cache warming configuration updated",
		zap.Duration("new_interval", s.warmingInterval),
		zap.Int("new_max_concurrent", s.maxConcurrentWarms))

	return nil
}
