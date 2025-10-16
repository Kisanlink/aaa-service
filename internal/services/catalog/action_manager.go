package catalog

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/actions"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"go.uber.org/zap"
)

// ActionManager handles action-related operations for the catalog service
type ActionManager struct {
	actionRepo *actions.ActionRepository
	logger     *zap.Logger
}

// NewActionManager creates a new action manager
func NewActionManager(actionRepo *actions.ActionRepository, logger *zap.Logger) *ActionManager {
	return &ActionManager{
		actionRepo: actionRepo,
		logger:     logger,
	}
}

// UpsertActions creates or updates actions using UPSERT pattern
// Returns the count of actions created/updated
func (am *ActionManager) UpsertActions(ctx context.Context, actionDefs []ActionDefinition, force bool) (int32, error) {
	var count int32

	for _, actionDef := range actionDefs {
		// Check if action already exists
		existing, err := am.getActionByName(ctx, actionDef.Name)

		if err == nil && existing != nil {
			// Action exists
			if !force {
				// Skip if not forcing update
				am.logger.Debug("Action already exists, skipping",
					zap.String("action", actionDef.Name))
				continue
			}

			// Update existing action
			existing.Description = actionDef.Description
			existing.Category = actionDef.Category
			existing.IsStatic = actionDef.IsStatic

			if err := am.actionRepo.Update(ctx, existing); err != nil {
				return count, fmt.Errorf("failed to update action %s: %w", actionDef.Name, err)
			}

			am.logger.Debug("Action updated",
				zap.String("action", actionDef.Name))
			count++
		} else {
			// Create new action
			action := models.NewActionWithCategory(
				actionDef.Name,
				actionDef.Description,
				actionDef.Category,
			)
			action.IsStatic = actionDef.IsStatic

			if err := am.actionRepo.Create(ctx, action); err != nil {
				return count, fmt.Errorf("failed to create action %s: %w", actionDef.Name, err)
			}

			am.logger.Debug("Action created",
				zap.String("action", actionDef.Name),
				zap.String("id", action.ID))
			count++
		}
	}

	return count, nil
}

// GetActionByName retrieves an action by name
func (am *ActionManager) GetActionByName(ctx context.Context, name string) (*models.Action, error) {
	return am.getActionByName(ctx, name)
}

// getActionByName is a private helper to retrieve action by name
func (am *ActionManager) getActionByName(ctx context.Context, name string) (*models.Action, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Build()

	actions, err := am.actionRepo.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find action by name: %w", err)
	}

	if len(actions) == 0 {
		return nil, nil
	}

	return actions[0], nil
}

// GetAllActions retrieves all actions
func (am *ActionManager) GetAllActions(ctx context.Context) ([]*models.Action, error) {
	filter := base.NewFilterBuilder().Build()
	return am.actionRepo.Find(ctx, filter)
}
