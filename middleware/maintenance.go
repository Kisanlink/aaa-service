package middleware

import (
	"net/http"
	"strings"

	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MaintenanceMiddleware handles maintenance mode checks
type MaintenanceMiddleware struct {
	maintenanceService interfaces.MaintenanceService
	responder          interfaces.Responder
	logger             interfaces.Logger
	bypassPaths        []string
}

// NewMaintenanceMiddleware creates a new maintenance middleware instance
func NewMaintenanceMiddleware(
	maintenanceService interfaces.MaintenanceService,
	responder interfaces.Responder,
	logger interfaces.Logger,
) *MaintenanceMiddleware {
	return &MaintenanceMiddleware{
		maintenanceService: maintenanceService,
		responder:          responder,
		logger:             logger,
		bypassPaths: []string{
			"/health",
			"/ready",
			"/live",
			"/api/v2/admin/maintenance", // Allow maintenance control endpoints
		},
	}
}

// Handler returns the maintenance mode middleware handler
func (m *MaintenanceMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if path should bypass maintenance mode
		if m.shouldBypassPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Check if system is in maintenance mode
		isMaintenanceMode, mode, err := m.maintenanceService.IsMaintenanceMode(c.Request.Context())
		if err != nil {
			m.logger.Error("Failed to check maintenance mode", zap.Error(err))
			// In case of error, allow the request to continue
			c.Next()
			return
		}

		if !isMaintenanceMode {
			c.Next()
			return
		}

		// System is in maintenance mode, check if user should be allowed
		userID := m.getUserID(c)
		isAdmin := m.isAdminUser(c)
		isReadOperation := m.isReadOperation(c.Request.Method)

		allowed, err := m.maintenanceService.IsUserAllowedDuringMaintenance(
			c.Request.Context(),
			userID,
			isAdmin,
			isReadOperation,
		)
		if err != nil {
			m.logger.Error("Failed to check user maintenance access", zap.Error(err))
			m.sendMaintenanceResponse(c, mode)
			return
		}

		if allowed {
			m.logger.Debug("User allowed during maintenance",
				zap.String("userID", userID),
				zap.Bool("isAdmin", isAdmin),
				zap.Bool("isReadOperation", isReadOperation))
			c.Next()
			return
		}

		// User not allowed, send maintenance response
		m.sendMaintenanceResponse(c, mode)
	}
}

// shouldBypassPath checks if a path should bypass maintenance mode
func (m *MaintenanceMiddleware) shouldBypassPath(path string) bool {
	for _, bypassPath := range m.bypassPaths {
		if strings.HasPrefix(path, bypassPath) {
			return true
		}
	}
	return false
}

// getUserID extracts user ID from context or JWT token
func (m *MaintenanceMiddleware) getUserID(c *gin.Context) string {
	// Try to get user ID from context (set by auth middleware)
	if userID, exists := c.Get("userID"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}

	// Could also extract from JWT token if needed
	return ""
}

// isAdminUser checks if the current user has admin privileges
func (m *MaintenanceMiddleware) isAdminUser(c *gin.Context) bool {
	// Check if user has admin role from context (set by auth middleware)
	if roles, exists := c.Get("userRoles"); exists {
		if roleList, ok := roles.([]string); ok {
			for _, role := range roleList {
				if role == "admin" || role == "superadmin" || role == "system_admin" {
					return true
				}
			}
		}
	}

	// Check admin flag from context
	if isAdmin, exists := c.Get("isAdmin"); exists {
		if admin, ok := isAdmin.(bool); ok {
			return admin
		}
	}

	return false
}

// isReadOperation checks if the HTTP method is a read operation
func (m *MaintenanceMiddleware) isReadOperation(method string) bool {
	readMethods := []string{"GET", "HEAD", "OPTIONS"}
	for _, readMethod := range readMethods {
		if method == readMethod {
			return true
		}
	}
	return false
}

// sendMaintenanceResponse sends the maintenance mode response
func (m *MaintenanceMiddleware) sendMaintenanceResponse(c *gin.Context, mode interface{}) {
	// Extract maintenance message and details
	message := "System is currently under maintenance. Please try again later."

	// Log the blocked request
	m.logger.Info("Request blocked due to maintenance mode",
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("user_agent", c.Request.UserAgent()),
		zap.String("remote_addr", c.ClientIP()))

	// Add retry-after header for maintenance mode
	c.Header("Retry-After", "3600") // Suggest retry after 1 hour

	m.responder.SendError(c, http.StatusServiceUnavailable, message, nil)
}

// MaintenanceMode returns a gin middleware function for maintenance mode
// This is a convenience function for easy integration
func MaintenanceMode(
	maintenanceService interfaces.MaintenanceService,
	responder interfaces.Responder,
	logger interfaces.Logger,
) gin.HandlerFunc {
	middleware := NewMaintenanceMiddleware(maintenanceService, responder, logger)
	return middleware.Handler()
}
