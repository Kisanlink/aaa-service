package grpc_server

import (
	"context"

	groupRequests "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/groups"
	groupResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/groups"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GroupHandler implements the GroupService gRPC service
type GroupHandler struct {
	pb.UnimplementedGroupServiceServer
	groupService interfaces.GroupService
	logger       *zap.Logger
}

// NewGroupHandler creates a new GroupHandler instance
func NewGroupHandler(groupService interfaces.GroupService, logger *zap.Logger) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
		logger:       logger,
	}
}

// CreateGroup creates a new group
func (h *GroupHandler) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupResponse, error) {
	h.logger.Info("gRPC CreateGroup request",
		zap.String("name", req.Name),
		zap.String("org_id", req.OrganizationId))

	// Validate request
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization_id is required")
	}

	// Build service request
	serviceReq := &groupRequests.CreateGroupRequest{
		Name:           req.Name,
		Description:    req.Description,
		OrganizationID: req.OrganizationId,
	}
	if req.ParentId != "" {
		serviceReq.ParentID = &req.ParentId
	}

	// Call service
	result, err := h.groupService.CreateGroup(ctx, serviceReq)
	if err != nil {
		h.logger.Error("Failed to create group", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert response
	groupResp, ok := result.(*groupResponses.GroupResponse)
	if !ok {
		return nil, status.Error(codes.Internal, "invalid response from group service")
	}

	return &pb.CreateGroupResponse{
		Group: &pb.Group{
			Id:             groupResp.ID,
			Name:           groupResp.Name,
			Description:    groupResp.Description,
			OrganizationId: groupResp.OrganizationID,
			IsActive:       groupResp.IsActive,
			CreatedAt:      timestamppb.New(*groupResp.CreatedAt),
			UpdatedAt:      timestamppb.New(*groupResp.UpdatedAt),
		},
	}, nil
}

// GetGroup retrieves a group by ID
func (h *GroupHandler) GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupResponse, error) {
	h.logger.Info("gRPC GetGroup request", zap.String("group_id", req.Id))

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "group ID is required")
	}

	result, err := h.groupService.GetGroup(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to get group", zap.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	groupResp, ok := result.(*groupResponses.GroupResponse)
	if !ok {
		return nil, status.Error(codes.Internal, "invalid response from group service")
	}

	return &pb.GetGroupResponse{
		Group: &pb.Group{
			Id:             groupResp.ID,
			Name:           groupResp.Name,
			Description:    groupResp.Description,
			OrganizationId: groupResp.OrganizationID,
			IsActive:       groupResp.IsActive,
			CreatedAt:      timestamppb.New(*groupResp.CreatedAt),
			UpdatedAt:      timestamppb.New(*groupResp.UpdatedAt),
		},
	}, nil
}

// ListGroups retrieves groups with pagination
func (h *GroupHandler) ListGroups(ctx context.Context, req *pb.ListGroupsRequest) (*pb.ListGroupsResponse, error) {
	h.logger.Info("gRPC ListGroups request",
		zap.String("org_id", req.OrganizationId),
		zap.Int32("page_size", req.PageSize))

	limit := int(req.PageSize)
	if limit <= 0 {
		limit = 10
	}
	offset := int(req.Page) * limit

	result, err := h.groupService.ListGroups(ctx, limit, offset, req.OrganizationId, req.IncludeInactive)
	if err != nil {
		h.logger.Error("Failed to list groups", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	groups, ok := result.([]*groupResponses.GroupResponse)
	if !ok {
		return nil, status.Error(codes.Internal, "invalid response from group service")
	}

	pbGroups := make([]*pb.Group, len(groups))
	for i, g := range groups {
		pbGroups[i] = &pb.Group{
			Id:             g.ID,
			Name:           g.Name,
			Description:    g.Description,
			OrganizationId: g.OrganizationID,
			IsActive:       g.IsActive,
		}
		if g.CreatedAt != nil {
			pbGroups[i].CreatedAt = timestamppb.New(*g.CreatedAt)
		}
		if g.UpdatedAt != nil {
			pbGroups[i].UpdatedAt = timestamppb.New(*g.UpdatedAt)
		}
	}

	return &pb.ListGroupsResponse{
		Groups: pbGroups,
	}, nil
}

// AddGroupMember adds a member to a group
func (h *GroupHandler) AddGroupMember(ctx context.Context, req *pb.AddGroupMemberRequest) (*pb.AddGroupMemberResponse, error) {
	h.logger.Info("gRPC AddGroupMember request",
		zap.String("group_id", req.GroupId),
		zap.String("principal_id", req.PrincipalId),
		zap.String("principal_type", req.PrincipalType))

	// Validate request
	if req.GroupId == "" {
		return &pb.AddGroupMemberResponse{
			StatusCode: 400,
			Message:    "group_id is required",
		}, status.Error(codes.InvalidArgument, "group_id is required")
	}
	if req.PrincipalId == "" {
		return &pb.AddGroupMemberResponse{
			StatusCode: 400,
			Message:    "principal_id is required",
		}, status.Error(codes.InvalidArgument, "principal_id is required")
	}

	// Default principal type to "user"
	principalType := req.PrincipalType
	if principalType == "" {
		principalType = "user"
	}

	// Extract user ID from context if available (set by auth middleware)
	addedByID := "system"
	if userIDValue := ctx.Value("user_id"); userIDValue != nil {
		if uid, ok := userIDValue.(string); ok && uid != "" {
			addedByID = uid
		}
	}

	// Build service request
	serviceReq := &groupRequests.AddMemberRequest{
		GroupID:       req.GroupId,
		PrincipalID:   req.PrincipalId,
		PrincipalType: principalType,
		AddedByID:     addedByID,
	}

	// Handle optional time bounds
	if req.StartsAt != nil {
		t := req.StartsAt.AsTime()
		serviceReq.StartsAt = &t
	}
	if req.EndsAt != nil {
		t := req.EndsAt.AsTime()
		serviceReq.EndsAt = &t
	}

	// Call service
	result, err := h.groupService.AddMemberToGroup(ctx, serviceReq)
	if err != nil {
		h.logger.Error("Failed to add member to group", zap.Error(err))
		return &pb.AddGroupMemberResponse{
			StatusCode: 500,
			Message:    err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	memberResp, ok := result.(*groupResponses.GroupMembershipResponse)
	if !ok {
		return &pb.AddGroupMemberResponse{
			StatusCode: 500,
			Message:    "invalid response from group service",
		}, status.Error(codes.Internal, "invalid response from group service")
	}

	h.logger.Info("Member added to group successfully",
		zap.String("group_id", req.GroupId),
		zap.String("principal_id", req.PrincipalId))

	return &pb.AddGroupMemberResponse{
		StatusCode: 200,
		Message:    "Member added to group successfully",
		Membership: &pb.GroupMembership{
			Id:            memberResp.ID,
			GroupId:       memberResp.GroupID,
			PrincipalId:   memberResp.PrincipalID,
			PrincipalType: memberResp.PrincipalType,
			IsActive:      memberResp.IsActive,
		},
	}, nil
}

// RemoveGroupMember removes a member from a group
func (h *GroupHandler) RemoveGroupMember(ctx context.Context, req *pb.RemoveGroupMemberRequest) (*pb.RemoveGroupMemberResponse, error) {
	h.logger.Info("gRPC RemoveGroupMember request",
		zap.String("group_id", req.GroupId),
		zap.String("principal_id", req.PrincipalId))

	if req.GroupId == "" {
		return &pb.RemoveGroupMemberResponse{
			StatusCode: 400,
			Message:    "group_id is required",
		}, status.Error(codes.InvalidArgument, "group_id is required")
	}
	if req.PrincipalId == "" {
		return &pb.RemoveGroupMemberResponse{
			StatusCode: 400,
			Message:    "principal_id is required",
		}, status.Error(codes.InvalidArgument, "principal_id is required")
	}

	// Use "system" as the removedBy for gRPC calls
	err := h.groupService.RemoveMemberFromGroup(ctx, req.GroupId, req.PrincipalId, "system")
	if err != nil {
		h.logger.Error("Failed to remove member from group", zap.Error(err))
		return &pb.RemoveGroupMemberResponse{
			StatusCode: 500,
			Message:    err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	h.logger.Info("Member removed from group successfully",
		zap.String("group_id", req.GroupId),
		zap.String("principal_id", req.PrincipalId))

	return &pb.RemoveGroupMemberResponse{
		StatusCode: 200,
		Message:    "Member removed from group successfully",
	}, nil
}

// ListGroupMembers retrieves members of a group
func (h *GroupHandler) ListGroupMembers(ctx context.Context, req *pb.ListGroupMembersRequest) (*pb.ListGroupMembersResponse, error) {
	h.logger.Info("gRPC ListGroupMembers request", zap.String("group_id", req.GroupId))

	if req.GroupId == "" {
		return nil, status.Error(codes.InvalidArgument, "group_id is required")
	}

	limit := int(req.PageSize)
	if limit <= 0 {
		limit = 10
	}
	offset := int(req.Page) * limit

	result, err := h.groupService.GetGroupMembers(ctx, req.GroupId, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list group members", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	members, ok := result.([]*groupResponses.GroupMembershipResponse)
	if !ok {
		return nil, status.Error(codes.Internal, "invalid response from group service")
	}

	pbMembers := make([]*pb.GroupMembership, len(members))
	for i, m := range members {
		pbMembers[i] = &pb.GroupMembership{
			Id:            m.ID,
			GroupId:       m.GroupID,
			PrincipalId:   m.PrincipalID,
			PrincipalType: m.PrincipalType,
			IsActive:      m.IsActive,
		}
	}

	return &pb.ListGroupMembersResponse{
		Memberships: pbMembers,
	}, nil
}

// UpdateGroup updates a group
func (h *GroupHandler) UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.UpdateGroupResponse, error) {
	h.logger.Info("gRPC UpdateGroup request", zap.String("group_id", req.Id))

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "group ID is required")
	}

	serviceReq := &groupRequests.UpdateGroupRequest{}
	if req.Name != "" {
		serviceReq.Name = &req.Name
	}
	if req.Description != "" {
		serviceReq.Description = &req.Description
	}

	result, err := h.groupService.UpdateGroup(ctx, req.Id, serviceReq)
	if err != nil {
		h.logger.Error("Failed to update group", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	groupResp, ok := result.(*groupResponses.GroupResponse)
	if !ok {
		return nil, status.Error(codes.Internal, "invalid response from group service")
	}

	return &pb.UpdateGroupResponse{
		Group: &pb.Group{
			Id:             groupResp.ID,
			Name:           groupResp.Name,
			Description:    groupResp.Description,
			OrganizationId: groupResp.OrganizationID,
			IsActive:       groupResp.IsActive,
		},
	}, nil
}

// DeleteGroup deletes a group
func (h *GroupHandler) DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.DeleteGroupResponse, error) {
	h.logger.Info("gRPC DeleteGroup request", zap.String("group_id", req.Id))

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "group ID is required")
	}

	err := h.groupService.DeleteGroup(ctx, req.Id, "system")
	if err != nil {
		h.logger.Error("Failed to delete group", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteGroupResponse{
		StatusCode: 200,
		Message:    "Group deleted successfully",
	}, nil
}

// LinkGroups links two groups (parent-child relationship)
func (h *GroupHandler) LinkGroups(ctx context.Context, req *pb.LinkGroupsRequest) (*pb.LinkGroupsResponse, error) {
	h.logger.Info("gRPC LinkGroups request",
		zap.String("parent_group_id", req.ParentGroupId),
		zap.String("child_group_id", req.ChildGroupId))

	// This would need to update the child group's parent_id
	// For now, return unimplemented
	return nil, status.Error(codes.Unimplemented, "LinkGroups not yet implemented")
}

// UnlinkGroups unlinks two groups
func (h *GroupHandler) UnlinkGroups(ctx context.Context, req *pb.UnlinkGroupsRequest) (*pb.UnlinkGroupsResponse, error) {
	h.logger.Info("gRPC UnlinkGroups request",
		zap.String("parent_group_id", req.ParentGroupId),
		zap.String("child_group_id", req.ChildGroupId))

	// This would need to remove the child group's parent_id
	// For now, return unimplemented
	return nil, status.Error(codes.Unimplemented, "UnlinkGroups not yet implemented")
}

// AssignRoleToGroup assigns a role to a group
func (h *GroupHandler) AssignRoleToGroup(ctx context.Context, req *pb.AssignRoleToGroupRequest) (*pb.AssignRoleToGroupResponse, error) {
	h.logger.Info("gRPC AssignRoleToGroup request",
		zap.String("group_id", req.GroupId),
		zap.String("role_id", req.RoleId))

	if req.GroupId == "" {
		return &pb.AssignRoleToGroupResponse{
			StatusCode: 400,
			Message:    "group_id is required",
		}, status.Error(codes.InvalidArgument, "group_id is required")
	}
	if req.RoleId == "" {
		return &pb.AssignRoleToGroupResponse{
			StatusCode: 400,
			Message:    "role_id is required",
		}, status.Error(codes.InvalidArgument, "role_id is required")
	}

	// Call service to assign role to group
	result, err := h.groupService.AssignRoleToGroup(ctx, req.GroupId, req.RoleId, "system")
	if err != nil {
		h.logger.Error("Failed to assign role to group", zap.Error(err))
		return &pb.AssignRoleToGroupResponse{
			StatusCode: 500,
			Message:    err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	groupRoleResp, ok := result.(*groupResponses.GroupRoleResponse)
	if !ok {
		return &pb.AssignRoleToGroupResponse{
			StatusCode: 500,
			Message:    "invalid response from group service",
		}, status.Error(codes.Internal, "invalid response from group service")
	}

	h.logger.Info("Role assigned to group successfully",
		zap.String("group_id", req.GroupId),
		zap.String("role_id", req.RoleId))

	return &pb.AssignRoleToGroupResponse{
		StatusCode: 200,
		Message:    "Role assigned to group successfully",
		GroupRole: &pb.GroupRole{
			GroupId:    groupRoleResp.GroupID,
			RoleId:     groupRoleResp.Role.ID,
			RoleName:   groupRoleResp.Role.Name,
			AssignedBy: groupRoleResp.AssignedBy,
			IsActive:   groupRoleResp.IsActive,
		},
	}, nil
}

// RemoveRoleFromGroup removes a role from a group
func (h *GroupHandler) RemoveRoleFromGroup(ctx context.Context, req *pb.RemoveRoleFromGroupRequest) (*pb.RemoveRoleFromGroupResponse, error) {
	h.logger.Info("gRPC RemoveRoleFromGroup request",
		zap.String("group_id", req.GroupId),
		zap.String("role_id", req.RoleId))

	if req.GroupId == "" {
		return &pb.RemoveRoleFromGroupResponse{
			StatusCode: 400,
			Message:    "group_id is required",
		}, status.Error(codes.InvalidArgument, "group_id is required")
	}
	if req.RoleId == "" {
		return &pb.RemoveRoleFromGroupResponse{
			StatusCode: 400,
			Message:    "role_id is required",
		}, status.Error(codes.InvalidArgument, "role_id is required")
	}

	// Call service to remove role from group
	err := h.groupService.RemoveRoleFromGroup(ctx, req.GroupId, req.RoleId)
	if err != nil {
		h.logger.Error("Failed to remove role from group", zap.Error(err))
		return &pb.RemoveRoleFromGroupResponse{
			StatusCode: 500,
			Message:    err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	h.logger.Info("Role removed from group successfully",
		zap.String("group_id", req.GroupId),
		zap.String("role_id", req.RoleId))

	return &pb.RemoveRoleFromGroupResponse{
		StatusCode: 200,
		Message:    "Role removed from group successfully",
	}, nil
}

// GetGroupRoles retrieves all roles assigned to a group
func (h *GroupHandler) GetGroupRoles(ctx context.Context, req *pb.GetGroupRolesRequest) (*pb.GetGroupRolesResponse, error) {
	h.logger.Info("gRPC GetGroupRoles request", zap.String("group_id", req.GroupId))

	if req.GroupId == "" {
		return &pb.GetGroupRolesResponse{
			StatusCode: 400,
			Message:    "group_id is required",
		}, status.Error(codes.InvalidArgument, "group_id is required")
	}

	// Call service to get group roles
	result, err := h.groupService.GetGroupRoles(ctx, req.GroupId)
	if err != nil {
		h.logger.Error("Failed to get group roles", zap.Error(err))
		return &pb.GetGroupRolesResponse{
			StatusCode: 500,
			Message:    err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	roles, ok := result.([]*groupResponses.GroupRoleDetail)
	if !ok {
		return &pb.GetGroupRolesResponse{
			StatusCode: 500,
			Message:    "invalid response from group service",
		}, status.Error(codes.Internal, "invalid response from group service")
	}

	pbRoles := make([]*pb.GroupRole, len(roles))
	for i, r := range roles {
		pbRoles[i] = &pb.GroupRole{
			Id:         r.ID,
			GroupId:    r.GroupID,
			RoleId:     r.RoleID,
			RoleName:   r.Role.Name,
			AssignedBy: r.AssignedBy,
			IsActive:   r.IsActive,
		}
		if r.StartsAt != nil {
			pbRoles[i].CreatedAt = timestamppb.New(*r.StartsAt)
		}
		if r.EndsAt != nil {
			pbRoles[i].UpdatedAt = timestamppb.New(*r.EndsAt)
		}
	}

	return &pb.GetGroupRolesResponse{
		StatusCode: 200,
		Message:    "Group roles retrieved successfully",
		Roles:      pbRoles,
	}, nil
}
