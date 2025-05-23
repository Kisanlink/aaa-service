package role

import (
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	roleService services.RoleServiceInterface
}

func NewRoleHandler(roleService services.RoleServiceInterface) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// CreateRoleWithPermissionsRestApi creates a new role with permissions
// @Summary Create a new role with permissions
// @Description Creates a new role with associated permissions
// @Tags Roles
// @Accept json
// @Produce json
// @Param request body model.CreateRoleRequest true "Role and permissions data"
// @Success 201 {object} helper.Response{data=model.Role} "Role created successfully"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 409 {object} helper.Response "Role already exists"
// @Failure 500 {object} helper.Response "Failed to create role"
// @Router /roles [post]
func (h *RoleHandler) CreateRoleWithPermissionsRestApi(c *gin.Context) {
	var req model.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{err.Error()})
		return
	}

	// Convert request to role and permissions
	role := &model.Role{
		Name:        req.Name,
		Description: req.Description,
		Source:      req.Source,
	}

	var permissions []model.Permission
	for _, perm := range req.Permissions {
		permissions = append(permissions, model.Permission{
			Resource: perm.Resource,
			Actions:  perm.Actions,
		})
	}

	// Check if role exists
	if err := h.roleService.CheckIfRoleExists(req.Name); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusConflict, []string{err.Error()})
		return
	}

	// Create role with permissions
	if err := h.roleService.CreateRoleWithPermissions(role, permissions); err != nil {
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

	// Get the created role with permissions
	createdRole, err := h.roleService.FindRoleByID(role.ID)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusCreated,
		"Role and authorization schema created successfully", createdRole)
}
