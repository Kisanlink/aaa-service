package routes

import (
	"net/http"
	"strconv"

	"github.com/Kisanlink/aaa-service/middleware"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RouteHandlers contains all handlers needed for AAA routes
type RouteHandlers struct {
	AuthService          *services.AuthService
	AuthorizationService *services.AuthorizationService
	AuditService         *services.AuditService
	AuthMiddleware       *middleware.AuthMiddleware
	Logger               *zap.Logger
}

// SetupAAA configures all AAA routes with proper authentication and authorization
func SetupAAA(router *gin.Engine, handlers RouteHandlers) {
	// Public routes (no authentication required)
	publicAPI := router.Group("/api/v2")
	{
		// Authentication endpoints
		auth := publicAPI.Group("/auth")
		{
			auth.POST("/login", createLoginHandler(handlers.AuthService, handlers.Logger))
			auth.POST("/register", createRegisterHandler(handlers.AuthService, handlers.Logger))
			auth.POST("/refresh", createRefreshHandler(handlers.AuthService, handlers.Logger))
		}

		// Health check
		publicAPI.GET("/health", createHealthHandler(handlers.Logger))
	}

	// Protected routes (authentication and authorization required)
	protectedAPI := router.Group("/api/v2")
	protectedAPI.Use(handlers.AuthMiddleware.HTTPAuthMiddleware())
	{
		// Authentication routes
		auth := protectedAPI.Group("/auth")
		{
			auth.POST("/logout", createLogoutHandler(handlers.AuthService, handlers.Logger))
		}

		// User management routes
		users := protectedAPI.Group("/users")
		{
			users.GET("", handlers.AuthMiddleware.RequirePermission("user", "read"), createGetUsersHandler(handlers.Logger))
			users.GET("/:id", handlers.AuthMiddleware.RequirePermission("user", "view"), createGetUserHandler(handlers.Logger))
			users.PUT("/:id", handlers.AuthMiddleware.RequirePermission("user", "update"), createUpdateUserHandler(handlers.Logger))
			users.DELETE("/:id", handlers.AuthMiddleware.RequirePermission("user", "delete"), createDeleteUserHandler(handlers.Logger))
		}

		// Role management routes
		roles := protectedAPI.Group("/roles")
		{
			roles.GET("", handlers.AuthMiddleware.RequirePermission("role", "read"), createGetRolesHandler(handlers.Logger))
			roles.POST("", handlers.AuthMiddleware.RequirePermission("role", "create"), createCreateRoleHandler(handlers.Logger))
			roles.GET("/:id", handlers.AuthMiddleware.RequirePermission("role", "view"), createGetRoleHandler(handlers.Logger))
			roles.PUT("/:id", handlers.AuthMiddleware.RequirePermission("role", "update"), createUpdateRoleHandler(handlers.Logger))
			roles.DELETE("/:id", handlers.AuthMiddleware.RequirePermission("role", "delete"), createDeleteRoleHandler(handlers.Logger))
		}

		// Permission management routes
		permissions := protectedAPI.Group("/permissions")
		{
			permissions.GET("", handlers.AuthMiddleware.RequirePermission("permission", "read"), createGetPermissionsHandler(handlers.Logger))
			permissions.POST("", handlers.AuthMiddleware.RequirePermission("permission", "create"), createCreatePermissionHandler(handlers.Logger))
		}

		// Authorization routes
		authz := protectedAPI.Group("/authz")
		{
			authz.POST("/check", createCheckPermissionHandler(handlers.AuthorizationService, handlers.Logger))
			authz.POST("/bulk-check", createBulkCheckPermissionHandler(handlers.AuthorizationService, handlers.Logger))
			authz.GET("/user/:id/permissions", createGetUserPermissionsHandler(handlers.AuthorizationService, handlers.Logger))
		}

		// Audit routes
		audit := protectedAPI.Group("/audit")
		audit.Use(handlers.AuthMiddleware.RequirePermission("audit_log", "read"))
		{
			audit.GET("/logs", createGetAuditLogsHandler(handlers.AuditService, handlers.Logger))
			audit.GET("/user/:id/trail", createGetUserAuditTrailHandler(handlers.AuditService, handlers.Logger))
			audit.GET("/resource/:type/:id/trail", createGetResourceAuditTrailHandler(handlers.AuditService, handlers.Logger))
			audit.GET("/security-events", createGetSecurityEventsHandler(handlers.AuditService, handlers.Logger))
			audit.GET("/statistics", createGetAuditStatisticsHandler(handlers.AuditService, handlers.Logger))
		}

		// Admin routes
		admin := protectedAPI.Group("/admin")
		admin.Use(handlers.AuthMiddleware.RequireRole("admin"))
		{
			admin.POST("/grant-permission", createGrantPermissionHandler(handlers.AuthorizationService, handlers.Logger))
			admin.POST("/revoke-permission", createRevokePermissionHandler(handlers.AuthorizationService, handlers.Logger))
			admin.POST("/assign-role", createAssignRoleHandler(handlers.AuthorizationService, handlers.Logger))
			admin.POST("/remove-role", createRemoveRoleHandler(handlers.AuthorizationService, handlers.Logger))
			admin.POST("/archive-logs", createArchiveLogsHandler(handlers.AuditService, handlers.Logger))
		}
	}
}

// Authentication handlers

// LoginV2 handles POST /v2/auth/login
// @Summary User login (V2)
// @Description Authenticate user with username and password and MFA
// @Tags authentication
// @Accept json
// @Produce json
// @Param credentials body SwaggerLoginRequest true "Login credentials"
// @Success 200 {object} SwaggerLoginResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/auth/login [post]
func createLoginHandler(authService *services.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
			return
		}

		response, err := authService.Login(c.Request.Context(), &req)
		if err != nil {
			logger.Error("Login failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// RegisterV2 handles POST /v2/auth/register
// @Summary User registration (V2)
// @Description Register a new user account with enhanced validation
// @Tags authentication
// @Accept json
// @Produce json
// @Param user body SwaggerRegisterRequest true "Registration data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/auth/register [post]
func createRegisterHandler(authService *services.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
			return
		}

		response, err := authService.Register(c.Request.Context(), &req)
		if err != nil {
			logger.Error("Registration failed", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "registration failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, response)
	}
}

// RefreshTokenV2 handles POST /v2/auth/refresh
// @Summary Refresh access token (V2)
// @Description Refresh access token using refresh token
// @Tags authentication
// @Accept json
// @Produce json
// @Param token body SwaggerRefreshTokenRequest true "Refresh token"
// @Success 200 {object} SwaggerLoginResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/auth/refresh [post]
func createRefreshHandler(authService *services.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
			return
		}

		response, err := authService.RefreshToken(c.Request.Context(), req.RefreshToken)
		if err != nil {
			logger.Error("Token refresh failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token refresh failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// LogoutV2 handles POST /v2/auth/logout
// @Summary User logout (V2)
// @Description Logout user and invalidate tokens
// @Tags authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/auth/logout [post]
func createLogoutHandler(authService *services.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error", "message": "user context not found"})
			return
		}

		err := authService.Logout(c.Request.Context(), userID.(string))
		if err != nil {
			logger.Error("Logout failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "logout failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
	}
}

// Health handler

// HealthCheckV2 handles GET /v2/health
// @Summary Health check (V2)
// @Description Basic health check for the AAA service
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v2/health [get]
func createHealthHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "aaa-service",
			"version": "2.0",
		})
	}
}

// User management handlers (placeholder implementations)

// GetUsersV2 handles GET /v2/users
// @Summary List all users (V2)
// @Description Get a list of all users with pagination
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/users [get]
func createGetUsersHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented", "message": "GetUsers not implemented yet"})
	}
}

// GetUserV2 handles GET /v2/users/:id
// @Summary Get user by ID (V2)
// @Description Get detailed information about a specific user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/users/{id} [get]
func createGetUserHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented", "message": "GetUser not implemented yet", "user_id": userID})
	}
}

// UpdateUserV2 handles PUT /v2/users/:id
// @Summary Update user (V2)
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param user body map[string]interface{} true "User update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/users/{id} [put]
func createUpdateUserHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented", "message": "UpdateUser not implemented yet", "user_id": userID})
	}
}

func createDeleteUserHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented", "message": "DeleteUser not implemented yet", "user_id": userID})
	}
}

// Role management handlers (placeholder implementations)

func createGetRolesHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented", "message": "GetRoles not implemented yet"})
	}
}

func createCreateRoleHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented", "message": "CreateRole not implemented yet"})
	}
}

func createGetRoleHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID := c.Param("id")
		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented", "message": "GetRole not implemented yet", "role_id": roleID})
	}
}

func createUpdateRoleHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID := c.Param("id")
		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented", "message": "UpdateRole not implemented yet", "role_id": roleID})
	}
}

func createDeleteRoleHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID := c.Param("id")
		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented", "message": "DeleteRole not implemented yet", "role_id": roleID})
	}
}

// Permission management handlers (placeholder implementations)

func createGetPermissionsHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented", "message": "GetPermissions not implemented yet"})
	}
}

func createCreatePermissionHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented", "message": "CreatePermission not implemented yet"})
	}
}

// Authorization handlers

// CheckPermissionV2 handles POST /v2/authz/check
// @Summary Check user permission (V2)
// @Description Check if a user has permission to perform an action on a resource
// @Tags authorization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param permission body services.Permission true "Permission check data"
// @Success 200 {object} services.PermissionResult
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/authz/check [post]
func createCheckPermissionHandler(authzService *services.AuthorizationService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.Permission
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
			return
		}

		result, err := authzService.CheckPermission(c.Request.Context(), &req)
		if err != nil {
			logger.Error("Permission check failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "permission check failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func createBulkCheckPermissionHandler(authzService *services.AuthorizationService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.BulkPermissionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
			return
		}

		result, err := authzService.CheckBulkPermissions(c.Request.Context(), &req)
		if err != nil {
			logger.Error("Bulk permission check failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "bulk permission check failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func createGetUserPermissionsHandler(authzService *services.AuthorizationService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		resourceType := c.Query("resource_type")
		if resourceType == "" {
			resourceType = "user" // default
		}

		permissions, err := authzService.GetUserPermissions(c.Request.Context(), userID, resourceType)
		if err != nil {
			logger.Error("Get user permissions failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "get user permissions failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":     userID,
			"permissions": permissions,
		})
	}
}

// Audit handlers

// GetAuditLogsV2 handles GET /v2/audit/logs
// @Summary Get audit logs (V2)
// @Description Retrieve audit logs with filtering and pagination
// @Tags audit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(50)
// @Param user_id query string false "Filter by user ID"
// @Param action query string false "Filter by action"
// @Param resource query string false "Filter by resource type"
// @Success 200 {object} SwaggerAuditQueryResult
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v2/audit/logs [get]
func createGetAuditLogsHandler(auditService *services.AuditService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

		query := &services.AuditQuery{
			UserID:   c.Query("user_id"),
			Action:   c.Query("action"),
			Resource: c.Query("resource"),
			Page:     page,
			PerPage:  perPage,
		}

		result, err := auditService.QueryAuditLogs(c.Request.Context(), query)
		if err != nil {
			logger.Error("Query audit logs failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "query audit logs failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func createGetUserAuditTrailHandler(auditService *services.AuditService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

		result, err := auditService.GetUserAuditTrail(c.Request.Context(), userID, days, page, perPage)
		if err != nil {
			logger.Error("Get user audit trail failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "get user audit trail failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func createGetResourceAuditTrailHandler(auditService *services.AuditService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceType := c.Param("type")
		resourceID := c.Param("id")
		days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

		result, err := auditService.GetResourceAuditTrail(c.Request.Context(), resourceType, resourceID, days, page, perPage)
		if err != nil {
			logger.Error("Get resource audit trail failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "get resource audit trail failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func createGetSecurityEventsHandler(auditService *services.AuditService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

		result, err := auditService.GetSecurityEvents(c.Request.Context(), days, page, perPage)
		if err != nil {
			logger.Error("Get security events failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "get security events failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func createGetAuditStatisticsHandler(auditService *services.AuditService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

		stats, err := auditService.GetAuditStatistics(c.Request.Context(), days)
		if err != nil {
			logger.Error("Get audit statistics failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "get audit statistics failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

// Admin handlers

func createGrantPermissionHandler(authzService *services.AuthorizationService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID     string `json:"user_id" binding:"required"`
			Resource   string `json:"resource" binding:"required"`
			ResourceID string `json:"resource_id" binding:"required"`
			Relation   string `json:"relation" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
			return
		}

		err := authzService.GrantPermission(c.Request.Context(), req.UserID, req.Resource, req.ResourceID, req.Relation)
		if err != nil {
			logger.Error("Grant permission failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "grant permission failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "permission granted successfully"})
	}
}

func createRevokePermissionHandler(authzService *services.AuthorizationService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID     string `json:"user_id" binding:"required"`
			Resource   string `json:"resource" binding:"required"`
			ResourceID string `json:"resource_id" binding:"required"`
			Relation   string `json:"relation" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
			return
		}

		err := authzService.RevokePermission(c.Request.Context(), req.UserID, req.Resource, req.ResourceID, req.Relation)
		if err != nil {
			logger.Error("Revoke permission failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "revoke permission failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "permission revoked successfully"})
	}
}

func createAssignRoleHandler(authzService *services.AuthorizationService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID string `json:"user_id" binding:"required"`
			RoleID string `json:"role_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
			return
		}

		err := authzService.AssignRoleToUser(c.Request.Context(), req.UserID, req.RoleID)
		if err != nil {
			logger.Error("Assign role failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "assign role failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "role assigned successfully"})
	}
}

func createRemoveRoleHandler(authzService *services.AuthorizationService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID string `json:"user_id" binding:"required"`
			RoleID string `json:"role_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
			return
		}

		err := authzService.RemoveRoleFromUser(c.Request.Context(), req.UserID, req.RoleID)
		if err != nil {
			logger.Error("Remove role failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "remove role failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "role removed successfully"})
	}
}

func createArchiveLogsHandler(auditService *services.AuditService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Days int `json:"days" binding:"required,min=1"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
			return
		}

		err := auditService.ArchiveOldLogs(c.Request.Context(), req.Days)
		if err != nil {
			logger.Error("Archive logs failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "archive logs failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "logs archived successfully"})
	}
}
