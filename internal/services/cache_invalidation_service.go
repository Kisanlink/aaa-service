package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/internal/services/groups"
	"github.com/Kisanlink/aaa-service/internal/services/organizations"
	"go.uber.org/zap"
)

// CacheInvalidationService handles coordinated cache invalidation across the system
type CacheInvalidationService struct {
	orgCache   *organizations.OrganizationCacheService
	groupCache *groups.GroupCacheService
	cache      interfaces.CacheService
	logger     *zap.Logger
	mu         sync.RWMutex
}

// InvalidationEvent represents a cache invalidation event
type InvalidationEvent struct {
	Type         string                 `json:"type"`
	ResourceID   string                 `json:"resource_id"`
	ResourceType string                 `json:"resource_type"`
	OrgID        string                 `json:"org_id,omitempty"`
	AffectedIDs  []string               `json:"affected_ids,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// InvalidationStrategy defines how to handle different types of cache invalidation
type InvalidationStrategy struct {
	EventType    string
	Handler      func(ctx context.Context, event InvalidationEvent) error
	Description  string
	AffectsUsers bool
}

// NewCacheInvalidationService creates a new cache invalidation service
func NewCacheInvalidationService(
	cache interfaces.CacheService,
	logger *zap.Logger,
) *CacheInvalidationService {
	orgCache := organizations.NewOrganizationCacheService(cache, logger)
	groupCache := groups.NewGroupCacheService(cache, logger)

	service := &CacheInvalidationService{
		orgCache:   orgCache,
		groupCache: groupCache,
		cache:      cache,
		logger:     logger,
	}

	return service
}

// GetInvalidationStrategies returns all available invalidation strategies
func (s *CacheInvalidationService) GetInvalidationStrategies() []InvalidationStrategy {
	return []InvalidationStrategy{
		{
			EventType:    "organization_created",
			Handler:      s.handleOrganizationCreated,
			Description:  "Invalidate parent organization caches when new organization is created",
			AffectsUsers: false,
		},
		{
			EventType:    "organization_updated",
			Handler:      s.handleOrganizationUpdated,
			Description:  "Invalidate organization and related hierarchy caches",
			AffectsUsers: false,
		},
		{
			EventType:    "organization_deleted",
			Handler:      s.handleOrganizationDeleted,
			Description:  "Invalidate organization and all related caches",
			AffectsUsers: true,
		},
		{
			EventType:    "organization_hierarchy_changed",
			Handler:      s.handleOrganizationHierarchyChanged,
			Description:  "Invalidate hierarchy caches for affected organizations",
			AffectsUsers: false,
		},
		{
			EventType:    "group_created",
			Handler:      s.handleGroupCreated,
			Description:  "Invalidate organization group caches when new group is created",
			AffectsUsers: false,
		},
		{
			EventType:    "group_updated",
			Handler:      s.handleGroupUpdated,
			Description:  "Invalidate group and related caches",
			AffectsUsers: false,
		},
		{
			EventType:    "group_deleted",
			Handler:      s.handleGroupDeleted,
			Description:  "Invalidate group and all related caches",
			AffectsUsers: true,
		},
		{
			EventType:    "group_hierarchy_changed",
			Handler:      s.handleGroupHierarchyChanged,
			Description:  "Invalidate hierarchy caches for affected groups",
			AffectsUsers: true,
		},
		{
			EventType:    "user_group_membership_changed",
			Handler:      s.handleUserGroupMembershipChanged,
			Description:  "Invalidate user group and effective roles caches",
			AffectsUsers: true,
		},
		{
			EventType:    "role_assigned_to_group",
			Handler:      s.handleRoleAssignedToGroup,
			Description:  "Invalidate group roles and user effective roles caches",
			AffectsUsers: true,
		},
		{
			EventType:    "role_removed_from_group",
			Handler:      s.handleRoleRemovedFromGroup,
			Description:  "Invalidate group roles and user effective roles caches",
			AffectsUsers: true,
		},
		{
			EventType:    "role_updated",
			Handler:      s.handleRoleUpdated,
			Description:  "Invalidate all caches containing the updated role",
			AffectsUsers: true,
		},
	}
}

// InvalidateCache processes a cache invalidation event
func (s *CacheInvalidationService) InvalidateCache(ctx context.Context, event InvalidationEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Processing cache invalidation event",
		zap.String("event_type", event.Type),
		zap.String("resource_id", event.ResourceID),
		zap.String("resource_type", event.ResourceType),
		zap.String("org_id", event.OrgID))

	strategies := s.GetInvalidationStrategies()
	for _, strategy := range strategies {
		if strategy.EventType == event.Type {
			if err := strategy.Handler(ctx, event); err != nil {
				s.logger.Error("Failed to execute invalidation strategy",
					zap.String("event_type", event.Type),
					zap.String("strategy", strategy.Description),
					zap.Error(err))
				return fmt.Errorf("invalidation strategy failed: %w", err)
			}

			s.logger.Debug("Cache invalidation strategy executed successfully",
				zap.String("event_type", event.Type),
				zap.String("strategy", strategy.Description),
				zap.Bool("affects_users", strategy.AffectsUsers))

			return nil
		}
	}

	s.logger.Warn("No invalidation strategy found for event type",
		zap.String("event_type", event.Type))

	return fmt.Errorf("no invalidation strategy found for event type: %s", event.Type)
}

// Organization-related invalidation handlers

func (s *CacheInvalidationService) handleOrganizationCreated(ctx context.Context, event InvalidationEvent) error {
	// When an organization is created, invalidate parent organization's children cache
	if parentID, exists := event.Metadata["parent_id"]; exists && parentID != nil {
		if parentIDStr, ok := parentID.(string); ok && parentIDStr != "" {
			return s.orgCache.InvalidateOrganizationCache(ctx, parentIDStr)
		}
	}
	return nil
}

func (s *CacheInvalidationService) handleOrganizationUpdated(ctx context.Context, event InvalidationEvent) error {
	// Invalidate the organization's own cache
	if err := s.orgCache.InvalidateOrganizationCache(ctx, event.ResourceID); err != nil {
		return err
	}

	// If hierarchy changed, invalidate related organizations
	if hierarchyChanged, exists := event.Metadata["hierarchy_changed"]; exists {
		if changed, ok := hierarchyChanged.(bool); ok && changed {
			return s.handleOrganizationHierarchyChanged(ctx, event)
		}
	}

	return nil
}

func (s *CacheInvalidationService) handleOrganizationDeleted(ctx context.Context, event InvalidationEvent) error {
	// Invalidate the organization's cache
	if err := s.orgCache.InvalidateOrganizationCache(ctx, event.ResourceID); err != nil {
		return err
	}

	// Invalidate parent and children caches if they exist
	if len(event.AffectedIDs) > 0 {
		for _, affectedID := range event.AffectedIDs {
			if err := s.orgCache.InvalidateOrganizationCache(ctx, affectedID); err != nil {
				s.logger.Warn("Failed to invalidate affected organization cache",
					zap.String("affected_org_id", affectedID),
					zap.Error(err))
			}
		}
	}

	return nil
}

func (s *CacheInvalidationService) handleOrganizationHierarchyChanged(ctx context.Context, event InvalidationEvent) error {
	// Invalidate hierarchy-related caches for the main organization and all affected ones
	affectedOrgIDs := append([]string{event.ResourceID}, event.AffectedIDs...)
	return s.orgCache.InvalidateHierarchyRelatedCache(ctx, event.ResourceID, affectedOrgIDs)
}

// Group-related invalidation handlers

func (s *CacheInvalidationService) handleGroupCreated(ctx context.Context, event InvalidationEvent) error {
	// Invalidate organization's group caches
	if event.OrgID != "" {
		return s.orgCache.InvalidateOrganizationCache(ctx, event.OrgID)
	}
	return nil
}

func (s *CacheInvalidationService) handleGroupUpdated(ctx context.Context, event InvalidationEvent) error {
	// Invalidate the group's cache
	if err := s.groupCache.InvalidateGroupCache(ctx, event.ResourceID); err != nil {
		return err
	}

	// Invalidate organization's group caches
	if event.OrgID != "" {
		return s.orgCache.InvalidateOrganizationCache(ctx, event.OrgID)
	}

	return nil
}

func (s *CacheInvalidationService) handleGroupDeleted(ctx context.Context, event InvalidationEvent) error {
	// Invalidate the group's cache
	if err := s.groupCache.InvalidateGroupCache(ctx, event.ResourceID); err != nil {
		return err
	}

	// Invalidate organization's group caches
	if event.OrgID != "" {
		if err := s.orgCache.InvalidateOrganizationCache(ctx, event.OrgID); err != nil {
			return err
		}
	}

	// Invalidate affected groups in hierarchy
	if len(event.AffectedIDs) > 0 {
		return s.groupCache.InvalidateHierarchyCache(ctx, event.ResourceID, event.AffectedIDs)
	}

	return nil
}

func (s *CacheInvalidationService) handleGroupHierarchyChanged(ctx context.Context, event InvalidationEvent) error {
	// Invalidate hierarchy caches for affected groups
	return s.groupCache.InvalidateHierarchyCache(ctx, event.ResourceID, event.AffectedIDs)
}

// User and role-related invalidation handlers

func (s *CacheInvalidationService) handleUserGroupMembershipChanged(ctx context.Context, event InvalidationEvent) error {
	userID, exists := event.Metadata["user_id"]
	if !exists {
		return fmt.Errorf("user_id not found in event metadata")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return fmt.Errorf("user_id is not a string")
	}

	// Invalidate user's group memberships and effective roles
	if err := s.groupCache.InvalidateUserEffectiveRolesCache(ctx, event.OrgID, userIDStr); err != nil {
		return err
	}

	// Invalidate group's member cache
	return s.groupCache.InvalidateGroupCache(ctx, event.ResourceID)
}

func (s *CacheInvalidationService) handleRoleAssignedToGroup(ctx context.Context, event InvalidationEvent) error {
	roleID, exists := event.Metadata["role_id"]
	if !exists {
		return fmt.Errorf("role_id not found in event metadata")
	}

	roleIDStr, ok := roleID.(string)
	if !ok {
		return fmt.Errorf("role_id is not a string")
	}

	// Invalidate role assignment caches
	return s.groupCache.InvalidateRoleAssignmentCache(ctx, event.OrgID, event.ResourceID, roleIDStr)
}

func (s *CacheInvalidationService) handleRoleRemovedFromGroup(ctx context.Context, event InvalidationEvent) error {
	roleID, exists := event.Metadata["role_id"]
	if !exists {
		return fmt.Errorf("role_id not found in event metadata")
	}

	roleIDStr, ok := roleID.(string)
	if !ok {
		return fmt.Errorf("role_id is not a string")
	}

	// Invalidate role assignment caches
	return s.groupCache.InvalidateRoleAssignmentCache(ctx, event.OrgID, event.ResourceID, roleIDStr)
}

func (s *CacheInvalidationService) handleRoleUpdated(ctx context.Context, event InvalidationEvent) error {
	// When a role is updated, we need to invalidate all caches that might contain this role
	// This is a broad invalidation but necessary for consistency

	// Invalidate all organization role-related caches
	if event.OrgID != "" {
		if err := s.orgCache.InvalidateRoleRelatedCache(ctx, event.OrgID); err != nil {
			return err
		}
	}

	// Invalidate all group role caches that might contain this role
	// This requires a pattern-based search
	rolePattern := fmt.Sprintf("*:role:%s:*", event.ResourceID)
	keys, err := s.cache.Keys(rolePattern)
	if err != nil {
		s.logger.Warn("Failed to get role-related cache keys",
			zap.String("role_id", event.ResourceID),
			zap.Error(err))
		return nil // Don't fail the operation
	}

	for _, key := range keys {
		if err := s.cache.Delete(key); err != nil {
			s.logger.Warn("Failed to delete role-related cache key",
				zap.String("key", key),
				zap.Error(err))
		}
	}

	return nil
}

// BatchInvalidateCache processes multiple invalidation events efficiently
func (s *CacheInvalidationService) BatchInvalidateCache(ctx context.Context, events []InvalidationEvent) error {
	s.logger.Info("Processing batch cache invalidation",
		zap.Int("event_count", len(events)))

	var errors []error
	for i, event := range events {
		if err := s.InvalidateCache(ctx, event); err != nil {
			s.logger.Error("Failed to process invalidation event in batch",
				zap.Int("event_index", i),
				zap.String("event_type", event.Type),
				zap.Error(err))
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("batch invalidation completed with %d errors", len(errors))
	}

	s.logger.Info("Batch cache invalidation completed successfully",
		zap.Int("event_count", len(events)))

	return nil
}

// GetInvalidationStats returns statistics about cache invalidation operations
func (s *CacheInvalidationService) GetInvalidationStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get cache key counts by pattern
	orgKeys, err := s.cache.Keys("org:*")
	if err == nil {
		stats["organization_cache_keys"] = len(orgKeys)
	}

	groupKeys, err := s.cache.Keys("group:*")
	if err == nil {
		stats["group_cache_keys"] = len(groupKeys)
	}

	userKeys, err := s.cache.Keys("*:user:*")
	if err == nil {
		stats["user_cache_keys"] = len(userKeys)
	}

	roleKeys, err := s.cache.Keys("*:role*")
	if err == nil {
		stats["role_cache_keys"] = len(roleKeys)
	}

	effectiveRolesKeys, err := s.cache.Keys("*:effective_roles*")
	if err == nil {
		stats["effective_roles_cache_keys"] = len(effectiveRolesKeys)
	}

	// Add strategy information
	strategies := s.GetInvalidationStrategies()
	stats["available_strategies"] = len(strategies)

	userAffectingStrategies := 0
	for _, strategy := range strategies {
		if strategy.AffectsUsers {
			userAffectingStrategies++
		}
	}
	stats["user_affecting_strategies"] = userAffectingStrategies

	return stats, nil
}

// ValidateInvalidationEvent validates an invalidation event structure
func (s *CacheInvalidationService) ValidateInvalidationEvent(event InvalidationEvent) error {
	if event.Type == "" {
		return fmt.Errorf("event type is required")
	}

	if event.ResourceID == "" {
		return fmt.Errorf("resource ID is required")
	}

	if event.ResourceType == "" {
		return fmt.Errorf("resource type is required")
	}

	// Validate that the event type is supported
	strategies := s.GetInvalidationStrategies()
	for _, strategy := range strategies {
		if strategy.EventType == event.Type {
			return nil // Valid event type
		}
	}

	return fmt.Errorf("unsupported event type: %s", event.Type)
}

// CreateInvalidationEvent creates a properly structured invalidation event
func (s *CacheInvalidationService) CreateInvalidationEvent(
	eventType, resourceID, resourceType, orgID string,
	affectedIDs []string,
	metadata map[string]interface{},
) InvalidationEvent {
	return InvalidationEvent{
		Type:         eventType,
		ResourceID:   resourceID,
		ResourceType: resourceType,
		OrgID:        orgID,
		AffectedIDs:  affectedIDs,
		Metadata:     metadata,
	}
}
