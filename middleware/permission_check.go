package middleware

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/repositories"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PermissionMiddleware struct {
	db *gorm.DB
}

func NewPermissionMiddleware(db *gorm.DB) *PermissionMiddleware {
	return &PermissionMiddleware{
		db: db,
	}
}

func (pm *PermissionMiddleware) GeneralPermissionCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract required headers
		userID := c.GetHeader("x-user-id")
		resourceName := c.GetHeader("x-resource-name")
		resourceID := c.GetHeader("x-principal-id")
		action := c.GetHeader("x-action")

		// Validate headers
		if userID == "" {
			helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Missing header: x-user-id"})
			c.Abort()
			return
		}
		if resourceName == "" {
			helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Missing header: x-resource-name"})
			c.Abort()
			return
		}
		if resourceID == "" {
			helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Missing header: x-principal-id"})
			c.Abort()
			return
		}
		if action == "" {
			helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Missing header: x-action"})
			c.Abort()
			return
		}

		// Check permissions using the injected client
		hasPermission, err := client.CheckPermission(
			userID,
			action,
			resourceName,
			resourceID,
		)

		if err != nil {
			helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
			c.Abort()
			return
		}

		if !hasPermission {
			helper.SendErrorResponse(c.Writer, http.StatusForbidden, []string{"Access denied: User cannot perform this action on this resource."})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (pm *PermissionMiddleware) CanCreatePermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract required headers
		userID := c.GetHeader("x-user-id")
		resourceName := c.GetHeader("x-resource-name")
		action := c.GetHeader("x-action")
		roleName := c.GetHeader("x-role-name")

		// Validate headers
		if userID == "" {
			helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Missing header: x-user-id"})
			c.Abort()
			return
		}
		if resourceName == "" {
			helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Missing header: x-resource-name"})
			c.Abort()
			return
		}
		if action == "" {
			helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Missing header: x-action"})
			c.Abort()
			return
		}
		if roleName == "" {
			helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Missing header: x-role-name"})
			c.Abort()
			return
		}

		// Initialize repositories and services
		userRepo := repositories.NewUserRepository(pm.db)
		roleRepo := repositories.NewRoleRepository(pm.db)
		userService := services.NewUserService(userRepo, roleRepo)

		// Get user roles with permissions
		rolesResponse, err := userService.GetUserRolesWithPermissions(userID)
		if err != nil {
			helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to fetch user roles: " + err.Error()})
			c.Abort()
			return
		}

		// Check permissions
		hasPermission := false

	roleLoop:
		for _, role := range rolesResponse.Roles {
			if role.RoleName == roleName {
				for _, permission := range role.Permissions {
					if permission.Resource == resourceName {
						for _, allowedAction := range permission.Actions {
							if allowedAction == action {
								hasPermission = true
								break roleLoop
							}
						}
					}
				}
			}
		}

		if !hasPermission {
			helper.SendErrorResponse(c.Writer, http.StatusForbidden, []string{
				"Access denied: User doesn't have permission to perform '" + action +
					"' on resource '" + resourceName + "' with role '" + roleName + "'",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
