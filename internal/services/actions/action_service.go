package actions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	actionRequests "github.com/Kisanlink/aaa-service/internal/entities/requests/actions"
	actionResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/actions"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	actionRepo "github.com/Kisanlink/aaa-service/internal/repositories/actions"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"go.uber.org/zap"
)

// ActionService handles business logic for Action entities
type ActionService struct {
	actionRepo actionRepo.ActionRepository
	cache      interfaces.CacheService
	logger     interfaces.Logger
	validator  interfaces.Validator
}

// NewActionService creates a new ActionService instance
func NewActionService(
	repo actionRepo.ActionRepository,
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

	// Convert to response format
	response := &actionResponses.ActionResponse{
		ID:          action.ID,
		Name:        action.Name,
		Description: action.Description,
		Category:    action.Category,
		IsStatic:    action.IsStatic,
		ServiceID:   action.ServiceID,
		Metadata:    action.Metadata,
		IsActive:    action.IsActive,
		CreatedAt:   action.CreatedAt,
		UpdatedAt:   action.UpdatedAt,
	}

	return response, nil
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
	response := &actionResponses.ActionResponse{
		ID:          action.ID,
		Name:        action.Name,
		Description: action.Description,
		Category:    action.Category,
		IsStatic:    action.IsStatic,
		ServiceID:   action.ServiceID,
		Metadata:    action.Metadata,
		IsActive:    action.IsActive,
		CreatedAt:   action.CreatedAt,
		UpdatedAt:   action.UpdatedAt,
	}

	// Cache the response
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
	cacheKey := fmt.Sprintf("action:%s", id)
	if err := s.cache.Delete(cacheKey); err != nil {
		s.logger.Warn("Failed to delete action cache", zap.Error(err))
	}

	// Convert to response format
	response := &actionResponses.ActionResponse{
		ID:          existingAction.ID,
		Name:        existingAction.Name,
		Description: existingAction.Description,
		Category:    existingAction.Category,
		IsStatic:    existingAction.IsStatic,
		ServiceID:   existingAction.ServiceID,
		Metadata:    existingAction.Metadata,
		IsActive:    existingAction.IsActive,
		CreatedAt:   existingAction.CreatedAt,
		UpdatedAt:   existingAction.UpdatedAt,
	}

	return response, nil
}

// DeleteAction deletes an action by ID
func (s *ActionService) DeleteAction(ctx context.Context, id string, deletedBy string) error {
	s.logger.Info("Deleting action", zap.String("id", id), zap.String("deleted_by", deletedBy))

	// Check if action exists
	exists, err := s.actionRepo.Exists(ctx, id)
	if err != nil {
		s.logger.Error("Failed to check action existence", zap.Error(err))
		return errors.NewInternalError(err)
	}

	if !exists {
		return errors.NewNotFoundError("action not found")
	}

	// Soft delete the action
	err = s.actionRepo.SoftDelete(ctx, id, deletedBy)
	if err != nil {
		s.logger.Error("Failed to delete action", zap.Error(err))
		return errors.NewInternalError(err)
	}

	s.logger.Info("Action deleted successfully", zap.String("action_id", id))

	// Clear cache
	cacheKey := fmt.Sprintf("action:%s", id)
	if err := s.cache.Delete(cacheKey); err != nil {
		s.logger.Warn("Failed to delete action cache", zap.Error(err))
	}

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
	actionResponseList := make([]*actionResponses.ActionResponse, len(actions))
	for i, action := range actions {
		actionResponseList[i] = &actionResponses.ActionResponse{
			ID:          action.ID,
			Name:        action.Name,
			Description: action.Description,
			Category:    action.Category,
			IsStatic:    action.IsStatic,
			ServiceID:   action.ServiceID,
			Metadata:    action.Metadata,
			IsActive:    action.IsActive,
			CreatedAt:   action.CreatedAt,
			UpdatedAt:   action.UpdatedAt,
		}
	}

	response := &actionResponses.ActionListResponse{
		Actions: actionResponseList,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}

	return response, nil
}

// GetActionsByService retrieves actions by service name
func (s *ActionService) GetActionsByService(ctx context.Context, serviceName string, limit, offset int) (*actionResponses.ActionListResponse, error) {
	s.logger.Info("Getting actions by service", zap.String("service", serviceName))

	// Get actions from repository
	actions, err := s.actionRepo.GetByServiceID(ctx, serviceName, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get actions by service", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Convert to response format
	actionResponseList := make([]*actionResponses.ActionResponse, len(actions))
	for i, action := range actions {
		actionResponseList[i] = &actionResponses.ActionResponse{
			ID:          action.ID,
			Name:        action.Name,
			Description: action.Description,
			Category:    action.Category,
			IsStatic:    action.IsStatic,
			ServiceID:   action.ServiceID,
			Metadata:    action.Metadata,
			IsActive:    action.IsActive,
			CreatedAt:   action.CreatedAt,
			UpdatedAt:   action.UpdatedAt,
		}
	}

	response := &actionResponses.ActionListResponse{
		Actions: actionResponseList,
		Total:   int64(len(actions)),
		Limit:   limit,
		Offset:  offset,
	}

	return response, nil
}
