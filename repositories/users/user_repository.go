package users

import (
	"context"
	"fmt"
	"time"

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

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if err := user.BeforeCreate(); err != nil {
		return fmt.Errorf("failed to prepare user for creation: %w", err)
	}

	return r.dbManager.Create(ctx, user)
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	if err := r.dbManager.GetByID(ctx, id, &user); err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	if err := user.BeforeUpdate(); err != nil {
		return fmt.Errorf("failed to prepare user for update: %w", err)
	}

	return r.dbManager.Update(ctx, user)
}

// Delete deletes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	return r.dbManager.Delete(ctx, id)
}

// List retrieves users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	var users []models.User

	if err := r.dbManager.List(ctx, []db.Filter{}, &users); err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Convert []models.User to []*models.User
	result := make([]*models.User, len(users))
	for i := range users {
		result[i] = &users[i]
	}

	return result, nil
}

// Count returns the total number of users
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var users []models.User
	if err := r.dbManager.List(ctx, []db.Filter{}, &users); err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return int64(len(users)), nil
}

// Exists checks if a user exists by ID
func (r *UserRepository) Exists(ctx context.Context, id string) (bool, error) {
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return false, nil // User doesn't exist
	}
	return true, nil
}

// SoftDelete soft deletes a user by ID
func (r *UserRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	user, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user for soft delete: %w", err)
	}

	if err := user.BeforeSoftDelete(); err != nil {
		return fmt.Errorf("failed to prepare user for soft delete: %w", err)
	}

	// Set deleted fields
	now := time.Now()
	user.DeletedAt = &now
	user.DeletedBy = &deletedBy

	return r.dbManager.Update(ctx, user)
}

// Restore restores a soft-deleted user
func (r *UserRepository) Restore(ctx context.Context, id string) error {
	user, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user for restore: %w", err)
	}

	// Clear deleted fields
	user.DeletedAt = nil
	user.DeletedBy = nil

	return r.dbManager.Update(ctx, user)
}

// ListWithDeleted retrieves users including soft-deleted ones
func (r *UserRepository) ListWithDeleted(ctx context.Context, limit, offset int) ([]*models.User, error) {
	// For now, same as List since we don't have soft delete filtering in our mock
	return r.List(ctx, limit, offset)
}

// CountWithDeleted returns count including soft-deleted users
func (r *UserRepository) CountWithDeleted(ctx context.Context) (int64, error) {
	// For now, same as Count since we don't have soft delete filtering in our mock
	return r.Count(ctx)
}

// ExistsWithDeleted checks if user exists including soft-deleted ones
func (r *UserRepository) ExistsWithDeleted(ctx context.Context, id string) (bool, error) {
	// For now, same as Exists since we don't have soft delete filtering in our mock
	return r.Exists(ctx, id)
}

// GetByCreatedBy gets users by creator
func (r *UserRepository) GetByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*models.User, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("created_by", db.FilterOpEqual, createdBy),
	}

	var users []models.User
	if err := r.dbManager.List(ctx, filters, &users); err != nil {
		return nil, fmt.Errorf("failed to get users by created_by: %w", err)
	}

	result := make([]*models.User, len(users))
	for i := range users {
		result[i] = &users[i]
	}

	return result, nil
}

// GetByUpdatedBy gets users by updater
func (r *UserRepository) GetByUpdatedBy(ctx context.Context, updatedBy string, limit, offset int) ([]*models.User, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("updated_by", db.FilterOpEqual, updatedBy),
	}

	var users []models.User
	if err := r.dbManager.List(ctx, filters, &users); err != nil {
		return nil, fmt.Errorf("failed to get users by updated_by: %w", err)
	}

	result := make([]*models.User, len(users))
	for i := range users {
		result[i] = &users[i]
	}

	return result, nil
}

// GetByDeletedBy gets users by deleter
func (r *UserRepository) GetByDeletedBy(ctx context.Context, deletedBy string, limit, offset int) ([]*models.User, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("deleted_by", db.FilterOpEqual, deletedBy),
	}

	var users []models.User
	if err := r.dbManager.List(ctx, filters, &users); err != nil {
		return nil, fmt.Errorf("failed to get users by deleted_by: %w", err)
	}

	result := make([]*models.User, len(users))
	for i := range users {
		result[i] = &users[i]
	}

	return result, nil
}

// CreateMany creates multiple users
func (r *UserRepository) CreateMany(ctx context.Context, users []*models.User) error {
	for _, user := range users {
		if err := r.Create(ctx, user); err != nil {
			return fmt.Errorf("failed to create user in batch: %w", err)
		}
	}
	return nil
}

// UpdateMany updates multiple users
func (r *UserRepository) UpdateMany(ctx context.Context, users []*models.User) error {
	for _, user := range users {
		if err := r.Update(ctx, user); err != nil {
			return fmt.Errorf("failed to update user in batch: %w", err)
		}
	}
	return nil
}

// DeleteMany deletes multiple users
func (r *UserRepository) DeleteMany(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := r.Delete(ctx, id); err != nil {
			return fmt.Errorf("failed to delete user in batch: %w", err)
		}
	}
	return nil
}

// SoftDeleteMany soft deletes multiple users
func (r *UserRepository) SoftDeleteMany(ctx context.Context, ids []string, deletedBy string) error {
	for _, id := range ids {
		if err := r.SoftDelete(ctx, id, deletedBy); err != nil {
			return fmt.Errorf("failed to soft delete user in batch: %w", err)
		}
	}
	return nil
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

// ListActive retrieves all active users
func (r *UserRepository) ListActive(ctx context.Context, limit, offset int) ([]*models.User, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("status", db.FilterOpEqual, "active"),
	}

	var users []models.User
	if err := r.dbManager.List(ctx, filters, &users); err != nil {
		return nil, fmt.Errorf("failed to list active users: %w", err)
	}

	// Convert []models.User to []*models.User
	result := make([]*models.User, len(users))
	for i := range users {
		result[i] = &users[i]
	}

	return result, nil
}

// CountActive returns the total number of active users
func (r *UserRepository) CountActive(ctx context.Context) (int64, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("status", db.FilterOpEqual, "active"),
	}

	var users []models.User
	if err := r.dbManager.List(ctx, filters, &users); err != nil {
		return 0, fmt.Errorf("failed to count active users: %w", err)
	}

	return int64(len(users)), nil
}

// GetWithRoles retrieves a user with their roles
func (r *UserRepository) GetWithRoles(ctx context.Context, userID string) (*models.User, error) {
	// For now, this is a simple implementation - in practice you'd use joins
	return r.GetByID(ctx, userID)
}

// GetWithAddress retrieves a user with their address
func (r *UserRepository) GetWithAddress(ctx context.Context, userID string) (*models.User, error) {
	// For now, this is a simple implementation - in practice you'd use joins
	return r.GetByID(ctx, userID)
}

// GetWithProfile retrieves a user with their profile
func (r *UserRepository) GetWithProfile(ctx context.Context, userID string) (*models.User, error) {
	// For now, this is a simple implementation - in practice you'd use joins
	return r.GetByID(ctx, userID)
}

// Search searches users by keyword in name, username, or mobile number
func (r *UserRepository) Search(ctx context.Context, keyword string, limit, offset int) ([]*models.User, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("username", db.FilterOpContains, keyword),
	}

	var users []models.User
	if err := r.dbManager.List(ctx, filters, &users); err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// Convert []models.User to []*models.User
	result := make([]*models.User, len(users))
	for i := range users {
		result[i] = &users[i]
	}

	return result, nil
}
