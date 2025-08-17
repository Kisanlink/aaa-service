package actions

import (
	"context"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// ActionRepository handles database operations for Action entities
type ActionRepository struct {
	*base.BaseFilterableRepository[*models.Action]
	dbManager db.DBManager
}

// NewActionRepository creates a new ActionRepository instance
func NewActionRepository(dbManager db.DBManager) *ActionRepository {
	baseRepo := base.NewBaseFilterableRepository[*models.Action]()
	baseRepo.SetDBManager(dbManager)
	return &ActionRepository{
		BaseFilterableRepository: baseRepo,
		dbManager:                dbManager,
	}
}

// Create creates a new action using the base repository
func (r *ActionRepository) Create(ctx context.Context, action *models.Action) error {
	return r.BaseFilterableRepository.Create(ctx, action)
}

// GetByID retrieves an action by ID using the base repository
func (r *ActionRepository) GetByID(ctx context.Context, id string) (*models.Action, error) {
	action := &models.Action{}
	return r.BaseFilterableRepository.GetByID(ctx, id, action)
}

// Update updates an existing action using the base repository
func (r *ActionRepository) Update(ctx context.Context, action *models.Action) error {
	return r.BaseFilterableRepository.Update(ctx, action)
}

// Delete deletes an action by ID using the base repository
func (r *ActionRepository) Delete(ctx context.Context, id string) error {
	action := &models.Action{}
	return r.BaseFilterableRepository.Delete(ctx, id, action)
}

// List retrieves actions with pagination using database-level filtering
func (r *ActionRepository) List(ctx context.Context, limit, offset int) ([]*models.Action, error) {
	filter := base.NewFilterBuilder().
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Count returns the total number of actions using database-level counting
func (r *ActionRepository) Count(ctx context.Context) (int64, error) {
	filter := base.NewFilter()
	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// Exists checks if an action exists by ID using the base repository
func (r *ActionRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.Exists(ctx, id)
}

// SoftDelete soft deletes an action by ID using the base repository
func (r *ActionRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	return r.BaseFilterableRepository.SoftDelete(ctx, id, deletedBy)
}

// Restore restores a soft-deleted action using the base repository
func (r *ActionRepository) Restore(ctx context.Context, id string) error {
	return r.BaseFilterableRepository.Restore(ctx, id)
}

// ListWithDeleted retrieves actions including soft-deleted ones using the base repository
func (r *ActionRepository) ListWithDeleted(ctx context.Context, limit, offset int) ([]*models.Action, error) {
	return r.BaseFilterableRepository.ListWithDeleted(ctx, limit, offset)
}

// CountWithDeleted returns count including soft-deleted actions using the base repository
func (r *ActionRepository) CountWithDeleted(ctx context.Context) (int64, error) {
	return r.BaseFilterableRepository.CountWithDeleted(ctx)
}

// ExistsWithDeleted checks if action exists including soft-deleted ones using the base repository
func (r *ActionRepository) ExistsWithDeleted(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.ExistsWithDeleted(ctx, id)
}

// GetByCreatedBy gets actions by creator using the base repository
func (r *ActionRepository) GetByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*models.Action, error) {
	return r.BaseFilterableRepository.GetByCreatedBy(ctx, createdBy, limit, offset)
}

// GetByUpdatedBy gets actions by updater using the base repository
func (r *ActionRepository) GetByUpdatedBy(ctx context.Context, updatedBy string, limit, offset int) ([]*models.Action, error) {
	return r.BaseFilterableRepository.GetByUpdatedBy(ctx, updatedBy, limit, offset)
}

// GetByName retrieves an action by name
func (r *ActionRepository) GetByName(ctx context.Context, name string) (*models.Action, error) {
	filter := base.NewFilterBuilder().
		Where("name", "=", name).
		Build()

	actions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(actions) == 0 {
		return nil, nil
	}

	return actions[0], nil
}

// GetByService retrieves actions by service name
func (r *ActionRepository) GetByService(ctx context.Context, serviceName string, limit, offset int) ([]*models.Action, error) {
	filter := base.NewFilterBuilder().
		Where("service_name", "=", serviceName).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}
