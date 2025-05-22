package routes

import (
	"github.com/Kisanlink/aaa-service/handler/action"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ActionRoutes(r *gin.RouterGroup, database *gorm.DB, s action.ActionHandler) {
	r.POST("/actions", s.CreateActionRestApi)
	r.GET("/actions", s.GetAllActionsRestApi)
	r.PUT("/actions/:id", s.UpdateActionRestApi)
	r.DELETE("/actions/:id", s.DeleteActionRestApi)
}
