package permissions

import (
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	permissionService "github.com/Kisanlink/aaa-service/internal/services/permissions"
)

// EvaluationResponse represents the result of a permission evaluation
// @Description Response structure for permission evaluation
type EvaluationResponse struct {
	Success   bool            `json:"success" example:"true"`
	Message   string          `json:"message" example:"Permission evaluated successfully"`
	Data      *EvaluationData `json:"data"`
	Timestamp time.Time       `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string          `json:"request_id" example:"req_abc123"`
}

// EvaluationData contains the permission evaluation result
type EvaluationData struct {
	Allowed          bool        `json:"allowed" example:"true"`
	Reason           string      `json:"reason" example:"Permission granted through role 'admin'"`
	EffectiveRoles   []*RoleInfo `json:"effective_roles,omitempty"`
	CacheHit         bool        `json:"cache_hit" example:"false"`
	EvaluationTimeMS int64       `json:"evaluation_time_ms" example:"5"`
	EvaluatedAt      time.Time   `json:"evaluated_at" example:"2024-01-01T00:00:00Z"`
}

// RoleInfo contains basic role information
type RoleInfo struct {
	ID          string `json:"id" example:"ROLE_abc123"`
	Name        string `json:"name" example:"admin"`
	Description string `json:"description,omitempty" example:"Administrator role"`
}

// NewEvaluationResponse creates a new EvaluationResponse from an EvaluationResult
func NewEvaluationResponse(
	result *permissionService.EvaluationResult,
	requestID string,
) *EvaluationResponse {
	if result == nil {
		return &EvaluationResponse{
			Success:   false,
			Message:   "Permission evaluation failed",
			Timestamp: time.Now(),
			RequestID: requestID,
		}
	}

	// Convert effective roles to RoleInfo
	effectiveRoles := make([]*RoleInfo, 0, len(result.EffectiveRoles))
	for _, role := range result.EffectiveRoles {
		effectiveRoles = append(effectiveRoles, &RoleInfo{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
		})
	}

	return &EvaluationResponse{
		Success: true,
		Message: "Permission evaluated successfully",
		Data: &EvaluationData{
			Allowed:          result.Allowed,
			Reason:           result.Reason,
			EffectiveRoles:   effectiveRoles,
			CacheHit:         result.CacheHit,
			EvaluationTimeMS: result.EvaluationTime.Milliseconds(),
			EvaluatedAt:      result.EvaluatedAt,
		},
		Timestamp: time.Now(),
		RequestID: requestID,
	}
}

// NewRoleInfo creates RoleInfo from a Role model
func NewRoleInfo(role *models.Role) *RoleInfo {
	if role == nil {
		return nil
	}
	return &RoleInfo{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
	}
}
