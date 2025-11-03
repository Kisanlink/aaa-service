package routes

import (
	"github.com/Kisanlink/aaa-service/v2/internal/handlers/addresses"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupAddressRoutes configures address management routes with proper service integration
func SetupAddressRoutes(
	protectedAPI *gin.RouterGroup,
	authMiddleware *middleware.AuthMiddleware,
	addressService interfaces.AddressService,
	validator interfaces.Validator,
	responder interfaces.Responder,
	logger *zap.Logger,
) {
	// Initialize handler with the provided services
	addressHandler := addresses.NewAddressHandler(addressService, validator, responder, logger)

	addressGroup := protectedAPI.Group("/addresses")
	{
		// Address CRUD operations
		addressGroup.POST("", authMiddleware.RequirePermission("address", "create"), addressHandler.CreateAddress)
		addressGroup.GET("/:id", authMiddleware.RequirePermission("address", "read"), addressHandler.GetAddress)
		addressGroup.PUT("/:id", authMiddleware.RequirePermission("address", "update"), addressHandler.UpdateAddress)
		addressGroup.DELETE("/:id", authMiddleware.RequirePermission("address", "delete"), addressHandler.DeleteAddress)

		// Address search operations
		addressGroup.GET("/search", authMiddleware.RequirePermission("address", "read"), addressHandler.SearchAddresses)
	}
}
