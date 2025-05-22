package routes

import (
	"github.com/Kisanlink/aaa-service/handler/action"
	"github.com/Kisanlink/aaa-service/handler/resource"
	"github.com/Kisanlink/aaa-service/handler/role"
	"github.com/Kisanlink/aaa-service/handler/user"
	"github.com/Kisanlink/aaa-service/repositories"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ApiRoutes(r *gin.RouterGroup, db *gorm.DB) {
	v1 := r.Group("/v1")
	userRepo := repositories.NewUserRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	actionRepo := repositories.NewActionRepository(db)
	resourceRepo := repositories.NewResourceRepository(db)
	userService := services.NewUserService(userRepo, roleRepo)
	roleService := services.NewRoleService(roleRepo)
	actionService := services.NewActionService(actionRepo)
	resourceService := services.NewResourceService(resourceRepo)
	actionHandler := action.NewActionHandler(actionService)
	resourceHandler := resource.NewResourceHandler(resourceService)
	userHandler := user.NewUserHandler(userService, roleService)
	rolHandler := role.NewRoleHandler(roleService)
	RolesRoutes(v1, db, *rolHandler)
	UserRoutes(v1, db, *userHandler)
	ActionRoutes(v1, db, *actionHandler)
	ResourceRoutes(v1, db, *resourceHandler)

}
