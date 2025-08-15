package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupHealthRoutes configures health check routes
func SetupHealthRoutes(publicAPI *gin.RouterGroup, logger *zap.Logger) {
	publicAPI.GET("/health", createHealthHandler(logger))
}

// HealthCheckV2 handles GET /v2/health
// @Summary Health check (V2)
// @Description Basic health check for the AAA service
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} responses.HealthCheckResponse
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
