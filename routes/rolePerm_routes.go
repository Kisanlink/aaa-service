package routes

import (
	rolepermission "github.com/Kisanlink/aaa-service/handler/role_permission"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RolePermRoutes(r *gin.RouterGroup, database *gorm.DB, s rolepermission.RolePermissionHandler) {
	r.POST("/assign-permissions", s.AssignPermissionToRoleRestApi)
	r.GET("/assign-permissions", s.GetAllRolesWithPermissionsRestApi)
	r.GET("/assign-permissions/:id", s.GetRolePermissionByIDRestApi)
	r.DELETE("/assign-permissions/:id", s.DeleteRolePermissionRestApi)
}
