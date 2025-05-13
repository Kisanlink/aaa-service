package routes

import (
	"github.com/Kisanlink/aaa-service/handler/permissions"
	rolepermission "github.com/Kisanlink/aaa-service/handler/rolePermission"
	"github.com/Kisanlink/aaa-service/handler/roles"
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
	permissionRepo := repositories.NewPermissionRepository(db)
	connectRolePermissionRepo := repositories.NewRolePermissionRepository(db)
	userService := services.NewUserService(userRepo)
	roleService := services.NewRoleService(roleRepo)
	permissionService := services.NewPermissionService(permissionRepo)
	connectRolePermissionService := services.NewRolePermissionService(connectRolePermissionRepo)

	userHandler := user.NewUserHandler(userService, roleService, permissionService, connectRolePermissionService)
	RoleHandler := roles.NewRoleHandler(roleService, permissionService)
	permissionHandler := permissions.NewPermissionHandler(roleService, permissionService)
	RolePermHandler := rolepermission.NewRolePermHandler(roleService, permissionService, connectRolePermissionService, userService)
	UserRoutes(v1, db, *userHandler)
	RolesRoutes(v1, db, *RoleHandler)
	PermissionRoutes(v1, db, *permissionHandler)
	AssignPermissionRoutes(v1, db, *RolePermHandler)

}
