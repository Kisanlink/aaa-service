package permissions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"go.uber.org/zap"
)

// EvaluatePermission is the core permission evaluation method
// It checks both permission models and applies role hierarchy
func (s *Service) EvaluatePermission(
	ctx context.Context,
	userID string,
	resourceType string,
	resourceID string,
	action string,
	evalCtx *EvaluationContext,
) (*EvaluationResult, error) {
	startTime := time.Now()

	if userID == "" {
		return &EvaluationResult{
			Allowed:        false,
			Reason:         "user ID is required",
			EvaluatedAt:    time.Now(),
			EvaluationTime: time.Since(startTime),
		}, nil
	}

	if resourceType == "" || resourceID == "" || action == "" {
		return &EvaluationResult{
			Allowed:        false,
			Reason:         "resource type, resource ID, and action are required",
			EvaluatedAt:    time.Now(),
			EvaluationTime: time.Since(startTime),
		}, nil
	}

	// 1. Check cache first
	cacheKey := s.buildEvaluationCacheKey(userID, resourceType, resourceID, action)
	if s.cache != nil {
		if cached, found := s.cache.Get(cacheKey); found {
			if result, ok := cached.(*EvaluationResult); ok {
				result.CacheHit = true
				result.EvaluationTime = time.Since(startTime)
				s.logger.Debug("Permission evaluation cache hit",
					zap.String("user_id", userID),
					zap.String("resource", resourceType),
					zap.String("action", action))
				return result, nil
			}
		}
	}

	// 2. Get user's effective roles (including inherited)
	effectiveRoles, err := s.getEffectiveRoles(ctx, userID, evalCtx)
	if err != nil {
		s.logger.Error("Failed to get effective roles",
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get effective roles: %w", err)
	}

	if len(effectiveRoles) == 0 {
		result := &EvaluationResult{
			Allowed:        false,
			Reason:         "user has no roles assigned",
			EffectiveRoles: effectiveRoles,
			CacheHit:       false,
			EvaluatedAt:    time.Now(),
			EvaluationTime: time.Since(startTime),
		}
		s.cacheEvaluationResult(ctx, cacheKey, result)
		return result, nil
	}

	// 3. Check Model 1: role_permissions → permissions
	hasPermModel1, reason1 := s.checkPermissionModel1(ctx, effectiveRoles, resourceType, resourceID, action)

	// 4. Check Model 2: resource_permissions (direct role → resource+action)
	hasPermModel2, reason2 := s.checkPermissionModel2(ctx, effectiveRoles, resourceType, resourceID, action)

	// 5. Apply context-based rules (time-based, IP-based, etc.)
	contextAllowed := true
	contextReason := ""
	if evalCtx != nil {
		contextAllowed, contextReason = s.applyContextRules(ctx, evalCtx)
	}

	// 6. Build final result
	allowed := (hasPermModel1 || hasPermModel2) && contextAllowed
	reason := s.buildEvaluationReason(hasPermModel1, hasPermModel2, contextAllowed, reason1, reason2, contextReason)

	result := &EvaluationResult{
		Allowed:        allowed,
		Reason:         reason,
		EffectiveRoles: effectiveRoles,
		CacheHit:       false,
		EvaluatedAt:    time.Now(),
		EvaluationTime: time.Since(startTime),
	}

	// 7. Cache result
	s.cacheEvaluationResult(ctx, cacheKey, result)

	// 8. Audit log
	if s.audit != nil {
		status := "denied"
		if allowed {
			status = "allowed"
		}
		s.audit.LogPermissionChange(ctx, userID, "evaluate", resourceType, resourceID, action,
			map[string]interface{}{
				"resource_type": resourceType,
				"resource_id":   resourceID,
				"action":        action,
				"result":        status,
				"reason":        reason,
				"duration_ms":   result.EvaluationTime.Milliseconds(),
			})
	}

	s.logger.Info("Permission evaluated",
		zap.String("user_id", userID),
		zap.String("resource", resourceType),
		zap.String("action", action),
		zap.Bool("allowed", allowed),
		zap.String("reason", reason),
		zap.Duration("duration", result.EvaluationTime))

	return result, nil
}

// checkPermissionModel1 checks permissions via role_permissions → permissions
func (s *Service) checkPermissionModel1(
	ctx context.Context,
	roles []*models.Role,
	resourceType string,
	resourceID string,
	action string,
) (bool, string) {
	for _, role := range roles {
		// Get all permissions for this role
		rolePerms, err := s.rolePermissionRepo.GetByRoleID(ctx, role.ID)
		if err != nil {
			continue
		}

		for _, rolePerm := range rolePerms {
			if rolePerm.Permission == nil && rolePerm.PermissionID != "" {
				// Lazy load permission
				perm, err := s.GetPermissionByID(ctx, rolePerm.PermissionID)
				if err != nil {
					continue
				}
				rolePerm.Permission = perm
			}

			if rolePerm.Permission == nil {
				continue
			}

			// Check if permission matches resource+action
			if s.permissionMatches(rolePerm.Permission, resourceType, resourceID, action) {
				return true, fmt.Sprintf("granted via role '%s' (Model 1: named permission)", role.Name)
			}
		}
	}

	return false, ""
}

// checkPermissionModel2 checks permissions via resource_permissions (direct assignment)
func (s *Service) checkPermissionModel2(
	ctx context.Context,
	roles []*models.Role,
	resourceType string,
	resourceID string,
	action string,
) (bool, string) {
	roleIDs := extractRoleIDs(roles)

	hasPermission, err := s.resourcePermissionRepo.CheckMultiplePermissions(ctx, roleIDs, resourceType, resourceID, action)
	if err == nil && hasPermission {
		return true, "granted via direct resource assignment (Model 2)"
	}

	return false, ""
}

// permissionMatches checks if a permission matches the requested resource and action
func (s *Service) permissionMatches(
	permission *models.Permission,
	resourceType string,
	resourceID string,
	action string,
) bool {
	// Permission must be active
	if !permission.IsActive {
		return false
	}

	// TODO: Implement proper resource and action matching
	// For now, we'll do basic validation
	// In a full implementation, this would:
	// 1. Load the Resource entity and check if its type matches resourceType
	// 2. Load the Action entity and check if its name matches action
	// 3. Handle wildcards and resource hierarchies

	return true // Placeholder
}

// getEffectiveRoles retrieves all effective roles for a user including inherited roles
func (s *Service) getEffectiveRoles(ctx context.Context, userID string, evalCtx *EvaluationContext) ([]*models.Role, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("user:%s:effective_roles", userID)
	if s.cache != nil {
		if cached, found := s.cache.Get(cacheKey); found {
			if roles, ok := cached.([]*models.Role); ok {
				return roles, nil
			}
		}
	}

	// Get direct roles from user_roles table
	// For now, we'll get roles directly. In full implementation, use UserRoleRepository
	// This is a placeholder - actual implementation would use proper user-role repository
	directRoles := []*models.Role{} // Placeholder

	// TODO: Use filter when proper user-role repository is integrated
	// filter := base.NewFilterBuilder().
	// 	Where("user_id", base.OpEqual, userID).
	// 	Where("is_active", base.OpEqual, true).
	// 	Build()

	// TODO: Get inherited roles from groups
	// if evalCtx != nil && evalCtx.GroupID != "" {
	//     groupRoles := s.getGroupRoles(ctx, evalCtx.GroupID)
	//     directRoles = append(directRoles, groupRoles...)
	// }

	// TODO: Get inherited roles from organization
	// if evalCtx != nil && evalCtx.OrganizationID != "" {
	//     orgRoles := s.getOrganizationRoles(ctx, evalCtx.OrganizationID)
	//     directRoles = append(directRoles, orgRoles...)
	// }

	// Get hierarchical roles (parent roles)
	allRoles := s.getHierarchicalRoles(ctx, directRoles)

	// Cache effective roles
	if s.cache != nil {
		if err := s.cache.Set(cacheKey, allRoles, 300); err != nil { // 5 minutes TTL
			s.logger.Warn("Failed to cache effective roles", zap.Error(err))
		}
	}

	return allRoles, nil
}

// getHierarchicalRoles traverses role hierarchy to get all parent roles
func (s *Service) getHierarchicalRoles(ctx context.Context, roles []*models.Role) []*models.Role {
	visited := make(map[string]bool)
	result := []*models.Role{}

	for _, role := range roles {
		s.addRoleWithParents(ctx, role, &result, visited)
	}

	return result
}

// addRoleWithParents recursively adds a role and its parent roles
func (s *Service) addRoleWithParents(ctx context.Context, role *models.Role, result *[]*models.Role, visited map[string]bool) {
	if role == nil || visited[role.ID] {
		return // Prevent cycles
	}

	visited[role.ID] = true
	*result = append(*result, role)

	// Add parent role if exists
	if role.ParentID != nil && *role.ParentID != "" {
		parent, err := s.roleRepo.GetByID(ctx, *role.ParentID, &models.Role{})
		if err == nil && parent != nil {
			s.addRoleWithParents(ctx, parent, result, visited)
		}
	}
}

// applyContextRules applies context-based permission rules
func (s *Service) applyContextRules(ctx context.Context, evalCtx *EvaluationContext) (bool, string) {
	// TODO: Implement context-based rules
	// - Time-based permissions (effective from/until)
	// - IP-based restrictions
	// - Custom attribute checks
	return true, ""
}

// buildEvaluationReason builds a human-readable reason string
func (s *Service) buildEvaluationReason(hasModel1, hasModel2, contextAllowed bool, reason1, reason2, contextReason string) string {
	var parts []string

	if !hasModel1 && !hasModel2 {
		parts = append(parts, "no matching permissions found")
	} else {
		if hasModel1 {
			parts = append(parts, reason1)
		}
		if hasModel2 {
			parts = append(parts, reason2)
		}
	}

	if !contextAllowed {
		parts = append(parts, fmt.Sprintf("context check failed: %s", contextReason))
	}

	return strings.Join(parts, "; ")
}

// Helper functions

func (s *Service) buildEvaluationCacheKey(userID, resourceType, resourceID, action string) string {
	return fmt.Sprintf("permission:%s:%s:%s:%s", userID, resourceType, resourceID, action)
}

func (s *Service) cacheEvaluationResult(ctx context.Context, cacheKey string, result *EvaluationResult) {
	if s.cache == nil {
		return
	}

	if err := s.cache.Set(cacheKey, result, 60); err != nil { // 1 minute TTL
		s.logger.Warn("Failed to cache evaluation result",
			zap.String("cache_key", cacheKey),
			zap.Error(err))
	}
}

func extractRoleIDs(roles []*models.Role) []string {
	roleIDs := make([]string, len(roles))
	for i, role := range roles {
		roleIDs[i] = role.ID
	}
	return roleIDs
}
