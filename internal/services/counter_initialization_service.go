package services

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CounterInitializationService handles initialization of ID counters from database
type CounterInitializationService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewCounterInitializationService creates a new counter initialization service
func NewCounterInitializationService(db *gorm.DB, logger *zap.Logger) *CounterInitializationService {
	return &CounterInitializationService{
		db:     db,
		logger: logger,
	}
}

// InitializeAllCounters initializes counters for all model types from the database
func (cis *CounterInitializationService) InitializeAllCounters(ctx context.Context) error {
	cis.logger.Info("Starting counter initialization from database")

	// Initialize counters for each model type
	if err := cis.initializeUserCounters(ctx); err != nil {
		cis.logger.Error("Failed to initialize user counters", zap.Error(err))
		return err
	}

	if err := cis.initializeRoleCounters(ctx); err != nil {
		cis.logger.Error("Failed to initialize role counters", zap.Error(err))
		return err
	}

	if err := cis.initializePermissionCounters(ctx); err != nil {
		cis.logger.Error("Failed to initialize permission counters", zap.Error(err))
		return err
	}

	if err := cis.initializeActionCounters(ctx); err != nil {
		cis.logger.Error("Failed to initialize action counters", zap.Error(err))
		return err
	}

	if err := cis.initializeOrganizationCounters(ctx); err != nil {
		cis.logger.Error("Failed to initialize organization counters", zap.Error(err))
		return err
	}

	if err := cis.initializeGroupCounters(ctx); err != nil {
		cis.logger.Error("Failed to initialize group counters", zap.Error(err))
		return err
	}

	if err := cis.initializeResourceCounters(ctx); err != nil {
		cis.logger.Error("Failed to initialize resource counters", zap.Error(err))
		return err
	}

	if err := cis.initializeAddressCounters(ctx); err != nil {
		cis.logger.Error("Failed to initialize address counters", zap.Error(err))
		return err
	}

	cis.logger.Info("Counter initialization completed successfully")
	return nil
}

// initializeUserCounters initializes user ID counters
func (cis *CounterInitializationService) initializeUserCounters(ctx context.Context) error {
	cis.logger.Info("Starting user counter initialization...")

	var users []models.User
	if err := cis.db.WithContext(ctx).Select("id").Find(&users).Error; err != nil {
		cis.logger.Error("Failed to fetch user IDs", zap.Error(err))
		return fmt.Errorf("failed to fetch user IDs: %w", err)
	}

	cis.logger.Info("Found users in database", zap.Int("count", len(users)))

	var existingIDs []string
	for _, user := range users {
		existingIDs = append(existingIDs, user.ID)
		cis.logger.Debug("Found user ID", zap.String("id", user.ID))
	}

	if len(existingIDs) > 0 {
		cis.logger.Info("Initializing user counters",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))

		// Call the hash package function
		hash.InitializeGlobalCountersFromDatabase("USER", existingIDs, hash.Medium)
		cis.logger.Info("User counters initialized with hash package")
	} else {
		cis.logger.Info("No existing users found, skipping user counter initialization")
	}

	return nil
}

// initializeRoleCounters initializes role ID counters
func (cis *CounterInitializationService) initializeRoleCounters(ctx context.Context) error {
	var roles []models.Role
	if err := cis.db.WithContext(ctx).Select("id").Find(&roles).Error; err != nil {
		return fmt.Errorf("failed to fetch role IDs: %w", err)
	}

	var existingIDs []string
	for _, role := range roles {
		existingIDs = append(existingIDs, role.ID)
	}

	if len(existingIDs) > 0 {
		cis.logger.Info("Initializing role counters",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))

		hash.InitializeGlobalCountersFromDatabase("ROLE", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing roles found, skipping role counter initialization")
	}

	return nil
}

// initializePermissionCounters initializes permission ID counters
func (cis *CounterInitializationService) initializePermissionCounters(ctx context.Context) error {
	var permissions []models.Permission
	if err := cis.db.WithContext(ctx).Select("id").Find(&permissions).Error; err != nil {
		return fmt.Errorf("failed to fetch permission IDs: %w", err)
	}

	var existingIDs []string
	for _, permission := range permissions {
		existingIDs = append(existingIDs, permission.ID)
	}

	if len(existingIDs) > 0 {
		cis.logger.Info("Initializing permission counters",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))

		hash.InitializeGlobalCountersFromDatabase("PERM", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing permissions found, skipping permission counter initialization")
	}

	return nil
}

// initializeActionCounters initializes action ID counters
func (cis *CounterInitializationService) initializeActionCounters(ctx context.Context) error {
	var actions []models.Action
	if err := cis.db.WithContext(ctx).Select("id").Find(&actions).Error; err != nil {
		return fmt.Errorf("failed to fetch action IDs: %w", err)
	}

	var existingIDs []string
	for _, action := range actions {
		existingIDs = append(existingIDs, action.ID)
	}

	if len(existingIDs) > 0 {
		cis.logger.Info("Initializing action counters",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))

		hash.InitializeGlobalCountersFromDatabase("ACTN", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing actions found, skipping action counter initialization")
	}

	return nil
}

// initializeOrganizationCounters initializes organization ID counters
func (cis *CounterInitializationService) initializeOrganizationCounters(ctx context.Context) error {
	var organizations []models.Organization
	if err := cis.db.WithContext(ctx).Select("id").Find(&organizations).Error; err != nil {
		return fmt.Errorf("failed to fetch organization IDs: %w", err)
	}

	var existingIDs []string
	for _, org := range organizations {
		existingIDs = append(existingIDs, org.ID)
	}

	if len(existingIDs) > 0 {
		cis.logger.Info("Initializing organization counters",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))

		hash.InitializeGlobalCountersFromDatabase("ORG", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing organizations found, skipping organization counter initialization")
	}

	return nil
}

// initializeGroupCounters initializes group ID counters
func (cis *CounterInitializationService) initializeGroupCounters(ctx context.Context) error {
	var groups []models.Group
	if err := cis.db.WithContext(ctx).Select("id").Find(&groups).Error; err != nil {
		return fmt.Errorf("failed to fetch group IDs: %w", err)
	}

	var existingIDs []string
	for _, group := range groups {
		existingIDs = append(existingIDs, group.ID)
	}

	if len(existingIDs) > 0 {
		cis.logger.Info("Initializing group counters",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))

		hash.InitializeGlobalCountersFromDatabase(groups[0].GetTableIdentifier(), existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing groups found, skipping group counter initialization")
	}

	return nil
}

// initializeResourceCounters initializes resource ID counters
func (cis *CounterInitializationService) initializeResourceCounters(ctx context.Context) error {
	var resources []models.Resource
	if err := cis.db.WithContext(ctx).Select("id").Find(&resources).Error; err != nil {
		return fmt.Errorf("failed to fetch resource IDs: %w", err)
	}

	var existingIDs []string
	for _, resource := range resources {
		existingIDs = append(existingIDs, resource.ID)
	}

	if len(existingIDs) > 0 {
		cis.logger.Info("Initializing resource counters",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))

		hash.InitializeGlobalCountersFromDatabase("RES", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing resources found, skipping resource counter initialization")
	}

	return nil
}

// initializeAddressCounters initializes address ID counters
func (cis *CounterInitializationService) initializeAddressCounters(ctx context.Context) error {
	var addresses []models.Address
	if err := cis.db.WithContext(ctx).Select("id").Find(&addresses).Error; err != nil {
		return fmt.Errorf("failed to fetch address IDs: %w", err)
	}

	var existingIDs []string
	for _, address := range addresses {
		existingIDs = append(existingIDs, address.ID)
	}

	if len(existingIDs) > 0 {
		cis.logger.Info("Initializing address counters",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))

		hash.InitializeGlobalCountersFromDatabase("ADDR", existingIDs, hash.Large)
	} else {
		cis.logger.Info("No existing addresses found, skipping address counter initialization")
	}

	return nil
}

// GetCounterStatus returns the current status of all counters for debugging
func (cis *CounterInitializationService) GetCounterStatus() map[string]interface{} {
	// This would need to be implemented in the hash package to expose counter values
	// For now, return a placeholder
	return map[string]interface{}{
		"status": "counters_initialized",
		"note":   "Counter values are not exposed by the hash package yet",
	}
}
