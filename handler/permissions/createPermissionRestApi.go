package permissions

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	roleService services.RoleServiceInterface
	permService services.PermissionServiceInterface
}

func NewPermissionHandler(
	roleService services.RoleServiceInterface,
	permService services.PermissionServiceInterface,
) *PermissionHandler {
	return &PermissionHandler{
		roleService: roleService,
		permService: permService,
	}
}

// CreatePermissionRestApi creates a new permission
// @Summary Create a new permission
// @Description Creates a new permission and updates the authorization schema with the new permission
// @Tags Permissions
// @Accept json
// @Produce json
// @Param request body model.CreatePermissionRequest true "Permission creation data"
// @Success 201 {object} helper.Response{data=model.Permission} "Permission created successfully"
// @Failure 400 {object} helper.Response "Invalid request or missing required fields"
// @Failure 409 {object} helper.Response "Permission already exists"
// @Failure 500 {object} helper.Response "Failed to create permission or update schema"
// @Router /permissions [post]
func (s *PermissionHandler) CreatePermissionRestApi(c *gin.Context) {
	var req model.Permission
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request body"})
		return
	}

	if req.Name == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Permission Name is required"})
		return
	}

	if err := s.permService.CheckIfPermissionExists(req.Name); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusConflict, []string{"Permission already exists"})
		return
	}

	newPermission := model.Permission{
		Name:           req.Name,
		Description:    req.Description,
		Action:         req.Action,
		Source:         req.Source,
		Resource:       req.Resource,
		ValidStartTime: time.Now(),
		ValidEndTime:   time.Now(),
	}

	if err := s.permService.CreatePermission(&newPermission); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to create permission"})
		return
	}

	roles, err := s.roleService.FindAllRoles()
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to retrieve roles"})
		return
	}

	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	permissions, err := s.permService.FindAllPermissions()
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to retrieve permissions"})
		return
	}

	var permissionNames []string
	actionSet := make(map[string]struct{})
	for _, permission := range permissions {
		permissionNames = append(permissionNames, permission.Name)
		actionSet[permission.Action] = struct{}{}
	}

	var allActions []string
	for action := range actionSet {
		allActions = append(allActions, strings.ToLower(action))
	}

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

	updated, err := client.UpdateSchema(roleNames, permissionNames, allActions)
	if err != nil {
		log.Printf("Failed to update schema: %v", err)

	}

	log.Printf("Schema updated successfully: %+v", updated)

	response := model.Permission{
		Base: model.Base{
			CreatedAt: newPermission.CreatedAt,
			UpdatedAt: newPermission.UpdatedAt,
			ID:        newPermission.ID,
		},
		Name:           newPermission.Name,
		Description:    newPermission.Description,
		Action:         newPermission.Action,
		Source:         newPermission.Source,
		Resource:       newPermission.Resource,
		ValidStartTime: newPermission.ValidStartTime,
		ValidEndTime:   newPermission.ValidEndTime,
	}

	helper.SendSuccessResponse(c.Writer, http.StatusCreated, "Permission created successfully", response)
}
