package user

import (
	"context"
	"time"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/kisanlink/protobuf/pb-aaa"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	existingUser, err := s.UserRepo.FindUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid password")
	}
	accessToken, err := helper.GenerateAccessToken(existingUser.ID, existingUser.Username, existingUser.IsValidated)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to generate access token")
	}
	refreshToken, err := helper.GenerateRefreshToken(existingUser.ID, existingUser.Username, existingUser.IsValidated)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to generate refresh token")
	}
	userRoles, err := s.UserRepo.FindUserRoles(ctx, existingUser.ID)
	if err != nil {
		return nil, err
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
