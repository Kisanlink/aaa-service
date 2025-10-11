package actions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// GetByID retrieves an action by ID using the database manager
func (r *ActionRepository) GetByID(ctx context.Context, id string) (*models.Action, error) {
	if id == "" {
		return nil, fmt.Errorf("action ID is required")
	}

	action := &models.Action{}
	err := r.dbManager.GetByID(ctx, id, action)
	if err != nil {
		return nil, fmt.Errorf("failed to get action by ID: %w", err)
	}
	return action, nil
}

// GetByName retrieves an action by name
func (r *ActionRepository) GetByName(ctx context.Context, name string) (*models.Action, error) {
	if name == "" {
		return nil, fmt.Errorf("action name is required")
	}

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

// List retrieves actions with pagination using database-level filtering
func (r *ActionRepository) List(ctx context.Context, filter *base.Filter) ([]*models.Action, error) {
	if filter == nil {
		filter = base.NewFilter()
	}

	actions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list actions: %w", err)
	}

	return actions, nil
}

// Count returns the total number of actions using database-level counting
func (r *ActionRepository) Count(ctx context.Context, filter *base.Filter) (int64, error) {
	if filter == nil {
		filter = base.NewFilter()
	}

	count, err := r.BaseFilterableRepository.CountWithFilter(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count actions: %w", err)
	}

	return count, nil
}

// Exists checks if an action exists by ID using the base repository
func (r *ActionRepository) Exists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, fmt.Errorf("action ID is required")
	}

	exists, err := r.BaseFilterableRepository.Exists(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to check if action exists: %w", err)
	}

	return exists, nil
}

// GetByCategory retrieves actions by category
func (r *ActionRepository) GetByCategory(ctx context.Context, category string, limit, offset int) ([]*models.Action, error) {
	if category == "" {
		return nil, fmt.Errorf("category is required")
	}

	filter := base.NewFilterBuilder().
		Where("category", base.OpEqual, category).
		Limit(limit, offset).
		Build()

	actions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions by category: %w", err)
	}

	return actions, nil
}

// GetStaticActions retrieves all static (built-in) actions
func (r *ActionRepository) GetStaticActions(ctx context.Context, limit, offset int) ([]*models.Action, error) {
	filter := base.NewFilterBuilder().
		Where("is_static", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	actions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get static actions: %w", err)
	}

	return actions, nil
}

// GetDynamicActions retrieves all dynamic (service-defined) actions
func (r *ActionRepository) GetDynamicActions(ctx context.Context, limit, offset int) ([]*models.Action, error) {
	filter := base.NewFilterBuilder().
		Where("is_static", base.OpEqual, false).
		Limit(limit, offset).
		Build()

	actions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get dynamic actions: %w", err)
	}

	return actions, nil
}

// GetByServiceID retrieves actions by service ID
func (r *ActionRepository) GetByServiceID(ctx context.Context, serviceID string, limit, offset int) ([]*models.Action, error) {
	if serviceID == "" {
		return nil, fmt.Errorf("service ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("service_id", base.OpEqual, serviceID).
		Limit(limit, offset).
		Build()

	actions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions by service ID: %w", err)
	}

	return actions, nil
}

// GetActiveActions retrieves all active actions
func (r *ActionRepository) GetActiveActions(ctx context.Context, limit, offset int) ([]*models.Action, error) {
	filter := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	actions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get active actions: %w", err)
	}

	return actions, nil
}
