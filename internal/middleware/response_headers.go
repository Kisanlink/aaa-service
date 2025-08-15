package middleware

import (
	"github.com/gin-gonic/gin"
)

// ResponseContextHeaders injects selected auth/context values into response headers
func ResponseContextHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Propagate request context into response headers early
		if uid, ok := c.Get("user_id"); ok {
			if s, ok := uid.(string); ok && s != "" {
				c.Writer.Header().Set("X-User-Id", s)
			}
		}
		if rid := c.GetString("request_id"); rid != "" {
			c.Writer.Header().Set("X-Request-Id", rid)
		}
		if authz := c.GetHeader("Authorization"); authz != "" {
			c.Writer.Header().Set("X-Authorization", authz)
		}

		c.Next()
	}
}
