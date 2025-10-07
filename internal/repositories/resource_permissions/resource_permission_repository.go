package resource_permissions

import (
	"context"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// ResourcePermissionRepository handles database operations for ResourcePermission (Model 2: Direct)
type ResourcePermissionRepository struct {
	*base.BaseFilterableRepository[*models.ResourcePermission]
	dbManager db.DBManager
}

// NewResourcePermissionRepository creates a new ResourcePermissionRepository instance
func NewResourcePermissionRepository(dbManager db.DBManager) *ResourcePermissionRepository {
	baseRepo := base.NewBaseFilterableRepository[*models.ResourcePermission]()
	baseRepo.SetDBManager(dbManager)
	return &ResourcePermissionRepository{
		BaseFilterableRepository: baseRepo,
		dbManager:                dbManager,
	}
}

// ResourcePermissionRepositoryInterface defines the contract for resource-permission operations
type ResourcePermissionRepositoryInterface interface {
	// Assignment operations
	Assign(ctx context.Context, roleID, resourceType, resourceID, action string) error
	AssignBatch(ctx context.Context, assignments []ResourcePermissionAssignment) error
	AssignActions(ctx context.Context, roleID, resourceType, resourceID string, actions []string) error

	// Revocation operations
	Revoke(ctx context.Context, roleID, resourceType, resourceID, action string) error
	RevokeBatch(ctx context.Context, roleID, resourceType, resourceID string, actions []string) error
	RevokeAllForRole(ctx context.Context, roleID string) error
	RevokeAllForResource(ctx context.Context, resourceType, resourceID string) error

	// Query operations
	GetByRoleID(ctx context.Context, roleID string) ([]*models.ResourcePermission, error)
	GetByResource(ctx context.Context, resourceType, resourceID string) ([]*models.ResourcePermission, error)
	GetByResourceType(ctx context.Context, resourceType string) ([]*models.ResourcePermission, error)
	GetByAction(ctx context.Context, action string) ([]*models.ResourcePermission, error)
	FindByFilter(ctx context.Context, filter *base.Filter) ([]*models.ResourcePermission, error)

	// Evaluation operations
	HasPermission(ctx context.Context, roleID, resourceType, resourceID, action string) (bool, error)
	GetRolePermissions(ctx context.Context, roleID, resourceType string) ([]*models.ResourcePermission, error)
	GetAllowedActions(ctx context.Context, roleID, resourceType, resourceID string) ([]string, error)
	CheckMultiplePermissions(ctx context.Context, roleIDs []string, resourceType, resourceID, action string) (bool, error)

	// Activation/Deactivation
	Activate(ctx context.Context, id string) error
	Deactivate(ctx context.Context, id string) error
	GetActive(ctx context.Context, roleID string) ([]*models.ResourcePermission, error)
}

// ResourcePermissionAssignment represents a batch assignment structure
type ResourcePermissionAssignment struct {
	RoleID       string
	ResourceType string
	ResourceID   string
	Action       string
}
