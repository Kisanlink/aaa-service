package routes

import (
	rolepermission "github.com/Kisanlink/aaa-service/handler/role_permission"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RolePermRoutes(r *gin.RouterGroup, database *gorm.DB, s rolepermission.RolePermissionHandler) {
	r.POST("/role-permissions", s.AssignPermissionToRoleRestApi)
	r.GET("/role-permissions", s.GetAllRolesWithPermissionsRestApi)
	r.GET("/role-permissions/:id", s.GetRolePermissionByIDRestApi)
	r.DELETE("/role-permissions/:id", s.DeleteRolePermissionRestApi)
}
