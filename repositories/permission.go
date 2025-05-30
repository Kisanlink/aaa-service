package repositories

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"gorm.io/gorm"
)

type PermissionRepositoryInterface interface {
	CreatePermission(permission *model.Permission) error
	FindPermissionByID(id string) (*model.Permission, error)
	FindPermissions(filter map[string]interface{}, page, limit int) ([]model.Permission, error)
	UpdatePermission(id string, permission model.Permission) error
	DeletePermission(id string) error
	CheckPermissionExists(resource string) error
}

type PermissionRepository struct {
	DB *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepositoryInterface {
	return &PermissionRepository{
		DB: db,
	}
}

func (repo *PermissionRepository) CreatePermission(permission *model.Permission) error {
	helper.PrettyJSON(permission)
	if err := repo.DB.Table("permissions").Create(permission).Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create permission: %w", err))
	}
	return nil
}

func (repo *PermissionRepository) FindPermissionByID(id string) (*model.Permission, error) {
	var permission model.Permission
	err := repo.DB.Where("id = ?", id).First(&permission).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, helper.NewAppError(http.StatusNotFound,
			fmt.Errorf("permission with ID %s not found", id))
	}
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError,
			fmt.Errorf("failed to query permission: %w", err))
	}
	return &permission, nil
}

func (repo *PermissionRepository) FindPermissions(filter map[string]interface{}, page, limit int) ([]model.Permission, error) {
	var permissions []model.Permission
	query := repo.DB.Model(&model.Permission{})

	// Apply filters
	if resource, ok := filter["resource"]; ok {
		if resourceStr, ok := resource.(string); ok {
			query = query.Where("resource ILIKE ?", "%"+resourceStr+"%") // Case-insensitive partial match
		}
	}
	if effect, ok := filter["effect"]; ok {
		if effectStr, ok := effect.(string); ok {
			query = query.Where("effect = ?", effectStr)
		}
	}
	if action, ok := filter["action"]; ok {
		if actionStr, ok := action.(string); ok {
			query = query.Where("? = ANY(actions)", actionStr)
		}
	}

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	err := query.Find(&permissions).Error
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError,
			fmt.Errorf("failed to retrieve permissions: %w", err))
	}
	return permissions, nil
}

func (repo *PermissionRepository) UpdatePermission(id string, permission model.Permission) error {
	result := repo.DB.Model(&model.Permission{}).Where("id = ?", id).Updates(permission)
	if result.Error != nil {
		return helper.NewAppError(http.StatusInternalServerError,
			fmt.Errorf("failed to update permission: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return helper.NewAppError(http.StatusNotFound,
			fmt.Errorf("permission with ID %s not found", id))
	}
	return nil
}

func (repo *PermissionRepository) DeletePermission(id string) error {
	result := repo.DB.Where("id = ?", id).Delete(&model.Permission{})
	if result.Error != nil {
		return helper.NewAppError(http.StatusInternalServerError,
			fmt.Errorf("failed to delete permission: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return helper.NewAppError(http.StatusNotFound,
			fmt.Errorf("permission with ID %s not found", id))
	}
	return nil
}

func (repo *PermissionRepository) CheckPermissionExists(resource string) error {
	var existingPermission model.Permission
	err := repo.DB.Where("resource = ?", resource).First(&existingPermission).Error
	if err == nil {
		return helper.NewAppError(http.StatusConflict,
			fmt.Errorf("permission for resource '%s' already exists", resource))
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return helper.NewAppError(http.StatusInternalServerError,
			fmt.Errorf("database error: %w", err))
	}
	return nil
}
