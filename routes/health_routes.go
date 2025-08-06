package routes

import (
	"github.com/Kisanlink/aaa-service/handlers/health"
	"github.com/gin-gonic/gin"
)

// SetupHealthRoutes configures health check and utility routes
func SetupHealthRoutes(router *gin.Engine, healthHandler *health.HealthHandler) {
	// Basic health check
	router.GET("/health", healthHandler.BasicHealth)

	// Ready check (for Kubernetes readiness probe)
	router.GET("/ready", healthHandler.ReadinessCheck)

	// Live check (for Kubernetes liveness probe)
	router.GET("/live", healthHandler.LivenessCheck)

	// API information
	router.GET("/info", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service":      "aaa-service",
			"description":  "Authentication, Authorization, and Accounting Service",
			"version":      "2.0.0",
			"api_versions": []string{"v2"},
			"endpoints": gin.H{
				"health":        "/health",
				"ready":         "/ready",
				"live":          "/live",
				"api_v2":        "/api/v2",
				"documentation": "/docs",
			},
		})
	})

	// API version discovery
	router.GET("/api", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"available_versions": []gin.H{
				{
					"version": "v2",
					"path":    "/api/v2",
					"status":  "current",
				},
			},
		})
	})
}
