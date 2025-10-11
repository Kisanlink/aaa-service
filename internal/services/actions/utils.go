package actions

import (
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	actionResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/actions"
	"go.uber.org/zap"
)

// toActionResponse converts a model to response DTO
func (s *ActionService) toActionResponse(action *models.Action) *actionResponses.ActionResponse {
	return &actionResponses.ActionResponse{
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

// toActionListResponse converts a list of models to list response DTO
func (s *ActionService) toActionListResponse(actions []*models.Action, total int64, limit, offset int) *actionResponses.ActionListResponse {
	actionResponseList := make([]*actionResponses.ActionResponse, len(actions))
	for i, action := range actions {
		actionResponseList[i] = s.toActionResponse(action)
	}

	return &actionResponses.ActionListResponse{
		Actions: actionResponseList,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}
}

// invalidateActionCache clears cache for a specific action
func (s *ActionService) invalidateActionCache(id string) {
	cacheKey := fmt.Sprintf("action:%s", id)
	if err := s.cache.Delete(cacheKey); err != nil {
		s.logger.Warn("Failed to delete action cache", zap.String("action_id", id), zap.Error(err))
	}
}

// invalidateActionListCache clears list and query caches
func (s *ActionService) invalidateActionListCache() {
	// Clear various list caches
	_ = s.cache.Delete("actions:list:*")
	_ = s.cache.Delete("actions:static:*")
	_ = s.cache.Delete("actions:dynamic:*")
	_ = s.cache.Delete("actions:active:*")
	_ = s.cache.Delete("actions:category:*")
}
