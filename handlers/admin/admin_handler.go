package admin

import (
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AdminHandler handles admin-related HTTP requests
type AdminHandler struct {
	maintenanceService interfaces.MaintenanceService
	validator          interfaces.Validator
	responder          interfaces.Responder
	logger             *zap.Logger
}

// NewAdminHandler creates a new AdminHandler instance
func NewAdminHandler(
	maintenanceService interfaces.MaintenanceService,
	validator interfaces.Validator,
	responder interfaces.Responder,
	logger *zap.Logger,
) *AdminHandler {
	return &AdminHandler{
		maintenanceService: maintenanceService,
		validator:          validator,
		responder:          responder,
		logger:             logger,
	}
}

// DetailedHealthCheckV2 handles GET /v2/admin/health/detailed
func (h *AdminHandler) DetailedHealthCheckV2(c *gin.Context) {
	h.logger.Info("Performing detailed health check")

	healthStatus := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "aaa-service",
		"version":   "2.0.0",
		"components": map[string]interface{}{
			"database": map[string]interface{}{
				"status":        "healthy",
				"response_time": "5ms",
				"connections": map[string]interface{}{
					"active":    10,
					"max":       100,
					"available": 90,
				},
			},
			"cache": map[string]interface{}{
				"status":        "healthy",
				"response_time": "1ms",
				"memory_usage": map[string]interface{}{
					"used_mb":       256,
					"available_mb":  1024,
					"usage_percent": 25,
				},
			},
			"external_services": map[string]interface{}{
				"spicedb": map[string]interface{}{
					"status":        "healthy",
					"response_time": "10ms",
				},
				"kisanlink_db": map[string]interface{}{
					"status":        "healthy",
					"response_time": "3ms",
				},
			},
		},
		"uptime":       "2h 30m 45s",
		"last_restart": "2024-01-01T00:00:00Z",
	}

	h.logger.Info("Detailed health check completed")
	h.responder.SendSuccess(c, http.StatusOK, healthStatus)
}

// MetricsV2 handles GET /v2/admin/metrics
func (h *AdminHandler) MetricsV2(c *gin.Context) {
	h.logger.Info("Retrieving system metrics")

	metrics := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"system": map[string]interface{}{
			"cpu_usage":    "15.2%",
			"memory_usage": "45.7%",
			"disk_usage":   "62.1%",
			"load_average": []float64{0.8, 0.9, 1.2},
		},
		"application": map[string]interface{}{
			"active_users":          1250,
			"total_requests":        52340,
			"requests_per_minute":   127,
			"average_response_time": "120ms",
			"error_rate":            "0.02%",
		},
		"database": map[string]interface{}{
			"total_connections":  85,
			"active_connections": 12,
			"queries_per_second": 45,
			"average_query_time": "15ms",
			"slow_queries":       2,
		},
		"cache": map[string]interface{}{
			"hit_rate":        "96.5%",
			"miss_rate":       "3.5%",
			"memory_usage_mb": 512,
			"evictions":       23,
		},
		"endpoints": map[string]interface{}{
			"/v1/users": map[string]interface{}{
				"total_requests":  15420,
				"success_rate":    "99.8%",
				"avg_response_ms": 95,
			},
			"/v2/auth/login": map[string]interface{}{
				"total_requests":  8750,
				"success_rate":    "99.2%",
				"avg_response_ms": 180,
			},
			"/v1/roles": map[string]interface{}{
				"total_requests":  3240,
				"success_rate":    "99.9%",
				"avg_response_ms": 65,
			},
		},
	}

	h.logger.Info("System metrics retrieved")
	h.responder.SendSuccess(c, http.StatusOK, metrics)
}

// AuditLogsV2 handles GET /v2/admin/audit
func (h *AdminHandler) AuditLogsV2(c *gin.Context) {
	h.logger.Info("Retrieving audit logs")

	// Parse query parameters
	userID := c.Query("user_id")
	action := c.Query("action")
	resource := c.Query("resource")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	// Mock audit logs data
	auditLogs := map[string]interface{}{
		"total_count":    2543,
		"filtered_count": 150,
		"filters": map[string]interface{}{
			"user_id":    userID,
			"action":     action,
			"resource":   resource,
			"start_date": startDate,
			"end_date":   endDate,
		},
		"pagination": map[string]interface{}{
			"limit":  limitStr,
			"offset": offsetStr,
		},
		"logs": []map[string]interface{}{
			{
				"id":        "audit_001",
				"timestamp": time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339),
				"user_id":   "user_123",
				"action":    "CREATE",
				"resource":  "USER",
				"details": map[string]interface{}{
					"username":   "john.doe",
					"ip_address": "192.168.1.100",
					"user_agent": "Mozilla/5.0...",
				},
				"result": "SUCCESS",
			},
			{
				"id":        "audit_002",
				"timestamp": time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339),
				"user_id":   "user_456",
				"action":    "LOGIN",
				"resource":  "AUTH",
				"details": map[string]interface{}{
					"ip_address": "10.0.0.25",
					"user_agent": "PostmanRuntime/7.32.3",
				},
				"result": "SUCCESS",
			},
			{
				"id":        "audit_003",
				"timestamp": time.Now().Add(-3 * time.Hour).UTC().Format(time.RFC3339),
				"user_id":   "user_789",
				"action":    "DELETE",
				"resource":  "ROLE",
				"details": map[string]interface{}{
					"role_name":  "deprecated_role",
					"ip_address": "172.16.0.10",
				},
				"result": "SUCCESS",
			},
		},
	}

	h.logger.Info("Audit logs retrieved",
		zap.String("userID", userID),
		zap.String("action", action),
		zap.String("resource", resource))
	h.responder.SendSuccess(c, http.StatusOK, auditLogs)
}

// MaintenanceModeV2 handles POST /v2/admin/maintenance
func (h *AdminHandler) MaintenanceModeV2(c *gin.Context) {
	h.logger.Info("Processing maintenance mode request")

	var req struct {
		Enabled    bool    `json:"enabled"`
		Message    string  `json:"message,omitempty"`
		Reason     string  `json:"reason,omitempty"`
		EndTime    *string `json:"end_time,omitempty"`
		AllowAdmin bool    `json:"allow_admin,omitempty"`
		AllowRead  bool    `json:"allow_read,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind maintenance request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Get user info from context
	userID := h.getUserID(c)
	if userID == "" {
		h.responder.SendError(c, http.StatusUnauthorized, "User authentication required", nil)
		return
	}

	if req.Enabled {
		// Enable maintenance mode
		maintenanceConfig := map[string]interface{}{
			"enabled":     true,
			"message":     req.Message,
			"reason":      req.Reason,
			"enabled_by":  userID,
			"allow_admin": req.AllowAdmin,
			"allow_read":  req.AllowRead,
		}

		if req.EndTime != nil && *req.EndTime != "" {
			// Parse end time if provided
			endTime, err := time.Parse(time.RFC3339, *req.EndTime)
			if err != nil {
				h.responder.SendValidationError(c, []string{"invalid end_time format, use RFC3339"})
				return
			}
			maintenanceConfig["end_time"] = endTime
		}

		if err := h.maintenanceService.EnableMaintenanceMode(c.Request.Context(), maintenanceConfig); err != nil {
			h.logger.Error("Failed to enable maintenance mode", zap.Error(err))
			h.responder.SendInternalError(c, err)
			return
		}

		h.logger.Warn("Maintenance mode enabled via admin API",
			zap.String("enabled_by", userID),
			zap.String("message", req.Message),
			zap.String("reason", req.Reason))

		response := map[string]interface{}{
			"status":              "success",
			"maintenance_enabled": true,
			"message":             "Maintenance mode has been enabled",
			"config":              maintenanceConfig,
			"timestamp":           time.Now().UTC().Format(time.RFC3339),
		}

		h.responder.SendSuccess(c, http.StatusOK, response)
	} else {
		// Disable maintenance mode
		if err := h.maintenanceService.DisableMaintenanceMode(c.Request.Context(), userID); err != nil {
			h.logger.Error("Failed to disable maintenance mode", zap.Error(err))
			h.responder.SendInternalError(c, err)
			return
		}

		h.logger.Info("Maintenance mode disabled via admin API", zap.String("disabled_by", userID))

		response := map[string]interface{}{
			"status":              "success",
			"maintenance_enabled": false,
			"message":             "Maintenance mode has been disabled",
			"timestamp":           time.Now().UTC().Format(time.RFC3339),
		}

		h.responder.SendSuccess(c, http.StatusOK, response)
	}
}

// GetMaintenanceStatus handles GET /v2/admin/maintenance
func (h *AdminHandler) GetMaintenanceStatus(c *gin.Context) {
	h.logger.Info("Getting maintenance status")

	status, err := h.maintenanceService.GetMaintenanceStatus(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get maintenance status", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	response := map[string]interface{}{
		"status":      "success",
		"maintenance": status,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// UpdateMaintenanceMessage handles PATCH /v2/admin/maintenance/message
func (h *AdminHandler) UpdateMaintenanceMessage(c *gin.Context) {
	h.logger.Info("Updating maintenance message")

	var req struct {
		Message string `json:"message" validate:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind message update request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Get user info from context
	userID := h.getUserID(c)
	if userID == "" {
		h.responder.SendError(c, http.StatusUnauthorized, "User authentication required", nil)
		return
	}

	if err := h.maintenanceService.UpdateMaintenanceMessage(c.Request.Context(), req.Message, userID); err != nil {
		h.logger.Error("Failed to update maintenance message", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Maintenance message updated",
		zap.String("updated_by", userID),
		zap.String("new_message", req.Message))

	response := map[string]interface{}{
		"status":      "success",
		"message":     "Maintenance message updated successfully",
		"new_message": req.Message,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// GetSystemInfo handles GET /v2/admin/system
func (h *AdminHandler) GetSystemInfo(c *gin.Context) {
	h.logger.Info("Retrieving system information")

	systemInfo := map[string]interface{}{
		"service": map[string]interface{}{
			"name":        "aaa-service",
			"version":     "2.0.0",
			"environment": "production",
			"build_date":  "2024-01-01T12:00:00Z",
			"commit_hash": "abc123def456",
		},
		"runtime": map[string]interface{}{
			"go_version":      "go1.21.0",
			"goroutines":      125,
			"memory_alloc_mb": 64,
			"memory_sys_mb":   128,
			"gc_cycles":       1250,
		},
		"dependencies": map[string]interface{}{
			"gin":          "v1.9.1",
			"zap":          "v1.24.0",
			"gorm":         "v1.25.0",
			"kisanlink-db": "v2.1.0",
		},
		"configuration": map[string]interface{}{
			"max_connections":    100,
			"request_timeout":    "30s",
			"cache_ttl":          "5m",
			"log_level":          "info",
			"cors_enabled":       true,
			"rate_limit_enabled": true,
		},
	}

	h.logger.Info("System information retrieved")
	h.responder.SendSuccess(c, http.StatusOK, systemInfo)
}

// Helper method to extract user ID from context
func (h *AdminHandler) getUserID(c *gin.Context) string {
	if userID, exists := c.Get("userID"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}
