package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// dummy handler for permissions list
func dummyListPermissions(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }

func TestPermissionsRoute_RequiresAuthFirst(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create a group mimicking /api/v2 with auth middleware
	api := r.Group("/api/v2")
	// Inject a lightweight auth that only sets user_id when Authorization header present
	api.Use(func(c *gin.Context) {
		if c.GetHeader("Authorization") != "" {
			c.Set("user_id", "u1")
		}
		c.Next()
	})

	// Register permissions route similar to production
	perms := api.Group("/permissions")
	perms.Use(func(c *gin.Context) {
		v, ok := c.Get("user_id")
		if !ok || v == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		c.Next()
	})
	perms.GET("", dummyListPermissions)

	// No Authorization -> expect 401
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v2/permissions", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}
