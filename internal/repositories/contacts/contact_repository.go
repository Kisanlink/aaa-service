package contacts

import (
	"context"

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
	baseRepo := base.NewBaseFilterableRepository[*models.Contact]()
	baseRepo.SetDBManager(dbManager)
	return &ContactRepository{
		BaseFilterableRepository: baseRepo,
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

// Update updates an existing contact using the base repository
func (r *ContactRepository) Update(ctx context.Context, contact *models.Contact) error {
	return r.BaseFilterableRepository.Update(ctx, contact)
}

// Delete deletes a contact by ID using the base repository
func (r *ContactRepository) Delete(ctx context.Context, id string) error {
	contact := &models.Contact{}
	return r.BaseFilterableRepository.Delete(ctx, id, contact)
}

// List retrieves contacts with pagination using database-level filtering
func (r *ContactRepository) List(ctx context.Context, limit, offset int) ([]*models.Contact, error) {
	filter := base.NewFilterBuilder().
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Count returns the total number of contacts using database-level counting
func (r *ContactRepository) Count(ctx context.Context) (int64, error) {
	filter := base.NewFilter()
	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// Exists checks if a contact exists by ID using the base repository
func (r *ContactRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.Exists(ctx, id)
}

// SoftDelete soft deletes a contact by ID using the base repository
func (r *ContactRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	return r.BaseFilterableRepository.SoftDelete(ctx, id, deletedBy)
}

// Restore restores a soft-deleted contact using the base repository
func (r *ContactRepository) Restore(ctx context.Context, id string) error {
	return r.BaseFilterableRepository.Restore(ctx, id)
}

// ListWithDeleted retrieves contacts including soft-deleted ones using the base repository
func (r *ContactRepository) ListWithDeleted(ctx context.Context, limit, offset int) ([]*models.Contact, error) {
	return r.BaseFilterableRepository.ListWithDeleted(ctx, limit, offset)
}

// CountWithDeleted returns count including soft-deleted contacts using the base repository
func (r *ContactRepository) CountWithDeleted(ctx context.Context) (int64, error) {
	return r.BaseFilterableRepository.CountWithDeleted(ctx)
}

// ExistsWithDeleted checks if contact exists including soft-deleted ones using the base repository
func (r *ContactRepository) ExistsWithDeleted(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.ExistsWithDeleted(ctx, id)
}

// GetByCreatedBy gets contacts by creator using the base repository
func (r *ContactRepository) GetByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*models.Contact, error) {
	return r.BaseFilterableRepository.GetByCreatedBy(ctx, createdBy, limit, offset)
}

// GetByUpdatedBy gets contacts by updater using the base repository
func (r *ContactRepository) GetByUpdatedBy(ctx context.Context, updatedBy string, limit, offset int) ([]*models.Contact, error) {
	return r.BaseFilterableRepository.GetByUpdatedBy(ctx, updatedBy, limit, offset)
}

// GetByUserID retrieves contacts by user ID
func (r *ContactRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Contact, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", base.OpEqual, userID).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByType retrieves contacts by type
func (r *ContactRepository) GetByType(ctx context.Context, contactType string, limit, offset int) ([]*models.Contact, error) {
	filter := base.NewFilterBuilder().
		Where("type", base.OpEqual, contactType).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}
