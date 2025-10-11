package permissions

import (
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	permissionService "github.com/Kisanlink/aaa-service/internal/services/permissions"
	roleAssignmentService "github.com/Kisanlink/aaa-service/internal/services/role_assignments"
	"go.uber.org/zap"
)

// PermissionHandler handles permission-related HTTP requests
type PermissionHandler struct {
	permissionService     permissionService.ServiceInterface
	roleAssignmentService roleAssignmentService.ServiceInterface
	validator             interfaces.Validator
	responder             interfaces.Responder
	logger                *zap.Logger
}

// NewPermissionHandler creates a new PermissionHandler instance
func NewPermissionHandler(
	permissionService permissionService.ServiceInterface,
	roleAssignmentService roleAssignmentService.ServiceInterface,
	validator interfaces.Validator,
	responder interfaces.Responder,
	logger *zap.Logger,
) *PermissionHandler {
	return &PermissionHandler{
		permissionService:     permissionService,
		roleAssignmentService: roleAssignmentService,
		validator:             validator,
		responder:             responder,
		logger:                logger.Named("permission_handler"),
	}
}

// getRequestID extracts the request ID from the Gin context
func (h *PermissionHandler) getRequestID(c interface {
	Get(string) (interface{}, bool)
}) string {
	if reqID, exists := c.Get("request_id"); exists {
		if id, ok := reqID.(string); ok {
			return id
		}
	}
	return ""
}
