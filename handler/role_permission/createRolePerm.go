package rolepermission

import (
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/gin-gonic/gin"
)

type RolePermissionHandler struct {
	rolePermissionService services.RolePermissionServiceInterface
	roleService           services.RoleServiceInterface
	permissionService     services.PermissionServiceInterface
}

func NewRolePermissionHandler(
	rolePermissionService services.RolePermissionServiceInterface, roleService services.RoleServiceInterface, permissionService services.PermissionServiceInterface,
) *RolePermissionHandler {
	return &RolePermissionHandler{
		rolePermissionService: rolePermissionService,
		roleService:           roleService,
		permissionService:     permissionService,
	}
}

// AssignPermissionToRoleRestApi assigns a permission to a role
// @Summary Assign permission to role
// @Description Creates an association between a role and permission
// @Tags RolePermissions
// @Accept json
// @Produce json
// @Param request body model.RolePermissionRequest true "Assignment data"
// @Success 201 {object} helper.Response{data=model.RolePermission} "Permission assigned successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid request or missing required fields"
// @Failure 404 {object} helper.ErrorResponse "Role or permission not found"
// @Failure 409 {object} helper.ErrorResponse "Association already exists"
// @Failure 500 {object} helper.ErrorResponse "Failed to assign permission"
// @Router /role-permissions [post]
func (h *RolePermissionHandler) AssignPermissionToRoleRestApi(c *gin.Context) {
	var req model.RolePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"invalid request body"})
		return
	}
	// Validate input
	if req.RoleID == "" {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"role ID is required"})
		return
	}

	if len(req.PermissionIDs) == 0 {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"at least one permission ID is required"})
		return
	}

	if err := h.rolePermissionService.CreateRolePermissions(req.RoleID, req.PermissionIDs); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}
	// Get all roles to build SpiceDB schema
	roles, err := h.roleService.FindRoles(map[string]interface{}{}, 0, 0)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	// Generate SpiceDB schema definitions
	schemaDefinitions := helper.GenerateSpiceDBSchema(roles)

	// Update SpiceDB schema
	_, err = client.UpdateSchema(schemaDefinitions)
	if err != nil {
		log.Printf("Failed to update SpiceDB schema: %v", err)

	}

	helper.SendSuccessResponse(c.Writer, http.StatusCreated, "Permission assigned to role successfully", req)
}
