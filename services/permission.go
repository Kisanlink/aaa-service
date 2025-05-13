package services

import (
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/repositories"
)

type PermissionServiceInterface interface {
	CheckIfPermissionExists(name string) error
	CreatePermission(newPermission *model.Permission) error
	FindPermissionByID(id string) (*model.Permission, error)
	FindPermissionByName(name string) (*model.Permission, error)
	DeletePermission(id string) error
	FindAllPermissions() ([]model.Permission, error)
	UpdatePermission(id string, updatedPermission model.Permission) error
}

type PermissionService struct {
	repo repositories.PermissionRepositoryInterface
}

func NewPermissionService(repo repositories.PermissionRepositoryInterface) PermissionServiceInterface {
	return &PermissionService{
		repo: repo,
	}
}

func (s *PermissionService) CheckIfPermissionExists(name string) error {
	err := s.repo.CheckIfPermissionExists(name)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to check if permission exists: %w", err))
	}
	return nil
}

func (s *PermissionService) CreatePermission(newPermission *model.Permission) error {
	err := s.repo.CreatePermission(newPermission)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create permission: %w", err))
	}
	return nil
}
func (s *PermissionService) FindPermissionByID(id string) (*model.Permission, error) {
	result, err := s.repo.FindPermissionByID(id)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get permission by ID: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("permission not found"))
	}
	return result, nil
}
func (s *PermissionService) FindPermissionByName(name string) (*model.Permission, error) {

	result, err := s.repo.FindPermissionByName(name)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get permission by name: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("permission not found"))
	}
	return result, nil
}
func (s *PermissionService) DeletePermission(id string) error {
	err := s.repo.DeletePermission(id)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete permission: %w", err))
	}
	return nil
}
func (s *PermissionService) FindAllPermissions() ([]model.Permission, error) {
	result, err := s.repo.FindAllPermissions()
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get all permissions: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("permissions not found"))
	}
	return result, nil
}
func (s *PermissionService) UpdatePermission(id string, updatedPermission model.Permission) error {
	err := s.repo.UpdatePermission(id, updatedPermission)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to update permission: %w", err))
	}
	return nil
}
