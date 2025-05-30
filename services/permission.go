package services

import (
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/repositories"
)

type PermissionServiceInterface interface {
	CreatePermission(permission *model.Permission) error
	GetPermissionByID(id string) (*model.Permission, error)
	FindPermissions(filter map[string]interface{}, page, limit int) ([]model.Permission, error)
	UpdatePermission(id string, permission model.Permission) error
	DeletePermission(id string) error
}

type PermissionService struct {
	repo repositories.PermissionRepositoryInterface
}

func NewPermissionService(
	repo repositories.PermissionRepositoryInterface,
) PermissionServiceInterface {
	return &PermissionService{
		repo: repo,
	}
}

func (s *PermissionService) CreatePermission(permission *model.Permission) error {
	if permission.Resource == "" {
		return helper.NewAppError(http.StatusBadRequest, fmt.Errorf("resource is required"))
	}
	if len(permission.Actions) == 0 {
		return helper.NewAppError(http.StatusBadRequest, fmt.Errorf("at least one action is required"))
	}

	// Check if permission already exists for this resource
	err := s.repo.CheckPermissionExists(permission.Resource)
	if err != nil {
		return err
	}

	err = s.repo.CreatePermission(permission)
	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}

	return nil
}

func (s *PermissionService) GetPermissionByID(id string) (*model.Permission, error) {
	if id == "" {
		return nil, helper.NewAppError(http.StatusBadRequest, fmt.Errorf("permission ID is required"))
	}

	permission, err := s.repo.FindPermissionByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return permission, nil
}

func (s *PermissionService) FindPermissions(filter map[string]interface{}, page, limit int) ([]model.Permission, error) {
	// Validate filters
	if resource, ok := filter["resource"]; ok {
		if resource == "" {
			return nil, helper.NewAppError(http.StatusBadRequest, fmt.Errorf("resource filter cannot be empty"))
		}
	}

	if effect, ok := filter["effect"]; ok {
		if effect == "" {
			return nil, helper.NewAppError(http.StatusBadRequest, fmt.Errorf("effect filter cannot be empty"))
		}
	}

	if action, ok := filter["action"]; ok {
		if action == "" {
			return nil, helper.NewAppError(http.StatusBadRequest, fmt.Errorf("action filter cannot be empty"))
		}
	}

	permissions, err := s.repo.FindPermissions(filter, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve permissions: %w", err)
	}

	return permissions, nil
}

func (s *PermissionService) UpdatePermission(id string, permission model.Permission) error {
	if id == "" {
		return helper.NewAppError(http.StatusBadRequest, fmt.Errorf("permission ID is required"))
	}
	if permission.Resource == "" {
		return helper.NewAppError(http.StatusBadRequest, fmt.Errorf("resource is required"))
	}
	if len(permission.Actions) == 0 {
		return helper.NewAppError(http.StatusBadRequest, fmt.Errorf("at least one action is required"))
	}

	// Check if the updated resource would conflict with an existing permission
	existing, err := s.repo.FindPermissionByID(id)
	if err != nil {
		return fmt.Errorf("failed to find existing permission: %w", err)
	}

	// Only check for conflict if the resource is being changed
	if existing.Resource != permission.Resource {
		err = s.repo.CheckPermissionExists(permission.Resource)
		if err != nil {
			return err
		}
	}

	err = s.repo.UpdatePermission(id, permission)
	if err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}

	return nil
}

func (s *PermissionService) DeletePermission(id string) error {
	if id == "" {
		return helper.NewAppError(http.StatusBadRequest, fmt.Errorf("permission ID is required"))
	}

	err := s.repo.DeletePermission(id)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	return nil
}
