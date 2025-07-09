package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupHealthRoutes configures health check and utility routes
func SetupHealthRoutes(router *gin.Engine) {
	// Basic health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "aaa-service",
			"version": "2.0.0",
		})
	})

	// Ready check (for Kubernetes readiness probe)
	router.GET("/ready", func(c *gin.Context) {
		// TODO: Add actual readiness checks (database connectivity, dependencies, etc.)
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
			"checks": gin.H{
				"database": "connected",
				"cache":    "connected",
			},
		})
	})

	// Live check (for Kubernetes liveness probe)
	router.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "alive",
		})
	})

	// API information
	router.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":      "aaa-service",
			"description":  "Authentication, Authorization, and Accounting Service",
			"version":      "2.0.0",
			"api_versions": []string{"v1", "v2"},
			"endpoints": gin.H{
				"health":        "/health",
				"ready":         "/ready",
				"live":          "/live",
				"api_v1":        "/api/v1",
				"api_v2":        "/api/v2",
				"documentation": "/docs",
			},
		})
	})

	// API version discovery
	router.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"available_versions": []gin.H{
				{
					"version": "v1",
					"path":    "/api/v1",
					"status":  "stable",
				},
				{
					"version": "v2",
					"path":    "/api/v2",
					"status":  "current",
				},
			},
		})
	})
}
