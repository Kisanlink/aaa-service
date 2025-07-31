package health

import (
	"context"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	dbManager    *db.DatabaseManager
	cacheService interfaces.CacheService
	responder    interfaces.Responder
	logger       *zap.Logger
}

// NewHealthHandler creates a new HealthHandler instance
func NewHealthHandler(
	dbManager *db.DatabaseManager,
	cacheService interfaces.CacheService,
	responder interfaces.Responder,
	logger *zap.Logger,
) *HealthHandler {
	return &HealthHandler{
		dbManager:    dbManager,
		cacheService: cacheService,
		responder:    responder,
		logger:       logger,
	}
}

// BasicHealth handles GET /health
func (h *HealthHandler) BasicHealth(c *gin.Context) {
	h.logger.Debug("Processing basic health check")

	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "aaa-service",
		"version":   "2.0.0",
		"timestamp": time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// ReadinessCheck handles GET /ready
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	h.logger.Debug("Processing readiness check")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	checks := map[string]interface{}{}
	allHealthy := true
	overallStatus := "ready"

	// Check database connectivity
	dbStatus, dbHealthy := h.checkDatabaseHealth(ctx)
	checks["database"] = dbStatus
	if !dbHealthy {
		allHealthy = false
	}

	// Check cache connectivity
	cacheStatus, cacheHealthy := h.checkCacheHealth(ctx)
	checks["cache"] = cacheStatus
	if !cacheHealthy {
		allHealthy = false
	}

	if !allHealthy {
		overallStatus = "not_ready"
	}

	response := map[string]interface{}{
		"status":    overallStatus,
		"checks":    checks,
		"timestamp": time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}

	statusCode := http.StatusOK
	if !allHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	h.responder.SendSuccess(c, statusCode, response)
}

// LivenessCheck handles GET /live
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	h.logger.Debug("Processing liveness check")

	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// Helper methods for health checks

func (h *HealthHandler) checkDatabaseHealth(ctx context.Context) (map[string]interface{}, bool) {
	startTime := time.Now()

	// Check all database managers and their connectivity status
	allConnected := true
	backends := map[string]bool{}

	// Check PostgreSQL
	if pgManager := h.dbManager.GetPostgresManager(); pgManager != nil {
		backends["postgres"] = pgManager.IsConnected()
		if !pgManager.IsConnected() {
			allConnected = false
		}
	}

	// Check DynamoDB
	if dynamoManager := h.dbManager.GetDynamoManager(); dynamoManager != nil {
		backends["dynamodb"] = dynamoManager.IsConnected()
		if !dynamoManager.IsConnected() {
			allConnected = false
		}
	}

	// Check SpiceDB
	if spiceManager := h.dbManager.GetSpiceManager(); spiceManager != nil {
		backends["spicedb"] = spiceManager.IsConnected()
		if !spiceManager.IsConnected() {
			allConnected = false
		}
	}

	responseTime := time.Since(startTime)

	if !allConnected {
		h.logger.Error("Database health check failed", zap.Any("backends", backends))
		return map[string]interface{}{
			"status":        "unhealthy",
			"backends":      backends,
			"response_time": responseTime.String(),
			"error":         "one or more backends are not connected",
		}, false
	}

	return map[string]interface{}{
		"status":        "healthy",
		"backends":      backends,
		"response_time": responseTime.String(),
	}, true
}

func (h *HealthHandler) checkCacheHealth(ctx context.Context) (map[string]interface{}, bool) {
	startTime := time.Now()

	// Test cache connectivity with a simple ping
	testKey := "health_check_" + time.Now().Format("20060102150405")
	testValue := "ping"

	err := h.cacheService.Set(testKey, testValue, 10) // 10 seconds TTL
	if err != nil {
		h.logger.Error("Cache set operation failed", zap.Error(err))
		responseTime := time.Since(startTime)
		return map[string]interface{}{
			"status":        "unhealthy",
			"error":         "set operation failed",
			"response_time": responseTime.String(),
		}, false
	}

	// Test get operation
	_, exists := h.cacheService.Get(testKey)
	if !exists {
		h.logger.Error("Cache get operation failed")
		responseTime := time.Since(startTime)
		return map[string]interface{}{
			"status":        "unhealthy",
			"error":         "get operation failed",
			"response_time": responseTime.String(),
		}, false
	}

	// Clean up test key
	if err := h.cacheService.Delete(testKey); err != nil {
		h.logger.Error("Failed to delete test key from cache", zap.Error(err))
	}

	responseTime := time.Since(startTime)
	return map[string]interface{}{
		"status":        "healthy",
		"response_time": responseTime.String(),
	}, true
}
