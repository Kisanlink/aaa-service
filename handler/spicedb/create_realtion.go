package spicedb

import (
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

// UpdateSpiceDbRelationships updates all user-role relationships in SpiceDB
// @Summary Update SpiceDB relationships
// @Description Updates all user-role relationships in SpiceDB by syncing with the current database state
// @Tags SpiceDB
// @Accept json
// @Produce json
// @Success 200 {object} helper.Response{data=model.User} "Relationships updated successfully"
// @Failure 500 {object} helper.ErrorResponse "Failed to update relationships"
// @Router /update/relations [get]
func (h *SpiceDBHandler) CreateRelation(c *gin.Context) {
	users, err := h.userService.GetUsers("", "", 0, 0)
	if err != nil {
		log.Printf("Failed to find users: %v", err)
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to retrieve users"})
		return
	}

	for _, user := range users {
		userRoles, err := h.userService.FindUserRoles(user.ID)
		if err != nil {
			log.Printf("Failed to find user roles for user %s: %v", user.Username, err)
			continue
		}

		roleNames := make([]string, 0, len(userRoles))
		for _, userRole := range userRoles {
			role, err := h.roleService.FindRoleByID(userRole.RoleID)
			if err != nil {
				log.Printf("Failed to find role by ID %s for user %s: %v", userRole.RoleID, user.Username, err)
				continue
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

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Relationships updated successfully", users)
}
