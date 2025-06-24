package routes

import (
	"github.com/Kisanlink/aaa-service/handler/user"
	"github.com/Kisanlink/aaa-service/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UsersRoutes(r *gin.RouterGroup, database *gorm.DB, s user.UserHandler) {
	r.POST("/login", s.LoginRestApi)
	r.POST("/register", s.CreateUserRestApi)
	r.GET("/users", s.GetUserRestApi)
	r.GET("/users/:id", s.GetUserByIdRestApi)
	r.PUT("/users/:id", s.UpdateUserRestApi)
	r.POST("/assign-role", s.AssignRoleRestApi)
	r.DELETE("/assign-role", s.DeleteAssignRoleRestApi)
	r.POST("/forgot-password", s.PasswordResetHandler)
	r.POST("token-transaction", s.TokenUsageHandler)
	permMiddleware := middleware.NewPermissionMiddleware(database)
	permTestHandler := middleware.NewPermissionTestHandler()
	r.GET("/test-permission-get",
		permMiddleware.GeneralPermissionCheck(),
		permTestHandler.TestPermissionGET)

	r.POST("/test-permission-post",
		permMiddleware.CanCreatePermission(),
		permTestHandler.TestPermissionPOST)
}
