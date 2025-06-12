package routes

import (
	"github.com/Kisanlink/aaa-service/handler/permission"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PermissionRoutes(r *gin.RouterGroup, database *gorm.DB, s permission.PermissionHandler) {
	r.POST("/permissions", s.CreatePermissionRestApi)
	r.GET("/permissions", s.GetAllPermissionsRestApi)
	r.PUT("/permissions/:id", s.UpdatePermissionRestApi)
	r.DELETE("/permissions/:id", s.DeletePermissionRestApi)
	r.GET("/permissions/:id", s.GetPermissionByIDRestApi)
}
