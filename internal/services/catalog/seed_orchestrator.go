package catalog

import (
	"context"
	"fmt"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SeedOrchestrator coordinates the seeding process with transaction management
type SeedOrchestrator struct {
	actionManager     *ActionManager
	resourceManager   *ResourceManager
	permissionManager *PermissionManager
	roleManager       *RoleManager
	providerRegistry  *SeedProviderRegistry
	dbManager         db.DBManager
	logger            *zap.Logger
}

// NewSeedOrchestrator creates a new seed orchestrator
func NewSeedOrchestrator(
	actionManager *ActionManager,
	resourceManager *ResourceManager,
	permissionManager *PermissionManager,
	roleManager *RoleManager,
	providerRegistry *SeedProviderRegistry,
	dbManager db.DBManager,
	logger *zap.Logger,
) *SeedOrchestrator {
	return &SeedOrchestrator{
		actionManager:     actionManager,
		resourceManager:   resourceManager,
		permissionManager: permissionManager,
		roleManager:       roleManager,
		providerRegistry:  providerRegistry,
		dbManager:         dbManager,
		logger:            logger,
	}
}

// getDB is a helper method to get the GORM database connection from the database manager
func (so *SeedOrchestrator) getDB(ctx context.Context, readOnly bool) (*gorm.DB, error) {
	// Try to get the database from the database manager
	if postgresMgr, ok := so.dbManager.(interface {
		GetDB(context.Context, bool) (*gorm.DB, error)
	}); ok {
		return postgresMgr.GetDB(ctx, readOnly)
	}

	return nil, fmt.Errorf("database manager does not support GetDB method")
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
// serviceID parameter is optional - if empty, uses default provider (farmers-module)
func (so *SeedOrchestrator) SeedRolesAndPermissions(ctx context.Context, serviceID string, force bool) (*SeedResult, error) {
	result := &SeedResult{}

	// Validate service_id (defense in depth - should be validated at handler level too)
	if err := ValidateServiceID(serviceID); err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("invalid service_id: %v", err)
		so.logger.Error("Invalid service_id in seed operation",
			zap.String("service_id", serviceID),
			zap.Error(err))
		return result, fmt.Errorf("invalid service_id: %w", err)
	}

	// Get the appropriate provider
	var provider SeedDataProvider
	var err error

	if serviceID == "" {
		// Use default provider (farmers-module) for backward compatibility
		provider = NewDefaultSeedProvider()
		so.logger.Info("Using default seed provider (farmers-module)")
	} else {
		// Get provider from registry
		provider, err = so.providerRegistry.Get(serviceID)
		if err != nil {
			result.Success = false
			result.ErrorMessage = fmt.Sprintf("provider not found for service: %s", serviceID)
			so.logger.Error("Provider not found", zap.String("service_id", serviceID), zap.Error(err))
			return result, fmt.Errorf("provider not found for service %s: %w", serviceID, err)
		}
		so.logger.Info("Using registered seed provider",
			zap.String("service_id", serviceID),
			zap.String("service_name", provider.GetServiceName()))
	}

	// Validate provider data before seeding
	if err := provider.Validate(ctx); err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("provider validation failed: %v", err)
		so.logger.Error("Provider validation failed", zap.Error(err))
		return result, fmt.Errorf("provider validation failed: %w", err)
	}

	// Wrap entire seeding operation in a transaction for atomicity
	// If any step fails, all changes are rolled back
	so.logger.Info("Starting transactional seed operation", zap.String("service", provider.GetServiceName()))

	// Get GORM DB instance for transaction support
	gormDB, err := so.getDB(ctx, false)
	if err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("failed to get database connection: %v", err)
		so.logger.Error("Failed to get database connection", zap.Error(err))
		return result, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Execute seed in transaction
	err = gormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create context with transaction for downstream operations
		txCtx := context.WithValue(ctx, "tx", tx)
		return so.executeSeeding(txCtx, provider, result, force)
	})

	if err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("seed operation failed (rolled back): %v", err)
		so.logger.Error("Seed operation failed and rolled back",
			zap.String("service", provider.GetServiceName()),
			zap.Error(err))
		return result, fmt.Errorf("seed operation failed: %w", err)
	}

	result.Success = true
	so.logger.Info("Successfully seeded roles and permissions (committed)",
		zap.String("service", provider.GetServiceName()),
		zap.Int32("actions", result.ActionsCreated),
		zap.Int32("resources", result.ResourcesCreated),
		zap.Int32("permissions", result.PermissionsCreated),
		zap.Int32("roles", result.RolesCreated))

	return result, nil
}

// executeSeeding performs the actual seeding work within a transaction
// This method should only be called from SeedRolesAndPermissions within a transaction context
func (so *SeedOrchestrator) executeSeeding(
	ctx context.Context,
	provider SeedDataProvider,
	result *SeedResult,
	force bool,
) error {
	// Step 1: Create/update actions
	so.logger.Info("Seeding actions", zap.String("service", provider.GetServiceName()))
	actionDefs := provider.GetActions()
	actionsCount, err := so.actionManager.UpsertActions(ctx, actionDefs, force)
	if err != nil {
		so.logger.Error("Failed to seed actions", zap.Error(err))
		return fmt.Errorf("failed to seed actions: %w", err)
	}
	result.ActionsCreated = actionsCount
	so.logger.Info("Actions seeded", zap.Int32("count", actionsCount))

	// Step 2: Create/update resources
	so.logger.Info("Seeding resources", zap.String("service", provider.GetServiceName()))
	resourceDefs := provider.GetResources()
	resourcesCount, err := so.resourceManager.UpsertResources(ctx, resourceDefs, force)
	if err != nil {
		so.logger.Error("Failed to seed resources", zap.Error(err))
		return fmt.Errorf("failed to seed resources: %w", err)
	}
	result.ResourcesCreated = resourcesCount
	so.logger.Info("Resources seeded", zap.Int32("count", resourcesCount))

	// Step 3: Create/update permissions (with wildcard expansion)
	so.logger.Info("Seeding permissions with wildcard expansion", zap.String("service", provider.GetServiceName()))
	roleDefs := provider.GetRoles()

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
		so.logger.Error("Failed to seed permissions", zap.Error(err))
		return fmt.Errorf("failed to seed permissions: %w", err)
	}
	result.PermissionsCreated = permsCount
	so.logger.Info("Permissions seeded with wildcard expansion",
		zap.Int32("count", permsCount),
		zap.Int("unique_permissions", len(permissionIDs)))

	// Step 4: Create/update roles and attach permissions with service mapping
	so.logger.Info("Seeding roles and attaching permissions", zap.String("service", provider.GetServiceName()))
	rolesCount, roleNames, err := so.roleManager.UpsertRolesWithPermissions(
		ctx,
		roleDefs,
		permissionIDs,
		provider.GetServiceID(),
		provider.GetServiceName(),
		force,
	)
	if err != nil {
		so.logger.Error("Failed to seed roles", zap.Error(err))
		return fmt.Errorf("failed to seed roles: %w", err)
	}
	result.RolesCreated = rolesCount
	result.CreatedRoleNames = roleNames
	so.logger.Info("Roles seeded with permissions attached",
		zap.Int32("count", rolesCount),
		zap.Strings("roles", roleNames))

	return nil
}
