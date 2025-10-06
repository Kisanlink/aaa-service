package actions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
)

// Create creates a new action using the database manager
func (r *ActionRepository) Create(ctx context.Context, action *models.Action) error {
	// Validate action
	if action == nil {
		return fmt.Errorf("action cannot be nil")
	}

	if action.Name == "" {
		return fmt.Errorf("action name is required")
	}

	if action.Category == "" {
		return fmt.Errorf("action category is required")
	}

	// Check if action with same name already exists
	existing, err := r.GetByName(ctx, action.Name)
	if err == nil && existing != nil {
		return fmt.Errorf("action with name '%s' already exists", action.Name)
	}

	// Create the action
	return r.dbManager.Create(ctx, action)
}

// CreateBatch creates multiple actions in a single transaction
func (r *ActionRepository) CreateBatch(ctx context.Context, actions []*models.Action) error {
	if len(actions) == 0 {
		return fmt.Errorf("no actions to create")
	}

	// Validate all actions before creating
	for i, action := range actions {
		if action == nil {
			return fmt.Errorf("action at index %d is nil", i)
		}

		if action.Name == "" {
			return fmt.Errorf("action at index %d has empty name", i)
		}

		if action.Category == "" {
			return fmt.Errorf("action at index %d has empty category", i)
		}
	}

	// Create all actions in batch
	for _, action := range actions {
		if err := r.Create(ctx, action); err != nil {
			return fmt.Errorf("failed to create action '%s': %w", action.Name, err)
		}
	}

	return nil
}

// CreateStaticAction creates a new static (built-in) action
func (r *ActionRepository) CreateStaticAction(ctx context.Context, name, description, category string) (*models.Action, error) {
	action := models.NewStaticAction(name, description, category)

	if err := r.Create(ctx, action); err != nil {
		return nil, fmt.Errorf("failed to create static action: %w", err)
	}

	return action, nil
}

// CreateDynamicAction creates a new dynamic (service-defined) action
func (r *ActionRepository) CreateDynamicAction(ctx context.Context, name, description, category, serviceID string) (*models.Action, error) {
	if serviceID == "" {
		return nil, fmt.Errorf("service ID is required for dynamic actions")
	}

	action := models.NewDynamicAction(name, description, category, serviceID)

	if err := r.Create(ctx, action); err != nil {
		return nil, fmt.Errorf("failed to create dynamic action: %w", err)
	}

	return action, nil
}
