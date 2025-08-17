package users

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// ContactRepository handles database operations for Contact entities
type ContactRepository struct {
	*base.BaseFilterableRepository[*models.Contact]
	dbManager db.DBManager
}

// NewContactRepository creates a new ContactRepository instance
func NewContactRepository(dbManager db.DBManager) *ContactRepository {
	return &ContactRepository{
		BaseFilterableRepository: base.NewBaseFilterableRepository[*models.Contact](),
		dbManager:                dbManager,
	}
}

// Create creates a new contact using the base repository
func (r *ContactRepository) Create(ctx context.Context, contact *models.Contact) error {
	return r.BaseFilterableRepository.Create(ctx, contact)
}

// GetByID retrieves a contact by ID using the base repository
func (r *ContactRepository) GetByID(ctx context.Context, id string) (*models.Contact, error) {
	contact := &models.Contact{}
	return r.BaseFilterableRepository.GetByID(ctx, id, contact)
}

// GetByUserID retrieves contacts by user ID using database-level filtering
func (r *ContactRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Contact, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByMobileNumber retrieves a contact by mobile number using database-level filtering
func (r *ContactRepository) GetByMobileNumber(ctx context.Context, mobileNumber uint64) (*models.Contact, error) {
	filter := base.NewFilterBuilder().
		Where("mobile_number", base.OpEqual, mobileNumber).
		Build()

	contacts, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact by mobile number: %w", err)
	}

	if len(contacts) == 0 {
		return nil, fmt.Errorf("contact not found with mobile number: %d", mobileNumber)
	}

	return contacts[0], nil
}

// Update updates an existing contact using the base repository
func (r *ContactRepository) Update(ctx context.Context, contact *models.Contact) error {
	return r.BaseFilterableRepository.Update(ctx, contact)
}

// Delete deletes a contact by ID using the base repository
func (r *ContactRepository) Delete(ctx context.Context, id string) error {
	contact := &models.Contact{}
	return r.BaseFilterableRepository.Delete(ctx, id, contact)
}

// List retrieves a list of contacts with pagination using database-level filtering
func (r *ContactRepository) List(ctx context.Context, limit, offset int) ([]*models.Contact, error) {
	filter := base.NewFilterBuilder().
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// ExistsByMobileNumber checks if a contact exists with the given mobile number
func (r *ContactRepository) ExistsByMobileNumber(ctx context.Context, mobileNumber uint64) (bool, error) {
	_, err := r.GetByMobileNumber(ctx, mobileNumber)
	if err != nil {
		return false, nil // Contact doesn't exist
	}
	return true, nil
}

// ExistsByUserIDAndMobileNumber checks if a contact exists for the given user ID and mobile number
func (r *ContactRepository) ExistsByUserIDAndMobileNumber(ctx context.Context, userID string, mobileNumber uint64) (bool, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("mobile_number", base.OpEqual, mobileNumber).
		Build()

	contacts, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check contact existence: %w", err)
	}

	return len(contacts) > 0, nil
}
