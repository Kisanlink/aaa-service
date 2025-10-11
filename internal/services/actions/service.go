package actions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	actionRequests "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/actions"
	actionResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/actions"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	actionRepo "github.com/Kisanlink/aaa-service/v2/internal/repositories/actions"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"go.uber.org/zap"
)

// ActionService defines the interface for action management operations
type ActionServiceInterface interface {
	// Basic CRUD
	CreateAction(ctx context.Context, req *actionRequests.CreateActionRequest) (*actionResponses.ActionResponse, error)
	GetAction(ctx context.Context, id string) (*actionResponses.ActionResponse, error)
	UpdateAction(ctx context.Context, id string, req *actionRequests.UpdateActionRequest) (*actionResponses.ActionResponse, error)
	DeleteAction(ctx context.Context, id string, deletedBy string) error
	ListActions(ctx context.Context, limit, offset int) (*actionResponses.ActionListResponse, error)

	// Lifecycle operations
	ActivateAction(ctx context.Context, id string) error
	DeactivateAction(ctx context.Context, id string) error
	IsStaticAction(ctx context.Context, id string) (bool, error)

	// Query operations
	GetActionsByCategory(ctx context.Context, category string, limit, offset int) (*actionResponses.ActionListResponse, error)
	GetActionsByService(ctx context.Context, serviceName string, limit, offset int) (*actionResponses.ActionListResponse, error)
	GetStaticActions(ctx context.Context, limit, offset int) (*actionResponses.ActionListResponse, error)
	GetDynamicActions(ctx context.Context, limit, offset int) (*actionResponses.ActionListResponse, error)
	GetActiveActions(ctx context.Context, limit, offset int) (*actionResponses.ActionListResponse, error)
}

// ActionService handles business logic for Action entities
type ActionService struct {
	actionRepo *actionRepo.ActionRepository
	cache      interfaces.CacheService
	logger     interfaces.Logger
	validator  interfaces.Validator
}

// NewActionService creates a new ActionService instance
func NewActionService(
	repo *actionRepo.ActionRepository,
	cache interfaces.CacheService,
	logger interfaces.Logger,
	validator interfaces.Validator,
) *ActionService {
	return &ActionService{
		actionRepo: repo,
		cache:      cache,
		logger:     logger,
		validator:  validator,
	}
}

// CreateAction creates a new action with proper validation and business logic
func (s *ActionService) CreateAction(ctx context.Context, req *actionRequests.CreateActionRequest) (*actionResponses.ActionResponse, error) {
	s.logger.Info("Creating new action", zap.String("name", req.Name))

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.Error("Action creation validation failed", zap.Error(err))
		return nil, errors.NewValidationError("invalid action data", err.Error())
	}

	// Check if action already exists by name
	existingAction, err := s.actionRepo.GetByName(ctx, req.Name)
	if err == nil && existingAction != nil {
		s.logger.Warn("Action already exists with name", zap.String("name", req.Name))
		return nil, errors.NewConflictError("action with this name already exists")
	}

	// Create action model
	action := &models.Action{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		IsStatic:    req.IsStatic,
		ServiceID:   req.ServiceID,
		Metadata:    req.Metadata,
		IsActive:    req.IsActive,
	}

	// Save action to repository
	err = s.actionRepo.Create(ctx, action)
	if err != nil {
		s.logger.Error("Failed to create action in repository", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	s.logger.Info("Action created successfully", zap.String("action_id", action.ID))

	// Clear list cache
	s.invalidateActionListCache()

	// Convert to response format
	return s.toActionResponse(action), nil
}

// GetAction retrieves an action by ID
func (s *ActionService) GetAction(ctx context.Context, id string) (*actionResponses.ActionResponse, error) {
	s.logger.Info("Getting action", zap.String("id", id))

	// Try to get from cache first
	cacheKey := fmt.Sprintf("action:%s", id)
	if cached, exists := s.cache.Get(cacheKey); exists && cached != nil {
		if action, ok := cached.(*actionResponses.ActionResponse); ok {
			s.logger.Debug("Action retrieved from cache", zap.String("id", id))
			return action, nil
		}
	}

	// Get from repository
	action, err := s.actionRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get action from repository", zap.Error(err))
		return nil, errors.NewNotFoundError("action not found")
	}

	if action == nil {
		return nil, errors.NewNotFoundError("action not found")
	}

	// Convert to response format
	response := s.toActionResponse(action)

	// Cache the response (TTL: 5 minutes)
	if err := s.cache.Set(cacheKey, response, 300); err != nil {
		s.logger.Warn("Failed to cache action response", zap.Error(err))
	}

	return response, nil
}

// UpdateAction updates an existing action
func (s *ActionService) UpdateAction(ctx context.Context, id string, req *actionRequests.UpdateActionRequest) (*actionResponses.ActionResponse, error) {
	s.logger.Info("Updating action", zap.String("id", id))

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.Error("Action update validation failed", zap.Error(err))
		return nil, errors.NewValidationError("invalid action data", err.Error())
	}

	// Get existing action
	existingAction, err := s.actionRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get action for update", zap.Error(err))
		return nil, errors.NewNotFoundError("action not found")
	}

	if existingAction == nil {
		return nil, errors.NewNotFoundError("action not found")
	}

	// Check if it's a static action and prevent certain updates
	if existingAction.IsStatic {
		// Static actions should not be made non-static or change certain core properties
		if req.IsStatic != nil && !*req.IsStatic {
			return nil, errors.NewValidationError("cannot change static action to dynamic", "static action protection")
		}
	}

	// Update fields if provided
	if req.Name != nil {
		existingAction.Name = *req.Name
	}
	if req.Description != nil {
		existingAction.Description = *req.Description
	}
	if req.Category != nil {
		existingAction.Category = *req.Category
	}
	if req.IsStatic != nil {
		existingAction.IsStatic = *req.IsStatic
	}
	if req.ServiceID != nil {
		existingAction.ServiceID = req.ServiceID
	}
	if req.Metadata != nil {
		existingAction.Metadata = req.Metadata
	}
	if req.IsActive != nil {
		existingAction.IsActive = *req.IsActive
	}

	// Save updated action
	err = s.actionRepo.Update(ctx, existingAction)
	if err != nil {
		s.logger.Error("Failed to update action in repository", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	s.logger.Info("Action updated successfully", zap.String("action_id", id))

	// Clear cache
	s.invalidateActionCache(id)
	s.invalidateActionListCache()

	return s.toActionResponse(existingAction), nil
}

// DeleteAction deletes an action by ID with static action protection
func (s *ActionService) DeleteAction(ctx context.Context, id string, deletedBy string) error {
	s.logger.Info("Deleting action", zap.String("id", id), zap.String("deleted_by", deletedBy))

	// Check if action exists and is static
	action, err := s.actionRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get action for deletion", zap.Error(err))
		return errors.NewNotFoundError("action not found")
	}

	// Protect static actions from deletion
	if action.IsStatic {
		s.logger.Warn("Attempted to delete static action", zap.String("action_id", id))
		return errors.NewValidationError("cannot delete static action", "static actions are protected")
	}

	// Soft delete the action
	err = s.actionRepo.SoftDelete(ctx, id, deletedBy)
	if err != nil {
		s.logger.Error("Failed to delete action", zap.Error(err))
		return errors.NewInternalError(err)
	}

	s.logger.Info("Action deleted successfully", zap.String("action_id", id))

	// Clear cache
	s.invalidateActionCache(id)
	s.invalidateActionListCache()

	return nil
}

// ListActions retrieves actions with pagination
func (s *ActionService) ListActions(ctx context.Context, limit, offset int) (*actionResponses.ActionListResponse, error) {
	s.logger.Info("Listing actions", zap.Int("limit", limit), zap.Int("offset", offset))

	// Get actions from repository
	filter := base.NewFilterBuilder().Limit(limit, offset).Build()
	actions, err := s.actionRepo.List(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to list actions", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Get total count
	total, err := s.actionRepo.Count(ctx, base.NewFilter())
	if err != nil {
		s.logger.Error("Failed to count actions", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Convert to response format
	return s.toActionListResponse(actions, total, limit, offset), nil
}
