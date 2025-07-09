package users

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// ContactRepository handles database operations for Contact entities
type ContactRepository struct {
	dbManager db.DBManager
}

// NewContactRepository creates a new ContactRepository instance
func NewContactRepository(dbManager db.DBManager) *ContactRepository {
	return &ContactRepository{
		dbManager: dbManager,
	}
}

// Create creates a new contact
func (r *ContactRepository) Create(ctx context.Context, contact *models.Contact) error {
	if err := contact.BeforeCreate(); err != nil {
		return fmt.Errorf("failed to prepare contact for creation: %w", err)
	}
	return r.dbManager.Create(ctx, contact)
}

// GetByID retrieves a contact by ID
func (r *ContactRepository) GetByID(ctx context.Context, id string) (*models.Contact, error) {
	var contact models.Contact
	if err := r.dbManager.GetByID(ctx, id, &contact); err != nil {
		return nil, fmt.Errorf("failed to get contact by ID: %w", err)
	}
	return &contact, nil
}

// GetByUserID retrieves contacts by user ID
func (r *ContactRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Contact, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("user_id", db.FilterOpEqual, userID),
	}

	var contacts []models.Contact
	if err := r.dbManager.List(ctx, filters, &contacts); err != nil {
		return nil, fmt.Errorf("failed to get contacts by user ID: %w", err)
	}

	// Convert []models.Contact to []*models.Contact
	results := make([]*models.Contact, len(contacts))
	for i, contact := range contacts {
		results[i] = &contact
	}

	return results, nil
}

// GetByMobileNumber retrieves a contact by mobile number
func (r *ContactRepository) GetByMobileNumber(ctx context.Context, mobileNumber uint64) (*models.Contact, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("mobile_number", db.FilterOpEqual, mobileNumber),
	}

	var contacts []models.Contact
	if err := r.dbManager.List(ctx, filters, &contacts); err != nil {
		return nil, fmt.Errorf("failed to get contact by mobile number: %w", err)
	}

	if len(contacts) == 0 {
		return nil, fmt.Errorf("contact not found with mobile number: %d", mobileNumber)
	}

	return &contacts[0], nil
}

// Update updates an existing contact
func (r *ContactRepository) Update(ctx context.Context, contact *models.Contact) error {
	if err := contact.BeforeUpdate(); err != nil {
		return fmt.Errorf("failed to prepare contact for update: %w", err)
	}
	return r.dbManager.Update(ctx, contact)
}

// Delete deletes a contact by ID
func (r *ContactRepository) Delete(ctx context.Context, id string) error {
	return r.dbManager.Delete(ctx, id)
}

// List retrieves a list of contacts with pagination
func (r *ContactRepository) List(ctx context.Context, filters []db.Filter, limit, offset int) ([]*models.Contact, error) {
	var contacts []models.Contact
	if err := r.dbManager.List(ctx, filters, &contacts); err != nil {
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}

	// Convert []models.Contact to []*models.Contact
	results := make([]*models.Contact, len(contacts))
	for i, contact := range contacts {
		results[i] = &contact
	}

	return results, nil
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
	filters := []db.Filter{
		r.dbManager.BuildFilter("user_id", db.FilterOpEqual, userID),
		r.dbManager.BuildFilter("mobile_number", db.FilterOpEqual, mobileNumber),
	}

	var contacts []models.Contact
	if err := r.dbManager.List(ctx, filters, &contacts); err != nil {
		return false, fmt.Errorf("failed to check contact existence: %w", err)
	}

	return len(contacts) > 0, nil
}
