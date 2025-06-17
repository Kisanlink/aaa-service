package routes

import (
	"github.com/Kisanlink/aaa-service/handler/action"
	"github.com/Kisanlink/aaa-service/handler/permission"
	"github.com/Kisanlink/aaa-service/handler/resource"
	"github.com/Kisanlink/aaa-service/handler/role"
	rolepermission "github.com/Kisanlink/aaa-service/handler/role_permission"
	"github.com/Kisanlink/aaa-service/handler/spicedb"
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
	permRepo := repositories.NewPermissionRepository(db)
	rolePermRepo := repositories.NewRolePermissionRepository(db)
	userService := services.NewUserService(userRepo, roleRepo)
	roleService := services.NewRoleService(roleRepo)
	actionService := services.NewActionService(actionRepo)
	resourceService := services.NewResourceService(resourceRepo)
	permissionService := services.NewPermissionService(permRepo)
	rolePermService := services.NewRolePermissionService(rolePermRepo)
	actionHandler := action.NewActionHandler(actionService)
	resourceHandler := resource.NewResourceHandler(resourceService)
	userHandler := user.NewUserHandler(userService, roleService)
	rolHandler := role.NewRoleHandler(roleService)
	permHandler := permission.NewPermissionHandler(permissionService)
	rolePermHandler := rolepermission.NewRolePermissionHandler(rolePermService, roleService, permissionService)
	spiceHandler := spicedb.NewSpiceDBHandler(roleService)
	RolesRoutes(v1, db, *rolHandler)
	ActionRoutes(v1, db, *actionHandler)
	ResourceRoutes(v1, db, *resourceHandler)
	RolePermRoutes(v1, db, *rolePermHandler)
	PermissionRoutes(v1, db, *permHandler)
	UsersRoutes(v1, db, *userHandler)
	SpiceDBRoutes(v1, db, *spiceHandler)

}
