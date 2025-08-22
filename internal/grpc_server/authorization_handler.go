package grpc_server

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/services"
	"github.com/Kisanlink/aaa-service/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthorizationHandler implements authorization-related gRPC services
type AuthorizationHandler struct {
	proto.UnimplementedAuthorizationServiceServer
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
func (h *AuthorizationHandler) Check(ctx context.Context, req *proto.CheckRequest) (*proto.CheckResponse, error) {
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

	return &proto.CheckResponse{
		Allowed:          result.Allowed,
		DecisionId:       result.DecisionID,
		ConsistencyToken: result.ConsistencyToken,
	}, nil
}

// BatchCheck implements the BatchCheck RPC method for authorization
func (h *AuthorizationHandler) BatchCheck(ctx context.Context, req *proto.BatchCheckRequest) (*proto.BatchCheckResponse, error) {
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
	results := make([]*proto.CheckResult, 0, len(req.Items))
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

		result := &proto.CheckResult{
			RequestId:  checkReq.RequestId,
			Allowed:    allowed,
			DecisionId: decisionId,
		}
		results = append(results, result)
	}

	h.logger.Info("Batch permission check completed",
		zap.Int("results_count", len(results)))

	return &proto.BatchCheckResponse{
		Results: results,
	}, nil
}

// LookupResources implements the LookupResources RPC method for authorization
func (h *AuthorizationHandler) LookupResources(ctx context.Context, req *proto.LookupResourcesRequest) (*proto.LookupResourcesResponse, error) {
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
	resourceAccess := make([]*proto.ResourceAccess, 0, len(resources))
	for _, resourceID := range resources {
		resourceAccess = append(resourceAccess, &proto.ResourceAccess{
			ResourceId:   resourceID,
			ResourceType: req.ResourceType,
		})
	}

	return &proto.LookupResourcesResponse{
		Resources: resourceAccess,
	}, nil
}

// CheckColumns implements the CheckColumns RPC method for authorization
func (h *AuthorizationHandler) CheckColumns(ctx context.Context, req *proto.CheckColumnsRequest) (*proto.CheckColumnsResponse, error) {
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

	return &proto.CheckColumnsResponse{
		Allowed:        result.Allowed,
		AllowedColumns: allowedColumns,
	}, nil
}

// ListAllowedColumns implements the ListAllowedColumns RPC method for authorization
func (h *AuthorizationHandler) ListAllowedColumns(ctx context.Context, req *proto.ListAllowedColumnsRequest) (*proto.ListAllowedColumnsResponse, error) {
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

	return &proto.ListAllowedColumnsResponse{
		AllowedColumns: allowedColumns,
	}, nil
}

// Explain implements the Explain RPC method for authorization
func (h *AuthorizationHandler) Explain(ctx context.Context, req *proto.ExplainRequest) (*proto.ExplainResponse, error) {
	h.logger.Info("gRPC Explain request",
		zap.String("principal_id", req.PrincipalId),
		zap.String("resource_type", req.ResourceType),
		zap.String("action", req.Action))

	// Check the permission to explain
	permission := &services.Permission{
		UserID:     req.PrincipalId,
		Resource:   req.ResourceType,
		ResourceID: req.ResourceId,
		Action:     req.Action,
	}

	result, err := h.authzService.CheckPermission(ctx, permission)
	if err != nil {
		h.logger.Error("Explain failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "explain failed: %v", err)
	}

	// Create a basic explanation
	var explanation string
	if result.Allowed {
		explanation = fmt.Sprintf("User %s is allowed to perform %s on %s:%s based on role-based permissions",
			req.PrincipalId, req.Action, req.ResourceType, req.ResourceId)
	} else {
		explanation = fmt.Sprintf("User %s is denied access to perform %s on %s:%s - insufficient permissions",
			req.PrincipalId, req.Action, req.ResourceType, req.ResourceId)
	}

	h.logger.Info("Explain completed",
		zap.String("principal_id", req.PrincipalId),
		zap.Bool("allowed", result.Allowed))

	return &proto.ExplainResponse{
		Allowed:          result.Allowed,
		DecisionId:       result.DecisionID,
		HumanExplanation: explanation,
	}, nil
}
