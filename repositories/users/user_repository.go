package users

import (
	"context"
	"fmt"

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
	// For now, we'll use the embedded BaseFilterableRepository's Count method
	return r.BaseFilterableRepository.Count(ctx)
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
