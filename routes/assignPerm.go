package routes

import (
	rolepermission "github.com/Kisanlink/aaa-service/controller/rolePermission"
	"github.com/Kisanlink/aaa-service/repositories"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AssignPermissionRoutes(r *gin.RouterGroup, database *gorm.DB) {
	rolePermRepo := repositories.NewRolePermissionRepository(database)
	roleRepo := repositories.NewRoleRepository(database)
	permRepo := repositories.NewPermissionRepository(database)
	userRepo := repositories.NewUserRepository(database)
	s := rolepermission.NewConnectRolePermissionServer(
		rolePermRepo,
		roleRepo,
		permRepo,
		userRepo,
	)

	r.POST("/assign-permission", s.AssignPermissionRestApi)
	r.GET("/permissions", s.GetRolePermissionByRoleNameRestApi)
}
