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

	if err := cis.initializeContactCounters(ctx); err != nil {
		cis.logger.Error("Failed to initialize contact counters", zap.Error(err))
		return err
	}

	if err := cis.initializeUserProfileCounters(ctx); err != nil {
		cis.logger.Error("Failed to initialize user profile counters", zap.Error(err))
		return err
	}

	if err := cis.initializeUserRoleCounters(ctx); err != nil {
		cis.logger.Error("Failed to initialize user role counters", zap.Error(err))
		return err
	}

	if err := cis.initializeGroupRoleCounters(ctx); err != nil {
		cis.logger.Error("Failed to initialize group role counters", zap.Error(err))
		return err
	}

	if err := cis.initializeGroupMembershipCounters(ctx); err != nil {
		cis.logger.Error("Failed to initialize group membership counters", zap.Error(err))
		return err
	}

	if err := cis.initializeGroupInheritanceCounters(ctx); err != nil {
		cis.logger.Error("Failed to initialize group inheritance counters", zap.Error(err))
		return err
	}

	if err := cis.initializeBindingCounters(ctx); err != nil {
		cis.logger.Error("Failed to initialize binding counters", zap.Error(err))
		return err
	}

	if err := cis.initializeBindingHistoryCounters(ctx); err != nil {
		cis.logger.Error("Failed to initialize binding history counters", zap.Error(err))
		return err
	}

	if err := cis.initializeAuditLogCounters(ctx); err != nil {
		cis.logger.Error("Failed to initialize audit log counters", zap.Error(err))
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
		cis.logger.Info("Initializing user counters from existing records",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))
		hash.InitializeGlobalCountersFromDatabase("USER", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing users found, initializing counter to start from 1")
		// Initialize with a zero ID so next ID will be 00000001
		hash.InitializeGlobalCountersFromDatabase("USER", []string{"USER00000000"}, hash.Medium)
	}

	cis.logger.Info("User counters initialized successfully")
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
		cis.logger.Info("Initializing role counters from existing records",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))
		hash.InitializeGlobalCountersFromDatabase("ROLE", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing roles found, initializing counter to start from 1")
		hash.InitializeGlobalCountersFromDatabase("ROLE", []string{"ROLE00000000"}, hash.Medium)
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
		cis.logger.Info("Initializing permission counters from existing records",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))
		hash.InitializeGlobalCountersFromDatabase("PERM", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing permissions found, initializing counter to start from 1")
		hash.InitializeGlobalCountersFromDatabase("PERM", []string{"PERM00000000"}, hash.Medium)
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
		cis.logger.Info("Initializing action counters from existing records",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))
		hash.InitializeGlobalCountersFromDatabase("ACTN", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing actions found, initializing counter to start from 1")
		hash.InitializeGlobalCountersFromDatabase("ACTN", []string{"ACTN00000000"}, hash.Medium)
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
		cis.logger.Info("Initializing organization counters from existing records",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))
		hash.InitializeGlobalCountersFromDatabase("ORGN", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing organizations found, initializing counter to start from 1")
		hash.InitializeGlobalCountersFromDatabase("ORGN", []string{"ORGN00000000"}, hash.Medium)
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
		cis.logger.Info("Initializing group counters from existing records",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))
		hash.InitializeGlobalCountersFromDatabase("GRPN", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing groups found, initializing counter to start from 1")
		hash.InitializeGlobalCountersFromDatabase("GRPN", []string{"GRPN00000000"}, hash.Medium)
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
		cis.logger.Info("Initializing resource counters from existing records",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))
		hash.InitializeGlobalCountersFromDatabase("RES", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing resources found, initializing counter to start from 1")
		hash.InitializeGlobalCountersFromDatabase("RES", []string{"RES00000000"}, hash.Medium)
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
		cis.logger.Info("Initializing address counters from existing records",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))
		hash.InitializeGlobalCountersFromDatabase("ADDR", existingIDs, hash.Large)
	} else {
		cis.logger.Info("No existing addresses found, initializing counter to start from 1")
		hash.InitializeGlobalCountersFromDatabase("ADDR", []string{"ADDR00000000"}, hash.Large)
	}

	return nil
}

// initializeContactCounters initializes contact ID counters
func (cis *CounterInitializationService) initializeContactCounters(ctx context.Context) error {
	var contacts []models.Contact
	if err := cis.db.WithContext(ctx).Select("id").Find(&contacts).Error; err != nil {
		return fmt.Errorf("failed to fetch contact IDs: %w", err)
	}

	var existingIDs []string
	for _, contact := range contacts {
		existingIDs = append(existingIDs, contact.ID)
	}

	if len(existingIDs) > 0 {
		cis.logger.Info("Initializing contact counters from existing records",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))
		hash.InitializeGlobalCountersFromDatabase("CONTACT", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing contacts found, initializing counter to start from 1")
		hash.InitializeGlobalCountersFromDatabase("CONTACT", []string{"CONTACT00000000"}, hash.Medium)
	}

	return nil
}

// initializeUserProfileCounters initializes user profile ID counters
func (cis *CounterInitializationService) initializeUserProfileCounters(ctx context.Context) error {
	var profiles []models.UserProfile
	if err := cis.db.WithContext(ctx).Select("id").Find(&profiles).Error; err != nil {
		return fmt.Errorf("failed to fetch user profile IDs: %w", err)
	}

	var existingIDs []string
	for _, profile := range profiles {
		existingIDs = append(existingIDs, profile.ID)
	}

	if len(existingIDs) > 0 {
		cis.logger.Info("Initializing user profile counters from existing records",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))
		hash.InitializeGlobalCountersFromDatabase("USR_PROF", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing user profiles found, initializing counter to start from 1")
		hash.InitializeGlobalCountersFromDatabase("USR_PROF", []string{"USR_PROF00000000"}, hash.Medium)
	}

	return nil
}

// initializeUserRoleCounters initializes user role ID counters
func (cis *CounterInitializationService) initializeUserRoleCounters(ctx context.Context) error {
	var userRoles []models.UserRole
	if err := cis.db.WithContext(ctx).Select("id").Find(&userRoles).Error; err != nil {
		return fmt.Errorf("failed to fetch user role IDs: %w", err)
	}

	var existingIDs []string
	for _, userRole := range userRoles {
		existingIDs = append(existingIDs, userRole.ID)
	}

	if len(existingIDs) > 0 {
		cis.logger.Info("Initializing user role counters from existing records",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))
		hash.InitializeGlobalCountersFromDatabase("USR_ROL", existingIDs, hash.Small)
	} else {
		cis.logger.Info("No existing user roles found, initializing counter to start from 1")
		hash.InitializeGlobalCountersFromDatabase("USR_ROL", []string{"USR_ROL00000000"}, hash.Small)
	}

	return nil
}

// initializeGroupRoleCounters initializes group role ID counters
func (cis *CounterInitializationService) initializeGroupRoleCounters(ctx context.Context) error {
	var groupRoles []models.GroupRole
	if err := cis.db.WithContext(ctx).Select("id").Find(&groupRoles).Error; err != nil {
		return fmt.Errorf("failed to fetch group role IDs: %w", err)
	}

	var existingIDs []string
	for _, groupRole := range groupRoles {
		existingIDs = append(existingIDs, groupRole.ID)
	}

	if len(existingIDs) > 0 {
		cis.logger.Info("Initializing group role counters from existing records",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))
		hash.InitializeGlobalCountersFromDatabase("GRPR", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing group roles found, initializing counter to start from 1")
		hash.InitializeGlobalCountersFromDatabase("GRPR", []string{"GRPR00000000"}, hash.Medium)
	}

	return nil
}

// initializeGroupMembershipCounters initializes group membership ID counters
func (cis *CounterInitializationService) initializeGroupMembershipCounters(ctx context.Context) error {
	var memberships []models.GroupMembership
	if err := cis.db.WithContext(ctx).Select("id").Find(&memberships).Error; err != nil {
		return fmt.Errorf("failed to fetch group membership IDs: %w", err)
	}

	var existingIDs []string
	for _, membership := range memberships {
		existingIDs = append(existingIDs, membership.ID)
	}

	if len(existingIDs) > 0 {
		cis.logger.Info("Initializing group membership counters from existing records",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))
		hash.InitializeGlobalCountersFromDatabase("GRPM", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing group memberships found, initializing counter to start from 1")
		hash.InitializeGlobalCountersFromDatabase("GRPM", []string{"GRPM00000000"}, hash.Medium)
	}

	return nil
}

// initializeGroupInheritanceCounters initializes group inheritance ID counters
func (cis *CounterInitializationService) initializeGroupInheritanceCounters(ctx context.Context) error {
	var inheritances []models.GroupInheritance
	if err := cis.db.WithContext(ctx).Select("id").Find(&inheritances).Error; err != nil {
		return fmt.Errorf("failed to fetch group inheritance IDs: %w", err)
	}

	var existingIDs []string
	for _, inheritance := range inheritances {
		existingIDs = append(existingIDs, inheritance.ID)
	}

	if len(existingIDs) > 0 {
		cis.logger.Info("Initializing group inheritance counters from existing records",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))
		hash.InitializeGlobalCountersFromDatabase("GRPI", existingIDs, hash.Small)
	} else {
		cis.logger.Info("No existing group inheritances found, initializing counter to start from 1")
		hash.InitializeGlobalCountersFromDatabase("GRPI", []string{"GRPI00000000"}, hash.Small)
	}

	return nil
}

// initializeBindingCounters initializes binding ID counters
func (cis *CounterInitializationService) initializeBindingCounters(ctx context.Context) error {
	var bindings []models.Binding
	if err := cis.db.WithContext(ctx).Select("id").Find(&bindings).Error; err != nil {
		return fmt.Errorf("failed to fetch binding IDs: %w", err)
	}

	var existingIDs []string
	for _, binding := range bindings {
		existingIDs = append(existingIDs, binding.ID)
	}

	if len(existingIDs) > 0 {
		cis.logger.Info("Initializing binding counters from existing records",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))
		hash.InitializeGlobalCountersFromDatabase("BND", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing bindings found, initializing counter to start from 1")
		hash.InitializeGlobalCountersFromDatabase("BND", []string{"BND00000000"}, hash.Medium)
	}

	return nil
}

// initializeBindingHistoryCounters initializes binding history ID counters
func (cis *CounterInitializationService) initializeBindingHistoryCounters(ctx context.Context) error {
	var histories []models.BindingHistory
	if err := cis.db.WithContext(ctx).Select("id").Find(&histories).Error; err != nil {
		return fmt.Errorf("failed to fetch binding history IDs: %w", err)
	}

	var existingIDs []string
	for _, history := range histories {
		existingIDs = append(existingIDs, history.ID)
	}

	if len(existingIDs) > 0 {
		cis.logger.Info("Initializing binding history counters from existing records",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))
		hash.InitializeGlobalCountersFromDatabase("BNH", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing binding histories found, initializing counter to start from 1")
		hash.InitializeGlobalCountersFromDatabase("BNH", []string{"BNH00000000"}, hash.Medium)
	}

	return nil
}

// initializeAuditLogCounters initializes audit log ID counters
func (cis *CounterInitializationService) initializeAuditLogCounters(ctx context.Context) error {
	var auditLogs []models.AuditLog
	if err := cis.db.WithContext(ctx).Select("id").Find(&auditLogs).Error; err != nil {
		return fmt.Errorf("failed to fetch audit log IDs: %w", err)
	}

	var existingIDs []string
	for _, auditLog := range auditLogs {
		existingIDs = append(existingIDs, auditLog.ID)
	}

	if len(existingIDs) > 0 {
		cis.logger.Info("Initializing audit log counters from existing records",
			zap.Int("count", len(existingIDs)),
			zap.String("sample_id", existingIDs[0]))
		hash.InitializeGlobalCountersFromDatabase("AUDIT", existingIDs, hash.Medium)
	} else {
		cis.logger.Info("No existing audit logs found, initializing counter to start from 1")
		hash.InitializeGlobalCountersFromDatabase("AUDIT", []string{"AUDIT00000000"}, hash.Medium)
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
