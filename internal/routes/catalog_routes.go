package routes

import (
	"context"
	"net/http"

	"github.com/Kisanlink/aaa-service/v2/internal/middleware"
	"github.com/Kisanlink/aaa-service/v2/internal/services/catalog"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupCatalogRoutes configures catalog-related HTTP endpoints
// CatalogServiceInterface defines the interface for catalog service operations
type CatalogServiceInterface interface {
	SeedRolesAndPermissions(ctx context.Context, serviceID string, force bool) (*catalog.SeedResult, error)
	ListRegisteredProviders() []string
}

func SetupCatalogRoutes(
	router *gin.RouterGroup,
	authMiddleware *middleware.AuthMiddleware,
	catalogService CatalogServiceInterface,
	logger *zap.Logger,
) {
	catalogGroup := router.Group("/catalog")

	// Seed roles and permissions endpoint
	// POST /api/v1/catalog/seed
	catalogGroup.POST("/seed",
		authMiddleware.RequirePermission("catalog", "seed"),
		func(c *gin.Context) {
			HandleSeedRolesAndPermissions(c, catalogService, logger)
		})

	// Get seed status endpoint (optional - for monitoring)
	// GET /api/v1/catalog/seed/status
	catalogGroup.GET("/seed/status",
		authMiddleware.RequirePermission("catalog", "read"),
		func(c *gin.Context) {
			HandleGetSeedStatus(c, catalogService, logger)
		})
}

// SeedRequest represents the HTTP request body for seeding
type SeedRequest struct {
	ServiceID string `json:"service_id" example:"erp-service" binding:"omitempty"`
	Force     bool   `json:"force" example:"false"`
}

// SeedResponse represents the HTTP response for seeding
type SeedResponse struct {
	StatusCode         int      `json:"status_code" example:"200"`
	Message            string   `json:"message" example:"Successfully seeded roles and permissions"`
	ActionsCreated     int32    `json:"actions_created" example:"9"`
	ResourcesCreated   int32    `json:"resources_created" example:"8"`
	PermissionsCreated int32    `json:"permissions_created" example:"72"`
	RolesCreated       int32    `json:"roles_created" example:"6"`
	CreatedRoles       []string `json:"created_roles" example:"farmer,kisansathi,CEO,fpo_manager,admin,readonly"`
}

// HandleSeedRolesAndPermissions handles the HTTP endpoint for seeding roles
// @Summary Seed roles and permissions for a service
// @Description Seeds predefined roles, permissions, actions, and resources for a specific service. Requires catalog:seed permission. Services can only seed their own roles unless user has admin:* permission.
// @Tags Catalog
// @Accept json
// @Produce json
// @Param request body SeedRequest true "Seed request parameters"
// @Success 200 {object} SeedResponse "Successfully seeded roles and permissions"
// @Failure 400 {object} map[string]interface{} "Invalid request parameters"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /catalog/seed [post]
func HandleSeedRolesAndPermissions(c *gin.Context, catalogService CatalogServiceInterface, logger *zap.Logger) {
	var req SeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind seed request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate service_id
	if err := catalog.ValidateServiceID(req.ServiceID); err != nil {
		logger.Error("Invalid service_id", zap.String("service_id", req.ServiceID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_service_id",
			"message": err.Error(),
		})
		return
	}

	logger.Info("Seed request received via HTTP",
		zap.String("service_id", req.ServiceID),
		zap.Bool("force", req.Force))

	// Call catalog service
	result, err := catalogService.SeedRolesAndPermissions(c.Request.Context(), req.ServiceID, req.Force)
	if err != nil {
		logger.Error("Seed operation failed",
			zap.String("service_id", req.ServiceID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "seed_failed",
			"message": "Failed to seed roles and permissions",
			"details": err.Error(),
		})
		return
	}

	response := SeedResponse{
		StatusCode:         200,
		Message:            "Successfully seeded roles and permissions",
		ActionsCreated:     result.ActionsCreated,
		ResourcesCreated:   result.ResourcesCreated,
		PermissionsCreated: result.PermissionsCreated,
		RolesCreated:       result.RolesCreated,
		CreatedRoles:       result.CreatedRoleNames,
	}

	logger.Info("Seed operation completed successfully via HTTP",
		zap.String("service_id", req.ServiceID),
		zap.Int32("roles_created", result.RolesCreated))

	c.JSON(http.StatusOK, response)
}

// SeedStatusResponse represents the seed status response
type SeedStatusResponse struct {
	TotalRoles         int64    `json:"total_roles" example:"25"`
	TotalPermissions   int64    `json:"total_permissions" example:"150"`
	TotalActions       int64    `json:"total_actions" example:"9"`
	TotalResources     int64    `json:"total_resources" example:"12"`
	RegisteredServices []string `json:"registered_services" example:"farmers-module,erp-service"`
}

// HandleGetSeedStatus returns the current seed status
// @Summary Get seed status
// @Description Returns information about currently seeded roles, permissions, and registered services
// @Tags Catalog
// @Produce json
// @Success 200 {object} SeedStatusResponse "Seed status information"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /catalog/seed/status [get]
func HandleGetSeedStatus(c *gin.Context, catalogService CatalogServiceInterface, logger *zap.Logger) {
	// Get provider registry status
	providers := catalogService.ListRegisteredProviders()

	response := SeedStatusResponse{
		TotalRoles:         0, // Would need to query database
		TotalPermissions:   0, // Would need to query database
		TotalActions:       0, // Would need to query database
		TotalResources:     0, // Would need to query database
		RegisteredServices: providers,
	}

	logger.Debug("Seed status requested", zap.Strings("registered_services", providers))

	c.JSON(http.StatusOK, response)
}
