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
	StatusCode int               `json:"status_code"`
	Success bool               `json:"success"`
	Message    string       `json:"message"`
	DataTimeStamp string             `json:"data_time_stamp"`
	Role       *RoleResponse `json:"role"`
}

func (s *RoleServer) CreateRoleRestApi(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role Name is required"})
		return
	}

	if err := s.RoleRepo.CheckIfRoleExists(c.Request.Context(), req.Name); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Role already exists"})
		return
	}

	newRole := model.Role{
		Name:        req.Name,
		Description: req.Description,
		Source:      req.Source,
	}

	if err := s.RoleRepo.CreateRole(c.Request.Context(), &newRole); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}
	roles, err := s.RoleRepo.FindAllRoles(c.Request.Context())
	if err != nil {
		log.Printf("Failed to fetch roles: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve roles"})
		return
	}

	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	permissions, err := s.PermissionRepo.FindAllPermissions(c.Request.Context())
	if err != nil {
		log.Printf("Failed to fetch permissions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve permissions"})
		return
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

	// Update schema
	updated, err := client.UpdateSchema(roleNames, permissionNames, allActions)
	if err != nil {
		log.Printf("Failed to update schema: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update schema"})
		return
	}

	log.Printf("Schema updated successfully: %+v", updated)

	response := CreateRoleResponse{
		StatusCode: http.StatusCreated,
		Success: true,
		Message:    "Role created successfully",
		DataTimeStamp: time.Now().Format(time.RFC3339), // Current time in RFC3339 string format
		Role: &RoleResponse{
			ID:          newRole.ID,
			Name:        newRole.Name,
			Description: newRole.Description,
			Source:      newRole.Source,
		},
	}

	c.JSON(http.StatusCreated, response)
}