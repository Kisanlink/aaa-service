package role_assignments

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

// Service handles role-permission assignment operations
type Service struct {
	roleRepo               *roles.RoleRepository
	rolePermissionRepo     *role_permissions.RolePermissionRepository
	resourcePermissionRepo *resource_permissions.ResourcePermissionRepository
	permissionRepo         *permissions.PermissionRepository
	auditRepo              interfaces.AuditRepository
	cache                  interfaces.CacheService
	audit                  interfaces.AuditService
	logger                 interfaces.Logger
}

// NewService creates a new Role Assignment service instance
func NewService(
	roleRepo *roles.RoleRepository,
	rolePermissionRepo *role_permissions.RolePermissionRepository,
	resourcePermissionRepo *resource_permissions.ResourcePermissionRepository,
	permissionRepo *permissions.PermissionRepository,
	auditRepo interfaces.AuditRepository,
	cache interfaces.CacheService,
	audit interfaces.AuditService,
	logger interfaces.Logger,
) *Service {
	return &Service{
		roleRepo:               roleRepo,
		rolePermissionRepo:     rolePermissionRepo,
		resourcePermissionRepo: resourcePermissionRepo,
		permissionRepo:         permissionRepo,
		auditRepo:              auditRepo,
		cache:                  cache,
		audit:                  audit,
		logger:                 logger.Named("role_assignment_service"),
	}
}

// ResourceActionAssignment represents a batch assignment of resource-actions
type ResourceActionAssignment struct {
	ResourceType string
	ResourceID   string
	Actions      []string
}

// PermissionAssignment represents a permission assignment with time constraints
type PermissionAssignment struct {
	PermissionIDs  []string
	EffectiveFrom  *time.Time
	EffectiveUntil *time.Time
}

// AssignmentHistory represents the history of a permission assignment
type AssignmentHistory struct {
	RoleID         string
	PermissionID   string
	AssignedAt     time.Time
	AssignedBy     string
	RevokedAt      *time.Time
	RevokedBy      *string
	AssignmentType string // "permission" or "resource"
}

// ServiceInterface defines the contract for role assignment operations
type ServiceInterface interface {
	// Model 1: Permission assignments
	AssignPermissionToRole(ctx context.Context, roleID, permissionID string, assignedBy string) error
	AssignPermissionsToRole(ctx context.Context, roleID string, permissionIDs []string, assignedBy string) error
	AssignPermissionToMultipleRoles(ctx context.Context, permissionID string, roleIDs []string, assignedBy string) error

	// Model 2: Resource-action assignments
	AssignResourceActionToRole(ctx context.Context, roleID, resourceType, resourceID, action string, assignedBy string) error
	AssignResourceActionsToRole(ctx context.Context, roleID string, assignments []ResourceActionAssignment, assignedBy string) error

	// Revocation operations
	RevokePermissionFromRole(ctx context.Context, roleID, permissionID string, revokedBy string) error
	RevokePermissionsFromRole(ctx context.Context, roleID string, permissionIDs []string, revokedBy string) error
	RevokeResourceActionFromRole(ctx context.Context, roleID, resourceType, resourceID, action string, revokedBy string) error
	RevokeAllPermissionsFromRole(ctx context.Context, roleID string, revokedBy string) error

	// Query operations
	GetRolePermissions(ctx context.Context, roleID string) ([]*models.Permission, error)
	GetRoleResources(ctx context.Context, roleID string) ([]*models.ResourcePermission, error)
	GetRolesWithPermission(ctx context.Context, permissionID string) ([]*models.Role, error)
	GetRolesWithResourceAccess(ctx context.Context, resourceType, resourceID, action string) ([]*models.Role, error)

	// Inheritance operations
	GetInheritedRoles(ctx context.Context, roleID string) ([]*models.Role, error)
	GetEffectiveRolesForUser(ctx context.Context, userID, orgID, groupID string) ([]*models.Role, error)
	GetEffectivePermissionsForUser(ctx context.Context, userID, orgID, groupID string) ([]*models.Permission, error)
	GetUserAccessToResource(ctx context.Context, userID, resourceType, resourceID string) ([]string, error)
}
