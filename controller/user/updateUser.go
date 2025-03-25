package user

import (
	"context"
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
	// if req.MobileNumber != "" {
	// 	existingUser.MobileNumber = &req.MobileNumber
	// }

	if err := s.UserRepo.UpdateUser(ctx, *existingUser); err != nil {
		return nil, err
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
	}

	return &pb.UpdateUserResponse{
		StatusCode: int32(codes.OK),
		Message:    "User updated successfully",
		User:       pbUser,
	}, nil
}
