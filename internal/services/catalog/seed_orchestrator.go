package catalog

import (
	"context"
	"fmt"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
)

// SeedOrchestrator coordinates the seeding process with transaction management
type SeedOrchestrator struct {
	actionManager     *ActionManager
	resourceManager   *ResourceManager
	permissionManager *PermissionManager
	roleManager       *RoleManager
	seedDataProvider  *SeedDataProvider
	dbManager         db.DBManager
	logger            *zap.Logger
}

// NewSeedOrchestrator creates a new seed orchestrator
func NewSeedOrchestrator(
	actionManager *ActionManager,
	resourceManager *ResourceManager,
	permissionManager *PermissionManager,
	roleManager *RoleManager,
	seedDataProvider *SeedDataProvider,
	dbManager db.DBManager,
	logger *zap.Logger,
) *SeedOrchestrator {
	return &SeedOrchestrator{
		actionManager:     actionManager,
		resourceManager:   resourceManager,
		permissionManager: permissionManager,
		roleManager:       roleManager,
		seedDataProvider:  seedDataProvider,
		dbManager:         dbManager,
		logger:            logger,
	}
}

// SeedResult contains the results of the seeding operation
type SeedResult struct {
	ActionsCreated     int32
	ResourcesCreated   int32
	PermissionsCreated int32
	RolesCreated       int32
	CreatedRoleNames   []string
	Success            bool
	ErrorMessage       string
}

// SeedRolesAndPermissions performs the complete seeding operation
func (so *SeedOrchestrator) SeedRolesAndPermissions(ctx context.Context, force bool) (*SeedResult, error) {
	result := &SeedResult{}

	// Step 1: Create/update actions
	so.logger.Info("Seeding actions")
	actionDefs := so.seedDataProvider.GetActions()
	actionsCount, err := so.actionManager.UpsertActions(ctx, actionDefs, force)
	if err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("failed to seed actions: %v", err)
		so.logger.Error("Failed to seed actions", zap.Error(err))
		return result, fmt.Errorf("failed to seed actions: %w", err)
	}
	result.ActionsCreated = actionsCount
	so.logger.Info("Actions seeded", zap.Int32("count", actionsCount))

	// Step 2: Create/update resources
	so.logger.Info("Seeding resources")
	resourceDefs := so.seedDataProvider.GetResources()
	resourcesCount, err := so.resourceManager.UpsertResources(ctx, resourceDefs, force)
	if err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("failed to seed resources: %v", err)
		so.logger.Error("Failed to seed resources", zap.Error(err))
		return result, fmt.Errorf("failed to seed resources: %w", err)
	}
	result.ResourcesCreated = resourcesCount
	so.logger.Info("Resources seeded", zap.Int32("count", resourcesCount))

	// Step 3: Create/update permissions (with wildcard expansion)
	so.logger.Info("Seeding permissions with wildcard expansion")
	roleDefs := so.seedDataProvider.GetRoles()

	// Collect all permission patterns
	allPatterns := make([]string, 0)
	for _, roleDef := range roleDefs {
		allPatterns = append(allPatterns, roleDef.Permissions...)
	}

	// Expand wildcards and create permissions
	permsCount, permissionIDs, err := so.permissionManager.ExpandAndUpsertPermissions(
		ctx,
		allPatterns,
		force,
	)
	if err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("failed to seed permissions: %v", err)
		so.logger.Error("Failed to seed permissions", zap.Error(err))
		return result, fmt.Errorf("failed to seed permissions: %w", err)
	}
	result.PermissionsCreated = permsCount
	so.logger.Info("Permissions seeded with wildcard expansion",
		zap.Int32("count", permsCount),
		zap.Int("unique_permissions", len(permissionIDs)))

	// Step 4: Create/update roles and attach permissions
	so.logger.Info("Seeding roles and attaching permissions")
	rolesCount, roleNames, err := so.roleManager.UpsertRolesWithPermissions(
		ctx,
		roleDefs,
		permissionIDs,
		force,
	)
	if err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("failed to seed roles: %v", err)
		so.logger.Error("Failed to seed roles", zap.Error(err))
		return result, fmt.Errorf("failed to seed roles: %w", err)
	}
	result.RolesCreated = rolesCount
	result.CreatedRoleNames = roleNames
	so.logger.Info("Roles seeded with permissions attached",
		zap.Int32("count", rolesCount),
		zap.Strings("roles", roleNames))

	result.Success = true
	so.logger.Info("Successfully seeded roles and permissions",
		zap.Int32("actions", result.ActionsCreated),
		zap.Int32("resources", result.ResourcesCreated),
		zap.Int32("permissions", result.PermissionsCreated),
		zap.Int32("roles", result.RolesCreated))

	return result, nil
}
