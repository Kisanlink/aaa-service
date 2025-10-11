package role_permissions

import (
	"context"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// RolePermissionRepository handles database operations for RolePermission join table
type RolePermissionRepository struct {
	*base.BaseFilterableRepository[*models.RolePermission]
	dbManager db.DBManager
}

// NewRolePermissionRepository creates a new RolePermissionRepository instance
func NewRolePermissionRepository(dbManager db.DBManager) *RolePermissionRepository {
	baseRepo := base.NewBaseFilterableRepository[*models.RolePermission]()
	baseRepo.SetDBManager(dbManager)
	return &RolePermissionRepository{
		BaseFilterableRepository: baseRepo,
		dbManager:                dbManager,
	}
}

// RolePermissionRepositoryInterface defines the contract for role-permission operations
type RolePermissionRepositoryInterface interface {
	// Assignment operations
	Assign(ctx context.Context, roleID, permissionID string) error
	AssignBatch(ctx context.Context, roleID string, permissionIDs []string) error
	AssignMultipleRoles(ctx context.Context, roleIDs []string, permissionID string) error

	// Revocation operations
	Revoke(ctx context.Context, roleID, permissionID string) error
	RevokeBatch(ctx context.Context, roleID string, permissionIDs []string) error
	RevokeAll(ctx context.Context, roleID string) error

	// Query operations
	GetByRoleID(ctx context.Context, roleID string) ([]*models.RolePermission, error)
	GetByPermissionID(ctx context.Context, permissionID string) ([]*models.RolePermission, error)
	GetByRoleAndPermission(ctx context.Context, roleID, permissionID string) (*models.RolePermission, error)
	Exists(ctx context.Context, roleID, permissionID string) (bool, error)
	CountByRole(ctx context.Context, roleID string) (int64, error)
	CountByPermission(ctx context.Context, permissionID string) (int64, error)

	// Bulk query operations
	GetPermissionsByRoles(ctx context.Context, roleIDs []string) ([]*models.RolePermission, error)
	GetRolesByPermissions(ctx context.Context, permissionIDs []string) ([]*models.RolePermission, error)

	// Active/Inactive operations
	Activate(ctx context.Context, roleID, permissionID string) error
	Deactivate(ctx context.Context, roleID, permissionID string) error
	GetActive(ctx context.Context, roleID string) ([]*models.RolePermission, error)
}
