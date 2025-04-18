package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Routes(router *gin.RouterGroup, db *gorm.DB) {
	api := router.Group("/v1")
	UserRoutes(api, db)
	RolesRoutes(api, db)
	PermissionRoutes(api, db)
	AssignPermissionRoutes(api, db)
	MediaRoutes(api, db)

}
