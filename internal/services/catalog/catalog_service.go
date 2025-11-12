package catalog

import (
	"context"

	"github.com/Kisanlink/aaa-service/v2/internal/repositories/actions"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/permissions"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/resources"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/role_permissions"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/roles"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/service_role_mappings"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
)

// CatalogService provides catalog management operations for actions, resources, roles, and permissions
type CatalogService struct {
	seedOrchestrator *SeedOrchestrator
	providerRegistry *SeedProviderRegistry
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
	serviceMappingRepo := service_role_mappings.NewServiceRoleMappingRepository(dbManager)

	// Initialize managers
	actionManager := NewActionManager(actionRepo, logger)
	resourceManager := NewResourceManager(resourceRepo, logger)
	permissionManager := NewPermissionManager(permissionRepo, actionManager, resourceManager, logger)
	roleManager := NewRoleManager(roleRepo, rolePermissionRepo, serviceMappingRepo, dbManager, logger)

	// Initialize provider registry and register default providers
	providerRegistry := NewSeedProviderRegistry()

	// Register default provider (Farmers-module)
	defaultProvider := NewDefaultSeedProvider()
	if err := providerRegistry.Register(defaultProvider); err != nil {
		logger.Warn("Failed to register default provider", zap.Error(err))
	} else {
		logger.Info("Registered default seed provider", zap.String("service", defaultProvider.GetServiceName()))
	}

	// Register ERP provider as example
	erpProvider := NewERPSeedProvider()
	if err := providerRegistry.Register(erpProvider); err != nil {
		logger.Warn("Failed to register ERP provider", zap.Error(err))
	} else {
		logger.Info("Registered ERP seed provider", zap.String("service", erpProvider.GetServiceName()))
	}

	// Initialize seed orchestrator
	seedOrchestrator := NewSeedOrchestrator(
		actionManager,
		resourceManager,
		permissionManager,
		roleManager,
		providerRegistry,
		dbManager,
		logger,
	)

	return &CatalogService{
		seedOrchestrator: seedOrchestrator,
		providerRegistry: providerRegistry,
		actionManager:    actionManager,
		resourceManager:  resourceManager,
		roleManager:      roleManager,
		logger:           logger,
	}
}

// SeedRolesAndPermissions seeds the database with predefined roles and permissions
// serviceID parameter is optional - if empty, uses default provider (farmers-module)
func (cs *CatalogService) SeedRolesAndPermissions(ctx context.Context, serviceID string, force bool) (*SeedResult, error) {
	cs.logger.Info("Starting seed roles and permissions operation",
		zap.String("service_id", serviceID),
		zap.Bool("force", force))

	result, err := cs.seedOrchestrator.SeedRolesAndPermissions(ctx, serviceID, force)
	if err != nil {
		cs.logger.Error("Seed roles and permissions failed",
			zap.String("service_id", serviceID),
			zap.Error(err))
		return result, err
	}

	cs.logger.Info("Seed roles and permissions completed successfully",
		zap.String("service_id", serviceID),
		zap.Int32("actions_created", result.ActionsCreated),
		zap.Int32("resources_created", result.ResourcesCreated),
		zap.Int32("permissions_created", result.PermissionsCreated),
		zap.Int32("roles_created", result.RolesCreated),
		zap.Strings("created_roles", result.CreatedRoleNames))

	return result, nil
}

// RegisterSeedProvider registers a new seed data provider
func (cs *CatalogService) RegisterSeedProvider(provider SeedDataProvider) error {
	if err := cs.providerRegistry.Register(provider); err != nil {
		cs.logger.Error("Failed to register seed provider",
			zap.String("service_id", provider.GetServiceID()),
			zap.Error(err))
		return err
	}

	cs.logger.Info("Registered seed provider",
		zap.String("service_id", provider.GetServiceID()),
		zap.String("service_name", provider.GetServiceName()))

	return nil
}

// UnregisterSeedProvider removes a seed data provider
func (cs *CatalogService) UnregisterSeedProvider(serviceID string) error {
	if err := cs.providerRegistry.Unregister(serviceID); err != nil {
		cs.logger.Error("Failed to unregister seed provider",
			zap.String("service_id", serviceID),
			zap.Error(err))
		return err
	}

	cs.logger.Info("Unregistered seed provider", zap.String("service_id", serviceID))
	return nil
}

// ListRegisteredProviders returns all registered seed providers
func (cs *CatalogService) ListRegisteredProviders() []SeedDataProvider {
	return cs.providerRegistry.GetAll()
}
