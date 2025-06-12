package routes

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	// Recovery middleware
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(*helper.AppError); ok {
			helper.SendErrorResponse(c.Writer, err.StatusCode, []string{err.Error()})
			c.Abort()
			return
		}

		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"internal server error"})
		c.Abort()
	}))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the AAA-SERVICE server!",
		})
	})
	api := r.Group("/api")

	ApiRoutes(api, db)

	return r

}
