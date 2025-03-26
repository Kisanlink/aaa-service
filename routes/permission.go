package routes

import (
	"github.com/Kisanlink/aaa-service/controller/permissions"
	"github.com/Kisanlink/aaa-service/repositories"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PermissionRoutes(r *gin.RouterGroup, database *gorm.DB) {
	permRepo := repositories.NewPermissionRepository(database)
	
	s := permissions.PermissionServer{PermissionRepo: permRepo}
	r.POST("/create-permission", s.CreatePermissionRestApi)
	r.GET("/fetch-permissions", s.GetAllPermissionsRestApi)
	r.GET("/permission/:id", s.GetPermissionByIdRestApi)
}