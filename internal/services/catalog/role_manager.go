package catalog

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/role_permissions"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/roles"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
)

// RoleManager handles role-related operations for the catalog service
type RoleManager struct {
	roleRepo           *roles.RoleRepository
	rolePermissionRepo *role_permissions.RolePermissionRepository
	dbManager          db.DBManager
	logger             *zap.Logger
}

// NewRoleManager creates a new role manager
func NewRoleManager(
	roleRepo *roles.RoleRepository,
	rolePermissionRepo *role_permissions.RolePermissionRepository,
	dbManager db.DBManager,
	logger *zap.Logger,
) *RoleManager {
	return &RoleManager{
		roleRepo:           roleRepo,
		rolePermissionRepo: rolePermissionRepo,
		dbManager:          dbManager,
		logger:             logger,
	}
}

// UpsertRolesWithPermissions creates or updates roles and attaches permissions
// Returns (rolesCreated, createdRoleNames, error)
func (rm *RoleManager) UpsertRolesWithPermissions(
	ctx context.Context,
	roleDefs []RoleDefinition,
	permissionIDs map[string]string,
	force bool,
) (int32, []string, error) {
	var count int32
	var createdRoleNames []string

	for _, roleDef := range roleDefs {
		// Check if role already exists
		existing, err := rm.roleRepo.GetByName(ctx, roleDef.Name)

		var role *models.Role

		if err == nil && existing != nil {
			// Role exists
			if !force {
				rm.logger.Debug("Role already exists, skipping",
					zap.String("role", roleDef.Name))
				continue
			}

			// Update existing role
			existing.Description = roleDef.Description
			existing.Scope = roleDef.Scope

			if err := rm.roleRepo.Update(ctx, existing); err != nil {
				return count, createdRoleNames, fmt.Errorf("failed to update role %s: %w", roleDef.Name, err)
			}

			role = existing
			rm.logger.Debug("Role updated",
				zap.String("role", roleDef.Name))
		} else {
			// Create new role
			role = models.NewRole(roleDef.Name, roleDef.Description, roleDef.Scope)

			if err := rm.roleRepo.Create(ctx, role); err != nil {
				return count, createdRoleNames, fmt.Errorf("failed to create role %s: %w", roleDef.Name, err)
			}

			rm.logger.Debug("Role created",
				zap.String("role", roleDef.Name),
				zap.String("id", role.ID))
		}

		// Attach permissions to role
		if err := rm.attachPermissionsToRole(ctx, role.ID, roleDef.Permissions, permissionIDs); err != nil {
			return count, createdRoleNames, fmt.Errorf("failed to attach permissions to role %s: %w", roleDef.Name, err)
		}

		count++
		createdRoleNames = append(createdRoleNames, roleDef.Name)
	}

	return count, createdRoleNames, nil
}

// attachPermissionsToRole attaches permissions to a role using the role_permissions join table
func (rm *RoleManager) attachPermissionsToRole(
	ctx context.Context,
	roleID string,
	permissionPatterns []string,
	permissionIDs map[string]string,
) error {
	// Get current permissions for the role
	currentPerms, err := rm.rolePermissionRepo.GetByRoleID(ctx, roleID)
	if err != nil {
		rm.logger.Warn("Failed to get current permissions for role, continuing",
			zap.String("roleID", roleID),
			zap.Error(err))
	}

	// Track which permissions should be attached
	shouldAttach := make(map[string]bool)

	// Expand patterns and collect permission IDs
	for _, pattern := range permissionPatterns {
		// Check if this is a wildcard pattern
		if pattern == "*:*" {
			// Attach all permissions
			for _, permID := range permissionIDs {
				shouldAttach[permID] = true
			}
		} else if pattern[len(pattern)-2:] == ":*" || pattern[:2] == "*:" {
			// Partial wildcard - find matching permissions
			for permName, permID := range permissionIDs {
				if matchesPattern(permName, pattern) {
					shouldAttach[permID] = true
				}
			}
		} else {
			// Exact match
			if permID, ok := permissionIDs[pattern]; ok {
				shouldAttach[permID] = true
			} else {
				rm.logger.Warn("Permission not found for pattern",
					zap.String("pattern", pattern))
			}
		}
	}

	// Create map of current permissions for quick lookup
	currentPermMap := make(map[string]bool)
	for _, rp := range currentPerms {
		currentPermMap[rp.PermissionID] = true
	}

	// Attach new permissions
	for permID := range shouldAttach {
		if !currentPermMap[permID] {
			// Create role_permission association using the model constructor
			rolePermission := models.NewRolePermission(roleID, permID)

			// Create the association
			if err := rm.dbManager.Create(ctx, rolePermission); err != nil {
				return fmt.Errorf("failed to create role_permission association: %w", err)
			}

			rm.logger.Debug("Permission attached to role",
				zap.String("roleID", roleID),
				zap.String("permissionID", permID))
		}
	}

	return nil
}

// matchesPattern checks if a permission name matches a pattern
func matchesPattern(permName, pattern string) bool {
	if pattern == "*:*" {
		return true
	}

	// Split both into resource:action
	permParts := splitPermission(permName)
	patternParts := splitPermission(pattern)

	if len(permParts) != 2 || len(patternParts) != 2 {
		return false
	}

	// Check resource match
	resourceMatch := patternParts[0] == "*" || patternParts[0] == permParts[0]

	// Check action match
	actionMatch := patternParts[1] == "*" || patternParts[1] == permParts[1]

	return resourceMatch && actionMatch
}

// splitPermission splits a permission name into [resource, action]
func splitPermission(perm string) []string {
	parts := make([]string, 2)
	idx := 0
	for i, ch := range perm {
		if ch == ':' {
			parts[0] = perm[:i]
			parts[1] = perm[i+1:]
			idx = 2
			break
		}
	}
	if idx == 0 {
		return []string{perm}
	}
	return parts
}
