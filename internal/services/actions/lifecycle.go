package actions

import (
	"context"

	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"go.uber.org/zap"
)

// ActivateAction activates a deactivated action
func (s *ActionService) ActivateAction(ctx context.Context, id string) error {
	s.logger.Info("Activating action", zap.String("action_id", id))

	// Get the action
	action, err := s.actionRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get action for activation", zap.Error(err))
		return errors.NewNotFoundError("action not found")
	}

	if action == nil {
		return errors.NewNotFoundError("action not found")
	}

	// Check if already active
	if action.IsActive {
		s.logger.Debug("Action is already active", zap.String("action_id", id))
		return nil
	}

	// Activate the action
	err = s.actionRepo.Activate(ctx, id)
	if err != nil {
		s.logger.Error("Failed to activate action", zap.Error(err))
		return errors.NewInternalError(err)
	}

	s.logger.Info("Action activated successfully", zap.String("action_id", id))

	// Clear cache
	s.invalidateActionCache(id)
	s.invalidateActionListCache()

	return nil
}

// DeactivateAction deactivates an active action
func (s *ActionService) DeactivateAction(ctx context.Context, id string) error {
	s.logger.Info("Deactivating action", zap.String("action_id", id))

	// Get the action
	action, err := s.actionRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get action for deactivation", zap.Error(err))
		return errors.NewNotFoundError("action not found")
	}

	if action == nil {
		return errors.NewNotFoundError("action not found")
	}

	// Protect static actions - they should typically remain active
	if action.IsStatic {
		s.logger.Warn("Attempted to deactivate static action", zap.String("action_id", id))
		return errors.NewValidationError("cannot deactivate static action", "static actions must remain active")
	}

	// Check if already inactive
	if !action.IsActive {
		s.logger.Debug("Action is already inactive", zap.String("action_id", id))
		return nil
	}

	// Deactivate the action
	err = s.actionRepo.Deactivate(ctx, id)
	if err != nil {
		s.logger.Error("Failed to deactivate action", zap.Error(err))
		return errors.NewInternalError(err)
	}

	s.logger.Info("Action deactivated successfully", zap.String("action_id", id))

	// Clear cache
	s.invalidateActionCache(id)
	s.invalidateActionListCache()

	return nil
}

// IsStaticAction checks if an action is static (built-in)
func (s *ActionService) IsStaticAction(ctx context.Context, id string) (bool, error) {
	s.logger.Debug("Checking if action is static", zap.String("action_id", id))

	// Get the action
	action, err := s.actionRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get action to check if static", zap.Error(err))
		return false, errors.NewNotFoundError("action not found")
	}

	if action == nil {
		return false, errors.NewNotFoundError("action not found")
	}

	return action.IsStatic, nil
}

// canModifyAction checks if an action can be modified
func (s *ActionService) canModifyAction(ctx context.Context, id string) (bool, error) {
	isStatic, err := s.IsStaticAction(ctx, id)
	if err != nil {
		return false, err
	}

	// Static actions have restrictions
	if isStatic {
		return false, errors.NewValidationError("cannot modify static action", "static action protection")
	}

	return true, nil
}

// canDeleteAction checks if an action can be deleted
func (s *ActionService) canDeleteAction(ctx context.Context, id string) (bool, error) {
	// Static actions cannot be deleted
	isStatic, err := s.IsStaticAction(ctx, id)
	if err != nil {
		return false, err
	}

	if isStatic {
		return false, errors.NewValidationError("cannot delete static action", "static actions are protected")
	}

	// Additional checks could be added here
	// For example, check if action is currently in use by permissions

	return true, nil
}
