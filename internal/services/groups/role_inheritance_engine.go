package groups

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/groups"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/roles"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"go.uber.org/zap"
)

// GroupRepositoryInterface defines the interface for group repository operations
type GroupRepositoryInterface interface {
	GetChildren(ctx context.Context, parentID string) ([]*models.Group, error)
}

// GroupRoleRepositoryInterface defines the interface for group role repository operations
type GroupRoleRepositoryInterface interface {
	GetByGroupID(ctx context.Context, groupID string) ([]*models.GroupRole, error)
}

// RoleRepositoryInterface defines the interface for role repository operations
type RoleRepositoryInterface interface {
	GetByID(ctx context.Context, id string, role *models.Role) (*models.Role, error)
}

// GroupMembershipRepositoryInterface defines the interface for group membership repository operations
type GroupMembershipRepositoryInterface interface {
	GetUserDirectGroups(ctx context.Context, orgID, userID string) ([]*models.Group, error)
}

// RoleInheritanceEngine handles role inheritance calculations for hierarchical groups.
//
// INHERITANCE MODEL: Bottom-Up (Upward) Inheritance Only
//
// This engine implements ONLY bottom-up inheritance where parent groups inherit roles
// from their child groups. This allows executives and managers to perform any action
// their subordinates can perform, enabling proper oversight and operational flexibility.
//
// Inheritance Flow:
//
//	CEO Group (Parent)
//	├── Manager Group (Child)
//	│   └── Employee Group (Grandchild)
//	└── Director Group (Child)
//
// If a user is in CEO Group, they inherit:
//   - CEO Group direct roles (distance 0, highest precedence)
//   - Manager Group roles (distance 1, inherited from child)
//   - Employee Group roles (distance 2, inherited from grandchild)
//   - Director Group roles (distance 1, inherited from child)
//
// Key Principles:
//  1. UPWARD ONLY: Roles flow UP the hierarchy (child → parent)
//  2. DISTANCE-BASED PRECEDENCE: Shorter distance = higher precedence
//  3. DIRECT WINS: Direct assignments (distance 0) always beat inherited roles
//  4. COMPREHENSIVE: All descendant roles are inherited, not just immediate children
//
// This model reflects real organizational hierarchies where executives need access
// to all systems and permissions their teams use for oversight and operational support.
type RoleInheritanceEngine struct {
	groupRepo           GroupRepositoryInterface
	groupRoleRepo       GroupRoleRepositoryInterface
	roleRepo            RoleRepositoryInterface
	groupMembershipRepo GroupMembershipRepositoryInterface
	cache               interfaces.CacheService
	logger              *zap.Logger
}

// NewRoleInheritanceEngine creates a new role inheritance engine
func NewRoleInheritanceEngine(
	groupRepo GroupRepositoryInterface,
	groupRoleRepo GroupRoleRepositoryInterface,
	roleRepo RoleRepositoryInterface,
	groupMembershipRepo GroupMembershipRepositoryInterface,
	cache interfaces.CacheService,
	logger *zap.Logger,
) *RoleInheritanceEngine {
	return &RoleInheritanceEngine{
		groupRepo:           groupRepo,
		groupRoleRepo:       groupRoleRepo,
		roleRepo:            roleRepo,
		groupMembershipRepo: groupMembershipRepo,
		cache:               cache,
		logger:              logger,
	}
}

// NewRoleInheritanceEngineWithRepos creates a new role inheritance engine with concrete repository types
func NewRoleInheritanceEngineWithRepos(
	groupRepo *groups.GroupRepository,
	groupRoleRepo *groups.GroupRoleRepository,
	roleRepo *roles.RoleRepository,
	groupMembershipRepo *groups.GroupMembershipRepository,
	cache interfaces.CacheService,
	logger *zap.Logger,
) *RoleInheritanceEngine {
	return &RoleInheritanceEngine{
		groupRepo:           groupRepo,
		groupRoleRepo:       groupRoleRepo,
		roleRepo:            roleRepo,
		groupMembershipRepo: groupMembershipRepo,
		cache:               cache,
		logger:              logger,
	}
}

// EffectiveRole represents a role with its inheritance path and precedence
type EffectiveRole struct {
	Role            *models.Role `json:"role"`
	GroupID         string       `json:"group_id"`
	GroupName       string       `json:"group_name"`
	InheritancePath []string     `json:"inheritance_path"` // Path from user's direct group to role source
	Distance        int          `json:"distance"`         // Distance from user's direct group (0 = direct, 1 = child, 2 = grandchild, etc.)
	IsDirectRole    bool         `json:"is_direct_role"`   // True if role is directly assigned to user's group
}

// CalculateEffectiveRoles calculates all effective roles for a user in an organization
// using BOTTOM-UP (UPWARD) INHERITANCE ONLY.
//
// Algorithm Steps:
//  1. Get user's direct group memberships (starting points)
//  2. For each direct group, recursively traverse ALL descendant groups
//  3. Collect roles from each group in the hierarchy with distance tracking
//  4. Apply conflict resolution (shortest distance wins)
//  5. Sort by precedence (distance ascending, then role name)
//
// Inheritance Direction: CHILD → PARENT (Bottom-Up Only)
//   - Parent groups inherit roles from ALL their descendant groups
//   - Child groups do NOT inherit roles from parent groups
//   - This enables executive oversight while maintaining security boundaries
//
// Conflict Resolution:
//   - Same role at multiple levels: keep the one with shortest distance
//   - Direct assignments (distance 0) always win over inherited roles
//   - Ties broken by role name for consistency
//
// Performance Optimizations:
//   - Aggressive caching (5-minute TTL)
//   - Early termination for inactive groups
//   - In-memory conflict resolution
func (r *RoleInheritanceEngine) CalculateEffectiveRoles(ctx context.Context, orgID, userID string) ([]*EffectiveRole, error) {
	r.logger.Info("Calculating effective roles for user",
		zap.String("org_id", orgID),
		zap.String("user_id", userID))

	// Check cache first
	cacheKey := fmt.Sprintf("org:%s:user:%s:effective_roles", orgID, userID)
	if cached, found := r.cache.Get(cacheKey); found {
		if effectiveRoles, ok := cached.([]*EffectiveRole); ok {
			r.logger.Debug("Returning cached effective roles", zap.String("user_id", userID))
			return effectiveRoles, nil
		}
	}

	// 1. Get user's direct group memberships
	directGroups, err := r.getUserDirectGroups(ctx, orgID, userID)
	if err != nil {
		r.logger.Error("Failed to get user's direct groups", zap.Error(err))
		return nil, err
	}

	if len(directGroups) == 0 {
		r.logger.Info("User has no group memberships", zap.String("user_id", userID))
		emptyRoles := []*EffectiveRole{}

		// Cache the empty result
		err = r.cache.Set(cacheKey, emptyRoles, 300) // Cache for 5 minutes
		if err != nil {
			r.logger.Warn("Failed to cache empty effective roles", zap.Error(err))
		}

		return emptyRoles, nil
	}

	// 2. For each direct group, collect roles through BOTTOM-UP inheritance
	// This implements UPWARD inheritance where parent groups inherit from children
	allEffectiveRoles := make(map[string]*EffectiveRole) // roleID -> EffectiveRole

	for _, directGroup := range directGroups {
		r.logger.Debug("Processing direct group for bottom-up inheritance",
			zap.String("group_id", directGroup.ID),
			zap.String("group_name", directGroup.Name),
			zap.String("user_id", userID))

		groupRoles, err := r.calculateBottomUpRoles(ctx, directGroup, 0)
		if err != nil {
			r.logger.Error("Failed to calculate roles for group hierarchy",
				zap.String("group_id", directGroup.ID),
				zap.Error(err))
			continue
		}

		r.logger.Debug("Collected roles from group hierarchy",
			zap.String("group_id", directGroup.ID),
			zap.Int("role_count", len(groupRoles)))

		// Merge roles with conflict resolution (most specific wins)
		// This ensures that when the same role exists at multiple levels,
		// the one with the shortest distance (most specific) is kept
		for roleID, effectiveRole := range groupRoles {
			if existing, exists := allEffectiveRoles[roleID]; exists {
				// Keep the role with the shortest distance (most specific)
				if effectiveRole.Distance < existing.Distance {
					r.logger.Debug("Role conflict resolved - keeping more specific role",
						zap.String("role_id", roleID),
						zap.Int("new_distance", effectiveRole.Distance),
						zap.Int("old_distance", existing.Distance),
						zap.String("winning_group", effectiveRole.GroupID))
					allEffectiveRoles[roleID] = effectiveRole
				} else {
					r.logger.Debug("Role conflict resolved - keeping existing role",
						zap.String("role_id", roleID),
						zap.Int("existing_distance", existing.Distance),
						zap.Int("new_distance", effectiveRole.Distance),
						zap.String("winning_group", existing.GroupID))
				}
			} else {
				allEffectiveRoles[roleID] = effectiveRole
			}
		}
	}

	// 3. Convert map to slice and sort by precedence
	effectiveRoles := make([]*EffectiveRole, 0, len(allEffectiveRoles))
	for _, role := range allEffectiveRoles {
		effectiveRoles = append(effectiveRoles, role)
	}

	// Sort by distance (most specific first), then by role name for consistency
	r.sortEffectiveRolesByPrecedence(effectiveRoles)

	// Cache the result
	err = r.cache.Set(cacheKey, effectiveRoles, 300) // Cache for 5 minutes
	if err != nil {
		r.logger.Warn("Failed to cache effective roles", zap.Error(err))
	}

	r.logger.Info("Calculated effective roles for user",
		zap.String("user_id", userID),
		zap.Int("role_count", len(effectiveRoles)))

	return effectiveRoles, nil
}

// calculateBottomUpRoles calculates roles for a group and all its descendant groups
// using bottom-up (upward) inheritance where parent groups inherit roles from their child groups.
//
// Inheritance Flow:
// 1. Start with a user's direct group membership
// 2. Collect roles directly assigned to that group (distance = 0)
// 3. Recursively traverse child groups and collect their roles
// 4. Child group roles are inherited by parent with increased distance
// 5. Most specific roles (shortest distance) take precedence in conflicts
//
// Example hierarchy: CEO Group -> Manager Group -> Employee Group
// If user is in CEO Group:
// - CEO Group direct roles: distance 0 (highest precedence)
// - Manager Group roles: distance 1 (inherited from child)
// - Employee Group roles: distance 2 (inherited from grandchild)
//
// This allows executives to inherit permissions from their subordinates,
// enabling them to perform any action their team members can perform.
func (r *RoleInheritanceEngine) calculateBottomUpRoles(ctx context.Context, group *models.Group, currentDistance int) (map[string]*EffectiveRole, error) {
	roles := make(map[string]*EffectiveRole)

	// 1. Get direct roles assigned to this group
	directRoles, err := r.getDirectGroupRoles(ctx, group.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get direct roles for group %s: %w", group.ID, err)
	}

	// Add direct roles
	for _, groupRole := range directRoles {
		role := &models.Role{}
		role, err = r.roleRepo.GetByID(ctx, groupRole.RoleID, role)
		if err != nil {
			r.logger.Warn("Failed to load role details",
				zap.String("role_id", groupRole.RoleID),
				zap.Error(err))
			continue
		}

		if role != nil && role.IsActive {
			effectiveRole := &EffectiveRole{
				Role:            role,
				GroupID:         group.ID,
				GroupName:       group.Name,
				InheritancePath: []string{group.ID},
				Distance:        currentDistance,
				IsDirectRole:    currentDistance == 0,
			}
			roles[role.ID] = effectiveRole
		}
	}

	// 2. Get child groups and inherit their roles (bottom-up inheritance)
	// This implements upward inheritance where parent groups inherit roles from children
	childGroups, err := r.groupRepo.GetChildren(ctx, group.ID)
	if err != nil {
		r.logger.Warn("Failed to get child groups",
			zap.String("group_id", group.ID),
			zap.Error(err))
		return roles, nil // Continue with direct roles even if children fail
	}

	// Recursively collect roles from child groups (bottom-up traversal)
	for _, childGroup := range childGroups {
		if !childGroup.IsActive {
			r.logger.Debug("Skipping inactive child group",
				zap.String("child_group_id", childGroup.ID),
				zap.String("parent_group_id", group.ID))
			continue
		}

		childRoles, err := r.calculateBottomUpRoles(ctx, childGroup, currentDistance+1)
		if err != nil {
			r.logger.Warn("Failed to calculate roles for child group",
				zap.String("child_group_id", childGroup.ID),
				zap.Error(err))
			continue
		}

		// Merge child roles with conflict resolution (most specific wins)
		for roleID, childRole := range childRoles {
			if existing, exists := roles[roleID]; exists {
				// Keep the role with the shortest distance (most specific)
				// Direct assignments (distance 0) always win over inherited roles
				if childRole.Distance < existing.Distance {
					// Update inheritance path to include current group at the front
					childRole.InheritancePath = append([]string{group.ID}, childRole.InheritancePath...)
					roles[roleID] = childRole
					r.logger.Debug("Role conflict resolved - keeping more specific role",
						zap.String("role_id", roleID),
						zap.Int("winning_distance", childRole.Distance),
						zap.Int("losing_distance", existing.Distance))
				}
			} else {
				// Update inheritance path to include current group at the front
				childRole.InheritancePath = append([]string{group.ID}, childRole.InheritancePath...)
				roles[roleID] = childRole
				r.logger.Debug("Inherited role from child group",
					zap.String("role_id", roleID),
					zap.String("child_group_id", childGroup.ID),
					zap.Int("distance", childRole.Distance))
			}
		}
	}

	return roles, nil
}

// getUserDirectGroups gets all groups that a user is directly a member of
func (r *RoleInheritanceEngine) getUserDirectGroups(ctx context.Context, orgID, userID string) ([]*models.Group, error) {
	r.logger.Debug("Getting user's direct groups",
		zap.String("org_id", orgID),
		zap.String("user_id", userID))

	// Check cache first
	cacheKey := fmt.Sprintf("org:%s:user:%s:groups", orgID, userID)
	if cached, found := r.cache.Get(cacheKey); found {
		if groups, ok := cached.([]*models.Group); ok {
			r.logger.Debug("Returning cached user groups",
				zap.String("user_id", userID),
				zap.Int("group_count", len(groups)))
			return groups, nil
		}
	}

	// Get user's direct group memberships using the repository
	groups, err := r.groupMembershipRepo.GetUserDirectGroups(ctx, orgID, userID)
	if err != nil {
		r.logger.Error("Failed to get user's direct groups from repository",
			zap.String("org_id", orgID),
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user's direct groups: %w", err)
	}

	r.logger.Debug("Retrieved user's direct groups",
		zap.String("org_id", orgID),
		zap.String("user_id", userID),
		zap.Int("group_count", len(groups)))

	// Cache the result for 5 minutes
	err = r.cache.Set(cacheKey, groups, 300)
	if err != nil {
		r.logger.Warn("Failed to cache user groups", zap.Error(err))
	}

	return groups, nil
}

// getDirectGroupRoles gets roles directly assigned to a group (not inherited)
func (r *RoleInheritanceEngine) getDirectGroupRoles(ctx context.Context, groupID string) ([]*models.GroupRole, error) {
	groupRoles, err := r.groupRoleRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Filter for currently effective roles
	now := time.Now()
	effectiveRoles := make([]*models.GroupRole, 0, len(groupRoles))
	for _, groupRole := range groupRoles {
		if groupRole.IsEffective(now) {
			effectiveRoles = append(effectiveRoles, groupRole)
		}
	}

	return effectiveRoles, nil
}

// sortEffectiveRolesByPrecedence sorts effective roles by precedence
// Most specific roles (shortest distance) come first
func (r *RoleInheritanceEngine) sortEffectiveRolesByPrecedence(roles []*EffectiveRole) {
	// Sort by distance (ascending), then by role name for consistency
	for i := 0; i < len(roles)-1; i++ {
		for j := i + 1; j < len(roles); j++ {
			// Primary sort: distance (most specific first)
			if roles[i].Distance > roles[j].Distance {
				roles[i], roles[j] = roles[j], roles[i]
			} else if roles[i].Distance == roles[j].Distance {
				// Secondary sort: role name for consistency
				if roles[i].Role.Name > roles[j].Role.Name {
					roles[i], roles[j] = roles[j], roles[i]
				}
			}
		}
	}
}

// InvalidateUserRoleCache invalidates the cached effective roles for a user
func (r *RoleInheritanceEngine) InvalidateUserRoleCache(ctx context.Context, orgID, userID string) error {
	cacheKey := fmt.Sprintf("org:%s:user:%s:effective_roles", orgID, userID)
	return r.cache.Delete(cacheKey)
}

// InvalidateGroupRoleCache invalidates cached effective roles for all users in a group hierarchy
func (r *RoleInheritanceEngine) InvalidateGroupRoleCache(ctx context.Context, orgID, groupID string) error {
	// This would need to invalidate cache for all users affected by changes to this group
	// For now, we'll use a simple pattern-based invalidation
	pattern := fmt.Sprintf("org:%s:user:*:effective_roles", orgID)
	keys, err := r.cache.Keys(pattern)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := r.cache.Delete(key); err != nil {
			r.logger.Warn("Failed to invalidate cache key", zap.String("key", key), zap.Error(err))
		}
	}

	return nil
}

// GetRoleInheritancePath returns the inheritance path for a specific role for a user
func (r *RoleInheritanceEngine) GetRoleInheritancePath(ctx context.Context, orgID, userID, roleID string) (*EffectiveRole, error) {
	effectiveRoles, err := r.CalculateEffectiveRoles(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}

	for _, effectiveRole := range effectiveRoles {
		if effectiveRole.Role.ID == roleID {
			return effectiveRole, nil
		}
	}

	return nil, errors.NewNotFoundError("role not found in user's effective roles")
}

// VerifyBottomUpInheritance verifies that the inheritance engine is correctly implementing
// bottom-up (upward) inheritance by checking that roles flow from child groups to parent groups.
//
// This method performs the following verifications:
//  1. Confirms that parent groups inherit roles from their child groups
//  2. Validates that inheritance paths show upward flow (child → parent)
//  3. Ensures distance increases as we go deeper into child hierarchies
//  4. Verifies that direct assignments have distance 0 and highest precedence
//
// Returns detailed verification results for debugging and validation purposes.
func (r *RoleInheritanceEngine) VerifyBottomUpInheritance(ctx context.Context, orgID, userID string) (*InheritanceVerificationResult, error) {
	r.logger.Info("Verifying bottom-up inheritance implementation",
		zap.String("org_id", orgID),
		zap.String("user_id", userID))

	effectiveRoles, err := r.CalculateEffectiveRoles(ctx, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate effective roles for verification: %w", err)
	}

	// Get user's direct groups for verification
	directGroups, err := r.getUserDirectGroups(ctx, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user's direct groups for verification: %w", err)
	}

	result := &InheritanceVerificationResult{
		UserID:              userID,
		OrganizationID:      orgID,
		DirectGroups:        make([]string, len(directGroups)),
		EffectiveRoles:      effectiveRoles,
		VerificationResults: make(map[string]*RoleVerification),
		IsBottomUpOnly:      true,
		Summary:             &VerificationSummary{},
	}

	// Record direct groups
	for i, group := range directGroups {
		result.DirectGroups[i] = group.ID
	}

	// Verify each effective role
	for _, effectiveRole := range effectiveRoles {
		verification := &RoleVerification{
			RoleID:          effectiveRole.Role.ID,
			RoleName:        effectiveRole.Role.Name,
			SourceGroupID:   effectiveRole.GroupID,
			SourceGroupName: effectiveRole.GroupName,
			Distance:        effectiveRole.Distance,
			InheritancePath: effectiveRole.InheritancePath,
			IsDirectRole:    effectiveRole.IsDirectRole,
			IsValidBottomUp: true,
			Issues:          []string{},
		}

		// Verify inheritance path shows upward flow
		if len(effectiveRole.InheritancePath) > 1 {
			// Path should start with user's direct group and end with role source
			userDirectGroupID := effectiveRole.InheritancePath[0]
			roleSourceGroupID := effectiveRole.InheritancePath[len(effectiveRole.InheritancePath)-1]

			// Verify user is actually in the first group of the path
			isUserInFirstGroup := false
			for _, directGroup := range directGroups {
				if directGroup.ID == userDirectGroupID {
					isUserInFirstGroup = true
					break
				}
			}

			if !isUserInFirstGroup {
				verification.IsValidBottomUp = false
				verification.Issues = append(verification.Issues,
					fmt.Sprintf("inheritance path starts with group %s but user is not a member", userDirectGroupID))
			}

			// Verify role source matches the last group in path
			if effectiveRole.GroupID != roleSourceGroupID {
				verification.IsValidBottomUp = false
				verification.Issues = append(verification.Issues,
					fmt.Sprintf("role source group %s doesn't match last group in path %s",
						effectiveRole.GroupID, roleSourceGroupID))
			}

			// Verify distance matches path length - 1
			expectedDistance := len(effectiveRole.InheritancePath) - 1
			if effectiveRole.Distance != expectedDistance {
				verification.IsValidBottomUp = false
				verification.Issues = append(verification.Issues,
					fmt.Sprintf("distance %d doesn't match path length %d",
						effectiveRole.Distance, expectedDistance))
			}
		}

		// Verify direct roles have distance 0
		if effectiveRole.IsDirectRole && effectiveRole.Distance != 0 {
			verification.IsValidBottomUp = false
			verification.Issues = append(verification.Issues,
				fmt.Sprintf("direct role has non-zero distance %d", effectiveRole.Distance))
		}

		// Verify non-direct roles have distance > 0
		if !effectiveRole.IsDirectRole && effectiveRole.Distance == 0 {
			verification.IsValidBottomUp = false
			verification.Issues = append(verification.Issues,
				"non-direct role has zero distance")
		}

		result.VerificationResults[effectiveRole.Role.ID] = verification

		// Update summary
		result.Summary.TotalRoles++
		if effectiveRole.IsDirectRole {
			result.Summary.DirectRoles++
		} else {
			result.Summary.InheritedRoles++
		}
		if !verification.IsValidBottomUp {
			result.Summary.InvalidRoles++
			result.IsBottomUpOnly = false
		}
	}

	// Verify roles are sorted by precedence (distance ascending)
	for i := 1; i < len(effectiveRoles); i++ {
		if effectiveRoles[i-1].Distance > effectiveRoles[i].Distance {
			result.IsBottomUpOnly = false
			result.Summary.SortingIssues++
		}
	}

	r.logger.Info("Bottom-up inheritance verification completed",
		zap.String("user_id", userID),
		zap.Bool("is_valid_bottom_up", result.IsBottomUpOnly),
		zap.Int("total_roles", result.Summary.TotalRoles),
		zap.Int("direct_roles", result.Summary.DirectRoles),
		zap.Int("inherited_roles", result.Summary.InheritedRoles),
		zap.Int("invalid_roles", result.Summary.InvalidRoles))

	return result, nil
}

// InheritanceVerificationResult contains the results of bottom-up inheritance verification
type InheritanceVerificationResult struct {
	UserID              string                       `json:"user_id"`
	OrganizationID      string                       `json:"organization_id"`
	DirectGroups        []string                     `json:"direct_groups"`
	EffectiveRoles      []*EffectiveRole             `json:"effective_roles"`
	VerificationResults map[string]*RoleVerification `json:"verification_results"`
	IsBottomUpOnly      bool                         `json:"is_bottom_up_only"`
	Summary             *VerificationSummary         `json:"summary"`
}

// RoleVerification contains verification details for a specific role
type RoleVerification struct {
	RoleID          string   `json:"role_id"`
	RoleName        string   `json:"role_name"`
	SourceGroupID   string   `json:"source_group_id"`
	SourceGroupName string   `json:"source_group_name"`
	Distance        int      `json:"distance"`
	InheritancePath []string `json:"inheritance_path"`
	IsDirectRole    bool     `json:"is_direct_role"`
	IsValidBottomUp bool     `json:"is_valid_bottom_up"`
	Issues          []string `json:"issues"`
}

// VerificationSummary provides a summary of verification results
type VerificationSummary struct {
	TotalRoles     int `json:"total_roles"`
	DirectRoles    int `json:"direct_roles"`
	InheritedRoles int `json:"inherited_roles"`
	InvalidRoles   int `json:"invalid_roles"`
	SortingIssues  int `json:"sorting_issues"`
}
