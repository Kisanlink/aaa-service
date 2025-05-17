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

type RolePermissionRepository struct {
	DB *gorm.DB
}

func NewRolePermissionRepository(db *gorm.DB) RolePermissionRepositoryInterface {
	return &RolePermissionRepository{
		DB: db,
	}
}

func (repo *RolePermissionRepository) CreateRolePermissions(rolePermissions []*model.RolePermission) error {
	if len(rolePermissions) == 0 {
		return nil
	}
	if err := repo.DB.Table("role_permissions").CreateInBatches(rolePermissions, 100).Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create RolePermissions: %w", err))
	}
	return nil
}

func (repo *RolePermissionRepository) DeleteRolePermissionByRoleID(id string) error {
	result := repo.DB.Table("role_permissions").Where("role_id = ?", id).Delete(&model.RolePermission{})
	if result.Error != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete RolePermission: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return helper.NewAppError(http.StatusNotFound, fmt.Errorf("rolePermission with role ID %s not found", id))
	}
	return nil
}

func (repo *RolePermissionRepository) DeleteRolePermissionByID(id string) error {
	result := repo.DB.Table("role_permissions").Where("id = ?", id).Delete(&model.RolePermission{})
	if result.Error != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete RolePermission: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return helper.NewAppError(http.StatusNotFound, fmt.Errorf("rolePermission with ID %s not found", id))
	}
	return nil
}

func (repo *RolePermissionRepository) GetAllRolePermissions() ([]model.RolePermission, error) {
	var rolePermissions []model.RolePermission
	if err := repo.DB.
		Preload("Role").
		Preload("Permission").
		Find(&rolePermissions).Error; err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get all role permissions: %w", err))
	}
	return rolePermissions, nil
}

func (repo *RolePermissionRepository) GetRolePermissionByID(id string) (*model.RolePermission, error) {
	var rolePermission model.RolePermission
	if err := repo.DB.
		Preload("Role").
		Preload("Permission").
		Where("id = ?", id).
		First(&rolePermission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("rolePermission with ID %s not found: %w", id, err))
		}
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get rolePermission by ID: %w", err))
	}
	return &rolePermission, nil
}

func (repo *RolePermissionRepository) GetRolePermissionsByRoleID(roleID string) ([]model.RolePermission, error) {
	var rolePermissions []model.RolePermission
	if err := repo.DB.
		Preload("Permission").
		Where("role_id = ?", roleID).
		Find(&rolePermissions).Error; err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get rolePermissions by role ID: %w", err))
	}
	return rolePermissions, nil
}

func (repo *RolePermissionRepository) GetRolePermissionsByRoleIDs(roleIDs []string) ([]model.RolePermission, error) {
	var rolePermissions []model.RolePermission
	if err := repo.DB.Where("role_id IN ?", roleIDs).Find(&rolePermissions).Error; err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get rolePermissions by role IDs: %w", err))
	}
	return rolePermissions, nil
}

func (repo *RolePermissionRepository) UpdateRolePermission(id string, updates model.RolePermission) error {
	result := repo.DB.Table("role_permissions").Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to update rolePermission: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return helper.NewAppError(http.StatusNotFound, fmt.Errorf("rolePermission with ID %s not found", id))
	}
	return nil
}

func (repo *RolePermissionRepository) GetRolePermissionByNames(roleName, permissionName string) (*model.RolePermission, error) {
	var rolePermission model.RolePermission

	err := repo.DB.
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("roles.name = ? AND permissions.name = ?", roleName, permissionName).
		Preload("Role").
		Preload("Permission").
		First(&rolePermission).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get rolePermission by names: %w", err))
	}

	return &rolePermission, nil
}
