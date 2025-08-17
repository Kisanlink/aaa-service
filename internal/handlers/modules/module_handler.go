package modules

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/internal/entities/requests"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/internal/services"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"github.com/Kisanlink/aaa-service/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ModuleHandler handles module-related HTTP requests
type ModuleHandler struct {
	moduleService *services.ModuleService
	logger        *zap.Logger
	validator     interfaces.Validator
	responder     *utils.Responder
}

// NewModuleHandler creates a new ModuleHandler instance
func NewModuleHandler(moduleService *services.ModuleService, logger *zap.Logger, validator interfaces.Validator, responder *utils.Responder) *ModuleHandler {
	return &ModuleHandler{
		moduleService: moduleService,
		logger:        logger,
		validator:     validator,
		responder:     responder,
	}
}

// RegisterModule handles POST /api/v2/modules/register
// @Summary Register a new module
// @Description Register a complete module with actions, roles, resources, and permissions
// @Tags modules
// @Accept json
// @Produce json
// @Param module body requests.ModuleRegistrationRequest true "Module registration request"
// @Success 201 {object} responses.ModuleDetailResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/modules/register [post]
func (h *ModuleHandler) RegisterModule(c *gin.Context) {
	h.logger.Info("Processing module registration request")

	var req requests.ModuleRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind module registration request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Additional validation using validator service
	if err := h.validator.ValidateStruct(&req); err != nil {
		h.logger.Error("Module registration validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Register the module
	response, err := h.moduleService.RegisterModule(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to register module", zap.Error(err))

		if validationErr, ok := err.(*errors.ValidationError); ok {
			h.responder.SendValidationError(c, []string{validationErr.Error()})
			return
		}

		if conflictErr, ok := err.(*errors.ConflictError); ok {
			h.responder.SendError(c, http.StatusConflict, conflictErr.Error(), conflictErr)
			return
		}

		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Module registered successfully",
		zap.String("service_name", req.ServiceName),
		zap.String("service_id", response.Data.ServiceID),
	)

	h.responder.SendSuccess(c, http.StatusCreated, response)
}

// GetModule handles GET /api/v2/modules/{service_name}
// @Summary Get module information
// @Description Get detailed information about a registered module
// @Tags modules
// @Produce json
// @Param service_name path string true "Service name"
// @Success 200 {object} responses.ModuleDetailResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/modules/{service_name} [get]
func (h *ModuleHandler) GetModule(c *gin.Context) {
	serviceName := c.Param("service_name")
	if serviceName == "" {
		h.responder.SendValidationError(c, []string{"service_name is required"})
		return
	}

	h.logger.Info("Getting module information", zap.String("service_name", serviceName))

	response, err := h.moduleService.GetModuleInfo(c.Request.Context(), serviceName)
	if err != nil {
		h.logger.Error("Failed to get module information", zap.String("service_name", serviceName), zap.Error(err))

		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusNotFound, notFoundErr.Error(), notFoundErr)
			return
		}

		h.responder.SendInternalError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// ListModules handles GET /api/v2/modules
// @Summary List all registered modules
// @Description Get a list of all registered modules
// @Tags modules
// @Produce json
// @Success 200 {object} responses.ModuleListResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/modules [get]
func (h *ModuleHandler) ListModules(c *gin.Context) {
	h.logger.Info("Listing all modules")

	modules, err := h.moduleService.ListModules(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list modules", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, map[string]interface{}{
		"modules": modules,
		"count":   len(modules),
	})
}

// ModuleHealthCheck handles GET /api/v2/modules/{service_name}/health
// @Summary Check module health
// @Description Check if a module is healthy and operational
// @Tags modules
// @Produce json
// @Param service_name path string true "Service name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/modules/{service_name}/health [get]
func (h *ModuleHandler) ModuleHealthCheck(c *gin.Context) {
	serviceName := c.Param("service_name")
	if serviceName == "" {
		h.responder.SendValidationError(c, []string{"service_name is required"})
		return
	}

	h.logger.Info("Checking module health", zap.String("service_name", serviceName))

	// Get module info to verify it exists
	moduleInfo, err := h.moduleService.GetModuleInfo(c.Request.Context(), serviceName)
	if err != nil {
		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusNotFound, notFoundErr.Error(), notFoundErr)
			return
		}

		h.responder.SendInternalError(c, err)
		return
	}

	healthStatus := map[string]interface{}{
		"service_name":    serviceName,
		"service_id":      moduleInfo.Data.ServiceID,
		"status":          "healthy",
		"actions_count":   len(moduleInfo.Data.Actions),
		"roles_count":     len(moduleInfo.Data.Roles),
		"resources_count": len(moduleInfo.Data.Resources),
		"checked_at":      gin.H{"timestamp": gin.H{}}, // This would be filled with actual timestamp
	}

	h.responder.SendSuccess(c, http.StatusOK, healthStatus)
}
