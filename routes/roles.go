package routes

import (
	"github.com/Kisanlink/aaa-service/controller/roles"
	"github.com/Kisanlink/aaa-service/repositories"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RolesRoutes(r *gin.RouterGroup, database *gorm.DB) {
	roleRepo := repositories.NewRoleRepository(database)
	
	s := roles.RoleServer{RoleRepo: roleRepo}
	r.POST("/create-role", s.CreateRoleRestApi)
	r.GET("/fetch-roles", s.GetAllRolesRestApi)
	r.GET("/role/:id", s.GetRoleByIdRestApi)
}