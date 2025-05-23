package services

import (
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/repositories"
)

type RoleServiceInterface interface {
	CheckIfRoleExists(roleName string) error
	CreateRoleWithPermissions(role *model.Role, permissions []model.Permission) error
	GetRoleByName(name string) (*model.Role, error)
	FindRoleByID(id string) (*model.Role, error)
	FindRoles(filter map[string]interface{}, page, limit int) ([]model.Role, error)
	UpdateRoleWithPermissions(id string, updatedRole model.Role, permissions []model.Permission) error
	DeleteRole(id string) error
}

type RoleService struct {
	repo repositories.RoleRepositoryInterface
}

func NewRoleService(repo repositories.RoleRepositoryInterface) RoleServiceInterface {
	return &RoleService{
		repo: repo,
	}
}

func (s *RoleService) CheckIfRoleExists(roleName string) error {
	if roleName == "" {
		return helper.NewAppError(http.StatusBadRequest, fmt.Errorf("role name cannot be empty"))
	}

	err := s.repo.CheckIfRoleExists(roleName)
	if err != nil {
		return fmt.Errorf("error checking role existence: %w", err)
	}
	return nil
}

func (s *RoleService) CreateRoleWithPermissions(role *model.Role, permissions []model.Permission) error {
	// Validate role name
	if role.Name == "" {
		return helper.NewAppError(http.StatusBadRequest, fmt.Errorf("role name is required"))
	}

	// Validate permissions
	for _, perm := range permissions {
		if perm.Resource == "" {
			return helper.NewAppError(http.StatusBadRequest, fmt.Errorf("permission resource is required"))
		}
		if len(perm.Actions) == 0 {
			return helper.NewAppError(http.StatusBadRequest, fmt.Errorf("permission actions cannot be empty"))
		}
	}

	err := s.repo.CreateRoleWithPermissions(role, permissions)
	if err != nil {
		return fmt.Errorf("failed to create role with permissions: %w", err)
	}
	return nil
}

func (s *RoleService) GetRoleByName(name string) (*model.Role, error) {
	if name == "" {
		return nil, helper.NewAppError(http.StatusBadRequest, fmt.Errorf("role name is required"))
	}

	role, err := s.repo.GetRoleByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}
	return role, nil
}

func (s *RoleService) FindRoleByID(id string) (*model.Role, error) {
	if id == "" {
		return nil, helper.NewAppError(http.StatusBadRequest, fmt.Errorf("role ID is required"))
	}

	role, err := s.repo.FindRoleByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find role by ID: %w", err)
	}
	return role, nil
}

func (s *RoleService) FindRoles(filter map[string]interface{}, page, limit int) ([]model.Role, error) {
	// Validate filter parameters
	if id, ok := filter["id"]; ok {
		if id == "" {
			return nil, helper.NewAppError(http.StatusBadRequest, fmt.Errorf("ID filter cannot be empty"))
		}
	}

	if name, ok := filter["name"]; ok {
		if name == "" {
			return nil, helper.NewAppError(http.StatusBadRequest, fmt.Errorf("name filter cannot be empty"))
		}
	}

	roles, err := s.repo.FindRoles(filter, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve roles: %w", err)
	}
	return roles, nil
}

func (s *RoleService) UpdateRoleWithPermissions(id string, updatedRole model.Role, permissions []model.Permission) error {
	if id == "" {
		return helper.NewAppError(http.StatusBadRequest, fmt.Errorf("role ID is required"))
	}

	// Validate permissions
	for _, perm := range permissions {
		if perm.Resource == "" {
			return helper.NewAppError(http.StatusBadRequest, fmt.Errorf("permission resource is required"))
		}
		if len(perm.Actions) == 0 {
			return helper.NewAppError(http.StatusBadRequest, fmt.Errorf("permission actions cannot be empty"))
		}
	}

	err := s.repo.UpdateRoleWithPermissions(id, updatedRole, permissions)
	if err != nil {
		return fmt.Errorf("failed to update role with permissions: %w", err)
	}
	return nil
}

func (s *RoleService) DeleteRole(id string) error {
	if id == "" {
		return helper.NewAppError(http.StatusBadRequest, fmt.Errorf("role ID is required"))
	}

	err := s.repo.DeleteRole(id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}
