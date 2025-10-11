package actions

import (
	"context"
	"fmt"

	actionResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/actions"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"go.uber.org/zap"
)

// GetActionsByCategory retrieves actions by category with pagination
func (s *ActionService) GetActionsByCategory(ctx context.Context, category string, limit, offset int) (*actionResponses.ActionListResponse, error) {
	s.logger.Info("Getting actions by category", zap.String("category", category))

	// Try cache first
	cacheKey := fmt.Sprintf("actions:category:%s:limit:%d:offset:%d", category, limit, offset)
	if cached, exists := s.cache.Get(cacheKey); exists && cached != nil {
		if response, ok := cached.(*actionResponses.ActionListResponse); ok {
			s.logger.Debug("Actions by category retrieved from cache", zap.String("category", category))
			return response, nil
		}
	}

	// Get actions from repository
	actions, err := s.actionRepo.GetByCategory(ctx, category, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get actions by category", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Create response
	response := s.toActionListResponse(actions, int64(len(actions)), limit, offset)

	// Cache the response (TTL: 5 minutes)
	if err := s.cache.Set(cacheKey, response, 300); err != nil {
		s.logger.Warn("Failed to cache actions by category", zap.Error(err))
	}

	return response, nil
}

// GetActionsByService retrieves actions by service ID with pagination
func (s *ActionService) GetActionsByService(ctx context.Context, serviceID string, limit, offset int) (*actionResponses.ActionListResponse, error) {
	s.logger.Info("Getting actions by service", zap.String("service_id", serviceID))

	// Try cache first
	cacheKey := fmt.Sprintf("actions:service:%s:limit:%d:offset:%d", serviceID, limit, offset)
	if cached, exists := s.cache.Get(cacheKey); exists && cached != nil {
		if response, ok := cached.(*actionResponses.ActionListResponse); ok {
			s.logger.Debug("Actions by service retrieved from cache", zap.String("service_id", serviceID))
			return response, nil
		}
	}

	// Get actions from repository
	actions, err := s.actionRepo.GetByServiceID(ctx, serviceID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get actions by service", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Create response
	response := s.toActionListResponse(actions, int64(len(actions)), limit, offset)

	// Cache the response (TTL: 5 minutes)
	if err := s.cache.Set(cacheKey, response, 300); err != nil {
		s.logger.Warn("Failed to cache actions by service", zap.Error(err))
	}

	return response, nil
}

// GetStaticActions retrieves all static (built-in) actions
func (s *ActionService) GetStaticActions(ctx context.Context, limit, offset int) (*actionResponses.ActionListResponse, error) {
	s.logger.Info("Getting static actions", zap.Int("limit", limit), zap.Int("offset", offset))

	// Try cache first
	cacheKey := fmt.Sprintf("actions:static:limit:%d:offset:%d", limit, offset)
	if cached, exists := s.cache.Get(cacheKey); exists && cached != nil {
		if response, ok := cached.(*actionResponses.ActionListResponse); ok {
			s.logger.Debug("Static actions retrieved from cache")
			return response, nil
		}
	}

	// Get static actions from repository
	actions, err := s.actionRepo.GetStaticActions(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get static actions", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Create response
	response := s.toActionListResponse(actions, int64(len(actions)), limit, offset)

	// Cache the response (TTL: 10 minutes - static actions change rarely)
	if err := s.cache.Set(cacheKey, response, 600); err != nil {
		s.logger.Warn("Failed to cache static actions", zap.Error(err))
	}

	return response, nil
}

// GetDynamicActions retrieves all dynamic (service-defined) actions
func (s *ActionService) GetDynamicActions(ctx context.Context, limit, offset int) (*actionResponses.ActionListResponse, error) {
	s.logger.Info("Getting dynamic actions", zap.Int("limit", limit), zap.Int("offset", offset))

	// Try cache first
	cacheKey := fmt.Sprintf("actions:dynamic:limit:%d:offset:%d", limit, offset)
	if cached, exists := s.cache.Get(cacheKey); exists && cached != nil {
		if response, ok := cached.(*actionResponses.ActionListResponse); ok {
			s.logger.Debug("Dynamic actions retrieved from cache")
			return response, nil
		}
	}

	// Get dynamic actions from repository
	actions, err := s.actionRepo.GetDynamicActions(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get dynamic actions", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Create response
	response := s.toActionListResponse(actions, int64(len(actions)), limit, offset)

	// Cache the response (TTL: 5 minutes)
	if err := s.cache.Set(cacheKey, response, 300); err != nil {
		s.logger.Warn("Failed to cache dynamic actions", zap.Error(err))
	}

	return response, nil
}

// GetActiveActions retrieves all active actions
func (s *ActionService) GetActiveActions(ctx context.Context, limit, offset int) (*actionResponses.ActionListResponse, error) {
	s.logger.Info("Getting active actions", zap.Int("limit", limit), zap.Int("offset", offset))

	// Try cache first
	cacheKey := fmt.Sprintf("actions:active:limit:%d:offset:%d", limit, offset)
	if cached, exists := s.cache.Get(cacheKey); exists && cached != nil {
		if response, ok := cached.(*actionResponses.ActionListResponse); ok {
			s.logger.Debug("Active actions retrieved from cache")
			return response, nil
		}
	}

	// Get active actions from repository
	actions, err := s.actionRepo.GetActiveActions(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get active actions", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Create response
	response := s.toActionListResponse(actions, int64(len(actions)), limit, offset)

	// Cache the response (TTL: 5 minutes)
	if err := s.cache.Set(cacheKey, response, 300); err != nil {
		s.logger.Warn("Failed to cache active actions", zap.Error(err))
	}

	return response, nil
}
