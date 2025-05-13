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
	CheckIfPermissionExists(name string) error
	CreatePermission(newPermission *model.Permission) error
	FindPermissionByID(id string) (*model.Permission, error)
	FindPermissionByName(name string) (*model.Permission, error)
	DeletePermission(id string) error
	FindAllPermissions() ([]model.Permission, error)
	UpdatePermission(id string, updatedPermission model.Permission) error
}

type PermissionRepository struct {
	DB *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepositoryInterface {
	return &PermissionRepository{DB: db}
}

func (repo *PermissionRepository) CheckIfPermissionExists(name string) error {
	existingPermission := model.Permission{}
	err := repo.DB.Table("permissions").Where("name = ?", name).First(&existingPermission).Error
	if err == nil {
		return helper.NewAppError(http.StatusConflict, fmt.Errorf("permission with name %s already exists", name))
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("database error: %w", err))
	}
	return nil
}

func (repo *PermissionRepository) CreatePermission(newPermission *model.Permission) error {
	if err := repo.DB.Table("permissions").Create(&newPermission).Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create permission: %w", err))
	}
	return nil
}

func (repo *PermissionRepository) FindPermissionByID(id string) (*model.Permission, error) {
	var permission model.Permission
	err := repo.DB.Table("permissions").Where("id = ?", id).First(&permission).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, helper.NewAppError(http.StatusNotFound, errors.New("permission not found"))
	} else if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to query permission: %w", err))
	}
	return &permission, nil
}

func (repo *PermissionRepository) FindPermissionByName(name string) (*model.Permission, error) {
	var permission model.Permission
	err := repo.DB.Table("permissions").Where("name = ?", name).First(&permission).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, helper.NewAppError(http.StatusNotFound, errors.New("permission not found"))
	} else if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to query permission: %w", err))
	}
	return &permission, nil
}

func (repo *PermissionRepository) DeletePermission(id string) error {
	result := repo.DB.Table("permissions").Where("id = ?", id).Delete(&model.Permission{})
	if result.Error != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete permission: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return helper.NewAppError(http.StatusNotFound, errors.New("permission not found"))
	}
	return nil
}

func (repo *PermissionRepository) FindAllPermissions() ([]model.Permission, error) {
	var permissions []model.Permission
	err := repo.DB.Table("permissions").Find(&permissions).Error
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to retrieve permissions: %w", err))
	}
	return permissions, nil
}

func (repo *PermissionRepository) UpdatePermission(id string, updatedPermission model.Permission) error {
	result := repo.DB.Table("permissions").Where("id = ?", id).Updates(updatedPermission)
	if result.Error != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to update permission: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return helper.NewAppError(http.StatusNotFound, errors.New("permission not found"))
	}
	return nil
}
