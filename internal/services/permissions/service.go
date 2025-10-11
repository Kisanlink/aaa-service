package permissions

import (
	"context"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/permissions"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/resource_permissions"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/role_permissions"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/roles"
)

// Service handles permission business logic and evaluation
type Service struct {
	permissionRepo         *permissions.PermissionRepository
	rolePermissionRepo     *role_permissions.RolePermissionRepository
	resourcePermissionRepo *resource_permissions.ResourcePermissionRepository
	roleRepo               *roles.RoleRepository
	cache                  interfaces.CacheService
	audit                  interfaces.AuditService
	logger                 interfaces.Logger
}

// NewService creates a new Permission service instance
func NewService(
	permissionRepo *permissions.PermissionRepository,
	rolePermissionRepo *role_permissions.RolePermissionRepository,
	resourcePermissionRepo *resource_permissions.ResourcePermissionRepository,
	roleRepo *roles.RoleRepository,
	cache interfaces.CacheService,
	audit interfaces.AuditService,
	logger interfaces.Logger,
) *Service {
	return &Service{
		permissionRepo:         permissionRepo,
		rolePermissionRepo:     rolePermissionRepo,
		resourcePermissionRepo: resourcePermissionRepo,
		roleRepo:               roleRepo,
		cache:                  cache,
		audit:                  audit,
		logger:                 logger.Named("permission_service"),
	}
}

// EvaluationContext contains contextual information for permission evaluation
type EvaluationContext struct {
	OrganizationID string
	GroupID        string
	IPAddress      string
	Timestamp      time.Time
	CustomAttrs    map[string]interface{}
}

// EvaluationResult contains the result of a permission evaluation
type EvaluationResult struct {
	Allowed        bool
	Reason         string
	EffectiveRoles []*models.Role
	CacheHit       bool
	EvaluatedAt    time.Time
	EvaluationTime time.Duration
}

// PermissionFilter contains filter criteria for listing permissions
type PermissionFilter struct {
	ResourceID *string
	ActionID   *string
	IsActive   *bool
	Limit      int
	Offset     int
}

// ServiceInterface defines the contract for permission operations
type ServiceInterface interface {
	// CRUD operations
	CreatePermission(ctx context.Context, permission *models.Permission) error
	GetPermissionByID(ctx context.Context, id string) (*models.Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*models.Permission, error)
	UpdatePermission(ctx context.Context, permission *models.Permission) error
	DeletePermission(ctx context.Context, id string) error
	ListPermissions(ctx context.Context, filter *PermissionFilter) ([]*models.Permission, error)

	// Query operations
	GetPermissionsForRole(ctx context.Context, roleID string) ([]*models.Permission, error)
	GetPermissionsForResource(ctx context.Context, resourceID string) ([]*models.Permission, error)
	GetPermissionsForAction(ctx context.Context, actionID string) ([]*models.Permission, error)

	// Evaluation operations
	EvaluatePermission(ctx context.Context, userID, resourceType, resourceID, action string, evalCtx *EvaluationContext) (*EvaluationResult, error)

	// Cache operations
	InvalidateUserCache(ctx context.Context, userID string) error
	InvalidateRoleCache(ctx context.Context, roleID string) error
	InvalidatePermissionCache(ctx context.Context, permissionID string) error
}
