package routes

import (
	"github.com/Kisanlink/aaa-service/controller/user"
	"github.com/Kisanlink/aaa-service/repositories"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UserRoutes(r *gin.RouterGroup, database *gorm.DB) {
	userRepo := repositories.NewUserRepository(database)
	roleRepo := repositories.NewRoleRepository(database)
	permRepo := repositories.NewPermissionRepository(database)
	role_permRepo := repositories.NewRolePermissionRepository(database)

	s := user.NewUserServer(userRepo, roleRepo, permRepo, role_permRepo)
	r.POST("/login", s.LoginRestApi)
	r.POST("/register", s.CreateUserRestApi)
	r.POST("/assign-role", s.AssignRoleRestApi)
	r.POST("/forgot-password", s.PasswordResetHandler)
}
