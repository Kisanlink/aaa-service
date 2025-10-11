package adapters

import (
	"context"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	userRepo "github.com/Kisanlink/aaa-service/v2/internal/repositories/users"
)

// UserRepositoryAdapter adapts the concrete user repository to the UserRepositoryInterface
type UserRepositoryAdapter struct {
	repo *userRepo.UserRepository
}

// NewUserRepositoryAdapter creates a new user repository adapter
func NewUserRepositoryAdapter(repo *userRepo.UserRepository) interfaces.UserRepositoryInterface {
	return &UserRepositoryAdapter{repo: repo}
}

// Create implements UserRepositoryInterface.Create
func (a *UserRepositoryAdapter) Create(ctx context.Context, user *models.User) error {
	return a.repo.Create(ctx, user)
}

// GetByID implements UserRepositoryInterface.GetByID
func (a *UserRepositoryAdapter) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user *models.User
	result, err := a.repo.GetByID(ctx, id, user)
	return result, err
}

// Update implements UserRepositoryInterface.Update
func (a *UserRepositoryAdapter) Update(ctx context.Context, user *models.User) error {
	return a.repo.Update(ctx, user)
}

// Delete implements UserRepositoryInterface.Delete (adapter method)
func (a *UserRepositoryAdapter) Delete(ctx context.Context, id string) error {
	// Create a zero value user for the base repository Delete method
	var user *models.User
	return a.repo.Delete(ctx, id, user)
}

// List implements UserRepositoryInterface.List
func (a *UserRepositoryAdapter) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return a.repo.List(ctx, limit, offset)
}

// Count implements UserRepositoryInterface.Count
func (a *UserRepositoryAdapter) Count(ctx context.Context) (int64, error) {
	return a.repo.Count(ctx)
}

// Exists implements UserRepositoryInterface.Exists
func (a *UserRepositoryAdapter) Exists(ctx context.Context, id string) (bool, error) {
	return a.repo.Exists(ctx, id)
}

// SoftDelete implements UserRepositoryInterface.SoftDelete
func (a *UserRepositoryAdapter) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	return a.repo.SoftDelete(ctx, id, deletedBy)
}

// Restore implements UserRepositoryInterface.Restore
func (a *UserRepositoryAdapter) Restore(ctx context.Context, id string) error {
	return a.repo.Restore(ctx, id)
}

// GetByEmail implements UserRepositoryInterface.GetByEmail
func (a *UserRepositoryAdapter) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return a.repo.GetByEmail(ctx, email)
}

// GetByPhoneNumber implements UserRepositoryInterface.GetByPhoneNumber
func (a *UserRepositoryAdapter) GetByPhoneNumber(ctx context.Context, phoneNumber, countryCode string) (*models.User, error) {
	return a.repo.GetByPhoneNumber(ctx, phoneNumber, countryCode)
}

// GetByAadhaarNumber implements UserRepositoryInterface.GetByAadhaarNumber
func (a *UserRepositoryAdapter) GetByAadhaarNumber(ctx context.Context, aadhaarNumber string) (*models.User, error) {
	return a.repo.GetByAadhaarNumber(ctx, aadhaarNumber)
}

// GetByUsername implements UserRepositoryInterface.GetByUsername
func (a *UserRepositoryAdapter) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return a.repo.GetByUsername(ctx, username)
}

// GetByMobileNumber implements UserRepositoryInterface.GetByMobileNumber
func (a *UserRepositoryAdapter) GetByMobileNumber(ctx context.Context, mobileNumber uint64) (*models.User, error) {
	return a.repo.GetByMobileNumber(ctx, mobileNumber)
}

// ListActive implements UserRepositoryInterface.ListActive
func (a *UserRepositoryAdapter) ListActive(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return a.repo.ListActive(ctx, limit, offset)
}

// CountActive implements UserRepositoryInterface.CountActive
func (a *UserRepositoryAdapter) CountActive(ctx context.Context) (int64, error) {
	return a.repo.CountActive(ctx)
}

// Search implements UserRepositoryInterface.Search
func (a *UserRepositoryAdapter) Search(ctx context.Context, keyword string, limit, offset int) ([]*models.User, error) {
	return a.repo.Search(ctx, keyword, limit, offset)
}

// ListAll implements UserRepositoryInterface.ListAll
func (a *UserRepositoryAdapter) ListAll(ctx context.Context) ([]*models.User, error) {
	return a.repo.ListAll(ctx)
}

// GetWithRoles implements UserRepositoryInterface.GetWithRoles
func (a *UserRepositoryAdapter) GetWithRoles(ctx context.Context, userID string) (*models.User, error) {
	return a.repo.GetWithRoles(ctx, userID)
}

// GetWithAddress implements UserRepositoryInterface.GetWithAddress
func (a *UserRepositoryAdapter) GetWithAddress(ctx context.Context, userID string) (*models.User, error) {
	return a.repo.GetWithAddress(ctx, userID)
}

// GetWithProfile implements UserRepositoryInterface.GetWithProfile
func (a *UserRepositoryAdapter) GetWithProfile(ctx context.Context, userID string) (*models.User, error) {
	return a.repo.GetWithProfile(ctx, userID)
}

// ExistsByEmail implements UserRepositoryInterface.ExistsByEmail
func (a *UserRepositoryAdapter) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	// Implement by trying to get the user by email
	user, err := a.repo.GetByEmail(ctx, email)
	if err != nil {
		return false, nil // User doesn't exist
	}
	return user != nil, nil
}

// ExistsByPhoneNumber implements UserRepositoryInterface.ExistsByPhoneNumber
func (a *UserRepositoryAdapter) ExistsByPhoneNumber(ctx context.Context, phoneNumber, countryCode string) (bool, error) {
	// Implement by trying to get the user by phone number
	user, err := a.repo.GetByPhoneNumber(ctx, phoneNumber, countryCode)
	if err != nil {
		return false, nil // User doesn't exist
	}
	return user != nil, nil
}

// ExistsByAadhaarNumber implements UserRepositoryInterface.ExistsByAadhaarNumber
func (a *UserRepositoryAdapter) ExistsByAadhaarNumber(ctx context.Context, aadhaarNumber string) (bool, error) {
	// Implement by trying to get the user by aadhaar number
	user, err := a.repo.GetByAadhaarNumber(ctx, aadhaarNumber)
	if err != nil {
		return false, nil // User doesn't exist
	}
	return user != nil, nil
}

// ExistsByUsername implements UserRepositoryInterface.ExistsByUsername
func (a *UserRepositoryAdapter) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	// Implement by trying to get the user by username
	user, err := a.repo.GetByUsername(ctx, username)
	if err != nil {
		return false, nil // User doesn't exist
	}
	return user != nil, nil
}

// UpdateLastLogin implements UserRepositoryInterface.UpdateLastLogin
func (a *UserRepositoryAdapter) UpdateLastLogin(ctx context.Context, userID string) error {
	// Get the user first
	var user *models.User
	user, err := a.repo.GetByID(ctx, userID, user)
	if err != nil {
		return err
	}

	// For now, this is a no-op since the User model doesn't have LastLoginAt field
	// TODO: Add LastLoginAt field to User model or implement in a separate table
	return nil
}

// UpdatePassword implements UserRepositoryInterface.UpdatePassword
func (a *UserRepositoryAdapter) UpdatePassword(ctx context.Context, userID, hashedPassword string) error {
	// Get the user first
	var user *models.User
	user, err := a.repo.GetByID(ctx, userID, user)
	if err != nil {
		return err
	}

	// Update the password
	user.Password = hashedPassword

	// Save the updated user
	return a.repo.Update(ctx, user)
}

// VerifyPassword implements UserRepositoryInterface.VerifyPassword
func (a *UserRepositoryAdapter) VerifyPassword(ctx context.Context, userID, password string) (bool, error) {
	// Get the user first
	var user *models.User
	user, err := a.repo.GetByID(ctx, userID, user)
	if err != nil {
		return false, err
	}

	// For now, just do a simple string comparison
	// TODO: Implement proper password hashing verification
	if user.Password != password {
		return false, nil
	}

	return true, nil
}
