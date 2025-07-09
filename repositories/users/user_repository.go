package users

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// UserRepository handles database operations for User entities
type UserRepository struct {
	dbManager db.DBManager
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(dbManager db.DBManager) *UserRepository {
	return &UserRepository{
		dbManager: dbManager,
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

// List retrieves users with optional filters
func (r *UserRepository) List(ctx context.Context, filters []db.Filter, limit, offset int) ([]models.User, error) {
	var users []models.User

	if err := r.dbManager.List(ctx, filters, &users); err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// ListActive retrieves all active users
func (r *UserRepository) ListActive(ctx context.Context, limit, offset int) ([]models.User, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("status", db.FilterOpEqual, "active"),
	}

	return r.List(ctx, filters, limit, offset)
}

// ListValidated retrieves all validated users
func (r *UserRepository) ListValidated(ctx context.Context, limit, offset int) ([]models.User, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("is_validated", db.FilterOpEqual, true),
	}

	return r.List(ctx, filters, limit, offset)
}

// SearchByKeyword searches users by keyword in name, username, or mobile number
func (r *UserRepository) SearchByKeyword(ctx context.Context, keyword string, limit, offset int) ([]models.User, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("name", db.FilterOpContains, keyword),
		r.dbManager.BuildFilter("username", db.FilterOpContains, keyword),
	}

	return r.List(ctx, filters, limit, offset)
}

// Exists checks if a user exists by ID
func (r *UserRepository) Exists(ctx context.Context, id string) (bool, error) {
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return false, nil // User doesn't exist
	}
	return true, nil
}

// ExistsByUsername checks if a user exists by username
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	_, err := r.GetByUsername(ctx, username)
	if err != nil {
		return false, nil // User doesn't exist
	}
	return true, nil
}

// ExistsByMobileNumber checks if a user exists by mobile number
func (r *UserRepository) ExistsByMobileNumber(ctx context.Context, mobileNumber uint64) (bool, error) {
	_, err := r.GetByMobileNumber(ctx, mobileNumber)
	if err != nil {
		return false, nil // User doesn't exist
	}
	return true, nil
}

// ExistsByAadhaarNumber checks if a user exists by Aadhaar number
func (r *UserRepository) ExistsByAadhaarNumber(ctx context.Context, aadhaarNumber string) (bool, error) {
	_, err := r.GetByAadhaarNumber(ctx, aadhaarNumber)
	if err != nil {
		return false, nil // User doesn't exist
	}
	return true, nil
}
