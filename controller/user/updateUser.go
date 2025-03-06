package user

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user ID not found in context")
	}
	log.Printf("User %s is updating user with ID %s", userID, req.GetId())
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}
	existingUser, err := s.UserRepo.FindExistingUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Username != "" {
		existingUser.Username = req.Username
	}
	if req.IsValidated != existingUser.IsValidated {
		existingUser.IsValidated = req.IsValidated
	}
	if err := s.UserRepo.UpdateUser(ctx, *existingUser); err != nil {
		return nil, err
	}
	if err := s.updateUserRoles(existingUser.ID, req.UserRoleIds); err != nil {
		return nil, err
	}
	userRoles, err := s.UserRepo.FindUserRoles(ctx, existingUser.ID)
	if err != nil {
		return nil, err
	}
	roles, permissions, err := s.UserRepo.FindUserRolesAndPermissions(ctx, existingUser.ID)
    if err != nil {
        log.Fatalf("Failed to fetch user roles and permissions: %v", err)
    }
	updated, err := client.CreateUserRoleRelationship(
		strings.ToLower(existingUser.Username), 
		LowerCaseSlice(roles), 
		LowerCaseSlice(permissions),
	)
		if err != nil {
		log.Fatalf("Error reading schema: %v", err)
	}
	log.Printf("create Relation  Response: %+v", updated)
	pbRoles := ConvertToPBUserRoles(userRoles)
	pbUser := &pb.User{
		Id:          existingUser.ID,
		CreatedAt:   existingUser.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   existingUser.UpdatedAt.Format(time.RFC3339Nano),
		Username:    existingUser.Username,
		IsValidated: existingUser.IsValidated,
		UserRoles:   pbRoles,
	}
	return &pb.UpdateUserResponse{
		StatusCode: int32(codes.OK),
		Message:    "User updated successfully",
		User:       pbUser,
	}, nil
}

func (s *Server) updateUserRoles(userID string, rolePermissionIDs []string) error {
	if err := s.UserRepo.DeleteUserRoles(context.Background(), userID); err != nil {
		return err
	}
	if len(rolePermissionIDs) == 0 {
		return nil
	}
	var userRoles []model.UserRole
	for _, rolePermissionID := range rolePermissionIDs {
		userRole := model.UserRole{
			UserID:           userID,
			RolePermissionID: rolePermissionID,
		}
		userRoles = append(userRoles, userRole)
	}

	if err := s.UserRepo.CreateUserRoles(context.Background(), userRoles); err != nil {
		return err
	}
	return nil
}
