package users

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
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

// Create creates a new user using the base repository
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return r.BaseFilterableRepository.Create(ctx, user)
}

// GetByID retrieves a user by ID using the base repository
func (r *UserRepository) GetByID(ctx context.Context, id string, user *models.User) (*models.User, error) {
	return r.BaseFilterableRepository.GetByID(ctx, id, user)
}

// Update updates an existing user using the base repository
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	return r.BaseFilterableRepository.Update(ctx, user)
}

// Delete deletes a user by ID using the base repository
func (r *UserRepository) Delete(ctx context.Context, id string, user *models.User) error {
	return r.BaseFilterableRepository.Delete(ctx, id, user)
}

// List retrieves users with pagination using database-level filtering
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	// Use base filterable repository for optimized database-level filtering
	filter := base.NewFilterBuilder().
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Count returns the total number of users using database-level counting
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	filter := base.NewFilter()
	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// Exists checks if a user exists by ID using the base repository
func (r *UserRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.Exists(ctx, id)
}

// SoftDelete soft deletes a user by ID using the base repository
func (r *UserRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	return r.BaseFilterableRepository.SoftDelete(ctx, id, deletedBy)
}

// Restore restores a soft-deleted user using the base repository
func (r *UserRepository) Restore(ctx context.Context, id string) error {
	return r.BaseFilterableRepository.Restore(ctx, id)
}

// ListWithDeleted retrieves users including soft-deleted ones using the base repository
func (r *UserRepository) ListWithDeleted(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return r.BaseFilterableRepository.ListWithDeleted(ctx, limit, offset)
}

// CountWithDeleted returns count including soft-deleted users using the base repository
func (r *UserRepository) CountWithDeleted(ctx context.Context) (int64, error) {
	return r.BaseFilterableRepository.CountWithDeleted(ctx)
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

// SoftDeleteMany soft deletes multiple users using the base repository's concurrent processing
func (r *UserRepository) SoftDeleteMany(ctx context.Context, ids []string, deletedBy string) error {
	return r.BaseFilterableRepository.SoftDeleteMany(ctx, ids, deletedBy)
}

// GetByUsername retrieves a user by username using kisanlink-db filters
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	filters := []base.FilterCondition{
		{Field: "username", Operator: base.OpEqual, Value: username},
	}

	var users []models.User
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: filters,
			Logic:      base.LogicAnd,
		},
	}
	if err := r.dbManager.List(ctx, filter, &users); err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found with username: %s", username)
	}

	return &users[0], nil
}

// GetByPhoneNumber retrieves a user by phone number using kisanlink-db filters
func (r *UserRepository) GetByPhoneNumber(ctx context.Context, phoneNumber string, countryCode string) (*models.User, error) {
	filters := []base.FilterCondition{
		{Field: "phone_number", Operator: base.OpEqual, Value: phoneNumber},
		{Field: "country_code", Operator: base.OpEqual, Value: countryCode},
	}

	var users []models.User
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: filters,
			Logic:      base.LogicAnd,
		},
	}
	if err := r.dbManager.List(ctx, filter, &users); err != nil {
		return nil, fmt.Errorf("failed to get user by phone number: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found with phone number: %s%s", countryCode, phoneNumber)
	}

	return &users[0], nil
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

// ListActive retrieves all active users using database-level filtering
func (r *UserRepository) ListActive(ctx context.Context, limit, offset int) ([]*models.User, error) {
	filter := base.NewFilterBuilder().
		Where("status", base.OpEqual, "active").
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// CountActive returns the total number of active users using database-level counting
func (r *UserRepository) CountActive(ctx context.Context) (int64, error) {
	filter := base.NewFilterBuilder().
		Where("status", base.OpEqual, "active").
		Build()

	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// GetWithRoles retrieves a user with their roles using proper joins with goroutines
func (r *UserRepository) GetWithRoles(ctx context.Context, userID string) (*models.User, error) {
	var user models.User

	// Get the GORM DB instance from the postgres manager
	db, err := r.dbManager.(*db.PostgresManager).GetDB(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Use goroutine to load user data concurrently with role data
	userChan := make(chan *models.User, 1)
	rolesChan := make(chan error, 1)

	// Load user data in goroutine
	go func() {
		var userData models.User
		err := db.Where("id = ?", userID).First(&userData).Error
		if err != nil {
			userChan <- nil
			return
		}
		userChan <- &userData
	}()

	// Load roles data in goroutine
	go func() {
		var roles []models.UserRole
		err := db.Preload("Role.Permissions").
			Where("user_id = ? AND is_active = ?", userID, true).
			Find(&roles).Error
		rolesChan <- err
	}()

	// Wait for user data
	userData := <-userChan
	if userData == nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	user = *userData

	// Wait for roles data
	if err := <-rolesChan; err != nil {
		return nil, fmt.Errorf("failed to load user roles: %w", err)
	}

	// Attach roles to user (this would need to be implemented based on your model structure)
	// For now, we'll use the preload approach as it's more efficient for this use case
	err = db.Preload("Roles.Role.Permissions").
		Preload("Roles", "is_active = ?", true).
		Where("id = ?", userID).
		First(&user).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user with roles: %w", err)
	}

	return &user, nil
}

// GetWithAddress retrieves a user with their addresses using proper joins with goroutines
func (r *UserRepository) GetWithAddress(ctx context.Context, userID string) (*models.User, error) {
	var user models.User

	// Get the GORM DB instance from the postgres manager
	db, err := r.dbManager.(*db.PostgresManager).GetDB(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Use goroutines to load different address types concurrently
	userChan := make(chan *models.User, 1)
	profileAddrChan := make(chan error, 1)
	contactAddrChan := make(chan error, 1)

	// Load user data in goroutine
	go func() {
		var userData models.User
		err := db.Where("id = ?", userID).First(&userData).Error
		if err != nil {
			userChan <- nil
			return
		}
		userChan <- &userData
	}()

	// Load profile address in goroutine
	go func() {
		var profile models.UserProfile
		err := db.Preload("Address").
			Where("user_id = ?", userID).
			First(&profile).Error
		profileAddrChan <- err
	}()

	// Load contact addresses in goroutine
	go func() {
		var contacts []models.Contact
		err := db.Preload("Address").
			Where("user_id = ?", userID).
			Find(&contacts).Error
		contactAddrChan <- err
	}()

	// Wait for user data
	userData := <-userChan
	if userData == nil {
		return nil, fmt.Errorf("failed to get user")
	}
	user = *userData

	// Wait for address data
	if err := <-profileAddrChan; err != nil {
		return nil, fmt.Errorf("failed to load profile address: %w", err)
	}

	if err := <-contactAddrChan; err != nil {
		return nil, fmt.Errorf("failed to load contact addresses: %w", err)
	}

	// Use preload for efficiency
	err = db.Preload("Profile.Address").
		Preload("Contacts.Address").
		Where("id = ?", userID).
		First(&user).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user with addresses: %w", err)
	}

	return &user, nil
}

// GetWithProfile retrieves a user with their profile using proper joins with goroutines
func (r *UserRepository) GetWithProfile(ctx context.Context, userID string) (*models.User, error) {
	var user models.User

	// Get the GORM DB instance from the postgres manager
	db, err := r.dbManager.(*db.PostgresManager).GetDB(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Use goroutines to load user and profile data concurrently
	userChan := make(chan *models.User, 1)
	profileChan := make(chan error, 1)

	// Load user data in goroutine
	go func() {
		var userData models.User
		err := db.Where("id = ?", userID).First(&userData).Error
		if err != nil {
			userChan <- nil
			return
		}
		userChan <- &userData
	}()

	// Load profile data in goroutine
	go func() {
		var profile models.UserProfile
		err := db.Preload("Address").
			Where("user_id = ?", userID).
			First(&profile).Error
		profileChan <- err
	}()

	// Wait for user data
	userData := <-userChan
	if userData == nil {
		return nil, fmt.Errorf("failed to get user")
	}
	user = *userData

	// Wait for profile data
	if err := <-profileChan; err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	// Use preload for efficiency
	err = db.Preload("Profile").
		Preload("Profile.Address").
		Where("id = ?", userID).
		First(&user).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user with profile: %w", err)
	}

	return &user, nil
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

	// Build filter for case-insensitive contains and pagination
	filter := base.NewFilterBuilder().
		Where("username", base.OpContains, keyword).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
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

// GetUserStats concurrently loads various user statistics using goroutines
func (r *UserRepository) GetUserStats(ctx context.Context) (map[string]int64, error) {
	// Get the GORM DB instance from the postgres manager
	db, err := r.dbManager.(*db.PostgresManager).GetDB(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Create channels for different stats
	totalChan := make(chan int64, 1)
	activeChan := make(chan int64, 1)
	pendingChan := make(chan int64, 1)
	validatedChan := make(chan int64, 1)

	// Load total users count
	go func() {
		var count int64
		err := db.Model(&models.User{}).Count(&count).Error
		if err != nil {
			count = 0
		}
		totalChan <- count
	}()

	// Load active users count
	go func() {
		var count int64
		err := db.Model(&models.User{}).Where("status = ?", "active").Count(&count).Error
		if err != nil {
			count = 0
		}
		activeChan <- count
	}()

	// Load pending users count
	go func() {
		var count int64
		err := db.Model(&models.User{}).Where("status = ?", "pending").Count(&count).Error
		if err != nil {
			count = 0
		}
		pendingChan <- count
	}()

	// Load validated users count
	go func() {
		var count int64
		err := db.Model(&models.User{}).Where("is_validated = ?", true).Count(&count).Error
		if err != nil {
			count = 0
		}
		validatedChan <- count
	}()

	// Collect results
	stats := make(map[string]int64)
	stats["total"] = <-totalChan
	stats["active"] = <-activeChan
	stats["pending"] = <-pendingChan
	stats["validated"] = <-validatedChan

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
	filter := base.NewFilterBuilder().
		Where("email", base.OpEqual, email).
		Build()

	users, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found with email: %s", email)
	}

	return users[0], nil
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
