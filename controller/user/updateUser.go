package user

import (
	"context"
	"net/http"
	"time"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
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
	if req.AadhaarNumber != "" {
		existingUser.AadhaarNumber = &req.AadhaarNumber
	}
	if req.Status != "" {
		existingUser.Status = &req.Status
	}
	if req.Name != "" {
		existingUser.Name = &req.Name
	}
	if req.CareOf != "" {
		existingUser.CareOf = &req.CareOf
	}
	if req.DateOfBirth != "" {
		existingUser.DateOfBirth = &req.DateOfBirth
	}
	if req.Photo != "" {
		existingUser.Photo = &req.Photo
	}
	if req.EmailHash != "" {
		existingUser.EmailHash = &req.EmailHash
	}
	if req.ShareCode != "" {
		existingUser.ShareCode = &req.ShareCode
	}
	if req.YearOfBirth != "" {
		existingUser.YearOfBirth = &req.YearOfBirth
	}
	if req.Message != "" {
		existingUser.Message = &req.Message
	}
	if req.MobileNumber == 0 { 
		existingUser.MobileNumber = req.MobileNumber
}

	if err := s.UserRepo.UpdateUser(ctx, *existingUser); err != nil {
		return nil, err
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
	pbUser := &pb.User{
		Id:           existingUser.ID,
		Username:     existingUser.Username,
		Password:     existingUser.Password,
		IsValidated:  existingUser.IsValidated,
		CreatedAt:    existingUser.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:    existingUser.UpdatedAt.Format(time.RFC3339Nano),
		AadhaarNumber: *existingUser.AadhaarNumber,
		Status:       *existingUser.Status,
		Name:         *existingUser.Name,
		CareOf:       *existingUser.CareOf,
		DateOfBirth:  *existingUser.DateOfBirth,
		Photo:        *existingUser.Photo,
		EmailHash:    *existingUser.EmailHash,
		ShareCode:    *existingUser.ShareCode,
		YearOfBirth:  *existingUser.YearOfBirth,
		Message:      *existingUser.Message,
		MobileNumber: existingUser.MobileNumber,
		UsageRight: userRoleResponse,
	}

	return &pb.UpdateUserResponse{
		StatusCode: http.StatusOK,
		Success: true,
		Message:    "User updated successfully",
		Data:       pbUser,
		DataTimeStamp: time.Now().Format(time.RFC3339), // Current time in RFC3339 string format
	}, nil
}
