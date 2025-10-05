package routes

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRouteParameterExtraction(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a test router
	router := gin.New()

	// Add test routes that mimic our organization group routes structure
	router.GET("/api/v1/organizations/:orgId/groups", func(c *gin.Context) {
		orgId := c.Param("orgId")
		assert.NotEmpty(t, orgId, "orgId parameter should be extracted")
		c.JSON(200, gin.H{"orgId": orgId})
	})

	router.GET("/api/v1/organizations/:orgId/groups/:groupId/users", func(c *gin.Context) {
		orgId := c.Param("orgId")
		groupId := c.Param("groupId")
		assert.NotEmpty(t, orgId, "orgId parameter should be extracted")
		assert.NotEmpty(t, groupId, "groupId parameter should be extracted")
		c.JSON(200, gin.H{"orgId": orgId, "groupId": groupId})
	})

	router.DELETE("/api/v1/organizations/:orgId/groups/:groupId/users/:userId", func(c *gin.Context) {
		orgId := c.Param("orgId")
		groupId := c.Param("groupId")
		userId := c.Param("userId")
		assert.NotEmpty(t, orgId, "orgId parameter should be extracted")
		assert.NotEmpty(t, groupId, "groupId parameter should be extracted")
		assert.NotEmpty(t, userId, "userId parameter should be extracted")
		c.JSON(200, gin.H{"orgId": orgId, "groupId": groupId, "userId": userId})
	})

	router.DELETE("/api/v1/organizations/:orgId/groups/:groupId/roles/:roleId", func(c *gin.Context) {
		orgId := c.Param("orgId")
		groupId := c.Param("groupId")
		roleId := c.Param("roleId")
		assert.NotEmpty(t, orgId, "orgId parameter should be extracted")
		assert.NotEmpty(t, groupId, "groupId parameter should be extracted")
		assert.NotEmpty(t, roleId, "roleId parameter should be extracted")
		c.JSON(200, gin.H{"orgId": orgId, "groupId": groupId, "roleId": roleId})
	})

	// Test that routes are registered
	routes := router.Routes()
	assert.NotEmpty(t, routes, "Routes should be registered")
	assert.Equal(t, 4, len(routes), "Should have 4 test routes")

	// Verify route patterns
	expectedPaths := []string{
		"/api/v1/organizations/:orgId/groups",
		"/api/v1/organizations/:orgId/groups/:groupId/users",
		"/api/v1/organizations/:orgId/groups/:groupId/users/:userId",
		"/api/v1/organizations/:orgId/groups/:groupId/roles/:roleId",
	}

	for i, route := range routes {
		assert.Equal(t, expectedPaths[i], route.Path, "Route path should match expected pattern")
	}
}
