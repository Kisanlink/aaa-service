package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kisanlink/aaa-service/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type RoleRepository struct {
	DB *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{
		DB: db,
	}
}

func (repo *RoleRepository) CheckIfRoleExists(ctx context.Context, roleName string) error {
	existingRole := model.Role{}
	err := repo.DB.Table("roles").Where("name = ?", roleName).First(&existingRole).Error
	if err == nil {
		return status.Error(codes.AlreadyExists, fmt.Sprintf("Role with name '%s' already exists", roleName))
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return status.Error(codes.Internal, "Database error: "+err.Error())
	}
	return nil
}

func (repo *RoleRepository) CreateRole(ctx context.Context, newRole *model.Role) error {
	if err := repo.DB.Table("roles").Create(newRole).Error; err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to create role: %v", err))
	}
	return nil
}

func (repo *RoleRepository) FindRoleByID(ctx context.Context, id string) (*model.Role, error) {
	var role model.Role
	err := repo.DB.Table("roles").Where("id = ?", id).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Role with ID %s not found", id))
	} else if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to query role: %v", err))
	}
	return &role, nil
}

func (repo *RoleRepository) FindAllRoles(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	err := repo.DB.Table("roles").Find(&roles).Error
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to retrieve roles: %v", err))
	}
	return roles, nil
}

func (repo *RoleRepository) UpdateRole(ctx context.Context, id string, updatedRole map[string]interface{}) error {
	result := repo.DB.Table("roles").Where("id = ?", id).Updates(updatedRole)
	if result.Error != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to update role: %v", result.Error))
	}
	if result.RowsAffected == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("Role with ID %s not found", id))
	}
	return nil
}

func (repo *RoleRepository) DeleteRole(ctx context.Context, id string) error {
	result := repo.DB.Table("roles").Where("id = ?", id).Delete(&model.Role{})
	if result.Error != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to delete role: %v", result.Error))
	}
	if result.RowsAffected == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("Role with ID %s not found", id))
	}
	return nil
}
