package user

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"go.uber.org/zap"
)

// OrganizationalContextProvider provides methods to fetch organizational context
type OrganizationalContextProvider interface {
	GetByID(ctx context.Context, id string) (*models.Group, error)
	GetByPrincipalID(ctx context.Context, principalID string, limit, offset int) ([]*models.GroupMembership, error)
}

// GetUserOrganizations retrieves all organizations the user belongs to through group memberships
// Returns organizations with their ID and name for JWT token context
func (s *Service) GetUserOrganizations(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	s.logger.Info("Getting user organizations", zap.String("user_id", userID))

	if userID == "" {
		return nil, errors.NewValidationError("user ID cannot be empty")
	}

	// Check if repositories are available
	if s.groupMembershipRepo == nil || s.groupRepo == nil || s.organizationRepo == nil {
		s.logger.Debug("Required repositories not injected for GetUserOrganizations, returning empty list", zap.String("user_id", userID))
		return []map[string]interface{}{}, nil
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user_organizations:%s", userID)
	if cachedOrgs, exists := s.cacheService.Get(cacheKey); exists {
		if orgs, ok := cachedOrgs.([]map[string]interface{}); ok {
			s.logger.Debug("User organizations found in cache", zap.String("user_id", userID))
			return orgs, nil
		}
	}

	// Type assert to get the actual repository types
	membershipRepo, ok := s.groupMembershipRepo.(interface {
		GetByPrincipalID(ctx context.Context, principalID string, limit, offset int) ([]*models.GroupMembership, error)
	})
	if !ok {
		s.logger.Warn("Group membership repository does not implement GetByPrincipalID", zap.String("user_id", userID))
		return []map[string]interface{}{}, nil
	}

	grpRepo, ok := s.groupRepo.(interface {
		GetByID(ctx context.Context, id string) (*models.Group, error)
	})
	if !ok {
		s.logger.Warn("Group repository does not implement GetByID", zap.String("user_id", userID))
		return []map[string]interface{}{}, nil
	}

	orgRepo, ok := s.organizationRepo.(interface {
		GetByID(ctx context.Context, id string) (*models.Organization, error)
	})
	if !ok {
		s.logger.Warn("Organization repository does not implement GetByID", zap.String("user_id", userID))
		return []map[string]interface{}{}, nil
	}

	// Get all active group memberships for the user
	memberships, err := membershipRepo.GetByPrincipalID(ctx, userID, 1000, 0)
	if err != nil {
		s.logger.Error("Failed to get group memberships", zap.String("user_id", userID), zap.Error(err))
		return []map[string]interface{}{}, nil // Return empty array on error instead of failing
	}

	// Collect unique organization IDs from active memberships
	orgIDMap := make(map[string]bool)
	now := time.Now()

	for _, membership := range memberships {
		// Only consider active memberships that are currently effective
		if membership.IsActive && membership.IsEffective(now) {
			// Get the group to find its organization
			group, err := grpRepo.GetByID(ctx, membership.GroupID)
			if err != nil {
				s.logger.Warn("Failed to get group for membership",
					zap.String("user_id", userID),
					zap.String("group_id", membership.GroupID),
					zap.Error(err))
				continue
			}

			// Only include active groups
			if group != nil && group.IsActive && group.DeletedAt == nil {
				orgIDMap[group.OrganizationID] = true
			}
		}
	}

	// Fetch organization details for each unique organization ID
	organizations := make([]map[string]interface{}, 0, len(orgIDMap))
	for orgID := range orgIDMap {
		org, err := orgRepo.GetByID(ctx, orgID)
		if err != nil {
			s.logger.Warn("Failed to get organization",
				zap.String("user_id", userID),
				zap.String("org_id", orgID),
				zap.Error(err))
			continue
		}

		// Only include active, non-deleted organizations
		if org.IsActive && org.DeletedAt == nil {
			organizations = append(organizations, map[string]interface{}{
				"id":   org.ID,
				"name": org.Name,
			})
		}
	}

	// Cache the result for 10 minutes
	if err := s.cacheService.Set(cacheKey, organizations, 600); err != nil {
		s.logger.Warn("Failed to cache user organizations", zap.Error(err))
	}

	s.logger.Info("User organizations retrieved successfully",
		zap.String("user_id", userID),
		zap.Int("count", len(organizations)))
	return organizations, nil
}

// GetUserGroups retrieves all groups the user is a member of
// Returns groups with their ID, name, and organization_id for JWT token context
func (s *Service) GetUserGroups(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	s.logger.Info("Getting user groups", zap.String("user_id", userID))

	if userID == "" {
		return nil, errors.NewValidationError("user ID cannot be empty")
	}

	// Check if repositories are available
	if s.groupMembershipRepo == nil || s.groupRepo == nil {
		s.logger.Debug("Required repositories not injected for GetUserGroups, returning empty list", zap.String("user_id", userID))
		return []map[string]interface{}{}, nil
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user_groups:%s", userID)
	if cachedGroups, exists := s.cacheService.Get(cacheKey); exists {
		if groups, ok := cachedGroups.([]map[string]interface{}); ok {
			s.logger.Debug("User groups found in cache", zap.String("user_id", userID))
			return groups, nil
		}
	}

	// Type assert to get the actual repository types
	membershipRepo, ok := s.groupMembershipRepo.(interface {
		GetByPrincipalID(ctx context.Context, principalID string, limit, offset int) ([]*models.GroupMembership, error)
	})
	if !ok {
		s.logger.Warn("Group membership repository does not implement GetByPrincipalID", zap.String("user_id", userID))
		return []map[string]interface{}{}, nil
	}

	grpRepo, ok := s.groupRepo.(interface {
		GetByID(ctx context.Context, id string) (*models.Group, error)
	})
	if !ok {
		s.logger.Warn("Group repository does not implement GetByID", zap.String("user_id", userID))
		return []map[string]interface{}{}, nil
	}

	// Get all active group memberships for the user
	memberships, err := membershipRepo.GetByPrincipalID(ctx, userID, 1000, 0)
	if err != nil {
		s.logger.Error("Failed to get group memberships", zap.String("user_id", userID), zap.Error(err))
		return []map[string]interface{}{}, nil // Return empty array on error instead of failing
	}

	// Collect groups from active memberships
	groups := make([]map[string]interface{}, 0)
	now := time.Now()

	for _, membership := range memberships {
		// Only consider active memberships that are currently effective
		if membership.IsActive && membership.IsEffective(now) {
			// Get the group details
			group, err := grpRepo.GetByID(ctx, membership.GroupID)
			if err != nil {
				s.logger.Warn("Failed to get group for membership",
					zap.String("user_id", userID),
					zap.String("group_id", membership.GroupID),
					zap.Error(err))
				continue
			}

			// Only include active, non-deleted groups
			if group != nil && group.IsActive && group.DeletedAt == nil {
				groups = append(groups, map[string]interface{}{
					"id":              group.ID,
					"name":            group.Name,
					"organization_id": group.OrganizationID,
				})
			}
		}
	}

	// Cache the result for 10 minutes
	if err := s.cacheService.Set(cacheKey, groups, 600); err != nil {
		s.logger.Warn("Failed to cache user groups", zap.Error(err))
	}

	s.logger.Info("User groups retrieved successfully",
		zap.String("user_id", userID),
		zap.Int("count", len(groups)))
	return groups, nil
}

// invalidateUserOrganizationalCache clears organizational context cache for a user
func (s *Service) invalidateUserOrganizationalCache(userID string) {
	cacheKeys := []string{
		fmt.Sprintf("user_organizations:%s", userID),
		fmt.Sprintf("user_groups:%s", userID),
	}

	for _, key := range cacheKeys {
		if err := s.cacheService.Delete(key); err != nil {
			s.logger.Warn("Failed to delete organizational cache key", zap.String("key", key), zap.Error(err))
		}
	}

	s.logger.Debug("User organizational cache invalidated", zap.String("user_id", userID))
}
