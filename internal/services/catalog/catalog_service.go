package catalog

import (
	"context"

	"github.com/Kisanlink/aaa-service/v2/internal/repositories/actions"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/permissions"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/resources"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/role_permissions"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/roles"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
)

// CatalogService provides catalog management operations for actions, resources, roles, and permissions
type CatalogService struct {
	seedOrchestrator *SeedOrchestrator
	actionManager    *ActionManager
	resourceManager  *ResourceManager
	roleManager      *RoleManager
	logger           *zap.Logger
}

// NewCatalogService creates a new catalog service
func NewCatalogService(
	dbManager db.DBManager,
	logger *zap.Logger,
) *CatalogService {
	// Initialize repositories
	actionRepo := actions.NewActionRepository(dbManager)
	resourceRepo := resources.NewResourceRepository(dbManager)
	permissionRepo := permissions.NewPermissionRepository(dbManager)
	roleRepo := roles.NewRoleRepository(dbManager)
	rolePermissionRepo := role_permissions.NewRolePermissionRepository(dbManager)

	// Initialize managers
	actionManager := NewActionManager(actionRepo, logger)
	resourceManager := NewResourceManager(resourceRepo, logger)
	permissionManager := NewPermissionManager(permissionRepo, actionManager, resourceManager, logger)
	roleManager := NewRoleManager(roleRepo, rolePermissionRepo, dbManager, logger)

	// Initialize seed data provider
	seedDataProvider := NewSeedDataProvider()

	// Initialize seed orchestrator
	seedOrchestrator := NewSeedOrchestrator(
		actionManager,
		resourceManager,
		permissionManager,
		roleManager,
		seedDataProvider,
		dbManager,
		logger,
	)

	return &CatalogService{
		seedOrchestrator: seedOrchestrator,
		actionManager:    actionManager,
		resourceManager:  resourceManager,
		roleManager:      roleManager,
		logger:           logger,
	}
}

// SeedRolesAndPermissions seeds the database with predefined roles and permissions
func (cs *CatalogService) SeedRolesAndPermissions(ctx context.Context, force bool) (*SeedResult, error) {
	cs.logger.Info("Starting seed roles and permissions operation",
		zap.Bool("force", force))

	result, err := cs.seedOrchestrator.SeedRolesAndPermissions(ctx, force)
	if err != nil {
		cs.logger.Error("Seed roles and permissions failed",
			zap.Error(err))
		return result, err
	}

	cs.logger.Info("Seed roles and permissions completed successfully",
		zap.Int32("actions_created", result.ActionsCreated),
		zap.Int32("resources_created", result.ResourcesCreated),
		zap.Int32("permissions_created", result.PermissionsCreated),
		zap.Int32("roles_created", result.RolesCreated),
		zap.Strings("created_roles", result.CreatedRoleNames))

	return result, nil
}
