package user

import (
	"context"
	"net/http"
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
	if req.Password == "" {
		return nil, status.Error(codes.NotFound,"Password is required")
	}
	if req.Username == "" {
		return nil, status.Error(codes.NotFound,"username is required")
	}
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid password")
	}

	roles, permissions, actions, err := s.UserRepo.FindUserRolesAndPermissions(ctx, existingUser.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user roles and permissions: %v", err)
	}
	userRole := &pb.UserRoleResponse{
		Roles:       roles,
		Permissions: permissions,
		Actions:     actions,
	}
	if err := helper.SetAuthHeadersWithTokens(
		ctx,
		existingUser.ID,
		existingUser.Username,
		existingUser.IsValidated,
	); err != nil {
		return nil, err
	}
	return &pb.LoginResponse{
		StatusCode: http.StatusOK,
		Success: true,
		Message:      "Login successful",
		User: &pb.User{
			Id:          existingUser.ID,
			CreatedAt:   existingUser.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:   existingUser.UpdatedAt.Format(time.RFC3339Nano),
			Username:    existingUser.Username,
			MobileNumber: existingUser.MobileNumber,
			CountryCode: *existingUser.CountryCode,
			IsValidated: existingUser.IsValidated,
			UsageRight:   userRole,
		},
	}, nil
}
