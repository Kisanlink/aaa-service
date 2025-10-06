package grpc_server

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/internal/services"
	pb "github.com/Kisanlink/aaa-service/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthorizationHandler implements authorization-related gRPC services
type AuthorizationHandler struct {
	pb.UnimplementedAuthorizationServiceServer
	authzService *services.AuthorizationService
	logger       *zap.Logger
}

// NewAuthorizationHandler creates a new authorization handler
func NewAuthorizationHandler(authzService *services.AuthorizationService, logger *zap.Logger) *AuthorizationHandler {
	return &AuthorizationHandler{
		authzService: authzService,
		logger:       logger,
	}
}

// Check implements the Check RPC method for authorization
func (h *AuthorizationHandler) Check(ctx context.Context, req *pb.CheckRequest) (*pb.CheckResponse, error) {
	h.logger.Info("gRPC Check request",
		zap.String("principal_id", req.PrincipalId),
		zap.String("resource_type", req.ResourceType),
		zap.String("resource_id", req.ResourceId),
		zap.String("action", req.Action))

	// Convert to internal permission format
	permission := &services.Permission{
		UserID:     req.PrincipalId,
		Resource:   req.ResourceType,
		ResourceID: req.ResourceId,
		Action:     req.Action,
	}

	result, err := h.authzService.CheckPermission(ctx, permission)
	if err != nil {
		h.logger.Error("Permission check failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "permission check failed: %v", err)
	}

	h.logger.Info("Permission check completed",
		zap.String("principal_id", req.PrincipalId),
		zap.Bool("allowed", result.Allowed))

	return &pb.CheckResponse{
		Allowed:          result.Allowed,
		DecisionId:       result.DecisionID,
		ConsistencyToken: result.ConsistencyToken,
	}, nil
}

// BatchCheck implements the BatchCheck RPC method for authorization
func (h *AuthorizationHandler) BatchCheck(ctx context.Context, req *pb.BatchCheckRequest) (*pb.BatchCheckResponse, error) {
	h.logger.Info("gRPC BatchCheck request",
		zap.Int("request_count", len(req.Items)))

	// Convert to internal format
	var permissions []services.Permission
	for _, checkReq := range req.Items {
		permission := &services.Permission{
			UserID:     checkReq.PrincipalId,
			Resource:   checkReq.ResourceType,
			ResourceID: checkReq.ResourceId,
			Action:     checkReq.Action,
		}
		permissions = append(permissions, *permission)
	}

	// Use the first item's principal ID as the user ID for bulk check
	if len(req.Items) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "no items in batch request")
	}

	request := &services.BulkPermissionRequest{
		UserID:      req.Items[0].PrincipalId,
		Permissions: permissions,
	}

	bulkResult, err := h.authzService.CheckBulkPermissions(ctx, request)
	if err != nil {
		h.logger.Error("Bulk permission check failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "bulk permission check failed: %v", err)
	}

	// Convert result to protobuf format
	results := make([]*pb.CheckResult, 0, len(req.Items))
	for _, checkReq := range req.Items {
		// Find corresponding result by creating a key
		key := fmt.Sprintf("%s:%s:%s:%s", checkReq.PrincipalId, checkReq.ResourceType, checkReq.ResourceId, checkReq.Action)
		permResult, exists := bulkResult.Results[key]

		var allowed bool
		var decisionId string
		if exists {
			allowed = permResult.Allowed
			decisionId = permResult.DecisionID
		}

		result := &pb.CheckResult{
			RequestId:  checkReq.RequestId,
			Allowed:    allowed,
			DecisionId: decisionId,
		}
		results = append(results, result)
	}

	h.logger.Info("Batch permission check completed",
		zap.Int("results_count", len(results)))

	return &pb.BatchCheckResponse{
		Results: results,
	}, nil
}

// LookupResources implements the LookupResources RPC method for authorization
func (h *AuthorizationHandler) LookupResources(ctx context.Context, req *pb.LookupResourcesRequest) (*pb.LookupResourcesResponse, error) {
	h.logger.Info("gRPC LookupResources request",
		zap.String("principal_id", req.PrincipalId),
		zap.String("resource_type", req.ResourceType),
		zap.String("action", req.Action))

	resources, err := h.authzService.LookupResources(ctx, req.PrincipalId, req.ResourceType, req.Action)
	if err != nil {
		h.logger.Error("LookupResources failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "lookup resources failed: %v", err)
	}

	h.logger.Info("LookupResources completed",
		zap.String("principal_id", req.PrincipalId),
		zap.Int("resource_count", len(resources)))

	// Convert string slice to ResourceAccess slice
	resourceAccess := make([]*pb.ResourceAccess, 0, len(resources))
	for _, resourceID := range resources {
		resourceAccess = append(resourceAccess, &pb.ResourceAccess{
			ResourceId:   resourceID,
			ResourceType: req.ResourceType,
		})
	}

	return &pb.LookupResourcesResponse{
		Resources: resourceAccess,
	}, nil
}

// CheckColumns implements the CheckColumns RPC method for authorization
func (h *AuthorizationHandler) CheckColumns(ctx context.Context, req *pb.CheckColumnsRequest) (*pb.CheckColumnsResponse, error) {
	h.logger.Info("gRPC CheckColumns request",
		zap.String("principal_id", req.PrincipalId),
		zap.String("table_name", req.TableName),
		zap.String("action", req.Action))

	// Check if user has permission on the table resource
	permission := &services.Permission{
		UserID:     req.PrincipalId,
		Resource:   "table",
		ResourceID: req.TableName,
		Action:     req.Action,
	}

	result, err := h.authzService.CheckPermission(ctx, permission)
	if err != nil {
		h.logger.Error("CheckColumns failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "check columns failed: %v", err)
	}

	// For now, if the user has table permission, allow all requested columns
	var allowedColumns []string
	if result.Allowed {
		allowedColumns = req.RequestedColumns
	}

	h.logger.Info("CheckColumns completed",
		zap.String("principal_id", req.PrincipalId),
		zap.Bool("allowed", result.Allowed),
		zap.Int("allowed_columns_count", len(allowedColumns)))

	return &pb.CheckColumnsResponse{
		Allowed:        result.Allowed,
		AllowedColumns: allowedColumns,
	}, nil
}

// ListAllowedColumns implements the ListAllowedColumns RPC method for authorization
func (h *AuthorizationHandler) ListAllowedColumns(ctx context.Context, req *pb.ListAllowedColumnsRequest) (*pb.ListAllowedColumnsResponse, error) {
	h.logger.Info("gRPC ListAllowedColumns request",
		zap.String("principal_id", req.PrincipalId),
		zap.String("table_name", req.TableName),
		zap.String("action", req.Action))

	// Check if user has permission on the table resource
	permission := &services.Permission{
		UserID:     req.PrincipalId,
		Resource:   "table",
		ResourceID: req.TableName,
		Action:     req.Action,
	}

	result, err := h.authzService.CheckPermission(ctx, permission)
	if err != nil {
		h.logger.Error("ListAllowedColumns failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "list allowed columns failed: %v", err)
	}

	// For now, return a basic set of allowed columns if user has table permission
	var allowedColumns []string
	if result.Allowed {
		// This would typically come from a database schema or configuration
		allowedColumns = []string{"id", "name", "created_at", "updated_at"}
	}

	h.logger.Info("ListAllowedColumns completed",
		zap.String("principal_id", req.PrincipalId),
		zap.Int("allowed_columns_count", len(allowedColumns)))

	return &pb.ListAllowedColumnsResponse{
		AllowedColumns: allowedColumns,
	}, nil
}

// EvaluatePermission implements the EvaluatePermission RPC method for authorization
func (h *AuthorizationHandler) EvaluatePermission(ctx context.Context, req *pb.EvaluatePermissionRequest) (*pb.EvaluatePermissionResponse, error) {
	h.logger.Info("gRPC EvaluatePermission request",
		zap.String("principal_id", req.PrincipalId),
		zap.String("permission", req.Permission),
		zap.String("resource_context", req.ResourceContext))

	// Parse permission format (e.g., "resource:action" or "resource.action")
	resourceType := req.ResourceContext
	action := req.ActionContext

	if req.Permission != "" {
		// Try to split permission string by : or .
		if strings.Contains(req.Permission, ":") {
			parts := strings.SplitN(req.Permission, ":", 2)
			if len(parts) == 2 {
				resourceType = parts[0]
				action = parts[1]
			}
		} else if strings.Contains(req.Permission, ".") {
			parts := strings.SplitN(req.Permission, ".", 2)
			if len(parts) == 2 {
				resourceType = parts[0]
				action = parts[1]
			}
		}
	}

	// Check the permission
	permission := &services.Permission{
		UserID:     req.PrincipalId,
		Resource:   resourceType,
		ResourceID: "*", // Default to wildcard
		Action:     action,
	}

	result, err := h.authzService.CheckPermission(ctx, permission)
	if err != nil {
		h.logger.Error("EvaluatePermission failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "evaluate permission failed: %v", err)
	}

	// Create response with reasons
	var reasons []string
	if result.Allowed {
		reasons = append(reasons, fmt.Sprintf("User %s has permission %s on %s based on role-based permissions",
			req.PrincipalId, action, resourceType))
	} else {
		reasons = append(reasons, fmt.Sprintf("User %s lacks permission %s on %s - insufficient permissions",
			req.PrincipalId, action, resourceType))
	}

	h.logger.Info("EvaluatePermission completed",
		zap.String("principal_id", req.PrincipalId),
		zap.Bool("allowed", result.Allowed))

	return &pb.EvaluatePermissionResponse{
		Allowed:         result.Allowed,
		DecisionId:      result.DecisionID,
		Reasons:         reasons,
		ConfidenceScore: 100, // Full confidence in RBAC decisions
	}, nil
}

// BulkEvaluatePermissions implements the BulkEvaluatePermissions RPC method for authorization
func (h *AuthorizationHandler) BulkEvaluatePermissions(ctx context.Context, req *pb.BulkEvaluatePermissionsRequest) (*pb.BulkEvaluatePermissionsResponse, error) {
	h.logger.Info("gRPC BulkEvaluatePermissions request",
		zap.String("principal_id", req.PrincipalId),
		zap.Int("permission_count", len(req.Permissions)))

	results := make([]*pb.PermissionResult, 0, len(req.Permissions))
	var allowedCount, deniedCount int32

	for _, permCheck := range req.Permissions {
		// Parse permission
		resourceType := permCheck.ResourceContext
		action := permCheck.ActionContext

		if permCheck.Permission != "" {
			// Try to parse permission string by : or .
			if strings.Contains(permCheck.Permission, ":") {
				parts := strings.SplitN(permCheck.Permission, ":", 2)
				if len(parts) == 2 {
					resourceType = parts[0]
					action = parts[1]
				}
			} else if strings.Contains(permCheck.Permission, ".") {
				parts := strings.SplitN(permCheck.Permission, ".", 2)
				if len(parts) == 2 {
					resourceType = parts[0]
					action = parts[1]
				}
			}
		}

		permission := &services.Permission{
			UserID:     req.PrincipalId,
			Resource:   resourceType,
			ResourceID: "*",
			Action:     action,
		}

		result, err := h.authzService.CheckPermission(ctx, permission)
		if err != nil {
			h.logger.Error("Permission check failed in bulk evaluation", zap.Error(err))
			deniedCount++
			results = append(results, &pb.PermissionResult{
				Permission:      permCheck.Permission,
				Allowed:         false,
				Reasons:         []string{fmt.Sprintf("Error checking permission: %v", err)},
				ConfidenceScore: 0,
			})
			continue
		}

		if result.Allowed {
			allowedCount++
		} else {
			deniedCount++
		}

		var reasons []string
		if result.Allowed {
			reasons = append(reasons, "Permission granted based on role-based access")
		} else {
			reasons = append(reasons, "Permission denied - insufficient privileges")
		}

		results = append(results, &pb.PermissionResult{
			Permission:      permCheck.Permission,
			Allowed:         result.Allowed,
			DecisionId:      result.DecisionID,
			Reasons:         reasons,
			ConfidenceScore: 100,
		})
	}

	h.logger.Info("BulkEvaluatePermissions completed",
		zap.String("principal_id", req.PrincipalId),
		zap.Int32("allowed", allowedCount),
		zap.Int32("denied", deniedCount))

	return &pb.BulkEvaluatePermissionsResponse{
		Results:       results,
		TotalChecks:   int32(len(req.Permissions)),
		AllowedChecks: allowedCount,
		DeniedChecks:  deniedCount,
	}, nil
}
