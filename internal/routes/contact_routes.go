package routes

import (
	"github.com/Kisanlink/aaa-service/v2/internal/handlers/contacts"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/middleware"
	contactService "github.com/Kisanlink/aaa-service/v2/internal/services/contacts"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupContactRoutes configures contact management routes with proper service integration
func SetupContactRoutes(
	protectedAPI *gin.RouterGroup,
	authMiddleware *middleware.AuthMiddleware,
	contactService *contactService.ContactService,
	validator interfaces.Validator,
	responder interfaces.Responder,
	logger *zap.Logger,
) {
	// Initialize handler with the provided services
	contactHandler := contacts.NewContactHandler(contactService, validator, responder, logger)

	contacts := protectedAPI.Group("/contacts")
	{
		// Contact CRUD operations
		contacts.POST("", authMiddleware.RequirePermission("contact", "create"), contactHandler.CreateContact)
		contacts.GET("", authMiddleware.RequirePermission("contact", "read"), contactHandler.ListContacts)
		contacts.GET("/:id", authMiddleware.RequirePermission("contact", "view"), contactHandler.GetContact)
		contacts.PUT("/:id", authMiddleware.RequirePermission("contact", "update"), contactHandler.UpdateContact)
		contacts.DELETE("/:id", authMiddleware.RequirePermission("contact", "delete"), contactHandler.DeleteContact)

		// User-specific contact operations
		contacts.GET("/user/:userID", authMiddleware.RequirePermission("contact", "read"), contactHandler.GetContactsByUser)
	}
}
