package user

import (
	"context"
	"errors"
	"time"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	pb "github.com/Kisanlink/aaa-service/proto"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	existingUser := model.User{}
	err := s.DB.Table("users").Where("username = ?", req.Username).First(&existingUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Error(codes.Internal, "Database error: "+err.Error())
	}
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid password")
	}
	accessToken, err := helper.GenerateAccessToken(existingUser.ID, existingUser.Roles, existingUser.Username, existingUser.IsValidated)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to generate access token")
	}
	refreshToken, err := helper.GenerateRefreshToken(existingUser.ID, existingUser.Roles, existingUser.Username, existingUser.IsValidated)
	if err != nil {

		return nil, status.Error(codes.Internal, "Failed to generate refresh token")
	}
	var userRoles []model.UserRole
	if err := s.DB.Table("user_roles").Where("user_id = ?", existingUser.ID).Find(&userRoles).Error; err != nil {
		return nil, status.Error(codes.Internal, "Failed to fetch updated roles")
	}

	pbUserRoles := make([]*pb.UserRole, len(userRoles))
	for i, role := range userRoles {
		pbUserRoles[i] = &pb.UserRole{
			Id:               role.ID,
			UserId:           role.UserID,
			RolePermissionId: role.RolePermissionID,
			CreatedAt:        role.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:        role.UpdatedAt.Format(time.RFC3339Nano),
		}
	}
	return &pb.LoginResponse{
		StatusCode:   200,
		Message:      "Login successful",
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: &pb.User{
			Id:          existingUser.ID,
			CreatedAt:   existingUser.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:   existingUser.UpdatedAt.Format(time.RFC3339Nano),
			Username:    existingUser.Username,
			IsValidated: existingUser.IsValidated,
			UserRoles:   pbUserRoles,
		},
	}, nil
}
