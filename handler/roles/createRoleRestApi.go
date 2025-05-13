package roles

import (
	"log"
	"net/http"
	"strings"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	roleService services.RoleServiceInterface
	permService services.PermissionServiceInterface
}

func NewRoleHandler(
	roleService services.RoleServiceInterface,
	permService services.PermissionServiceInterface,
) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		permService: permService,
	}
}

// CreateRoleRestApi creates a new role
// @Summary Create a new role
// @Description Creates a new role with the provided details and updates the authorization schema
// @Tags Roles
// @Accept json
// @Produce json
// @Param request body model.CreateRoleRequest true "Role creation data"
// @Success 201 {object} helper.Response{data=model.Role} "Role created successfully"
// @Failure 400 {object} helper.Response "Invalid request or missing required fields"
// @Failure 409 {object} helper.Response "Role already exists"
// @Failure 500 {object} helper.Response "Failed to create role or update schema"
// @Router /roles [post]
func (s *RoleHandler) CreateRoleRestApi(c *gin.Context) {
	var req model.Role
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request body"})
		return
	}

	if req.Name == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Role Name is required"})
		return
	}

	// Check if role already exists
	if err := s.roleService.CheckIfRoleExists(req.Name); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusConflict, []string{"Role already exists"})
		return
	}

	// Create new role
	newRole := model.Role{
		Name:        req.Name,
		Description: req.Description,
		Source:      req.Source,
	}

	if err := s.roleService.CreateRole(&newRole); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to create role"})
		return
	}

	roles, err := s.roleService.FindAllRoles()
	if err != nil {
		log.Printf("Failed to fetch roles: %v", err)
	}

	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	// Get all permissions for schema update
	permissions, err := s.permService.FindAllPermissions()
	if err != nil {
		log.Printf("Failed to fetch permissions: %v", err)
	}

	var permissionNames []string
	var allActions []string
	actionSet := make(map[string]struct{})

	for _, permission := range permissions {
		permissionNames = append(permissionNames, permission.Name)
		actionSet[permission.Action] = struct{}{}
	}

	for action := range actionSet {
		allActions = append(allActions, action)
	}

	for i, action := range allActions {
		allActions[i] = strings.ToLower(action)
	}

	// Set defaults if no roles/permissions/actions found
	defaultRoles := []string{"test role"}
	defaultPermissions := []string{"test permission"}
	defaultActions := []string{"test action"}

	if len(roleNames) == 0 {
		roleNames = defaultRoles
	}
	if len(permissionNames) == 0 {
		permissionNames = defaultPermissions
	}
	if len(allActions) == 0 {
		allActions = defaultActions
	}

	// Update schema in client service
	updated, err := client.UpdateSchema(roleNames, permissionNames, allActions)
	if err != nil {
		log.Printf("Failed to update schema: %v", err)
	}

	log.Printf("Schema updated successfully: %+v", updated)

	// Prepare response
	roleResponse := &model.Role{
		Base: model.Base{
			ID:        newRole.ID,
			CreatedAt: newRole.CreatedAt,
			UpdatedAt: newRole.UpdatedAt,
		},
		Name:        newRole.Name,
		Description: newRole.Description,
		Source:      newRole.Source,
	}

	helper.SendSuccessResponse(c.Writer, http.StatusCreated, "Role created successfully", roleResponse)
}
