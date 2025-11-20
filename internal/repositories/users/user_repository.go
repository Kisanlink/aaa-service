package users

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRepository handles database operations for User entities
type UserRepository struct {
	*base.BaseFilterableRepository[*models.User]
	dbManager db.DBManager
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(dbManager db.DBManager) *UserRepository {
	baseRepo := base.NewBaseFilterableRepository[*models.User]()
	baseRepo.SetDBManager(dbManager) // Connect the base repository to the actual database
	return &UserRepository{
		BaseFilterableRepository: baseRepo,
		dbManager:                dbManager,
	}
}

// GetDBManager returns the database manager
func (r *UserRepository) GetDBManager() db.DBManager {
	return r.dbManager
}

// Create creates a new user using the base repository
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return r.BaseFilterableRepository.Create(ctx, user)
}

// GetByID retrieves a user by ID with active roles preloaded
func (r *UserRepository) GetByID(ctx context.Context, id string, user *models.User) (*models.User, error) {
	// Use GetWithActiveRoles for efficient loading with preloaded roles
	return r.GetWithActiveRoles(ctx, id)
}

// Update updates an existing user using the base repository
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	return r.BaseFilterableRepository.Update(ctx, user)
}

// Delete deletes a user by ID using the base repository
func (r *UserRepository) Delete(ctx context.Context, id string, user *models.User) error {
	return r.BaseFilterableRepository.Delete(ctx, id, user)
}

// SoftDelete soft deletes a user by setting deleted_at and deleted_by fields
func (r *UserRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	// Use the database manager directly for this operation
	// Since we know we're working with the users table, we can implement this properly
	db, err := r.getDB(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// First check if the user exists and is not already deleted
	var count int64
	if err := db.WithContext(ctx).Table("users").Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("user with id %s not found", id)
	}

	// Update the user to mark as deleted
	result := db.WithContext(ctx).Table("users").Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]interface{}{
		"deleted_at": time.Now(),
		"deleted_by": deletedBy,
		"updated_at": time.Now(),
	})

	if result.Error != nil {
		return fmt.Errorf("failed to soft delete user: %w", result.Error)
	}

	return nil
}

// Restore restores a soft-deleted user using the base repository
func (r *UserRepository) Restore(ctx context.Context, id string) error {
	return r.BaseFilterableRepository.Restore(ctx, id)
}

// List retrieves active (non-deleted) users with pagination and preloaded active roles
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	// Use BaseFilterableRepository with preloading for active roles
	filter := base.NewFilterBuilder().
		WhereNull("deleted_at").                      // Only get users that are not soft-deleted
		Preload("Roles", "is_active = ?", true).      // Preload only active user roles
		Preload("Roles.Role", "is_active = ?", true). // Preload only active roles
		Limit(limit, offset).
		Build()

	users, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list users with roles: %w", err)
	}

	// Filter out any user roles where the role itself is inactive (safety check)
	for _, user := range users {
		activeUserRoles := make([]models.UserRole, 0, len(user.Roles))
		for _, userRole := range user.Roles {
			if userRole.IsActive && userRole.Role.IsActive {
				activeUserRoles = append(activeUserRoles, userRole)
			}
		}
		user.Roles = activeUserRoles
	}

	return users, nil
}

// ListAll retrieves all active (non-deleted) users with preloaded active roles
func (r *UserRepository) ListAll(ctx context.Context) ([]*models.User, error) {
	// Use a large page size to get all active users with roles preloaded
	filter := base.NewFilterBuilder().
		WhereNull("deleted_at").                      // Only get users that are not soft-deleted
		Preload("Roles", "is_active = ?", true).      // Preload only active user roles
		Preload("Roles.Role", "is_active = ?", true). // Preload only active roles
		Page(1, 1000).                                // Get up to 1000 users
		Build()

	users, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Filter out any user roles where the role itself is inactive (safety check)
	for _, user := range users {
		activeUserRoles := make([]models.UserRole, 0, len(user.Roles))
		for _, userRole := range user.Roles {
			if userRole.IsActive && userRole.Role.IsActive {
				activeUserRoles = append(activeUserRoles, userRole)
			}
		}
		user.Roles = activeUserRoles
	}

	return users, nil
}

// Count returns the total number of users using database-level counting
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	filter := base.NewFilter()
	return r.BaseFilterableRepository.Count(ctx, filter, models.User{})
}

// CountWithDeleted returns count including soft-deleted users
func (r *UserRepository) CountWithDeleted(ctx context.Context) (int64, error) {
	return r.BaseFilterableRepository.CountWithDeleted(ctx, &models.User{})
}

// Exists checks if a user exists by ID using the base repository
func (r *UserRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.Exists(ctx, id)
}

// SoftDeleteMany soft deletes multiple users using the base repository's concurrent processing
func (r *UserRepository) SoftDeleteMany(ctx context.Context, ids []string, deletedBy string) error {
	return r.BaseFilterableRepository.SoftDeleteMany(ctx, ids, deletedBy)
}

// ExistsWithDeleted checks if user exists including soft-deleted ones using the base repository
func (r *UserRepository) ExistsWithDeleted(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.ExistsWithDeleted(ctx, id)
}

// GetByCreatedBy gets users by creator using the base repository
func (r *UserRepository) GetByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*models.User, error) {
	return r.BaseFilterableRepository.GetByCreatedBy(ctx, createdBy, limit, offset)
}

// GetByUpdatedBy gets users by updater using the base repository
func (r *UserRepository) GetByUpdatedBy(ctx context.Context, updatedBy string, limit, offset int) ([]*models.User, error) {
	return r.BaseFilterableRepository.GetByUpdatedBy(ctx, updatedBy, limit, offset)
}

// GetByDeletedBy gets users by deleter using the base repository
func (r *UserRepository) GetByDeletedBy(ctx context.Context, deletedBy string, limit, offset int) ([]*models.User, error) {
	return r.BaseFilterableRepository.GetByDeletedBy(ctx, deletedBy, limit, offset)
}

// CreateMany creates multiple users using the base repository's concurrent processing
func (r *UserRepository) CreateMany(ctx context.Context, users []*models.User) error {
	// Convert []*models.User to []*models.User for base repository (same type)
	return r.BaseFilterableRepository.CreateMany(ctx, users)
}

// UpdateMany updates multiple users using the base repository's concurrent processing
func (r *UserRepository) UpdateMany(ctx context.Context, users []*models.User) error {
	// Convert []*models.User to []*models.User for base repository (same type)
	return r.BaseFilterableRepository.UpdateMany(ctx, users)
}

// DeleteMany deletes multiple users using the base repository's concurrent processing
func (r *UserRepository) DeleteMany(ctx context.Context, ids []string) error {
	return r.BaseFilterableRepository.DeleteMany(ctx, ids)
}

// getDB is a helper method to get the database connection from the database manager
func (r *UserRepository) getDB(ctx context.Context, readOnly bool) (*gorm.DB, error) {
	// Try to get the database from the database manager
	if postgresMgr, ok := r.dbManager.(interface {
		GetDB(context.Context, bool) (*gorm.DB, error)
	}); ok {
		return postgresMgr.GetDB(ctx, readOnly)
	}

	return nil, fmt.Errorf("database manager does not support GetDB method")
}

// GetByUsername retrieves an active (non-deleted) user by username with active roles preloaded
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("username", base.OpEqual, username).
		WhereNull("deleted_at"). // Only get users that are not soft-deleted
		Build()

	// Use the base repository's Find method
	users, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found with username: %s", username)
	}

	// Get the user with active roles preloaded
	return r.GetWithActiveRoles(ctx, users[0].ID)
}

// GetByPhoneNumber retrieves an active (non-deleted) user by phone number with active roles preloaded
func (r *UserRepository) GetByPhoneNumber(ctx context.Context, phoneNumber string, countryCode string) (*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("phone_number", base.OpEqual, phoneNumber).
		Where("country_code", base.OpEqual, countryCode).
		WhereNull("deleted_at"). // Only get users that are not soft-deleted
		Build()

	// Use the base repository's Find method
	users, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by phone number: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found with phone number: %s%s", countryCode, phoneNumber)
	}

	// Get the user with active roles preloaded
	return r.GetWithActiveRoles(ctx, users[0].ID)
}

// GetByMobileNumber retrieves a user by mobile number using database-level filtering
func (r *UserRepository) GetByMobileNumber(ctx context.Context, mobileNumber uint64) (*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("mobile_number", base.OpEqual, mobileNumber).
		Build()

	users, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by mobile number: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found with mobile number: %d", mobileNumber)
	}

	return users[0], nil
}

// GetByAadhaarNumber retrieves a user by Aadhaar number using database-level filtering
func (r *UserRepository) GetByAadhaarNumber(ctx context.Context, aadhaarNumber string) (*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("aadhaar_number", base.OpEqual, aadhaarNumber).
		Build()

	users, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by Aadhaar number: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found with Aadhaar number: %s", aadhaarNumber)
	}

	return users[0], nil
}

// ListActive retrieves all active users with preloaded active roles
func (r *UserRepository) ListActive(ctx context.Context, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("status", base.OpEqual, "active").
		Preload("Roles", "is_active = ?", true).      // Preload only active user roles
		Preload("Roles.Role", "is_active = ?", true). // Preload only active roles
		Limit(limit, offset).
		Build()

	users, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Filter out any user roles where the role itself is inactive (safety check)
	for _, user := range users {
		activeUserRoles := make([]models.UserRole, 0, len(user.Roles))
		for _, userRole := range user.Roles {
			if userRole.IsActive && userRole.Role.IsActive {
				activeUserRoles = append(activeUserRoles, userRole)
			}
		}
		user.Roles = activeUserRoles
	}

	return users, nil
}

// CountActive returns the total number of active users using database-level counting
func (r *UserRepository) CountActive(ctx context.Context) (int64, error) {
	filter := base.NewFilterBuilder().
		Where("status", base.OpEqual, "active").
		Build()

	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// GetWithAddress retrieves a user with their addresses using efficient preloading
func (r *UserRepository) GetWithAddress(ctx context.Context, userID string) (*models.User, error) {
	// Get the GORM DB instance from the postgres manager
	db, err := r.getDB(ctx, true) // Use read-only for efficiency
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Initialize user pointer with BaseModel to allow GORM to scan into it
	user := &models.User{
		BaseModel: &base.BaseModel{},
	}

	// Use efficient single query with preloading for all address relationships
	err = db.WithContext(ctx).
		Preload("Profile.Address").
		Preload("Contacts.Address").
		Where("id = ? AND deleted_at IS NULL", userID).
		First(user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user with id %s not found", userID)
		}
		return nil, fmt.Errorf("failed to get user with addresses: %w", err)
	}

	return user, nil
}

// GetWithProfile retrieves a user with their profile using efficient preloading
func (r *UserRepository) GetWithProfile(ctx context.Context, userID string) (*models.User, error) {
	// Get the GORM DB instance from the postgres manager
	db, err := r.getDB(ctx, true) // Use read-only for efficiency
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Initialize user pointer with BaseModel to allow GORM to scan into it
	user := &models.User{
		BaseModel: &base.BaseModel{},
	}

	// Use efficient single query with preloading for profile, address, contacts, and active roles
	err = db.WithContext(ctx).
		Preload("Profile").
		Preload("Profile.Address").
		Preload("Contacts").
		Preload("Roles", "is_active = ?", true).
		Preload("Roles.Role", "is_active = ?", true).
		Where("id = ? AND deleted_at IS NULL", userID).
		First(user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user with id %s not found", userID)
		}
		return nil, fmt.Errorf("failed to get user with profile: %w", err)
	}

	return user, nil
}

// Search searches for users by keyword in username using the BaseFilterableRepository
// with case-insensitive contains and proper pagination.
func (r *UserRepository) Search(ctx context.Context, keyword string, limit, offset int) ([]*models.User, error) {
	// Sanitize pagination
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10
	}

	// If no keyword, just list with pagination
	if strings.TrimSpace(keyword) == "" {
		return r.List(ctx, limit, offset)
	}

	// Use database-level search with BaseFilterableRepository and preload roles
	// Create a filter that searches in username and phone_number fields (only fields that exist)
	// Since we need OR logic, we'll create multiple conditions and use the Or method
	filter := base.NewFilterBuilder().
		Or(
			base.FilterCondition{Field: "username", Operator: base.OpContains, Value: keyword},
			base.FilterCondition{Field: "phone_number", Operator: base.OpContains, Value: keyword},
		).
		WhereNull("deleted_at").                      // Only get users that are not soft-deleted
		Preload("Roles", "is_active = ?", true).      // Preload only active user roles
		Preload("Roles.Role", "is_active = ?", true). // Preload only active roles
		Limit(limit, offset).
		Build()

	// Use the base repository's Find method for database-level search
	users, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// Filter out any user roles where the role itself is inactive (safety check)
	for _, user := range users {
		activeUserRoles := make([]models.UserRole, 0, len(user.Roles))
		for _, userRole := range user.Roles {
			if userRole.IsActive && userRole.Role.IsActive {
				activeUserRoles = append(activeUserRoles, userRole)
			}
		}
		user.Roles = activeUserRoles
	}

	return users, nil
}

// GetUsersWithRelationships efficiently loads multiple users with their relationships using goroutines
func (r *UserRepository) GetUsersWithRelationships(ctx context.Context, userIDs []string, includeRoles, includeProfile, includeAddresses bool) ([]*models.User, error) {
	if len(userIDs) == 0 {
		return []*models.User{}, nil
	}

	// Get the GORM DB instance from the postgres manager
	db, err := r.dbManager.(*db.PostgresManager).GetDB(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Load users in goroutine
	usersChan := make(chan []*models.User, 1)
	usersErrChan := make(chan error, 1)

	go func() {
		var users []models.User
		query := db.Where("id IN ?", userIDs)

		// Add preloads based on requested relationships
		if includeRoles {
			query = query.Preload("Roles.Role.Permissions")
		}
		if includeProfile {
			query = query.Preload("Profile")
			if includeAddresses {
				query = query.Preload("Profile.Address")
			}
		}
		if includeAddresses {
			query = query.Preload("Contacts.Address")
		}

		err := query.Find(&users).Error
		if err != nil {
			usersErrChan <- err
			return
		}

		// Convert to pointers
		userPtrs := make([]*models.User, len(users))
		for i := range users {
			userPtrs[i] = &users[i]
		}

		usersChan <- userPtrs
	}()

	// Wait for results
	select {
	case users := <-usersChan:
		return users, nil
	case err := <-usersErrChan:
		return nil, fmt.Errorf("failed to load users with relationships: %w", err)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetUserStats retrieves various user statistics using BaseFilterableRepository.Count() methods
func (r *UserRepository) GetUserStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total users count (including soft-deleted for historical stats)
	totalFilter := base.NewFilter()
	total, err := r.BaseFilterableRepository.Count(ctx, totalFilter, models.User{})
	if err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}
	stats["total"] = total

	// Active users count
	activeFilter := base.NewFilterBuilder().
		Where("status", base.OpEqual, "active").
		Build()
	active, err := r.BaseFilterableRepository.CountWithFilter(ctx, activeFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}
	stats["active"] = active

	// Pending users count
	pendingFilter := base.NewFilterBuilder().
		Where("status", base.OpEqual, "pending").
		Build()
	pending, err := r.BaseFilterableRepository.CountWithFilter(ctx, pendingFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending users: %w", err)
	}
	stats["pending"] = pending

	// Validated users count
	validatedFilter := base.NewFilterBuilder().
		Where("is_validated", base.OpEqual, true).
		Build()
	validated, err := r.BaseFilterableRepository.CountWithFilter(ctx, validatedFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to count validated users: %w", err)
	}
	stats["validated"] = validated

	return stats, nil
}

// BulkValidateUsers concurrently validates multiple users using goroutines
func (r *UserRepository) BulkValidateUsers(ctx context.Context, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	// Get the GORM DB instance from the postgres manager
	db, err := r.dbManager.(*db.PostgresManager).GetDB(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Use worker pool pattern for concurrent validation
	const maxWorkers = 10
	workerCount := min(maxWorkers, len(userIDs))

	// Create channels for coordination
	jobs := make(chan string, len(userIDs))
	results := make(chan error, len(userIDs))

	// Start workers
	for i := 0; i < workerCount; i++ {
		go func() {
			for userID := range jobs {
				// Update user validation status
				err := db.Model(&models.User{}).
					Where("id = ?", userID).
					Updates(map[string]interface{}{
						"is_validated": true,
						"status":       "active",
					}).Error
				results <- err
			}
		}()
	}

	// Send jobs
	for _, userID := range userIDs {
		jobs <- userID
	}
	close(jobs)

	// Collect results
	for i := 0; i < len(userIDs); i++ {
		if err := <-results; err != nil {
			return fmt.Errorf("failed to validate user in batch: %w", err)
		}
	}

	return nil
}

// GetByEmail retrieves a user by email using database-level filtering
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	// Get database connection
	db, err := r.getDB(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Search for user by email in contacts table
	var contact models.Contact
	err = db.WithContext(ctx).
		Where("type = ? AND value = ?", "email", email).
		First(&contact).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
		return nil, fmt.Errorf("failed to search email in contacts: %w", err)
	}

	// Get the user by ID - Initialize user pointer with BaseModel
	user := &models.User{
		BaseModel: &base.BaseModel{},
	}
	if err := r.dbManager.GetByID(ctx, contact.UserID, user); err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByStatus retrieves users by status using database-level filtering
func (r *UserRepository) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("status", base.OpEqual, status).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByValidationStatus retrieves users by validation status using database-level filtering
func (r *UserRepository) GetByValidationStatus(ctx context.Context, isValidated bool, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("is_validated", base.OpEqual, isValidated).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByDateRange retrieves users created within a date range using database-level filtering
func (r *UserRepository) GetByDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUpdatedDateRange retrieves users updated within a date range using database-level filtering
func (r *UserRepository) GetByUpdatedDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		WhereBetween("updated_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByDeletedDateRange retrieves users deleted within a date range using database-level filtering
func (r *UserRepository) GetByDeletedDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		WhereBetween("deleted_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUsernameAndStatus retrieves users by username and status using database-level filtering
func (r *UserRepository) GetByUsernameAndStatus(ctx context.Context, username, status string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("username", base.OpEqual, username).
		Where("status", base.OpEqual, status).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUsernameAndValidationStatus retrieves users by username and validation status using database-level filtering
func (r *UserRepository) GetByUsernameAndValidationStatus(ctx context.Context, username string, isValidated bool, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("username", base.OpEqual, username).
		Where("is_validated", base.OpEqual, isValidated).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUsernameAndDateRange retrieves users by username and date range using database-level filtering
func (r *UserRepository) GetByUsernameAndDateRange(ctx context.Context, username, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("username", base.OpEqual, username).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUsernameAndUpdatedDateRange retrieves users by username and updated date range using database-level filtering
func (r *UserRepository) GetByUsernameAndUpdatedDateRange(ctx context.Context, username, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("username", base.OpEqual, username).
		WhereBetween("updated_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUsernameAndDeletedDateRange retrieves users by username and deleted date range using database-level filtering
func (r *UserRepository) GetByUsernameAndDeletedDateRange(ctx context.Context, username, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("username", base.OpEqual, username).
		WhereBetween("deleted_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByStatusAndValidationStatus retrieves users by status and validation status using database-level filtering
func (r *UserRepository) GetByStatusAndValidationStatus(ctx context.Context, status string, isValidated bool, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("status", base.OpEqual, status).
		Where("is_validated", base.OpEqual, isValidated).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByStatusAndDateRange retrieves users by status and date range using database-level filtering
func (r *UserRepository) GetByStatusAndDateRange(ctx context.Context, status, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("status", base.OpEqual, status).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByStatusAndUpdatedDateRange retrieves users by status and updated date range using database-level filtering
func (r *UserRepository) GetByStatusAndUpdatedDateRange(ctx context.Context, status, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("status", base.OpEqual, status).
		WhereBetween("updated_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByStatusAndDeletedDateRange retrieves users by status and deleted date range using database-level filtering
func (r *UserRepository) GetByStatusAndDeletedDateRange(ctx context.Context, status, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("status", base.OpEqual, status).
		WhereBetween("deleted_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByValidationStatusAndDateRange retrieves users by validation status and date range using database-level filtering
func (r *UserRepository) GetByValidationStatusAndDateRange(ctx context.Context, isValidated bool, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("is_validated", base.OpEqual, isValidated).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByValidationStatusAndUpdatedDateRange retrieves users by validation status and updated date range using database-level filtering
func (r *UserRepository) GetByValidationStatusAndUpdatedDateRange(ctx context.Context, isValidated bool, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("is_validated", base.OpEqual, isValidated).
		WhereBetween("updated_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByValidationStatusAndDeletedDateRange retrieves users by validation status and deleted date range using database-level filtering
func (r *UserRepository) GetByValidationStatusAndDeletedDateRange(ctx context.Context, isValidated bool, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("is_validated", base.OpEqual, isValidated).
		WhereBetween("deleted_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// SoftDeleteWithCascade performs a soft delete of a user with proper cascade operations for related entities
// This method handles role cleanup, contact deactivation, and profile archiving in a transaction
func (r *UserRepository) SoftDeleteWithCascade(ctx context.Context, userID, deletedBy string) error {
	// Get the GORM DB instance for transaction support
	db, err := r.getDB(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Use transaction to ensure all cascade operations succeed or fail together
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First check if the user exists and is not already deleted
		var count int64
		if err := tx.Table("users").Where("id = ? AND deleted_at IS NULL", userID).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check user existence: %w", err)
		}

		if count == 0 {
			return fmt.Errorf("user with id %s not found or already deleted", userID)
		}

		now := time.Now()

		// 1. Deactivate all user role assignments
		if err := tx.Table("user_roles").
			Where("user_id = ? AND is_active = ?", userID, true).
			Updates(map[string]interface{}{
				"is_active":  false,
				"updated_at": now,
				"deleted_by": deletedBy,
			}).Error; err != nil {
			return fmt.Errorf("failed to deactivate user roles: %w", err)
		}

		// 2. Soft delete user contacts
		if err := tx.Table("contacts").
			Where("user_id = ? AND deleted_at IS NULL", userID).
			Updates(map[string]interface{}{
				"deleted_at": now,
				"deleted_by": deletedBy,
				"updated_at": now,
			}).Error; err != nil {
			return fmt.Errorf("failed to soft delete user contacts: %w", err)
		}

		// 3. Soft delete user profile
		if err := tx.Table("user_profiles").
			Where("user_id = ? AND deleted_at IS NULL", userID).
			Updates(map[string]interface{}{
				"deleted_at": now,
				"deleted_by": deletedBy,
				"updated_at": now,
			}).Error; err != nil {
			return fmt.Errorf("failed to soft delete user profile: %w", err)
		}

		// 4. Finally, soft delete the user
		if err := tx.Table("users").
			Where("id = ? AND deleted_at IS NULL", userID).
			Updates(map[string]interface{}{
				"deleted_at": now,
				"deleted_by": deletedBy,
				"updated_at": now,
			}).Error; err != nil {
			return fmt.Errorf("failed to soft delete user: %w", err)
		}

		return nil
	})
}

// GetWithActiveRoles retrieves a user with their active roles and role details efficiently loaded
// This method uses optimized queries with proper preloading to minimize database round trips
func (r *UserRepository) GetWithActiveRoles(ctx context.Context, userID string) (*models.User, error) {
	// Get the GORM DB instance for complex queries with preloading
	db, err := r.getDB(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Initialize user pointer with BaseModel to allow GORM to scan into it
	user := &models.User{
		BaseModel: &base.BaseModel{},
	}

	// Use a single query with preloading to efficiently load user and active roles
	err = db.WithContext(ctx).
		Preload("Roles", "is_active = ?", true).
		Preload("Roles.Role", "is_active = ?", true).
		Where("id = ? AND deleted_at IS NULL", userID).
		First(user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user with id %s not found", userID)
		}
		return nil, fmt.Errorf("failed to get user with active roles: %w", err)
	}

	// Filter out any user roles where the role itself is inactive
	// This is a safety check in case the preload condition doesn't work as expected
	activeUserRoles := make([]models.UserRole, 0, len(user.Roles))
	for _, userRole := range user.Roles {
		if userRole.IsActive && userRole.Role.IsActive {
			activeUserRoles = append(activeUserRoles, userRole)
		}
	}
	user.Roles = activeUserRoles

	return user, nil
}

// VerifyMPin verifies a user's MPIN against the stored hash
// This method uses secure comparison and proper error handling for authentication
func (r *UserRepository) VerifyMPin(ctx context.Context, userID, plainMPin string) error {
	// Get the GORM DB instance for direct query
	db, err := r.getDB(ctx, true) // Use read-only connection for verification
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Initialize user pointer with BaseModel to allow GORM to scan into it
	user := &models.User{
		BaseModel: &base.BaseModel{},
	}
	err = db.WithContext(ctx).
		Select("id, m_pin").
		Where("id = ? AND deleted_at IS NULL", userID).
		First(user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("user with id %s not found", userID)
		}
		return fmt.Errorf("failed to get user for MPIN verification: %w", err)
	}

	// Check if user has MPIN set
	if !user.HasMPin() {
		return fmt.Errorf("user does not have MPIN set")
	}

	// Use bcrypt to compare the plain MPIN with the stored hash
	// Note: This assumes MPIN is stored as a bcrypt hash like passwords
	// If a different hashing method is used, this should be updated accordingly
	if err := bcrypt.CompareHashAndPassword([]byte(*user.MPin), []byte(plainMPin)); err != nil {
		return fmt.Errorf("invalid MPIN")
	}

	return nil
}

// GetByUsernameAndStatusAndValidationStatus retrieves users by username, status and validation status using database-level filtering
func (r *UserRepository) GetByUsernameAndStatusAndValidationStatus(ctx context.Context, username, status string, isValidated bool, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("username", base.OpEqual, username).
		Where("status", base.OpEqual, status).
		Where("is_validated", base.OpEqual, isValidated).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUsernameAndStatusAndDateRange retrieves users by username, status and date range using database-level filtering
func (r *UserRepository) GetByUsernameAndStatusAndDateRange(ctx context.Context, username, status, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("username", base.OpEqual, username).
		Where("status", base.OpEqual, status).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUsernameAndStatusAndUpdatedDateRange retrieves users by username, status and updated date range using database-level filtering
func (r *UserRepository) GetByUsernameAndStatusAndUpdatedDateRange(ctx context.Context, username, status, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("username", base.OpEqual, username).
		Where("status", base.OpEqual, status).
		WhereBetween("updated_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUsernameAndStatusAndDeletedDateRange retrieves users by username, status and deleted date range using database-level filtering
func (r *UserRepository) GetByUsernameAndStatusAndDeletedDateRange(ctx context.Context, username, status, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("username", base.OpEqual, username).
		Where("status", base.OpEqual, status).
		WhereBetween("deleted_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUsernameAndValidationStatusAndDateRange retrieves users by username, validation status and date range using database-level filtering
func (r *UserRepository) GetByUsernameAndValidationStatusAndDateRange(ctx context.Context, username string, isValidated bool, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("username", base.OpEqual, username).
		Where("is_validated", base.OpEqual, isValidated).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUsernameAndValidationStatusAndUpdatedDateRange retrieves users by username, validation status and updated date range using database-level filtering
func (r *UserRepository) GetByUsernameAndValidationStatusAndUpdatedDateRange(ctx context.Context, username string, isValidated bool, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("username", base.OpEqual, username).
		Where("is_validated", base.OpEqual, isValidated).
		WhereBetween("updated_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUsernameAndValidationStatusAndDeletedDateRange retrieves users by username, validation status and deleted date range using database-level filtering
func (r *UserRepository) GetByUsernameAndValidationStatusAndDeletedDateRange(ctx context.Context, username string, isValidated bool, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("username", base.OpEqual, username).
		Where("is_validated", base.OpEqual, isValidated).
		WhereBetween("deleted_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByStatusAndValidationStatusAndDateRange retrieves users by status, validation status and date range using database-level filtering
func (r *UserRepository) GetByStatusAndValidationStatusAndDateRange(ctx context.Context, status string, isValidated bool, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("status", base.OpEqual, status).
		Where("is_validated", base.OpEqual, isValidated).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByStatusAndValidationStatusAndUpdatedDateRange retrieves users by status, validation status and updated date range using database-level filtering
func (r *UserRepository) GetByStatusAndValidationStatusAndUpdatedDateRange(ctx context.Context, status string, isValidated bool, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("status", base.OpEqual, status).
		Where("is_validated", base.OpEqual, isValidated).
		WhereBetween("updated_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByStatusAndValidationStatusAndDeletedDateRange retrieves users by status, validation status and deleted date range using database-level filtering
func (r *UserRepository) GetByStatusAndValidationStatusAndDeletedDateRange(ctx context.Context, status string, isValidated bool, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("status", base.OpEqual, status).
		Where("is_validated", base.OpEqual, isValidated).
		WhereBetween("deleted_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUsernameAndStatusAndValidationStatusAndDateRange retrieves users by username, status, validation status and date range using database-level filtering
func (r *UserRepository) GetByUsernameAndStatusAndValidationStatusAndDateRange(ctx context.Context, username, status string, isValidated bool, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("username", base.OpEqual, username).
		Where("status", base.OpEqual, status).
		Where("is_validated", base.OpEqual, isValidated).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUsernameAndStatusAndValidationStatusAndUpdatedDateRange retrieves users by username, status, validation status and updated date range using database-level filtering
func (r *UserRepository) GetByUsernameAndStatusAndValidationStatusAndUpdatedDateRange(ctx context.Context, username, status string, isValidated bool, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("username", base.OpEqual, username).
		Where("status", base.OpEqual, status).
		Where("is_validated", base.OpEqual, isValidated).
		WhereBetween("updated_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUsernameAndStatusAndValidationStatusAndDeletedDateRange retrieves users by username, status, validation status and deleted date range using database-level filtering
func (r *UserRepository) GetByUsernameAndStatusAndValidationStatusAndDeletedDateRange(ctx context.Context, username, status string, isValidated bool, startDate, endDate string, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("username", base.OpEqual, username).
		Where("status", base.OpEqual, status).
		Where("is_validated", base.OpEqual, isValidated).
		WhereBetween("deleted_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}
