package catalog

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/permissions"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"go.uber.org/zap"
)

// PermissionManager handles permission-related operations including wildcard expansion
type PermissionManager struct {
	permissionRepo  *permissions.PermissionRepository
	actionManager   *ActionManager
	resourceManager *ResourceManager
	logger          *zap.Logger
}

// NewPermissionManager creates a new permission manager
func NewPermissionManager(
	permissionRepo *permissions.PermissionRepository,
	actionManager *ActionManager,
	resourceManager *ResourceManager,
	logger *zap.Logger,
) *PermissionManager {
	return &PermissionManager{
		permissionRepo:  permissionRepo,
		actionManager:   actionManager,
		resourceManager: resourceManager,
		logger:          logger,
	}
}

// ExpandAndUpsertPermissions expands wildcard patterns and creates/updates permissions
// Returns the count of permissions created/updated
func (pm *PermissionManager) ExpandAndUpsertPermissions(
	ctx context.Context,
	permissionPatterns []string,
	force bool,
) (int32, map[string]string, error) {
	// Get all resources and actions for wildcard expansion
	allResources, err := pm.resourceManager.GetAllResources(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get resources for expansion: %w", err)
	}

	allActions, err := pm.actionManager.GetAllActions(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get actions for expansion: %w", err)
	}

	// Expand patterns into concrete permissions
	expandedPerms := pm.expandWildcardPatterns(permissionPatterns, allResources, allActions)

	// Create/update permissions and track their IDs
	var count int32
	permissionIDs := make(map[string]string) // permissionName -> permissionID

	for _, perm := range expandedPerms {
		permID, created, err := pm.upsertPermission(ctx, perm.Resource, perm.Action, force)
		if err != nil {
			return count, permissionIDs, err
		}

		if created {
			count++
		}

		permName := fmt.Sprintf("%s:%s", perm.Resource.Name, perm.Action.Name)
		permissionIDs[permName] = permID
	}

	return count, permissionIDs, nil
}

// ExpandedPermission represents a concrete permission after wildcard expansion
type ExpandedPermission struct {
	Resource *models.Resource
	Action   *models.Action
}

// expandWildcardPatterns expands wildcard patterns into concrete permissions
func (pm *PermissionManager) expandWildcardPatterns(
	patterns []string,
	allResources []*models.Resource,
	allActions []*models.Action,
) []ExpandedPermission {
	var result []ExpandedPermission
	seen := make(map[string]bool) // Track duplicates

	for _, pattern := range patterns {
		parts := strings.Split(pattern, ":")
		if len(parts) != 2 {
			pm.logger.Warn("Invalid permission pattern, skipping",
				zap.String("pattern", pattern))
			continue
		}

		resourcePattern := parts[0]
		actionPattern := parts[1]

		// Handle wildcards
		var resources []*models.Resource
		var actions []*models.Action

		if resourcePattern == "*" {
			resources = allResources
		} else {
			// Find matching resource
			for _, r := range allResources {
				if r.Name == resourcePattern {
					resources = append(resources, r)
					break
				}
			}
		}

		if actionPattern == "*" {
			actions = allActions
		} else {
			// Find matching action
			for _, a := range allActions {
				if a.Name == actionPattern {
					actions = append(actions, a)
					break
				}
			}
		}

		// Create expanded permissions
		for _, resource := range resources {
			for _, action := range actions {
				key := fmt.Sprintf("%s:%s", resource.Name, action.Name)
				if !seen[key] {
					result = append(result, ExpandedPermission{
						Resource: resource,
						Action:   action,
					})
					seen[key] = true
				}
			}
		}
	}

	return result
}

// upsertPermission creates or updates a single permission
// Returns (permissionID, wasCreated, error)
func (pm *PermissionManager) upsertPermission(
	ctx context.Context,
	resource *models.Resource,
	action *models.Action,
	force bool,
) (string, bool, error) {
	permName := fmt.Sprintf("%s:%s", resource.Name, action.Name)

	// Check if permission exists
	existing, err := pm.getPermissionByName(ctx, permName)

	if err == nil && existing != nil {
		// Permission exists
		if !force {
			pm.logger.Debug("Permission already exists, skipping",
				zap.String("permission", permName))
			return existing.ID, false, nil
		}

		// Update existing permission
		existing.ResourceID = &resource.ID
		existing.ActionID = &action.ID
		existing.Description = fmt.Sprintf("%s permission on %s", action.Name, resource.Name)

		if err := pm.permissionRepo.Update(ctx, existing); err != nil {
			return "", false, fmt.Errorf("failed to update permission %s: %w", permName, err)
		}

		pm.logger.Debug("Permission updated",
			zap.String("permission", permName))
		return existing.ID, true, nil
	}

	// Create new permission
	permission := models.NewPermissionWithResourceAndAction(
		permName,
		fmt.Sprintf("%s permission on %s", action.Name, resource.Name),
		resource.ID,
		action.ID,
	)

	if err := pm.permissionRepo.Create(ctx, permission); err != nil {
		return "", false, fmt.Errorf("failed to create permission %s: %w", permName, err)
	}

	pm.logger.Debug("Permission created",
		zap.String("permission", permName),
		zap.String("id", permission.ID))
	return permission.ID, true, nil
}

// getPermissionByName retrieves a permission by name
func (pm *PermissionManager) getPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Build()

	permissions, err := pm.permissionRepo.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find permission by name: %w", err)
	}

	if len(permissions) == 0 {
		return nil, nil
	}

	return permissions[0], nil
}

// GetPermissionByName retrieves a permission by name (public method)
func (pm *PermissionManager) GetPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	return pm.getPermissionByName(ctx, name)
}
