package routes

import (
	"github.com/Kisanlink/aaa-service/handler/roles"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RolesRoutes(r *gin.RouterGroup, database *gorm.DB, s roles.RoleHandler) {
	r.POST("/roles", s.CreateRoleRestApi)
	r.GET("/roles", s.GetAllRolesRestApi)
	r.GET("/roles/:id", s.GetRoleByIdRestApi)
}
