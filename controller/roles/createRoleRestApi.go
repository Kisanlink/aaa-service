package roles

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

type RoleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

type CreateRoleResponse struct {
	StatusCode    int           `json:"status_code"`
	Success       bool          `json:"success"`
	Message       string        `json:"message"`
	Data          *RoleResponse `json:"data"`
	DataTimeStamp string        `json:"data_time_stamp"`
}

func (s *RoleServer) CreateRoleRestApi(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": http.StatusBadRequest,
			"success":     false,
			"message":     "Invalid request body",
			"data":        nil,
		})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": http.StatusBadRequest,
			"success":     false,
			"message":     "Role Name is required",
			"data":        nil,
		})
		return
	}

	ctx := c.Request.Context()

	// Check if role already exists
	if err := s.RoleRepo.CheckIfRoleExists(ctx, req.Name); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"status_code": http.StatusConflict,
			"success":     false,
			"message":     "Role already exists",
			"data":        nil,
		})
		return
	}

	// Create new role
	newRole := model.Role{
		Name:        req.Name,
		Description: req.Description,
		Source:      req.Source,
	}

	if err := s.RoleRepo.CreateRole(ctx, &newRole); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": http.StatusInternalServerError,
			"success":     false,
			"message":     "Failed to create role",
			"data":        nil,
		})
		return
	}
	roles, err := s.RoleRepo.FindAllRoles(ctx)
	if err != nil {
		log.Printf("Failed to fetch roles: %v", err)
	}

	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	// Get all permissions for schema update
	permissions, err := s.PermissionRepo.FindAllPermissions(ctx)
	if err != nil {
		log.Printf("Failed to fetch permissions: %v", err)
	}
	log.Println(permissions)
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
	roleResponse := &RoleResponse{
		ID:          newRole.ID,
		Name:        newRole.Name,
		Description: newRole.Description,
		Source:      newRole.Source,
	}

	response := &CreateRoleResponse{
		StatusCode:    http.StatusCreated,
		Success:       true,
		Message:       "Role created successfully",
		Data:          roleResponse,
		DataTimeStamp: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, response)
}
