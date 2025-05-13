package services

import (
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/repositories"
)

type RolePermissionServiceInterface interface {
	CreateRolePermissions(rolePermissions []*model.RolePermission) error
	DeleteRolePermissionByRoleID(id string) error
	DeleteRolePermissionByID(id string) error
	GetAllRolePermissions() ([]model.RolePermission, error)
	GetRolePermissionByID(id string) (*model.RolePermission, error)
	GetRolePermissionsByRoleID(roleID string) ([]model.RolePermission, error)
	GetRolePermissionsByRoleIDs(roleIDs []string) ([]model.RolePermission, error)
	UpdateRolePermission(id string, updates model.RolePermission) error
	GetRolePermissionByNames(roleName, permissionName string) (*model.RolePermission, error)
}

type RolePermissionService struct {
	repo repositories.RolePermissionRepositoryInterface
}

func NewRolePermissionService(repo repositories.RolePermissionRepositoryInterface) RolePermissionServiceInterface {
	return &RolePermissionService{
		repo: repo,
	}
}
func (s *RolePermissionService) CreateRolePermissions(rolePermissions []*model.RolePermission) error {
	if len(rolePermissions) == 0 {
		return nil
	}
	err := s.repo.CreateRolePermissions(rolePermissions)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create role permissions: %w", err))
	}
	return nil
}
func (s *RolePermissionService) DeleteRolePermissionByRoleID(id string) error {
	err := s.repo.DeleteRolePermissionByRoleID(id)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete role permission by role ID: %w", err))
	}
	return nil
}

func (s *RolePermissionService) DeleteRolePermissionByID(id string) error {
	err := s.repo.DeleteRolePermissionByID(id)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete role permission by ID: %w", err))
	}
	return nil
}
func (s *RolePermissionService) GetAllRolePermissions() ([]model.RolePermission, error) {
	rolePermissions, err := s.repo.GetAllRolePermissions()
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get all role permissions: %w", err))
	}
	return rolePermissions, nil
}
func (s *RolePermissionService) GetRolePermissionByID(id string) (*model.RolePermission, error) {
	rolePermission, err := s.repo.GetRolePermissionByID(id)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get role permission by ID: %w", err))
	}
	if rolePermission == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("role permission not found"))
	}
	return rolePermission, nil
}
func (s *RolePermissionService) GetRolePermissionsByRoleID(roleID string) ([]model.RolePermission, error) {
	rolePermissions, err := s.repo.GetRolePermissionsByRoleID(roleID)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get role permissions by role ID: %w", err))
	}
	if len(rolePermissions) == 0 {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("role permissions not found"))
	}
	return rolePermissions, nil
}

func (s *RolePermissionService) GetRolePermissionsByRoleIDs(roleIDs []string) ([]model.RolePermission, error) {
	rolePermissions, err := s.repo.GetRolePermissionsByRoleIDs(roleIDs)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get role permissions by role IDs: %w", err))
	}
	if len(rolePermissions) == 0 {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("role permissions not found"))
	}
	return rolePermissions, nil
}
func (s *RolePermissionService) UpdateRolePermission(id string, updates model.RolePermission) error {
	err := s.repo.UpdateRolePermission(id, updates)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to update role permission: %w", err))
	}
	return nil
}

func (s *RolePermissionService) GetRolePermissionByNames(roleName, permissionName string) (*model.RolePermission, error) {
	rolePermission, err := s.repo.GetRolePermissionByNames(roleName, permissionName)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get role permission by names: %w", err))
	}
	if rolePermission == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("role permission not found"))
	}
	return rolePermission, nil
}
