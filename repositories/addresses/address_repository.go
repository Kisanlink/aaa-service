package addresses

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// AddressRepository handles database operations for Address entities
type AddressRepository struct {
	*base.BaseFilterableRepository[*models.Address]
	dbManager db.DBManager
}

// NewAddressRepository creates a new AddressRepository instance
func NewAddressRepository(dbManager db.DBManager) *AddressRepository {
	return &AddressRepository{
		BaseFilterableRepository: base.NewBaseFilterableRepository[*models.Address](),
		dbManager:                dbManager,
	}
}

// Create creates a new address
func (r *AddressRepository) Create(ctx context.Context, address *models.Address) error {
	if err := address.BeforeCreate(); err != nil {
		return fmt.Errorf("failed to prepare address for creation: %w", err)
	}
	return r.dbManager.Create(ctx, address)
}

// GetByID retrieves an address by ID
func (r *AddressRepository) GetByID(ctx context.Context, id string) (*models.Address, error) {
	var address models.Address
	if err := r.dbManager.GetByID(ctx, id, &address); err != nil {
		return nil, fmt.Errorf("failed to get address by ID: %w", err)
	}
	return &address, nil
}

// Update updates an existing address
func (r *AddressRepository) Update(ctx context.Context, address *models.Address) error {
	if err := address.BeforeUpdate(); err != nil {
		return fmt.Errorf("failed to prepare address for update: %w", err)
	}
	return r.dbManager.Update(ctx, address)
}

// Delete deletes an address by ID
func (r *AddressRepository) Delete(ctx context.Context, id string) error {
	return r.dbManager.Delete(ctx, id)
}

// List retrieves a list of addresses with pagination
func (r *AddressRepository) List(ctx context.Context, limit, offset int) ([]*models.Address, error) {
	var addresses []models.Address
	if err := r.dbManager.List(ctx, []db.Filter{}, &addresses); err != nil {
		return nil, fmt.Errorf("failed to list addresses: %w", err)
	}

	// Convert []models.Address to []*models.Address
	result := make([]*models.Address, len(addresses))
	for i := range addresses {
		result[i] = &addresses[i]
	}

	return result, nil
}

// Count returns the total number of addresses
func (r *AddressRepository) Count(ctx context.Context) (int64, error) {
	return r.BaseFilterableRepository.Count(ctx)
}

// GetByUserID retrieves addresses by user ID
func (r *AddressRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Address, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("user_id", db.FilterOpEqual, userID),
	}

	var addresses []models.Address
	if err := r.dbManager.List(ctx, filters, &addresses); err != nil {
		return nil, fmt.Errorf("failed to get addresses by user ID: %w", err)
	}

	// Convert []models.Address to []*models.Address
	result := make([]*models.Address, len(addresses))
	for i := range addresses {
		result[i] = &addresses[i]
	}

	return result, nil
}

// Search searches addresses by keyword
func (r *AddressRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.Address, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("full_address", db.FilterOpContains, query),
	}

	var addresses []models.Address
	if err := r.dbManager.List(ctx, filters, &addresses); err != nil {
		return nil, fmt.Errorf("failed to search addresses: %w", err)
	}

	// Convert []models.Address to []*models.Address
	result := make([]*models.Address, len(addresses))
	for i := range addresses {
		result[i] = &addresses[i]
	}

	return result, nil
}
