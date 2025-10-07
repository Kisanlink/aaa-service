package actions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
)

// Update updates an existing action using the database manager
func (r *ActionRepository) Update(ctx context.Context, action *models.Action) error {
	if action == nil {
		return fmt.Errorf("action cannot be nil")
	}

	if action.ID == "" {
		return fmt.Errorf("action ID is required")
	}

	// Check if action exists
	exists, err := r.Exists(ctx, action.ID)
	if err != nil {
		return fmt.Errorf("failed to check if action exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("action with ID '%s' not found", action.ID)
	}

	// Validate required fields
	if action.Name == "" {
		return fmt.Errorf("action name is required")
	}

	if action.Category == "" {
		return fmt.Errorf("action category is required")
	}

	// Update the action
	if err := r.dbManager.Update(ctx, action); err != nil {
		return fmt.Errorf("failed to update action: %w", err)
	}

	return nil
}

// UpdateName updates the name of an action
func (r *ActionRepository) UpdateName(ctx context.Context, id, name string) error {
	if id == "" {
		return fmt.Errorf("action ID is required")
	}

	if name == "" {
		return fmt.Errorf("action name is required")
	}

	action, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get action: %w", err)
	}

	action.Name = name
	return r.Update(ctx, action)
}

// UpdateDescription updates the description of an action
func (r *ActionRepository) UpdateDescription(ctx context.Context, id, description string) error {
	if id == "" {
		return fmt.Errorf("action ID is required")
	}

	action, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get action: %w", err)
	}

	action.Description = description
	return r.Update(ctx, action)
}

// UpdateCategory updates the category of an action
func (r *ActionRepository) UpdateCategory(ctx context.Context, id, category string) error {
	if id == "" {
		return fmt.Errorf("action ID is required")
	}

	if category == "" {
		return fmt.Errorf("action category is required")
	}

	action, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get action: %w", err)
	}

	action.Category = category
	return r.Update(ctx, action)
}

// Activate activates an action
func (r *ActionRepository) Activate(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("action ID is required")
	}

	action, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get action: %w", err)
	}

	if action.IsActive {
		return nil // Already active
	}

	action.IsActive = true
	return r.Update(ctx, action)
}

// Deactivate deactivates an action
func (r *ActionRepository) Deactivate(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("action ID is required")
	}

	action, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get action: %w", err)
	}

	if !action.IsActive {
		return nil // Already inactive
	}

	action.IsActive = false
	return r.Update(ctx, action)
}
