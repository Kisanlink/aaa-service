package user

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) AssignRole(ctx context.Context, req *pb.AssignRoleToUserRequest) (*pb.AssignRoleToUserResponse, error) {
	// Validate user exists
	_, err := s.userService.GetUserByID(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "User with ID %s not found", req.UserId)
	}

	// Validate role exists
	roleName := req.GetRole()
	role, err := s.roleService.GetRoleByName(roleName)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Role with name %s not found", roleName)
	}

	// Create user-role relationship
	userRole := model.UserRole{
		UserID:   req.UserId,
		RoleID:   role.ID,
		IsActive: true,
	}

	if err := s.userService.CreateUserRoles(userRole); err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.AlreadyExists {
			return nil, status.Errorf(codes.AlreadyExists, "user already has role '%s' assigned", roleName)
		}
		return nil, status.Errorf(codes.Internal, "failed to assign role to user: %v", err)
	}
	// Get updated user details
	updatedUser, err := s.userService.GetUserByID(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch user details: %v", err)
	}

	// Get roles and permissions for relationship updates
	roles, permissions, actions, err := s.userService.FindUserRolesAndPermissions(updatedUser.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user roles and permissions: %v", err)
	}

	// Update relationships in external service
	_, err = client.DeleteUserRoleRelationship(
		updatedUser.Username,
		helper.LowerCaseSlice(roles),
		helper.LowerCaseSlice(permissions),
		helper.LowerCaseSlice(actions),
	)
	if err != nil {
		log.Printf("Failed to delete relationships: %v", err)
	}

	_, err = client.CreateUserRoleRelationship(
		updatedUser.Username,
		helper.LowerCaseSlice(roles),
		helper.LowerCaseSlice(permissions),
		helper.LowerCaseSlice(actions),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create relation: %v", err)
	}

	// Get role permissions in the correct format
	rolePermissions, err := s.userService.FindUsageRights(updatedUser.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user roles and permissions: %v", err)
	}

	// Convert role permissions to protobuf format and remove duplicates
	var pbRolePermissions []*pb.RolePermissions
	for role, permissions := range rolePermissions {
		// Use a map to track unique permissions
		uniquePerms := make(map[string]*pb.PermissionResponse)
		for _, perm := range permissions {
			key := perm.Name + ":" + perm.Action + ":" + perm.Resource
			if _, exists := uniquePerms[key]; !exists {
				uniquePerms[key] = &pb.PermissionResponse{
					Name:        perm.Name,
					Description: perm.Description,
					Action:      perm.Action,
					Source:      perm.Source,
					Resource:    perm.Resource,
				}
			}
		}

		// Convert unique permissions map to slice
		var pbPermissions []*pb.PermissionResponse
		for _, perm := range uniquePerms {
			pbPermissions = append(pbPermissions, perm)
		}

		pbRolePermissions = append(pbRolePermissions, &pb.RolePermissions{
			RoleName:    role,
			Permissions: pbPermissions,
		})
	}

	// Format timestamps
	createdAt := ""
	if !updatedUser.CreatedAt.IsZero() {
		createdAt = updatedUser.CreatedAt.Format(time.RFC3339Nano)
	}

	updatedAt := ""
	if !updatedUser.UpdatedAt.IsZero() {
		updatedAt = updatedUser.UpdatedAt.Format(time.RFC3339Nano)
	}

	// Build response
	return &pb.AssignRoleToUserResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "Role assigned to user successfully",
		DataTimeStamp: time.Now().Format(time.RFC3339),
		Data: &pb.AssignRolePermission{
			Id:              updatedUser.ID,
			Username:        updatedUser.Username,
			Password:        "", // Explicitly empty for security
			IsValidated:     updatedUser.IsValidated,
			CreatedAt:       createdAt,
			UpdatedAt:       updatedAt,
			RolePermissions: pbRolePermissions,
		},
	}, nil
}
