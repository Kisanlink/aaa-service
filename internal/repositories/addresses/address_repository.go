package addresses

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
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
func (r *AddressRepository) GetByID(ctx context.Context, id string, address *models.Address) (*models.Address, error) {
	if err := r.dbManager.GetByID(ctx, id, address); err != nil {
		return nil, fmt.Errorf("failed to get address by ID: %w", err)
	}
	return address, nil
}

// Update updates an existing address
func (r *AddressRepository) Update(ctx context.Context, address *models.Address) error {
	if err := address.BeforeUpdate(); err != nil {
		return fmt.Errorf("failed to prepare address for update: %w", err)
	}
	return r.dbManager.Update(ctx, address)
}

// Delete deletes an address by ID
func (r *AddressRepository) Delete(ctx context.Context, id string, address *models.Address) error {
	return r.BaseFilterableRepository.Delete(ctx, id, address)
}

// List retrieves a list of addresses with pagination using database-level filtering
func (r *AddressRepository) List(ctx context.Context, limit, offset int) ([]*models.Address, error) {
	// Use base filterable repository for optimized database-level filtering
	filter := base.NewFilterBuilder().
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Count returns the total number of addresses using database-level counting
func (r *AddressRepository) Count(ctx context.Context) (int64, error) {
	filter := base.NewFilter()
	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// GetByUserID retrieves addresses by user ID using database-level filtering
func (r *AddressRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Search searches addresses by keyword using database-level filtering
func (r *AddressRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("full_address", base.OpContains, query).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByType retrieves addresses by type using database-level filtering
func (r *AddressRepository) GetByType(ctx context.Context, addressType string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("type", base.OpEqual, addressType).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByCity retrieves addresses by city using database-level filtering
func (r *AddressRepository) GetByCity(ctx context.Context, city string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("city", base.OpEqual, city).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByState retrieves addresses by state using database-level filtering
func (r *AddressRepository) GetByState(ctx context.Context, state string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("state", base.OpEqual, state).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByPostalCode retrieves addresses by postal code using database-level filtering
func (r *AddressRepository) GetByPostalCode(ctx context.Context, postalCode string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("postal_code", base.OpEqual, postalCode).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByCountry retrieves addresses by country using database-level filtering
func (r *AddressRepository) GetByCountry(ctx context.Context, country string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("country", base.OpEqual, country).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetActiveAddresses retrieves active addresses using database-level filtering
func (r *AddressRepository) GetActiveAddresses(ctx context.Context, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetDefaultAddresses retrieves default addresses using database-level filtering
func (r *AddressRepository) GetDefaultAddresses(ctx context.Context, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("is_default", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByDateRange retrieves addresses created within a date range using database-level filtering
func (r *AddressRepository) GetAddressesByDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndType retrieves addresses by user ID and type using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndType(ctx context.Context, userID, addressType string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("type", base.OpEqual, addressType).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndCity retrieves addresses by user ID and city using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndCity(ctx context.Context, userID, city string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("city", base.OpEqual, city).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndState retrieves addresses by user ID and state using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndState(ctx context.Context, userID, state string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("state", base.OpEqual, state).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndCountry retrieves addresses by user ID and country using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndCountry(ctx context.Context, userID, country string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("country", base.OpEqual, country).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndPostalCode retrieves addresses by user ID and postal code using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndPostalCode(ctx context.Context, userID, postalCode string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("postal_code", base.OpEqual, postalCode).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndActive retrieves active addresses by user ID using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndActive(ctx context.Context, userID string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndDefault retrieves default addresses by user ID using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndDefault(ctx context.Context, userID string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("is_default", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndDateRange retrieves addresses by user ID and date range using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndDateRange(ctx context.Context, userID, startDate, endDate string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndSearch retrieves addresses by user ID and search query using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndSearch(ctx context.Context, userID, query string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("full_address", base.OpContains, query).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndTypeAndActive retrieves active addresses by user ID and type using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndTypeAndActive(ctx context.Context, userID, addressType string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("type", base.OpEqual, addressType).
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndTypeAndDefault retrieves default addresses by user ID and type using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndTypeAndDefault(ctx context.Context, userID, addressType string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("type", base.OpEqual, addressType).
		Where("is_default", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndTypeAndDateRange retrieves addresses by user ID, type and date range using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndTypeAndDateRange(ctx context.Context, userID, addressType, startDate, endDate string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("type", base.OpEqual, addressType).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndTypeAndSearch retrieves addresses by user ID, type and search query using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndTypeAndSearch(ctx context.Context, userID, addressType, query string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("type", base.OpEqual, addressType).
		Where("full_address", base.OpContains, query).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndTypeAndActiveAndDefault retrieves active default addresses by user ID and type using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndTypeAndActiveAndDefault(ctx context.Context, userID, addressType string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("type", base.OpEqual, addressType).
		Where("is_active", base.OpEqual, true).
		Where("is_default", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndTypeAndActiveAndDateRange retrieves active addresses by user ID, type and date range using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndTypeAndActiveAndDateRange(ctx context.Context, userID, addressType, startDate, endDate string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("type", base.OpEqual, addressType).
		Where("is_active", base.OpEqual, true).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndTypeAndActiveAndSearch retrieves active addresses by user ID, type and search query using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndTypeAndActiveAndSearch(ctx context.Context, userID, addressType, query string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("type", base.OpEqual, addressType).
		Where("is_active", base.OpEqual, true).
		Where("full_address", base.OpContains, query).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndTypeAndDefaultAndDateRange retrieves default addresses by user ID, type and date range using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndTypeAndDefaultAndDateRange(ctx context.Context, userID, addressType, startDate, endDate string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("type", base.OpEqual, addressType).
		Where("is_default", base.OpEqual, true).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndTypeAndDefaultAndSearch retrieves default addresses by user ID, type and search query using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndTypeAndDefaultAndSearch(ctx context.Context, userID, addressType, query string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("type", base.OpEqual, addressType).
		Where("is_default", base.OpEqual, true).
		Where("full_address", base.OpContains, query).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndTypeAndDateRangeAndSearch retrieves addresses by user ID, type, date range and search query using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndTypeAndDateRangeAndSearch(ctx context.Context, userID, addressType, startDate, endDate, query string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("type", base.OpEqual, addressType).
		WhereBetween("created_at", startDate, endDate).
		Where("full_address", base.OpContains, query).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndTypeAndActiveAndDefaultAndDateRange retrieves active default addresses by user ID, type and date range using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndTypeAndActiveAndDefaultAndDateRange(ctx context.Context, userID, addressType, startDate, endDate string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("type", base.OpEqual, addressType).
		Where("is_active", base.OpEqual, true).
		Where("is_default", base.OpEqual, true).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndTypeAndActiveAndDefaultAndSearch retrieves active default addresses by user ID, type and search query using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndTypeAndActiveAndDefaultAndSearch(ctx context.Context, userID, addressType, query string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("type", base.OpEqual, addressType).
		Where("is_active", base.OpEqual, true).
		Where("is_default", base.OpEqual, true).
		Where("full_address", base.OpContains, query).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndTypeAndActiveAndDateRangeAndSearch retrieves active addresses by user ID, type, date range and search query using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndTypeAndActiveAndDateRangeAndSearch(ctx context.Context, userID, addressType, startDate, endDate, query string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("type", base.OpEqual, addressType).
		Where("is_active", base.OpEqual, true).
		WhereBetween("created_at", startDate, endDate).
		Where("full_address", base.OpContains, query).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndTypeAndDefaultAndDateRangeAndSearch retrieves default addresses by user ID, type, date range and search query using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndTypeAndDefaultAndDateRangeAndSearch(ctx context.Context, userID, addressType, startDate, endDate, query string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("type", base.OpEqual, addressType).
		Where("is_default", base.OpEqual, true).
		WhereBetween("created_at", startDate, endDate).
		Where("full_address", base.OpContains, query).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetAddressesByUserAndTypeAndActiveAndDefaultAndDateRangeAndSearch retrieves active default addresses by user ID, type, date range and search query using database-level filtering
func (r *AddressRepository) GetAddressesByUserAndTypeAndActiveAndDefaultAndDateRangeAndSearch(ctx context.Context, userID, addressType, startDate, endDate, query string, limit, offset int) ([]*models.Address, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Where("type", base.OpEqual, addressType).
		Where("is_active", base.OpEqual, true).
		Where("is_default", base.OpEqual, true).
		WhereBetween("created_at", startDate, endDate).
		Where("full_address", base.OpContains, query).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}
