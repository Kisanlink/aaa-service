package repositories

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"gorm.io/gorm"
)

type RolePermissionRepositoryInterface interface {
	CreateRolePermissions(roleID string, permissionIDs []string) error
	FetchAll(filter map[string]interface{}, page, limit int) ([]model.RolePermission, error)
	DeleteByID(id string) error
	FetchByID(id string) (*model.RolePermission, error)
}

type RolePermissionRepository struct {
	DB *gorm.DB
}

func NewRolePermissionRepository(db *gorm.DB) RolePermissionRepositoryInterface {
	return &RolePermissionRepository{
		DB: db,
	}
}

func (repo *RolePermissionRepository) CreateRolePermissions(roleID string, permissionIDs []string) error {
	err := repo.DB.Transaction(func(tx *gorm.DB) error {
		// First check for existing associations to prevent partial creation
		var existing []model.RolePermission
		err := tx.Model(&model.RolePermission{}).
			Where("role_id = ? AND permission_id IN ?", roleID, permissionIDs).
			Find(&existing).
			Error

		if err != nil {
			return helper.NewAppError(http.StatusInternalServerError,
				fmt.Errorf("failed to check existing role-permission associations: %w", err))
		}

		if len(existing) > 0 {
			existingIDs := make([]string, 0, len(existing))
			for _, rp := range existing {
				existingIDs = append(existingIDs, rp.PermissionID)
			}
			return helper.NewAppError(http.StatusConflict,
				fmt.Errorf("role-permission associations already exist for role %s and permissions: %v",
					roleID, existingIDs))
		}

		// Prepare batch insert
		rolePermissions := make([]model.RolePermission, 0, len(permissionIDs))
		for _, permissionID := range permissionIDs {
			rolePermissions = append(rolePermissions, model.RolePermission{
				RoleID:       roleID,
				PermissionID: permissionID,
			})
		}

		// Batch create
		if err := tx.CreateInBatches(rolePermissions, 100).Error; err != nil {
			return helper.NewAppError(http.StatusInternalServerError,
				fmt.Errorf("failed to create role-permission associations: %w", err))
		}

		return nil
	})

	return err
}

func (repo *RolePermissionRepository) FetchAll(filter map[string]interface{}, page, limit int) ([]model.RolePermission, error) {
	var rolePermissions []model.RolePermission

	query := repo.DB.Table("role_permissions")

	// Apply filters if provided
	if roleID, ok := filter["role_id"]; ok {
		if roleIDStr, ok := roleID.(string); ok && roleIDStr != "" {
			query = query.Where("role_id = ?", roleIDStr)
		}
	}

	if permissionID, ok := filter["permission_id"]; ok {
		if permissionIDStr, ok := permissionID.(string); ok && permissionIDStr != "" {
			query = query.Where("permission_id = ?", permissionIDStr) // Fixed typo (was "permission_id")
		}
	}

	// Apply pagination if page and limit are valid
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	// Execute the query
	if err := query.Find(&rolePermissions).Error; err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError,
			fmt.Errorf("failed to retrieve role permissions: %w", err))
	}

	return rolePermissions, nil
}

func (repo *RolePermissionRepository) FetchByID(id string) (*model.RolePermission, error) {
	var rolePermission model.RolePermission

	err := repo.DB.
		Where("id = ?", id).
		First(&rolePermission).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if not found (no error)
		}
		return nil, helper.NewAppError(
			http.StatusInternalServerError,
			fmt.Errorf("failed to fetch role permission: %w", err),
		)
	}

	return &rolePermission, nil
}

func (repo *RolePermissionRepository) DeleteByID(id string) error {
	if err := repo.DB.Table("role_permissions").Delete(&model.RolePermission{}, "id = ?", id).Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete role permission: %w", err))
	}
	return nil
}
