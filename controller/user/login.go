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
	if !helper.IsValidUsername(req.Username) {
        return nil, status.Errorf(
            codes.InvalidArgument,
            "username '%s' contains invalid characters. Only alphanumeric (a-z, A-Z, 0-9), /, _, |, -, =, + are allowed, and spaces are prohibited",
            req.Username,
        )
    }
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid password")
	}
	roles, permissions, err := s.UserRepo.FindUsageRights(ctx, existingUser.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user roles and permissions: %v", err)
	}
	pbPermissions := make([]*pb.PermissionResponse, len(permissions))
	for i, perm := range permissions {
		pbPermissions[i] = &pb.PermissionResponse{
			Name:        perm.Name,
			Description: perm.Description,
			Action:      perm.Action,
			Source:      perm.Source,
			Resource:    perm.Resource,
		}
	}
	userRoleResponse := &pb.UserRoleResponse{
		Roles:       roles,
		Permissions: pbPermissions,
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
		DataTimeStamp: time.Now().Format(time.RFC3339),
		Data: &pb.AssignRolePermission{
			Id:          existingUser.ID,
			CreatedAt:   existingUser.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:   existingUser.UpdatedAt.Format(time.RFC3339Nano),
			Username:    existingUser.Username,
			IsValidated: existingUser.IsValidated,
			UsageRight:   userRoleResponse,
		},

	}, nil
}
