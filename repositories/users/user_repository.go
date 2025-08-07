package users

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/entities/models"
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
	return &UserRepository{
		BaseFilterableRepository: base.NewBaseFilterableRepository[*models.User](),
		dbManager:                dbManager,
	}
}

// Create creates a new user using the base repository
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return r.BaseFilterableRepository.Create(ctx, user)
}

// GetByID retrieves a user by ID using the base repository
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	return r.BaseFilterableRepository.GetByID(ctx, id)
}

// Update updates an existing user using the base repository
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	return r.BaseFilterableRepository.Update(ctx, user)
}

// Delete deletes a user by ID using the base repository
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	return r.BaseFilterableRepository.Delete(ctx, id)
}

// List retrieves users with pagination using the base repository
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return r.BaseFilterableRepository.List(ctx, limit, offset)
}

// Count returns the total number of users using the base repository
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	return r.BaseFilterableRepository.Count(ctx)
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

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("username", db.FilterOpEqual, username),
	}

	var users []models.User
	if err := r.dbManager.List(ctx, filters, &users); err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found with username: %s", username)
	}

	return &users[0], nil
}

// GetByPhoneNumber retrieves a user by phone number
func (r *UserRepository) GetByPhoneNumber(ctx context.Context, phoneNumber string, countryCode string) (*models.User, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("phone_number", db.FilterOpEqual, phoneNumber),
		r.dbManager.BuildFilter("country_code", db.FilterOpEqual, countryCode),
	}

	var users []models.User
	if err := r.dbManager.List(ctx, filters, &users); err != nil {
		return nil, fmt.Errorf("failed to get user by phone number: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found with phone number: %s%s", countryCode, phoneNumber)
	}

	return &users[0], nil
}

// GetByMobileNumber retrieves a user by mobile number
func (r *UserRepository) GetByMobileNumber(ctx context.Context, mobileNumber uint64) (*models.User, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("mobile_number", db.FilterOpEqual, mobileNumber),
	}

	var users []models.User
	if err := r.dbManager.List(ctx, filters, &users); err != nil {
		return nil, fmt.Errorf("failed to get user by mobile number: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found with mobile number: %d", mobileNumber)
	}

	return &users[0], nil
}

// GetByAadhaarNumber retrieves a user by Aadhaar number
func (r *UserRepository) GetByAadhaarNumber(ctx context.Context, aadhaarNumber string) (*models.User, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("aadhaar_number", db.FilterOpEqual, aadhaarNumber),
	}

	var users []models.User
	if err := r.dbManager.List(ctx, filters, &users); err != nil {
		return nil, fmt.Errorf("failed to get user by Aadhaar number: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found with Aadhaar number: %s", aadhaarNumber)
	}

	return &users[0], nil
}

// ListActive retrieves all active users using the base repository
func (r *UserRepository) ListActive(ctx context.Context, limit, offset int) ([]*models.User, error) {
	// Get all users from base repository
	allUsers, err := r.BaseFilterableRepository.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Filter for active users
	var activeUsers []*models.User
	for _, user := range allUsers {
		if user.Status != nil && *user.Status == "active" {
			activeUsers = append(activeUsers, user)
		}
	}

	return activeUsers, nil
}

// CountActive returns the total number of active users using the base repository
func (r *UserRepository) CountActive(ctx context.Context) (int64, error) {
	// Get all users from base repository
	allUsers, err := r.BaseFilterableRepository.List(ctx, 1000, 0) // Get all users for counting
	if err != nil {
		return 0, fmt.Errorf("failed to list users: %w", err)
	}

	// Count active users
	var count int64
	for _, user := range allUsers {
		if user.Status != nil && *user.Status == "active" {
			count++
		}
	}

	return count, nil
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

// Search searches users by keyword in name, username, or mobile number
// Search searches for users by keyword in username using the base repository
func (r *UserRepository) Search(ctx context.Context, keyword string, limit, offset int) ([]*models.User, error) {
	// Get all users from base repository
	allUsers, err := r.BaseFilterableRepository.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Filter for users matching the keyword
	var matchingUsers []*models.User
	for _, user := range allUsers {
		if user.Username != nil && strings.Contains(strings.ToLower(*user.Username), strings.ToLower(keyword)) {
			matchingUsers = append(matchingUsers, user)
		}
	}

	return matchingUsers, nil
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
