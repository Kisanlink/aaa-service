package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	authzedpb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AuthorizationService provides authorization services using SpiceDB
type AuthorizationService struct {
	spicedbClient *authzed.Client
	cacheService  interfaces.CacheService
	auditService  *AuditService
	logger        *zap.Logger
}

// AuthorizationServiceConfig contains configuration for AuthorizationService
type AuthorizationServiceConfig struct {
	SpiceDBToken string
	SpiceDBAddr  string
}

// NewAuthorizationService creates a new authorization service
func NewAuthorizationService(
	cacheService interfaces.CacheService,
	auditService *AuditService,
	config *AuthorizationServiceConfig,
	logger *zap.Logger,
) (*AuthorizationService, error) {
	// Initialize SpiceDB client
	spicedbClient, err := authzed.NewClient(
		config.SpiceDBAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(config.SpiceDBToken),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SpiceDB client: %w", err)
	}

	return &AuthorizationService{
		spicedbClient: spicedbClient,
		cacheService:  cacheService,
		auditService:  auditService,
		logger:        logger,
	}, nil
}

// Permission represents a permission check request
type Permission struct {
	UserID     string `json:"user_id"`
	Resource   string `json:"resource"`
	ResourceID string `json:"resource_id"`
	Action     string `json:"action"`
}

// PermissionResult represents the result of a permission check
type PermissionResult struct {
	Allowed     bool     `json:"allowed"`
	Reason      string   `json:"reason,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// BulkPermissionRequest represents a bulk permission check request
type BulkPermissionRequest struct {
	UserID      string       `json:"user_id"`
	Permissions []Permission `json:"permissions"`
}

// BulkPermissionResult represents the result of a bulk permission check
type BulkPermissionResult struct {
	Results map[string]*PermissionResult `json:"results"`
}

// CheckPermission checks if a user has permission to perform an action on a resource
func (s *AuthorizationService) CheckPermission(ctx context.Context, perm *Permission) (*PermissionResult, error) {
	// Create cache key for permission check
	cacheKey := fmt.Sprintf("permission:%s:%s:%s:%s", perm.UserID, perm.Resource, perm.ResourceID, perm.Action)

	// Try to get result from cache first
	if cachedResult, exists := s.cacheService.Get(cacheKey); exists {
		// In a real implementation, you'd deserialize the cached result
		// For now, we'll skip cache and always check SpiceDB
		_ = cachedResult
	}

	// Check permission in SpiceDB
	req := &authzedpb.CheckPermissionRequest{
		Resource: &authzedpb.ObjectReference{
			ObjectType: fmt.Sprintf("aaa/%s", perm.Resource),
			ObjectId:   perm.ResourceID,
		},
		Permission: perm.Action,
		Subject: &authzedpb.SubjectReference{
			Object: &authzedpb.ObjectReference{
				ObjectType: "aaa/user",
				ObjectId:   perm.UserID,
			},
		},
	}

	resp, err := s.spicedbClient.CheckPermission(ctx, req)
	if err != nil {
		s.logger.Error("Failed to check permission in SpiceDB",
			zap.String("user_id", perm.UserID),
			zap.String("resource", perm.Resource),
			zap.String("resource_id", perm.ResourceID),
			zap.String("action", perm.Action),
			zap.Error(err))
		return nil, errors.NewInternalError(fmt.Errorf("failed to check permission: %w", err))
	}

	result := &PermissionResult{
		Allowed: resp.Permissionship == authzedpb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION,
	}

	if !result.Allowed {
		result.Reason = "insufficient permissions"
	}

	// Cache the result for a short time
	if err := s.cacheService.Set(cacheKey, result, 300); err != nil { // 5 minutes cache
		s.logger.Warn("Failed to cache permission result", zap.Error(err))
	}

	// Audit permission check if denied
	if !result.Allowed && s.auditService != nil {
		s.auditService.LogAccessDenied(ctx, perm.UserID, perm.Action, perm.Resource, perm.ResourceID, "insufficient permissions")
	}

	return result, nil
}

// CheckBulkPermissions checks multiple permissions for a user
func (s *AuthorizationService) CheckBulkPermissions(ctx context.Context, req *BulkPermissionRequest) (*BulkPermissionResult, error) {
	results := make(map[string]*PermissionResult)

	for i, perm := range req.Permissions {
		perm.UserID = req.UserID // Ensure user ID is set
		result, err := s.CheckPermission(ctx, &perm)
		if err != nil {
			s.logger.Error("Failed to check permission in bulk request",
				zap.Int("permission_index", i),
				zap.String("user_id", perm.UserID),
				zap.Error(err))
			result = &PermissionResult{
				Allowed: false,
				Reason:  "error checking permission",
			}
		}

		// Use a composite key for the result
		key := fmt.Sprintf("%s:%s:%s", perm.Resource, perm.ResourceID, perm.Action)
		results[key] = result
	}

	return &BulkPermissionResult{
		Results: results,
	}, nil
}

// GetUserPermissions retrieves all permissions for a user on a specific resource type
func (s *AuthorizationService) GetUserPermissions(ctx context.Context, userID, resourceType string) ([]string, error) {
	// Use SpiceDB's LookupResources to find all resources of a type the user has access to
	actions := []string{"view", "edit", "delete", "manage", "create", "read", "update"}
	var permissions []string

	for _, action := range actions {
		req := &authzedpb.LookupResourcesRequest{
			ResourceObjectType: fmt.Sprintf("aaa/%s", resourceType),
			Permission:         action,
			Subject: &authzedpb.SubjectReference{
				Object: &authzedpb.ObjectReference{
					ObjectType: "aaa/user",
					ObjectId:   userID,
				},
			},
		}

		stream, err := s.spicedbClient.LookupResources(ctx, req)
		if err != nil {
			s.logger.Error("Failed to lookup resources in SpiceDB",
				zap.String("user_id", userID),
				zap.String("resource_type", resourceType),
				zap.String("action", action),
				zap.Error(err))
			continue
		}

		hasPermission := false
		for {
			resp, err := stream.Recv()
			if err != nil {
				break // End of stream or error
			}
			if resp.ResourceObjectId != "" {
				hasPermission = true
				break
			}
		}

		if hasPermission {
			permissions = append(permissions, fmt.Sprintf("%s:%s", resourceType, action))
		}
	}

	return permissions, nil
}

// GrantPermission grants a permission to a user by creating a relationship in SpiceDB
func (s *AuthorizationService) GrantPermission(ctx context.Context, userID, resource, resourceID, relation string) error {
	req := &authzedpb.WriteRelationshipsRequest{
		Updates: []*authzedpb.RelationshipUpdate{
			{
				Operation: authzedpb.RelationshipUpdate_OPERATION_CREATE,
				Relationship: &authzedpb.Relationship{
					Resource: &authzedpb.ObjectReference{
						ObjectType: fmt.Sprintf("aaa/%s", resource),
						ObjectId:   resourceID,
					},
					Relation: relation,
					Subject: &authzedpb.SubjectReference{
						Object: &authzedpb.ObjectReference{
							ObjectType: "aaa/user",
							ObjectId:   userID,
						},
					},
				},
			},
		},
	}

	_, err := s.spicedbClient.WriteRelationships(ctx, req)
	if err != nil {
		s.logger.Error("Failed to grant permission in SpiceDB",
			zap.String("user_id", userID),
			zap.String("resource", resource),
			zap.String("resource_id", resourceID),
			zap.String("relation", relation),
			zap.Error(err))
		return errors.NewInternalError(fmt.Errorf("failed to grant permission: %w", err))
	}

	// Audit permission grant
	if s.auditService != nil {
		s.auditService.LogPermissionChange(ctx, userID, "grant", resource, resourceID, relation, map[string]interface{}{
			"granted_by": userID, // In practice, this would be the admin user ID
		})
	}

	s.logger.Info("Permission granted",
		zap.String("user_id", userID),
		zap.String("resource", resource),
		zap.String("resource_id", resourceID),
		zap.String("relation", relation))

	return nil
}

// RevokePermission revokes a permission from a user by deleting a relationship in SpiceDB
func (s *AuthorizationService) RevokePermission(ctx context.Context, userID, resource, resourceID, relation string) error {
	req := &authzedpb.WriteRelationshipsRequest{
		Updates: []*authzedpb.RelationshipUpdate{
			{
				Operation: authzedpb.RelationshipUpdate_OPERATION_DELETE,
				Relationship: &authzedpb.Relationship{
					Resource: &authzedpb.ObjectReference{
						ObjectType: fmt.Sprintf("aaa/%s", resource),
						ObjectId:   resourceID,
					},
					Relation: relation,
					Subject: &authzedpb.SubjectReference{
						Object: &authzedpb.ObjectReference{
							ObjectType: "aaa/user",
							ObjectId:   userID,
						},
					},
				},
			},
		},
	}

	_, err := s.spicedbClient.WriteRelationships(ctx, req)
	if err != nil {
		s.logger.Error("Failed to revoke permission in SpiceDB",
			zap.String("user_id", userID),
			zap.String("resource", resource),
			zap.String("resource_id", resourceID),
			zap.String("relation", relation),
			zap.Error(err))
		return errors.NewInternalError(fmt.Errorf("failed to revoke permission: %w", err))
	}

	// Audit permission revocation
	if s.auditService != nil {
		s.auditService.LogPermissionChange(ctx, userID, "revoke", resource, resourceID, relation, map[string]interface{}{
			"revoked_by": userID, // In practice, this would be the admin user ID
		})
	}

	s.logger.Info("Permission revoked",
		zap.String("user_id", userID),
		zap.String("resource", resource),
		zap.String("resource_id", resourceID),
		zap.String("relation", relation))

	return nil
}

// AssignRoleToUser assigns a role to a user in SpiceDB
func (s *AuthorizationService) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	req := &authzedpb.WriteRelationshipsRequest{
		Updates: []*authzedpb.RelationshipUpdate{
			{
				Operation: authzedpb.RelationshipUpdate_OPERATION_CREATE,
				Relationship: &authzedpb.Relationship{
					Resource: &authzedpb.ObjectReference{
						ObjectType: "aaa/user",
						ObjectId:   userID,
					},
					Relation: "role",
					Subject: &authzedpb.SubjectReference{
						Object: &authzedpb.ObjectReference{
							ObjectType: "aaa/role",
							ObjectId:   roleID,
						},
					},
				},
			},
		},
	}

	_, err := s.spicedbClient.WriteRelationships(ctx, req)
	if err != nil {
		s.logger.Error("Failed to assign role to user in SpiceDB",
			zap.String("user_id", userID),
			zap.String("role_id", roleID),
			zap.Error(err))
		return errors.NewInternalError(fmt.Errorf("failed to assign role: %w", err))
	}

	// Audit role assignment
	if s.auditService != nil {
		s.auditService.LogRoleChange(ctx, userID, "assign", roleID, map[string]interface{}{
			"assigned_by": userID, // In practice, this would be the admin user ID
		})
	}

	s.logger.Info("Role assigned to user",
		zap.String("user_id", userID),
		zap.String("role_id", roleID))

	return nil
}

// RemoveRoleFromUser removes a role from a user in SpiceDB
func (s *AuthorizationService) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	req := &authzedpb.WriteRelationshipsRequest{
		Updates: []*authzedpb.RelationshipUpdate{
			{
				Operation: authzedpb.RelationshipUpdate_OPERATION_DELETE,
				Relationship: &authzedpb.Relationship{
					Resource: &authzedpb.ObjectReference{
						ObjectType: "aaa/user",
						ObjectId:   userID,
					},
					Relation: "role",
					Subject: &authzedpb.SubjectReference{
						Object: &authzedpb.ObjectReference{
							ObjectType: "aaa/role",
							ObjectId:   roleID,
						},
					},
				},
			},
		},
	}

	_, err := s.spicedbClient.WriteRelationships(ctx, req)
	if err != nil {
		s.logger.Error("Failed to remove role from user in SpiceDB",
			zap.String("user_id", userID),
			zap.String("role_id", roleID),
			zap.Error(err))
		return errors.NewInternalError(fmt.Errorf("failed to remove role: %w", err))
	}

	// Audit role removal
	if s.auditService != nil {
		s.auditService.LogRoleChange(ctx, userID, "remove", roleID, map[string]interface{}{
			"removed_by": userID, // In practice, this would be the admin user ID
		})
	}

	s.logger.Info("Role removed from user",
		zap.String("user_id", userID),
		zap.String("role_id", roleID))

	return nil
}

// ValidateAPIEndpointAccess validates access to API endpoints
func (s *AuthorizationService) ValidateAPIEndpointAccess(ctx context.Context, userID, method, endpoint string) (*PermissionResult, error) {
	// Parse endpoint to determine resource and action
	resource, action := s.parseEndpointToResourceAction(method, endpoint)

	perm := &Permission{
		UserID:     userID,
		Resource:   resource,
		ResourceID: resource, // For API endpoints, we use the resource type as ID
		Action:     action,
	}

	return s.CheckPermission(ctx, perm)
}

// parseEndpointToResourceAction parses HTTP method and endpoint to determine resource and action
func (s *AuthorizationService) parseEndpointToResourceAction(method, endpoint string) (resource, action string) {
	// Remove leading slash and split by slash
	endpoint = strings.TrimPrefix(endpoint, "/")
	parts := strings.Split(endpoint, "/")

	if len(parts) == 0 {
		return "api", "access"
	}

	// Skip "api" prefix if present
	if parts[0] == "api" && len(parts) > 1 {
		parts = parts[1:]
	}

	// Skip version prefix (v1, v2, etc.) if present
	if len(parts) > 0 && strings.HasPrefix(parts[0], "v") {
		if len(parts) > 1 {
			parts = parts[1:]
		}
	}

	if len(parts) == 0 {
		return "api", "access"
	}

	resource = parts[0]

	// Map HTTP methods to actions
	switch strings.ToUpper(method) {
	case "GET":
		if len(parts) > 1 && parts[1] != "" {
			action = "view" // GET /resource/id
		} else {
			action = "read" // GET /resource (list)
		}
	case "POST":
		action = "create"
	case "PUT", "PATCH":
		action = "update"
	case "DELETE":
		action = "delete"
	default:
		action = "access"
	}

	return resource, action
}

// GetResourceHierarchy gets the hierarchy of a resource for inheritance checks
func (s *AuthorizationService) GetResourceHierarchy(ctx context.Context, resourceType, resourceID string) ([]string, error) {
	// This would implement traversal of resource hierarchy in SpiceDB
	// For now, return a simple hierarchy
	hierarchy := []string{resourceID}

	// Example: for user:123, hierarchy might be [user:123, organization:456, system]
	// This would be implemented based on your specific resource relationships

	return hierarchy, nil
}

// EvaluatePolicy evaluates a policy against user attributes and context
func (s *AuthorizationService) EvaluatePolicy(ctx context.Context, userID string, policy map[string]interface{}) (*PermissionResult, error) {
	// Extract policy components
	resource, ok := policy["resource"].(string)
	if !ok {
		return &PermissionResult{
			Allowed: false,
			Reason:  "invalid policy: missing resource",
		}, nil
	}

	action, ok := policy["action"].(string)
	if !ok {
		return &PermissionResult{
			Allowed: false,
			Reason:  "invalid policy: missing action",
		}, nil
	}

	resourceID, _ := policy["resource_id"].(string)
	if resourceID == "" {
		resourceID = "default"
	}

	// First check basic permission
	perm := &Permission{
		UserID:     userID,
		Resource:   resource,
		ResourceID: resourceID,
		Action:     action,
	}

	basicResult, err := s.CheckPermission(ctx, perm)
	if err != nil {
		return &PermissionResult{
			Allowed: false,
			Reason:  fmt.Sprintf("permission check failed: %v", err),
		}, err
	}

	if !basicResult.Allowed {
		return &PermissionResult{
			Allowed: false,
			Reason:  "basic permission denied",
		}, nil
	}

	// Evaluate additional conditions if they exist
	conditions, ok := policy["conditions"].(map[string]interface{})
	if !ok || len(conditions) == 0 {
		// No conditions to evaluate, return basic permission result
		return basicResult, nil
	}

	// Evaluate each condition
	for conditionType, conditionValue := range conditions {
		switch conditionType {
		case "time_range":
			if !s.evaluateTimeRangeCondition(ctx, conditionValue) {
				return &PermissionResult{
					Allowed: false,
					Reason:  "time range condition not met",
				}, nil
			}

		case "ip_range":
			if !s.evaluateIPRangeCondition(ctx, conditionValue) {
				return &PermissionResult{
					Allowed: false,
					Reason:  "IP range condition not met",
				}, nil
			}

		case "mfa_required":
			if conditionValue.(bool) && !s.evaluateMFACondition(ctx, userID) {
				return &PermissionResult{
					Allowed: false,
					Reason:  "MFA required but not verified",
				}, nil
			}

		case "user_status":
			if !s.evaluateUserStatusCondition(ctx, userID, conditionValue) {
				return &PermissionResult{
					Allowed: false,
					Reason:  "user status condition not met",
				}, nil
			}

		case "resource_owner":
			if conditionValue.(bool) && !s.evaluateResourceOwnerCondition(ctx, userID, resource, resourceID) {
				return &PermissionResult{
					Allowed: false,
					Reason:  "resource ownership condition not met",
				}, nil
			}

		case "max_requests_per_hour":
			if !s.evaluateRateLimitCondition(ctx, userID, conditionValue) {
				return &PermissionResult{
					Allowed: false,
					Reason:  "rate limit exceeded",
				}, nil
			}

		default:
			s.logger.Warn("Unknown policy condition",
				zap.String("condition_type", conditionType),
				zap.String("user_id", userID))
		}
	}

	return &PermissionResult{
		Allowed: true,
		Reason:  "policy evaluation passed with all conditions met",
	}, nil
}

// evaluateTimeRangeCondition checks if current time is within allowed range
func (s *AuthorizationService) evaluateTimeRangeCondition(ctx context.Context, condition interface{}) bool {
	timeRange, ok := condition.(string)
	if !ok {
		return false
	}

	switch timeRange {
	case "business_hours":
		// 9 AM to 5 PM weekdays
		now := time.Now()
		hour := now.Hour()
		weekday := now.Weekday()
		return weekday >= time.Monday && weekday <= time.Friday && hour >= 9 && hour < 17
	case "24x7":
		return true
	case "weekdays":
		now := time.Now()
		weekday := now.Weekday()
		return weekday >= time.Monday && weekday <= time.Friday
	default:
		return false
	}
}

// evaluateIPRangeCondition checks if request comes from allowed IP range
func (s *AuthorizationService) evaluateIPRangeCondition(ctx context.Context, condition interface{}) bool {
	ipRange, ok := condition.(string)
	if !ok {
		return false
	}

	// Get IP from context
	clientIP, ok := ctx.Value("client_ip").(string)
	if !ok {
		return false
	}

	switch ipRange {
	case "internal":
		// Check if IP is in internal ranges (RFC 1918)
		return strings.HasPrefix(clientIP, "10.") ||
			strings.HasPrefix(clientIP, "172.16.") ||
			strings.HasPrefix(clientIP, "192.168.") ||
			clientIP == "127.0.0.1"
	case "any":
		return true
	default:
		// Custom IP range checking would go here
		return false
	}
}

// evaluateMFACondition checks if user has completed MFA for this session
func (s *AuthorizationService) evaluateMFACondition(ctx context.Context, userID string) bool {
	// Check if user has MFA token in session
	mfaVerified, ok := ctx.Value("mfa_verified").(bool)
	if !ok {
		return false
	}
	return mfaVerified
}

// evaluateUserStatusCondition checks user status requirements
func (s *AuthorizationService) evaluateUserStatusCondition(ctx context.Context, userID string, condition interface{}) bool {
	requiredStatus, ok := condition.(string)
	if !ok {
		return false
	}

	// This would query user status from database in practice
	// For now, assume user is active
	userStatus := "active"

	return userStatus == requiredStatus
}

// evaluateResourceOwnerCondition checks if user owns the resource
func (s *AuthorizationService) evaluateResourceOwnerCondition(ctx context.Context, userID, resourceType, resourceID string) bool {
	// Check ownership in SpiceDB
	req := &authzedpb.CheckPermissionRequest{
		Resource: &authzedpb.ObjectReference{
			ObjectType: fmt.Sprintf("aaa/%s", resourceType),
			ObjectId:   resourceID,
		},
		Permission: "owner",
		Subject: &authzedpb.SubjectReference{
			Object: &authzedpb.ObjectReference{
				ObjectType: "aaa/user",
				ObjectId:   userID,
			},
		},
	}

	resp, err := s.spicedbClient.CheckPermission(ctx, req)
	if err != nil {
		s.logger.Error("Failed to check resource ownership",
			zap.String("user_id", userID),
			zap.String("resource_type", resourceType),
			zap.String("resource_id", resourceID),
			zap.Error(err))
		return false
	}

	return resp.Permissionship == authzedpb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION
}

// evaluateRateLimitCondition checks if user is within rate limits
func (s *AuthorizationService) evaluateRateLimitCondition(ctx context.Context, userID string, condition interface{}) bool {
	maxRequests, ok := condition.(float64)
	if !ok {
		return false
	}

	// Check current request count from cache
	cacheKey := fmt.Sprintf("rate_limit:%s:hour", userID)
	currentCount, exists := s.cacheService.Get(cacheKey)
	if !exists {
		// First request in this hour
		s.cacheService.Set(cacheKey, 1, 3600) // 1 hour TTL
		return true
	}

	count, ok := currentCount.(int)
	if !ok {
		return false
	}

	if float64(count) >= maxRequests {
		return false
	}

	// Increment counter
	s.cacheService.Set(cacheKey, count+1, 3600)
	return true
}

// SpiceDB resource type helper methods using our model constants

// CheckUserPermission checks if a user has a specific permission on another user
func (s *AuthorizationService) CheckUserPermission(ctx context.Context, subjectUserID, objectUserID, permission string) (bool, error) {
	perm := &Permission{
		UserID:     subjectUserID,
		Resource:   "user",
		ResourceID: objectUserID,
		Action:     permission,
	}
	result, err := s.CheckPermission(ctx, perm)
	if err != nil {
		return false, err
	}
	return result.Allowed, nil
}

// CheckRolePermission checks if a user has a specific permission on a role
func (s *AuthorizationService) CheckRolePermission(ctx context.Context, userID, roleID, permission string) (bool, error) {
	perm := &Permission{
		UserID:     userID,
		Resource:   "role",
		ResourceID: roleID,
		Action:     permission,
	}
	result, err := s.CheckPermission(ctx, perm)
	if err != nil {
		return false, err
	}
	return result.Allowed, nil
}

// CheckSystemPermission checks if a user has a system-level permission
func (s *AuthorizationService) CheckSystemPermission(ctx context.Context, userID, permission string) (bool, error) {
	perm := &Permission{
		UserID:     userID,
		Resource:   "system",
		ResourceID: "system",
		Action:     permission,
	}
	result, err := s.CheckPermission(ctx, perm)
	if err != nil {
		return false, err
	}
	return result.Allowed, nil
}

// CheckAuditLogPermission checks if a user has permission to access audit logs
func (s *AuthorizationService) CheckAuditLogPermission(ctx context.Context, userID, auditLogID, permission string) (bool, error) {
	perm := &Permission{
		UserID:     userID,
		Resource:   "audit_log",
		ResourceID: auditLogID,
		Action:     permission,
	}
	result, err := s.CheckPermission(ctx, perm)
	if err != nil {
		return false, err
	}
	return result.Allowed, nil
}

// CheckAPIEndpointPermission checks if a user has permission to access an API endpoint
func (s *AuthorizationService) CheckAPIEndpointPermission(ctx context.Context, userID, endpoint, httpMethod string) (bool, error) {
	perm := &Permission{
		UserID:     userID,
		Resource:   "api_endpoint",
		ResourceID: endpoint,
		Action:     httpMethod,
	}
	result, err := s.CheckPermission(ctx, perm)
	if err != nil {
		return false, err
	}
	return result.Allowed, nil
}

// GrantUserRole grants a role to a user in SpiceDB
func (s *AuthorizationService) GrantUserRole(ctx context.Context, userID, roleID string) error {
	relationship := &authzedpb.Relationship{
		Resource: &authzedpb.ObjectReference{
			ObjectType: models.ResourceTypeUser,
			ObjectId:   userID,
		},
		Relation: "role",
		Subject: &authzedpb.SubjectReference{
			Object: &authzedpb.ObjectReference{
				ObjectType: models.ResourceTypeRole,
				ObjectId:   roleID,
			},
		},
	}

	return s.writeRelationship(ctx, relationship, authzedpb.RelationshipUpdate_OPERATION_CREATE)
}

// RevokeUserRole revokes a role from a user in SpiceDB
func (s *AuthorizationService) RevokeUserRole(ctx context.Context, userID, roleID string) error {
	relationship := &authzedpb.Relationship{
		Resource: &authzedpb.ObjectReference{
			ObjectType: models.ResourceTypeUser,
			ObjectId:   userID,
		},
		Relation: "role",
		Subject: &authzedpb.SubjectReference{
			Object: &authzedpb.ObjectReference{
				ObjectType: models.ResourceTypeRole,
				ObjectId:   roleID,
			},
		},
	}

	return s.writeRelationship(ctx, relationship, authzedpb.RelationshipUpdate_OPERATION_DELETE)
}

// GrantRolePermission grants a permission to a role in SpiceDB
func (s *AuthorizationService) GrantRolePermission(ctx context.Context, roleID, permissionID string) error {
	relationship := &authzedpb.Relationship{
		Resource: &authzedpb.ObjectReference{
			ObjectType: models.ResourceTypeRole,
			ObjectId:   roleID,
		},
		Relation: "perms",
		Subject: &authzedpb.SubjectReference{
			Object: &authzedpb.ObjectReference{
				ObjectType: models.ResourceTypePermission,
				ObjectId:   permissionID,
			},
		},
	}

	return s.writeRelationship(ctx, relationship, authzedpb.RelationshipUpdate_OPERATION_CREATE)
}

// SetResourceOwner sets the owner of a resource in SpiceDB
func (s *AuthorizationService) SetResourceOwner(ctx context.Context, resourceType, resourceID, ownerUserID string) error {
	relationship := &authzedpb.Relationship{
		Resource: &authzedpb.ObjectReference{
			ObjectType: resourceType,
			ObjectId:   resourceID,
		},
		Relation: "owner",
		Subject: &authzedpb.SubjectReference{
			Object: &authzedpb.ObjectReference{
				ObjectType: models.ResourceTypeUser,
				ObjectId:   ownerUserID,
			},
		},
	}

	return s.writeRelationship(ctx, relationship, authzedpb.RelationshipUpdate_OPERATION_CREATE)
}

// SetResourceParent establishes a parent-child relationship between resources
func (s *AuthorizationService) SetResourceParent(ctx context.Context, resourceType, childResourceID, parentResourceID string) error {
	relationship := &authzedpb.Relationship{
		Resource: &authzedpb.ObjectReference{
			ObjectType: resourceType,
			ObjectId:   childResourceID,
		},
		Relation: "parent",
		Subject: &authzedpb.SubjectReference{
			Object: &authzedpb.ObjectReference{
				ObjectType: resourceType,
				ObjectId:   parentResourceID,
			},
		},
	}

	return s.writeRelationship(ctx, relationship, authzedpb.RelationshipUpdate_OPERATION_CREATE)
}

// writeRelationship is a helper method to write a relationship to SpiceDB
func (s *AuthorizationService) writeRelationship(ctx context.Context, relationship *authzedpb.Relationship, operation authzedpb.RelationshipUpdate_Operation) error {
	update := &authzedpb.RelationshipUpdate{
		Operation:    operation,
		Relationship: relationship,
	}

	request := &authzedpb.WriteRelationshipsRequest{
		Updates: []*authzedpb.RelationshipUpdate{update},
	}

	_, err := s.spicedbClient.WriteRelationships(ctx, request)
	if err != nil {
		s.logger.Error("Failed to write relationship to SpiceDB",
			zap.Error(err),
			zap.String("resource_type", relationship.Resource.ObjectType),
			zap.String("resource_id", relationship.Resource.ObjectId),
			zap.String("relation", relationship.Relation),
			zap.String("subject_type", relationship.Subject.Object.ObjectType),
			zap.String("subject_id", relationship.Subject.Object.ObjectId),
		)
		return fmt.Errorf("failed to write relationship to SpiceDB: %w", err)
	}

	return nil
}
