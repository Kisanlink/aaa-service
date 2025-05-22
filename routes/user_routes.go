package routes

import (
	"github.com/Kisanlink/aaa-service/handler/user"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UserRoutes(r *gin.RouterGroup, database *gorm.DB, s user.UserHandler) {
	r.POST("/login", s.LoginRestApi)
	r.POST("/register", s.CreateUserRestApi)
	r.GET("/users", s.GetUserRestApi)
	r.GET("/users/:id", s.GetUserByIdRestApi)
	r.PATCH("/users/:id", s.UpdateUserRestApi)
	r.POST("/assign-role", s.AssignRoleRestApi)
	r.DELETE("/remove/:role/by/:userID", s.DeleteAssignRoleRestApi)
	r.POST("/forgot-password", s.PasswordResetHandler)
	r.POST("token-transaction", s.TokenUsageHandler)
}
