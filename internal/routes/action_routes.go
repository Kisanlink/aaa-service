package routes

import (
	"github.com/Kisanlink/aaa-service/internal/handlers/actions"
	"github.com/gin-gonic/gin"
)

// RegisterActionRoutes registers all action-related routes
func RegisterActionRoutes(router *gin.RouterGroup, handler *actions.ActionHandler) {
	actionsGroup := router.Group("/actions")
	{
		actionsGroup.POST("", handler.CreateAction)
		actionsGroup.GET("", handler.ListActions)
		actionsGroup.GET("/:id", handler.GetAction)
		actionsGroup.PUT("/:id", handler.UpdateAction)
		actionsGroup.DELETE("/:id", handler.DeleteAction)
		actionsGroup.GET("/service/:serviceName", handler.GetActionsByService)
	}
}
