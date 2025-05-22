package routes

import (
	"github.com/Kisanlink/aaa-service/handler/resource"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ResourceRoutes(r *gin.RouterGroup, database *gorm.DB, s resource.ResourceHandler) {
	r.POST("/resources", s.CreateResourceRestApi)
	r.GET("/resources", s.GetResourcesRestApi)
	r.PUT("/resources/:id", s.UpdateResourceRestApi)
	r.DELETE("/resources/:id", s.DeleteResourceRestApi)
}
