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
	CreateRole(newRole *model.Role) error
	GetRoleByName(name string) (*model.Role, error)
	FindRoleByID(id string) (*model.Role, error)
	FindAllRoles() ([]model.Role, error)
	UpdateRole(id string, updatedRole model.Role) error
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
	existingRole := model.Role{}
	err := repo.DB.Table("roles").Where("name = ?", roleName).First(&existingRole).Error
	if err == nil {
		return helper.NewAppError(http.StatusConflict, fmt.Errorf("role with name '%s' already exists", roleName))
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("database error: %w", err))
	}
	return nil
}

func (repo *RoleRepository) CreateRole(newRole *model.Role) error {
	if err := repo.DB.Table("roles").Create(newRole).Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create role: %w", err))
	}
	return nil
}

func (repo *RoleRepository) GetRoleByName(name string) (*model.Role, error) {
	var role model.Role
	err := repo.DB.Table("roles").Where("name ILIKE ?", name).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("role with name %s not found", name))
	} else if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to query role: %w", err))
	}
	return &role, nil
}

func (repo *RoleRepository) FindRoleByID(id string) (*model.Role, error) {
	var role model.Role
	err := repo.DB.Table("roles").Where("id = ?", id).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("role with ID %s not found", id))
	} else if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to query role: %w", err))
	}
	return &role, nil
}

func (repo *RoleRepository) FindAllRoles() ([]model.Role, error) {
	var roles []model.Role
	err := repo.DB.Table("roles").Find(&roles).Error
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to retrieve roles: %w", err))
	}
	return roles, nil
}

func (repo *RoleRepository) UpdateRole(id string, updatedRole model.Role) error {
	result := repo.DB.Table("roles").Where("id = ?", id).Updates(updatedRole)
	if result.Error != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to update role: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return helper.NewAppError(http.StatusNotFound, fmt.Errorf("role with ID %s not found", id))
	}
	return nil
}

func (repo *RoleRepository) DeleteRole(id string) error {
	result := repo.DB.Table("roles").Where("id = ?", id).Delete(&model.Role{})
	if result.Error != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete role: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return helper.NewAppError(http.StatusNotFound, fmt.Errorf("role with ID %s not found", id))
	}
	return nil
}
