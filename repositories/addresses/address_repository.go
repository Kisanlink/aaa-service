package addresses

import (
	"context"
	"errors"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"gorm.io/gorm"
)

// AddressRepository handles database operations for Address entities
type AddressRepository struct {
	base.Repository[models.Address]
	dbManager *db.Manager
}

// NewAddressRepository creates a new AddressRepository instance
func NewAddressRepository(dbManager *db.Manager) *AddressRepository {
	return &AddressRepository{
		Repository: base.NewRepository[models.Address](dbManager),
		dbManager:  dbManager,
	}
}

// Create creates a new address
func (r *AddressRepository) Create(ctx context.Context, address *models.Address) error {
	return r.Repository.Create(ctx, address)
}

// GetByID retrieves an address by ID
func (r *AddressRepository) GetByID(ctx context.Context, id string) (*models.Address, error) {
	var address models.Address
	err := r.dbManager.GetDB().WithContext(ctx).
		Where("id = ?", id).
		First(&address).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, base.ErrNotFound
		}
		return nil, err
	}
	return &address, nil
}

// GetByUserID retrieves addresses by user ID
func (r *AddressRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Address, error) {
	var addresses []*models.Address
	err := r.dbManager.GetDB().WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&addresses).Error
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

// Update updates an existing address
func (r *AddressRepository) Update(ctx context.Context, address *models.Address) error {
	return r.Repository.Update(ctx, address)
}

// Delete deletes an address by ID
func (r *AddressRepository) Delete(ctx context.Context, id string) error {
	return r.Repository.Delete(ctx, id)
}

// List retrieves a list of addresses with pagination
func (r *AddressRepository) List(ctx context.Context, filters *base.Filters) ([]*models.Address, error) {
	var addresses []*models.Address
	query := r.dbManager.GetDB().WithContext(ctx)

	if filters != nil {
		query = filters.Apply(query)
	}

	err := query.Find(&addresses).Error
	if err != nil {
		return nil, err
	}

	return addresses, nil
}

// GetByPincode retrieves addresses by pincode
func (r *AddressRepository) GetByPincode(ctx context.Context, pincode string) ([]*models.Address, error) {
	var addresses []*models.Address
	err := r.dbManager.GetDB().WithContext(ctx).
		Where("pincode = ?", pincode).
		Find(&addresses).Error
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

// GetByState retrieves addresses by state
func (r *AddressRepository) GetByState(ctx context.Context, state string) ([]*models.Address, error) {
	var addresses []*models.Address
	err := r.dbManager.GetDB().WithContext(ctx).
		Where("state = ?", state).
		Find(&addresses).Error
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

// GetByDistrict retrieves addresses by district
func (r *AddressRepository) GetByDistrict(ctx context.Context, district string) ([]*models.Address, error) {
	var addresses []*models.Address
	err := r.dbManager.GetDB().WithContext(ctx).
		Where("district = ?", district).
		Find(&addresses).Error
	if err != nil {
		return nil, err
	}
	return addresses, nil
}
