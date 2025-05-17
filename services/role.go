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
	CreateRole(newRole *model.Role) error
	GetRoleByName(name string) (*model.Role, error)
	FindRoleByID(id string) (*model.Role, error)
	FindAllRoles() ([]model.Role, error)
	UpdateRole(id string, updatedRole model.Role) error
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
	err := s.repo.CheckIfRoleExists(roleName)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to check if role exists: %w", err))
	}
	return nil
}
func (s *RoleService) CreateRole(newRole *model.Role) error {
	err := s.repo.CreateRole(newRole)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create role: %w", err))
	}
	return nil
}
func (s *RoleService) GetRoleByName(name string) (*model.Role, error) {
	result, err := s.repo.GetRoleByName(name)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get role by name: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("role not found"))
	}
	return result, nil
}

func (s *RoleService) FindRoleByID(id string) (*model.Role, error) {
	result, err := s.repo.FindRoleByID(id)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get role by ID: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("role not found"))
	}
	return result, nil
}

func (s *RoleService) FindAllRoles() ([]model.Role, error) {
	result, err := s.repo.FindAllRoles()
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get roles: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("roles not found"))
	}
	return result, nil
}

func (s *RoleService) UpdateRole(id string, updatedRole model.Role) error {
	err := s.repo.UpdateRole(id, updatedRole)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to update role: %w", err))
	}
	return nil
}
func (s *RoleService) DeleteRole(id string) error {
	err := s.repo.DeleteRole(id)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete role: %w", err))
	}
	return nil
}
