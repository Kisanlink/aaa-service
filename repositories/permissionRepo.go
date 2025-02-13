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

type PermissionRepository struct {
	DB *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{
		DB: db,
	}
}

func (repo *PermissionRepository) CheckIfPermissionExists(ctx context.Context, name string) error {
	existingPermission := model.Permission{}
	err := repo.DB.Table("permissions").Where("name = ?", name).First(&existingPermission).Error
	if err == nil {
		return status.Error(codes.AlreadyExists, fmt.Sprintf("Permission with name %s already exists", name))
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return status.Error(codes.Internal, "Database error: "+err.Error())
	}
	return nil
}

func (repo *PermissionRepository) CreatePermission(ctx context.Context, newPermission *model.Permission) error {
	if err := repo.DB.Table("permissions").Create(&newPermission).Error; err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to create permission: %v", err))
	}
	return nil
}

func (repo *PermissionRepository) FindPermissionByID(ctx context.Context, id string) (*model.Permission, error) {
	var permission model.Permission
	err := repo.DB.Table("permissions").Where("id = ?", id).First(&permission).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Permission with ID %s not found", id))
	} else if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to query permission: %v", err))
	}
	return &permission, nil
}

func (repo *PermissionRepository) DeletePermission(ctx context.Context, id string) error {
	result := repo.DB.Table("permissions").Where("id = ?", id).Delete(&model.Permission{})
	if result.Error != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to delete permission: %v", result.Error))
	}
	if result.RowsAffected == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("Permission with ID %s not found", id))
	}
	return nil
}

func (repo *PermissionRepository) FindAllPermissions(ctx context.Context) ([]model.Permission, error) {
	var permissions []model.Permission
	err := repo.DB.Table("permissions").Find(&permissions).Error
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to retrieve permissions: %v", err))
	}
	return permissions, nil
}

func (repo *PermissionRepository) UpdatePermission(ctx context.Context, id string, updatedPermission map[string]interface{}) error {
	result := repo.DB.Table("permissions").Where("id = ?", id).Updates(updatedPermission)
	if result.Error != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to update permission: %v", result.Error))
	}
	if result.RowsAffected == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("Permission with ID %s not found", id))
	}
	return nil
}
