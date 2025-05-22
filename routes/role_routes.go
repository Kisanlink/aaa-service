package routes

import (
	"github.com/Kisanlink/aaa-service/handler/role"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RolesRoutes(r *gin.RouterGroup, database *gorm.DB, s role.RoleHandler) {
	r.POST("/roles", s.CreateRoleWithPermissionsRestApi)
	r.GET("/roles", s.GetAllRolesRestApi)
	r.PUT("/roles/:id", s.UpdateRoleWithPermissionsRestApi)
	r.DELETE("/roles/:id", s.DeleteRoleRestApi)
	r.GET("/update/schema", s.UpdateSpiceDb)
}
