package users

import (
	"context"
	"errors"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"gorm.io/gorm"
)

// ContactRepository handles database operations for Contact entities
type ContactRepository struct {
	base.Repository[models.Contact]
	dbManager *db.Manager
}

// NewContactRepository creates a new ContactRepository instance
func NewContactRepository(dbManager *db.Manager) *ContactRepository {
	return &ContactRepository{
		Repository: base.NewRepository[models.Contact](dbManager),
		dbManager:  dbManager,
	}
}

// Create creates a new contact
func (r *ContactRepository) Create(ctx context.Context, contact *models.Contact) error {
	return r.Repository.Create(ctx, contact)
}

// GetByID retrieves a contact by ID
func (r *ContactRepository) GetByID(ctx context.Context, id string) (*models.Contact, error) {
	var contact models.Contact
	err := r.dbManager.GetDB().WithContext(ctx).
		Preload("Address").
		Where("id = ?", id).
		First(&contact).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, base.ErrNotFound
		}
		return nil, err
	}
	return &contact, nil
}

// GetByUserID retrieves contacts by user ID
func (r *ContactRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Contact, error) {
	var contacts []*models.Contact
	err := r.dbManager.GetDB().WithContext(ctx).
		Preload("Address").
		Where("user_id = ?", userID).
		Find(&contacts).Error
	if err != nil {
		return nil, err
	}
	return contacts, nil
}

// GetByMobileNumber retrieves a contact by mobile number
func (r *ContactRepository) GetByMobileNumber(ctx context.Context, mobileNumber uint64) (*models.Contact, error) {
	var contact models.Contact
	err := r.dbManager.GetDB().WithContext(ctx).
		Preload("Address").
		Where("mobile_number = ?", mobileNumber).
		First(&contact).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, base.ErrNotFound
		}
		return nil, err
	}
	return &contact, nil
}

// Update updates an existing contact
func (r *ContactRepository) Update(ctx context.Context, contact *models.Contact) error {
	return r.Repository.Update(ctx, contact)
}

// Delete deletes a contact by ID
func (r *ContactRepository) Delete(ctx context.Context, id string) error {
	return r.Repository.Delete(ctx, id)
}

// List retrieves a list of contacts with pagination
func (r *ContactRepository) List(ctx context.Context, filters *base.Filters) ([]*models.Contact, error) {
	var contacts []*models.Contact
	query := r.dbManager.GetDB().WithContext(ctx).Preload("Address")

	if filters != nil {
		query = filters.Apply(query)
	}

	err := query.Find(&contacts).Error
	if err != nil {
		return nil, err
	}

	return contacts, nil
}

// ExistsByMobileNumber checks if a contact exists with the given mobile number
func (r *ContactRepository) ExistsByMobileNumber(ctx context.Context, mobileNumber uint64) (bool, error) {
	var count int64
	err := r.dbManager.GetDB().WithContext(ctx).
		Model(&models.Contact{}).
		Where("mobile_number = ?", mobileNumber).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByUserIDAndMobileNumber checks if a contact exists for the given user ID and mobile number
func (r *ContactRepository) ExistsByUserIDAndMobileNumber(ctx context.Context, userID string, mobileNumber uint64) (bool, error) {
	var count int64
	err := r.dbManager.GetDB().WithContext(ctx).
		Model(&models.Contact{}).
		Where("user_id = ? AND mobile_number = ?", userID, mobileNumber).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
