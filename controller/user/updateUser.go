package user

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/model"
	pb "github.com/Kisanlink/aaa-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}
	var existingUser model.User
	err := s.DB.Table("users").Where("id = ?", id).First(&existingUser).Error
	if err != nil {
		if err.Error() == "record not found" {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to fetch user: %v", err))
	}
	if req.Username != "" {
		existingUser.Username = req.Username
	}
	if req.IsValidated != existingUser.IsValidated {
		existingUser.IsValidated = req.IsValidated
	}
	if err := s.DB.Table("users").Save(&existingUser).Error; err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to update user: %v", err))
	}

	rolePermissionIDs := req.UserRoleIds
	if err := s.updateUserRoles(id, rolePermissionIDs); err != nil {
		return nil, err
	}

	var userRoles []model.UserRole
	if err := s.DB.Table("user_roles").Where("user_id = ?", id).Find(&userRoles).Error; err != nil {
		return nil, status.Error(codes.Internal, "Failed to fetch updated roles")
	}

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
	if err := s.DB.Table("user_roles").Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
		return status.Error(codes.Internal, "Failed to delete existing UserRole entries")
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

	if err := s.DB.Table("user_roles").Create(&userRoles).Error; err != nil {
		return status.Error(codes.Internal, "Failed to create new UserRole entries")
	}

	return nil
}
