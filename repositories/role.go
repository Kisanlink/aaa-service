package repositories

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"gorm.io/gorm"
)

type RoleRepositoryInterface interface {
	CheckIfRoleExists(roleName string) error
	CreateRoleWithPermissions(role *model.Role, permissions []model.Permission) error
	GetRoleByName(name string) (*model.Role, error)
	FindRoleByID(id string) (*model.Role, error)
	FindRoles(filter map[string]interface{}) ([]model.Role, error)
	UpdateRoleWithPermissions(id string, updatedRole model.Role, permissions []model.Permission) error
	DeleteRole(id string) error
}

type RoleRepository struct {
	DB *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepositoryInterface {
	return &RoleRepository{
		DB: db,
	}
}

func (repo *RoleRepository) CheckIfRoleExists(roleName string) error {
	var existingRole model.Role
	err := repo.DB.Where("name = ?", roleName).First(&existingRole).Error
	if err == nil {
		return helper.NewAppError(http.StatusConflict, fmt.Errorf("role with name '%s' already exists", roleName))
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("database error: %w", err))
	}
	return nil
}

func (repo *RoleRepository) CreateRoleWithPermissions(role *model.Role, permissions []model.Permission) error {
	tx := repo.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, err)
	}

	// Create the role first
	if err := tx.Create(role).Error; err != nil {
		tx.Rollback()
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create role: %w", err))
	}

	// Associate permissions with the role
	for i := range permissions {
		permissions[i].RoleID = role.ID
		if err := tx.Create(&permissions[i]).Error; err != nil {
			tx.Rollback()
			return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create permission: %w", err))
		}
	}

	if err := tx.Commit().Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("transaction commit failed: %w", err))
	}

	return nil
}

func (repo *RoleRepository) GetRoleByName(name string) (*model.Role, error) {
	var role model.Role
	err := repo.DB.Preload("Permissions").Where("name = ?", name).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("role with name %s not found", name))
	}
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to query role: %w", err))
	}
	return &role, nil
}

func (repo *RoleRepository) FindRoleByID(id string) (*model.Role, error) {
	var role model.Role
	err := repo.DB.Preload("Permissions").Where("id = ?", id).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("role with ID %s not found", id))
	}
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to query role: %w", err))
	}
	return &role, nil
}

// FindRoles retrieves roles with optional filtering by ID or name
func (repo *RoleRepository) FindRoles(filter map[string]interface{}) ([]model.Role, error) {
	var roles []model.Role
	query := repo.DB.Preload("Permissions")

	// Apply filters if provided
	if id, ok := filter["id"]; ok {
		query = query.Where("id = ?", id)
	}
	if name, ok := filter["name"]; ok {
		if nameStr, ok := name.(string); ok {
			query = query.Where("name ILIKE ?", "%"+nameStr+"%") // Case-insensitive partial match
		}
	}

	err := query.Find(&roles).Error
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError,
			fmt.Errorf("failed to retrieve roles: %w", err))
	}
	return roles, nil
}

func (repo *RoleRepository) UpdateRoleWithPermissions(id string, updatedRole model.Role, permissions []model.Permission) error {
	tx := repo.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, err)
	}

	// Update role fields
	result := tx.Model(&model.Role{}).Where("id = ?", id).Updates(updatedRole)
	if result.Error != nil {
		tx.Rollback()
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to update role: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return helper.NewAppError(http.StatusNotFound, fmt.Errorf("role with ID %s not found", id))
	}

	// Delete existing permissions
	if err := tx.Where("role_id = ?", id).Delete(&model.Permission{}).Error; err != nil {
		tx.Rollback()
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete old permissions: %w", err))
	}

	// Create new permissions
	for i := range permissions {
		permissions[i].RoleID = id
		if err := tx.Create(&permissions[i]).Error; err != nil {
			tx.Rollback()
			return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create permission: %w", err))
		}
	}

	if err := tx.Commit().Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("transaction commit failed: %w", err))
	}

	return nil
}

func (repo *RoleRepository) DeleteRole(id string) error {
	tx := repo.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, err)
	}

	// First delete all permissions associated with the role
	if err := tx.Where("role_id = ?", id).Delete(&model.Permission{}).Error; err != nil {
		tx.Rollback()
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete permissions: %w", err))
	}

	// Then delete the role
	result := tx.Where("id = ?", id).Delete(&model.Role{})
	if result.Error != nil {
		tx.Rollback()
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete role: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return helper.NewAppError(http.StatusNotFound, fmt.Errorf("role with ID %s not found", id))
	}

	if err := tx.Commit().Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("transaction commit failed: %w", err))
	}

	return nil
}
