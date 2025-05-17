package routes

import (
	"github.com/Kisanlink/aaa-service/handler/permissions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PermissionRoutes(r *gin.RouterGroup, database *gorm.DB, s permissions.PermissionHandler) {

	r.POST("/permissions", s.CreatePermissionRestApi)
	r.GET("/permissions", s.GetAllPermissionsRestApi)
	r.GET("/permissions/:id", s.GetPermissionByIdRestApi)
}
