package routes

import (
	"github.com/Kisanlink/aaa-service/controller/user"
	"github.com/Kisanlink/aaa-service/repositories"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UserRoutes(r *gin.RouterGroup, database *gorm.DB) {
	userRepo := repositories.NewUserRepository(database)
	
	s := user.Server{UserRepo: userRepo}
	r.POST("/login", s.LoginRestApi)
	r.POST("/register", s.CreateUserRestApi)
}