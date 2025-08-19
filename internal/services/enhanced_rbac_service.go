package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// EnhancedRBACService implements the enhanced RBAC gRPC service
type EnhancedRBACService struct {
	pb.UnimplementedEnhancedRBACServiceServer
	db *gorm.DB
}

// NewEnhancedRBACService creates a new EnhancedRBACService instance
func NewEnhancedRBACService(db *gorm.DB) *EnhancedRBACService {
	return &EnhancedRBACService{
		db: db,
	}
}

// CreateAction creates a new action
func (s *EnhancedRBACService) CreateAction(ctx context.Context, req *pb.CreateActionRequest) (*pb.CreateActionResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "action name is required")
	}

	action := models.NewActionWithCategory(req.Name, req.Description, req.Category)

	if err := s.db.Create(action).Error; err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create action: %v", err))
	}

	return &pb.CreateActionResponse{
		StatusCode: int32(codes.OK),
		Message:    "Action created successfully",
		Action: &pb.Action{
			Id:          action.ID,
			Name:        action.Name,
			Description: action.Description,
			Category:    action.Category,
			IsActive:    action.IsActive,
			CreatedAt:   action.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   action.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// GetAction retrieves an action by ID
func (s *EnhancedRBACService) GetAction(ctx context.Context, req *pb.GetActionRequest) (*pb.Action, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "action ID is required")
	}

	var action models.Action
	if err := s.db.Where("id = ?", req.Id).First(&action).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "action not found")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get action: %v", err))
	}

	return &pb.Action{
		Id:          action.ID,
		Name:        action.Name,
		Description: action.Description,
		Category:    action.Category,
		IsActive:    action.IsActive,
		CreatedAt:   action.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   action.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GetAllActions streams all actions
func (s *EnhancedRBACService) GetAllActions(req *pb.GetAllActionsRequest, stream pb.EnhancedRBACService_GetAllActionsServer) error {
	query := s.db.Model(&models.Action{})

	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}

	var actions []models.Action
	if err := query.Find(&actions).Error; err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to get actions: %v", err))
	}

	for _, action := range actions {
		pbAction := &pb.Action{
			Id:          action.ID,
			Name:        action.Name,
			Description: action.Description,
			Category:    action.Category,
			IsActive:    action.IsActive,
			CreatedAt:   action.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   action.UpdatedAt.Format(time.RFC3339),
		}

		if err := stream.Send(pbAction); err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("failed to send action: %v", err))
		}
	}

	return nil
}

// UpdateAction updates an existing action
func (s *EnhancedRBACService) UpdateAction(ctx context.Context, req *pb.UpdateActionRequest) (*pb.Action, error) {
	if req.Action == nil || req.Action.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "action and action ID are required")
	}

	var action models.Action
	if err := s.db.Where("id = ?", req.Action.Id).First(&action).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "action not found")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get action: %v", err))
	}

	// Update fields
	if req.Action.Name != "" {
		action.Name = req.Action.Name
	}
	if req.Action.Description != "" {
		action.Description = req.Action.Description
	}
	if req.Action.Category != "" {
		action.Category = req.Action.Category
	}
	action.IsActive = req.Action.IsActive

	if err := s.db.Save(&action).Error; err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update action: %v", err))
	}

	return &pb.Action{
		Id:          action.ID,
		Name:        action.Name,
		Description: action.Description,
		Category:    action.Category,
		IsActive:    action.IsActive,
		CreatedAt:   action.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   action.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// DeleteAction deletes an action
func (s *EnhancedRBACService) DeleteAction(ctx context.Context, req *pb.DeleteActionRequest) (*pb.DeleteActionResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "action ID is required")
	}

	if err := s.db.Where("id = ?", req.Id).Delete(&models.Action{}).Error; err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete action: %v", err))
	}

	return &pb.DeleteActionResponse{
		StatusCode: int32(codes.OK),
		Message:    "Action deleted successfully",
		Success:    true,
	}, nil
}

// CreateResource creates a new resource
func (s *EnhancedRBACService) CreateResource(ctx context.Context, req *pb.CreateResourceRequest) (*pb.CreateResourceResponse, error) {
	if req.Name == "" || req.Type == "" {
		return nil, status.Error(codes.InvalidArgument, "resource name and type are required")
	}

	resource := models.NewResource(req.Name, req.Type, req.Description)

	if req.ParentId != "" {
		resource.ParentID = &req.ParentId
	}
	if req.OwnerId != "" {
		resource.OwnerID = &req.OwnerId
	}

	if err := s.db.Create(resource).Error; err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create resource: %v", err))
	}

	return &pb.CreateResourceResponse{
		StatusCode: int32(codes.OK),
		Message:    "Resource created successfully",
		Resource: &pb.Resource{
			Id:          resource.ID,
			Name:        resource.Name,
			Type:        resource.Type,
			Description: resource.Description,
			ParentId:    getStringValue(resource.ParentID),
			OwnerId:     getStringValue(resource.OwnerID),
			IsActive:    resource.IsActive,
			CreatedAt:   resource.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   resource.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// GetResource retrieves a resource by ID
func (s *EnhancedRBACService) GetResource(ctx context.Context, req *pb.GetResourceRequest) (*pb.Resource, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "resource ID is required")
	}

	var resource models.Resource
	if err := s.db.Where("id = ?", req.Id).First(&resource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get resource: %v", err))
	}

	return &pb.Resource{
		Id:          resource.ID,
		Name:        resource.Name,
		Type:        resource.Type,
		Description: resource.Description,
		ParentId:    getStringValue(resource.ParentID),
		OwnerId:     getStringValue(resource.OwnerID),
		IsActive:    resource.IsActive,
		CreatedAt:   resource.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   resource.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GetAllResources streams all resources
func (s *EnhancedRBACService) GetAllResources(req *pb.GetAllResourcesRequest, stream pb.EnhancedRBACService_GetAllResourcesServer) error {
	query := s.db.Model(&models.Resource{})

	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}

	var resources []models.Resource
	if err := query.Find(&resources).Error; err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to get resources: %v", err))
	}

	for _, resource := range resources {
		pbResource := &pb.Resource{
			Id:          resource.ID,
			Name:        resource.Name,
			Type:        resource.Type,
			Description: resource.Description,
			ParentId:    getStringValue(resource.ParentID),
			OwnerId:     getStringValue(resource.OwnerID),
			IsActive:    resource.IsActive,
			CreatedAt:   resource.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   resource.UpdatedAt.Format(time.RFC3339),
		}

		if err := stream.Send(pbResource); err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("failed to send resource: %v", err))
		}
	}

	return nil
}

// UpdateResource updates an existing resource
func (s *EnhancedRBACService) UpdateResource(ctx context.Context, req *pb.UpdateResourceRequest) (*pb.Resource, error) {
	if req.Resource == nil || req.Resource.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "resource and resource ID are required")
	}

	var resource models.Resource
	if err := s.db.Where("id = ?", req.Resource.Id).First(&resource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get resource: %v", err))
	}

	// Update fields
	if req.Resource.Name != "" {
		resource.Name = req.Resource.Name
	}
	if req.Resource.Type != "" {
		resource.Type = req.Resource.Type
	}
	if req.Resource.Description != "" {
		resource.Description = req.Resource.Description
	}
	if req.Resource.ParentId != "" {
		resource.ParentID = &req.Resource.ParentId
	}
	if req.Resource.OwnerId != "" {
		resource.OwnerID = &req.Resource.OwnerId
	}
	resource.IsActive = req.Resource.IsActive

	if err := s.db.Save(&resource).Error; err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update resource: %v", err))
	}

	return &pb.Resource{
		Id:          resource.ID,
		Name:        resource.Name,
		Type:        resource.Type,
		Description: resource.Description,
		ParentId:    getStringValue(resource.ParentID),
		OwnerId:     getStringValue(resource.OwnerID),
		IsActive:    resource.IsActive,
		CreatedAt:   resource.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   resource.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// DeleteResource deletes a resource
func (s *EnhancedRBACService) DeleteResource(ctx context.Context, req *pb.DeleteResourceRequest) (*pb.DeleteResourceResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "resource ID is required")
	}

	if err := s.db.Where("id = ?", req.Id).Delete(&models.Resource{}).Error; err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete resource: %v", err))
	}

	return &pb.DeleteResourceResponse{
		StatusCode: int32(codes.OK),
		Message:    "Resource deleted successfully",
		Success:    true,
	}, nil
}

// Helper function to safely get string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
