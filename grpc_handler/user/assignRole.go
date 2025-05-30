package user

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) AssignRole(ctx context.Context, req *pb.AssignRoleToUserRequest) (*pb.AssignRoleToUserResponse, error) {
	// Validate user exists
	user, err := s.userService.GetUserByID(req.UserId)
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
	userRoles, err := s.userService.FindUserRoles(user.ID)
	if err != nil {

		return nil, status.Errorf(codes.Internal, "failed to fetch user role: %v", err)
	}
	roleNames := make([]string, 0, len(userRoles))

	for _, userRole := range userRoles {
		role, err := s.roleService.FindRoleByID(userRole.RoleID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to find role with ID %s: %w", userRole.RoleID, err)
		}
		roleNames = append(roleNames, role.Name)
	}

	err = client.DeleteRelationships(
		roleNames,
		user.Username,
		user.ID,
	)

	if err != nil {
		log.Printf("Error deleting relationships: %v", err)
	}

	err = client.CreateRelationships(
		roleNames,
		user.Username,
		user.ID,
	)

	if err != nil {
		log.Printf("Error creating relationships: %v", err)
	}
	// Get updated user details with roles and permissions
	rolesResponse, err := s.userService.GetUserRolesWithPermissions(user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch user roles: %v", err)
	}

	// Build response
	return &pb.AssignRoleToUserResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "Role assigned to user successfully",
		DataTimeStamp: time.Now().Format(time.RFC3339),
		Data: &pb.AssignRolePermission{
			Id:          user.ID,
			Username:    user.Username,
			IsValidated: user.IsValidated,
			CreatedAt:   user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
			Roles:       convertRoleResponseToPB(rolesResponse),
		},
	}, nil
}

// convertRoleResponseToPB converts RoleResponse to protobuf Role format
func convertRoleResponseToPB(rolesResponse *model.RoleResponse) []*pb.Role {
	var pbRoles []*pb.Role
	for _, roleDetail := range rolesResponse.Roles {
		pbRoles = append(pbRoles, convertRoleDetailToPB(roleDetail))
	}
	return pbRoles
}

// convertRoleDetailToPB converts a single RoleDetail to protobuf Role format
func convertRoleDetailToPB(roleDetail model.RoleDetail) *pb.Role {
	return &pb.Role{
		RoleName:    roleDetail.RoleName,
		Permissions: convertPermissionsToPB(roleDetail.Permissions),
	}
}

// convertPermissionsToPB converts RolePermissions to protobuf Permissions
func convertPermissionsToPB(permissions []model.RolePermissionRes) []*pb.Permission {
	var pbPermissions []*pb.Permission
	for _, perm := range permissions {
		pbPermissions = append(pbPermissions, &pb.Permission{
			Resource: perm.Resource,
			Actions:  perm.Actions,
		})
	}
	return pbPermissions
}
