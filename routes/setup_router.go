package routes

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, router *gin.Engine) {
	// Recovery middleware
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(*helper.AppError); ok {
			helper.SendErrorResponse(c.Writer, err.StatusCode, []string{err.Error()})
			c.Abort()
			return
		}

		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"internal server error"})
		c.Abort()
	}))

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the AAA-SERVICE server!",
		})
	})

	// API routes under /api/v1
	api := router.Group("/api")
	ApiRoutes(api, db)
}
