package routes

import (
	rolepermission "github.com/Kisanlink/aaa-service/handler/rolePermission"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AssignPermissionRoutes(r *gin.RouterGroup, database *gorm.DB, s rolepermission.RolePermHandler) {
	r.POST("/assign-permissions", s.AssignPermissionRestApi)
	r.GET("/assign-permissions/by", s.GetRolePermissionByRoleNameRestApi)
	r.GET("/assign-permissions", s.GetAllRolePermissionsRestApi)
}
