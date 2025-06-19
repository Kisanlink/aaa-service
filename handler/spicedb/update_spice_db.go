package spicedb

import (
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/gin-gonic/gin"
)

type SpiceDBHandler struct {
	roleService     services.RoleServiceInterface
	resourceService services.ResourceServiceInterface
	userService     services.UserServiceInterface
}

func NewSpiceDBHandler(roleService services.RoleServiceInterface, resourceService services.ResourceServiceInterface, userService services.UserServiceInterface) *SpiceDBHandler {
	return &SpiceDBHandler{
		roleService:     roleService,
		resourceService: resourceService,
		userService:     userService,
	}
}

// UpdateSpiceDb schema
// @Summary update spice db schema
// @Description update schema by Retrieves all roles
// @Tags SpiceDB
// @Accept json
// @Produce json
// @Success 200 {object} helper.Response{data=[]model.Role} "Roles retrieved successfully"
// @Failure 500 {object} helper.ErrorResponse "Failed to retrieve roles"
// @Router /update/schema [get]
func (h *SpiceDBHandler) UpdateSpiceDb(c *gin.Context) {
	filter := make(map[string]interface{})
	roles, err := h.roleService.FindRoles(filter, 0, 0)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}
	resource, err := h.resourceService.FindResources(map[string]interface{}{}, 0, 0)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}
	// Generate SpiceDB schema definitions
	schemaDefinitions := helper.GenerateSpiceDBSchema(roles, resource)

	// Update SpiceDB schema
	_, err = client.UpdateSchema(schemaDefinitions)
	if err != nil {
		log.Printf("Failed to update SpiceDB schema: %v", err)

	}
	users, err := h.userService.GetUsers("", "", 0, 0)
	if err != nil {
		log.Printf("Failed to find users: %v", err)
	}
	for _, user := range users {
		userRoles, err := h.userService.FindUserRoles(user.ID)
		if err != nil {
			log.Printf("Failed to find user roles for user %s: %v", user.Username, err)
		}
		roleNames := make([]string, 0, len(userRoles))
		for _, userRole := range userRoles {
			role, err := h.roleService.FindRoleByID(userRole.RoleID)
			if err != nil {
				log.Printf("Failed to find role by ID %s for user %s: %v", userRole.RoleID, user.Username, err)
				continue // Skip this role if it cannot be found
			}
			roleNames = append(roleNames, role.Name)
		}

		err = client.DeleteRelationships(
			roleNames,
			user.Username,
			user.ID,
		)
		if err != nil {
			log.Printf("Error deleting relationships: %v", err)
		}

		err = client.CreateRelationships(
			roleNames,
			user.Username,
			user.ID,
		)

		if err != nil {
			log.Printf("Error creating relationships: %v", err)
		}
	}
	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Schema updated successfully", roles)
}

func UpdateSpiceDBData(roleService services.RoleServiceInterface, userService services.UserServiceInterface) {
	users, err := userService.GetUsers("", "", 0, 0)
	if err != nil {
		log.Printf("Failed to find users: %v", err)
	}
	for _, user := range users {
		userRoles, err := userService.FindUserRoles(user.ID)
		if err != nil {
			log.Printf("Failed to find user roles for user %s: %v", user.Username, err)
		}
		roleNames := make([]string, 0, len(userRoles))
		for _, userRole := range userRoles {
			role, err := roleService.FindRoleByID(userRole.RoleID)
			if err != nil {
				log.Printf("Failed to find role by ID %s for user %s: %v", userRole.RoleID, user.Username, err)
				continue // Skip this role if it cannot be found
			}
			roleNames = append(roleNames, role.Name)
		}

		err = client.DeleteRelationships(
			roleNames,
			user.Username,
			user.ID,
		)
		if err != nil {
			log.Printf("Error deleting relationships: %v", err)
		}

		err = client.CreateRelationships(
			roleNames,
			user.Username,
			user.ID,
		)

		if err != nil {
			log.Printf("Error creating relationships: %v", err)
		}
	}
	log.Println("spice db relationships updated successfully")
}
