package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type RolePermissionRepository struct {
	DB *gorm.DB
}

func NewRolePermissionRepository(db *gorm.DB) *RolePermissionRepository {
	return &RolePermissionRepository{
		DB: db,
	}
}

type QueryResult struct {
	ID          string `json:"id"`
	CreatedAt   time.Time
	Role        string
	Permissions string
}

func (repo *RolePermissionRepository) CreateRolePermissions(ctx context.Context, rolePermissions []*model.RolePermission) error {
	if len(rolePermissions) == 0 {
		return nil
	}
	if err := repo.DB.Table("role_permissions").CreateInBatches(rolePermissions, 100).Error; err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to create RolePermissions: %v", err))
	}

	return nil
}

func (repo *RolePermissionRepository) DeleteRolePermissionByRoleID(ctx context.Context, id string) error {
	result := repo.DB.Table("role_permissions").Where("role_id = ?", id).Delete(&model.RolePermission{})
	if result.Error != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to delete RolePermission: %v", result.Error))
	}
	if result.RowsAffected == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("RolePermission with ID %s not found", id))
	}
	return nil
}
func (repo *RolePermissionRepository) DeleteRolePermissionByID(ctx context.Context, id string) error {
	result := repo.DB.Table("role_permissions").Where("id = ?", id).Delete(&model.RolePermission{})
	if result.Error != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to delete RolePermission: %v", result.Error))
	}
	if result.RowsAffected == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("RolePermission with ID %s not found", id))
	}
	return nil
}

func (repo *RolePermissionRepository) GetAllRolePermissions(ctx context.Context) ([]model.RolePermission, error) {
	var rolePermissions []model.RolePermission
	if err := repo.DB.
		Preload("Role").
		Preload("Permission").
		Find(&rolePermissions).Error; err != nil {
		return nil, err
	}

	return rolePermissions, nil
}

func (repo *RolePermissionRepository) GetRolePermissionByID(ctx context.Context, id string) (*model.RolePermission, error) {
	var rolePermission model.RolePermission
	if err := repo.DB.
		Preload("Role").
		Preload("Permission").
		Where("id = ?", id).
		First(&rolePermission).Error; err != nil {
		return nil, err
	}

	return &rolePermission, nil
}

func (repo *RolePermissionRepository) GetRolePermissionsByRoleID(ctx context.Context, roleID string) ([]model.RolePermission, error) {
	var rolePermissions []model.RolePermission
	if err := repo.DB.
		Preload("Permission").
		Where("role_id = ?", roleID).
		Find(&rolePermissions).Error; err != nil {
		return nil, err 
	}

	return rolePermissions, nil
}
func (repo *RolePermissionRepository) GetRolePermissionsByRoleIDs(ctx context.Context, roleIDs []string) ([]model.RolePermission, error) {
	var rolePermissions []model.RolePermission
	if err := repo.DB.Where("role_id IN ?", roleIDs).Find(&rolePermissions).Error; err != nil {
		return nil, err
	}
	return rolePermissions, nil
}

func (repo *RolePermissionRepository) UpdateRolePermission(ctx context.Context, id string, updates map[string]interface{}) error {
	result := repo.DB.Table("role_permissions").Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to update RolePermission: %v", result.Error))
	}
	if result.RowsAffected == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("RolePermission with ID %s not found", id))
	}
	return nil
}


// Add this method to your RolePermissionRepository
func (repo *RolePermissionRepository) GetRolePermissionByNames(ctx context.Context, roleName, permissionName string) (*model.RolePermission, error) {
    var rolePermission model.RolePermission
    
    err := repo.DB.WithContext(ctx).
        Joins("JOIN roles ON roles.id = role_permissions.role_id").
        Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
        Where("roles.name = ? AND permissions.name = ?", roleName, permissionName).
        Preload("Role").
        Preload("Permission").
        First(&rolePermission).Error
        
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil // Return nil if not found
        }
        return nil, err
    }
    
    return &rolePermission, nil
}