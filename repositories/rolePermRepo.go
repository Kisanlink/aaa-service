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

// CreateRolePermission creates a new RolePermission entry
func (repo *RolePermissionRepository) CreateRolePermission(ctx context.Context, rolePermission *model.RolePermission) error {

	if err := repo.DB.Table("role_permissions").Create(&rolePermission).Error; err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to create RolePermission: %v", err))
	}
	return nil
}

func (repo *RolePermissionRepository) CreatePermissionOnRole(ctx context.Context, permissionOnRole *model.PermissionOnRole) error {
	if err := repo.DB.Table("permission_on_roles").Create(&permissionOnRole).Error; err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to create PermissionOnRole: %v", err))
	}
	return nil
}

func (repo *RolePermissionRepository) DeletePermissionOnRoleByUserRoleID(ctx context.Context, userRoleID string) error {
	result := repo.DB.Table("permission_on_roles").Where("user_role_id = ?", userRoleID).Delete(&model.PermissionOnRole{})
	if result.Error != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to delete associated PermissionOnRole entries: %v", result.Error))
	}
	if result.RowsAffected == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("No PermissionOnRole entries found for user_role_id %s", userRoleID))
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

func (repo *RolePermissionRepository) GetAllRolePermissions(ctx context.Context) ([]QueryResult, error) {
	var queryResults []QueryResult

	err := repo.DB.Raw(`
   SELECT
    rp.id AS id,
    rp.created_at AS created_at,
    json_build_object(
        'id', r.id,
        'name', r.name,
        'description', r.description,
        'created_at', r.created_at
    ) AS role,
    COALESCE(
        json_agg(
            json_build_object(
                'id', por.id,
                'created_at', por.created_at,
                'permission_id', p.id,
                'user_role_id', rp.id,
                'permission', json_build_object(
                    'id', p.id,
                    'name', p.name,
                    'description', p.description
                )
            )
        ) FILTER (WHERE por.id IS NOT NULL),
        '[]'
    ) AS permissions
FROM
    role_permissions rp
JOIN
    roles r ON rp.role_id = r.id
LEFT JOIN
    permission_on_roles por ON rp.id = por.user_role_id
LEFT JOIN
    permissions p ON por.permission_id = p.id
GROUP BY
    rp.id, rp.created_at, r.id, r.name, r.description, r.created_at;
    `).Scan(&queryResults).Error

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to retrieve role permissions: %v", err))
	}

	return queryResults, nil
}

func (repo *RolePermissionRepository) FindRolePermissionByID(ctx context.Context, id string) (*model.RolePermission, error) {
	var rolePermission model.RolePermission
	err := repo.DB.Table("role_permissions").
		Preload("PermissionOnRoles", func(db *gorm.DB) *gorm.DB {
			return db.Table("permission_on_roles")
		}).
		Preload("PermissionOnRoles.Permission", func(db *gorm.DB) *gorm.DB {
			return db.Table("permissions")
		}).
		Preload("Role", func(db *gorm.DB) *gorm.DB {
			return db.Table("roles")
		}).
		Where("id = ?", id).First(&rolePermission).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("RolePermission with ID %s not found", id))
	} else if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to fetch RolePermission: %v", err))
	}
	return &rolePermission, nil
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
