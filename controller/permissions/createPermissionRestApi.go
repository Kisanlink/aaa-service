package permissions

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

type CreatePermissionRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Action      string `json:"action"`
	Source      string `json:"source"`
	Resource      string `json:"resource"`
}

type PermissionResponse struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Action        string  `json:"action"`
	Source        string  `json:"source"`
	Resource      string  `json:"resource"`
	ValidStartTime string  `json:"valid_start_time"`
	ValidEndTime   string  `json:"valid_end_time"`
}

type CreatePermissionResponse struct {
	StatusCode int               `json:"status_code"`
	Success bool               `json:"success"`
	Message    string            `json:"message"`
	Permission *PermissionResponse `json:"permission"`
}

func (s *PermissionServer) CreatePermissionRestApi(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Permission Name is required"})
		return
	}

	if err := s.PermissionRepo.CheckIfPermissionExists(c.Request.Context(), req.Name); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Permission already exists"})
		return
	}

	newPermission := model.Permission{
		Name:           req.Name,
		Description:    req.Description,
		Action:         req.Action,
		Source: req.Source,
		Resource: req.Resource,
		ValidStartTime: time.Now(),
		ValidEndTime:   time.Now().AddDate(1, 0, 0), // Default to 1 year validity
	}

	if err := s.PermissionRepo.CreatePermission(c.Request.Context(), &newPermission); err != nil {
		log.Printf("Failed to create permission in database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create permission"})
		return
	}

	// Update schema with all roles and permissions
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

	// Default values if no roles/permissions/actions exist
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

	response := CreatePermissionResponse{
		StatusCode: http.StatusCreated,
		Success: true,
		Message:    "Permission created successfully",
		Permission: &PermissionResponse{
			ID:            newPermission.ID,
			Name:          newPermission.Name,
			Description:   newPermission.Description,
			Action:        newPermission.Action,
			ValidStartTime: newPermission.ValidStartTime.Format(time.RFC3339Nano),
			ValidEndTime:   newPermission.ValidEndTime.Format(time.RFC3339Nano),
		},
	}

	c.JSON(http.StatusCreated, response)
}