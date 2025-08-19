package actions

import (
	"context"
	"fmt"

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
	baseRepo.SetDBManager(dbManager) // Connect the base repository to the actual database
	return &ActionRepository{
		BaseFilterableRepository: baseRepo,
		dbManager:                dbManager,
	}
}

// Create creates a new action using the database manager
func (r *ActionRepository) Create(ctx context.Context, action *models.Action) error {
	return r.dbManager.Create(ctx, action)
}

// GetByID retrieves an action by ID using the database manager
func (r *ActionRepository) GetByID(ctx context.Context, id string) (*models.Action, error) {
	action := &models.Action{}
	err := r.dbManager.GetByID(ctx, id, action)
	if err != nil {
		return nil, err
	}
	return action, nil
}

// Update updates an existing action using the database manager
func (r *ActionRepository) Update(ctx context.Context, action *models.Action) error {
	return r.dbManager.Update(ctx, action)
}

// Delete deletes an action by ID using the database manager
func (r *ActionRepository) Delete(ctx context.Context, id string) error {
	// Get the action first to pass as model parameter
	action, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get action for deletion: %w", err)
	}

	// Use the BaseFilterableRepository which now properly handles table names
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
		Where("name", base.OpEqual, name).
		Build()

	actions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get action by name: %w", err)
	}

	if len(actions) == 0 {
		return nil, fmt.Errorf("action not found with name: %s", name)
	}

	return actions[0], nil
}

// GetByServiceName retrieves actions by service name
func (r *ActionRepository) GetByServiceName(ctx context.Context, serviceName string, limit, offset int) ([]*models.Action, error) {
	filter := base.NewFilterBuilder().
		Where("service_name", base.OpEqual, serviceName).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}
