package services

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/repositories"
)

type RolePermissionServiceInterface interface {
	CreateRolePermissions(roleID string, permissionIDs []string) error
	FetchAll(filterParams map[string]interface{}, page, limit int) ([]model.RolePermission, error)
	FetchByID(id string) (*model.RolePermission, error)
	DeleteByID(id string) error
}

type RolePermissionService struct {
	repo repositories.RolePermissionRepositoryInterface
}

func NewRolePermissionService(
	repo repositories.RolePermissionRepositoryInterface,
) RolePermissionServiceInterface {
	return &RolePermissionService{
		repo: repo,
	}
}

func (s *RolePermissionService) CreateRolePermissions(roleID string, permissionIDs []string) error {
	// Validate input at service layer
	if roleID == "" {
		return helper.NewAppError(http.StatusBadRequest,
			fmt.Errorf("role ID cannot be empty"))
	}

	if len(permissionIDs) == 0 {
		return helper.NewAppError(http.StatusBadRequest,
			fmt.Errorf("at least one permission ID is required"))
	}

	// Check for empty permission IDs
	for _, pid := range permissionIDs {
		if pid == "" {
			return helper.NewAppError(http.StatusBadRequest,
				fmt.Errorf("permission ID cannot be empty"))
		}
	}

	// Business logic validation could be added here
	// For example:
	// - Check if role exists
	// - Check if permissions exist
	// - Check if user has permission to perform this action

	// Call repository
	if err := s.repo.CreateRolePermissions(roleID, permissionIDs); err != nil {
		// Preserve the original error type if it's already an AppError
		var appErr *helper.AppError
		if errors.As(err, &appErr) {
			return err
		}
		return helper.NewAppError(http.StatusInternalServerError,
			fmt.Errorf("failed to assign permissions to role: %w", err))
	}

	return nil
}

func (s *RolePermissionService) FetchAll(filterParams map[string]interface{}, page, limit int) ([]model.RolePermission, error) {
	return s.repo.FetchAll(filterParams, page, limit)
}

func (s *RolePermissionService) FetchByID(id string) (*model.RolePermission, error) {
	return s.repo.FetchByID(id)
}

func (s *RolePermissionService) DeleteByID(id string) error {
	return s.repo.DeleteByID(id)
}
