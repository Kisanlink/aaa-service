package user

import (
	"context"
	"time"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	user, err := s.UserRepo.FindExistingUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	roles, permissions, actions, err := s.UserRepo.FindUserRolesAndPermissions(ctx, user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user roles and permissions: %v", err)
	}
	address, err := s.UserRepo.GetAddressByID(ctx, *user.AddressID)
	if err != nil {
		return nil, err
	}
	userRole := &pb.UserRoleResponse{
		Roles:       roles,
		Permissions: permissions,
		Actions:     actions,
	}
	pbUser := &pb.User{
		Id:           user.ID,
		Username:     user.Username,
		Password:     user.Password,
		IsValidated:  user.IsValidated,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:    user.UpdatedAt.Format(time.RFC3339Nano),
		UsageRight:   userRole,
		AadhaarNumber: *user.AadhaarNumber,
		Status:       *user.Status,
		Name:         *user.Name,
		CareOf:       *user.CareOf,
		DateOfBirth:  *user.DateOfBirth,
		Photo:        *user.Photo,
		EmailHash:    *user.EmailHash,
		ShareCode:    *user.ShareCode,
		YearOfBirth:  *user.YearOfBirth,
		Message:      *user.Message,
		MobileNumber: user.MobileNumber,
		Address: &pb.Address{
			Id:          address.ID,
			Plot:        *address.Plot,
				Street:      *address.Street,
				Landmark:    *address.Landmark,
				PostOffice:  *address.PostOffice,
				Subdistrict: *address.Subdistrict,
				District:    *address.District,
				Vtc:         *address.VTC,
				State:       *address.State,
				Country:     *address.Country,
				Pincode:     *address.Pincode,
				FullAddress: *address.FullAddress,
		},
	}
	return &pb.GetUserByIdResponse{
		StatusCode: int32(codes.OK),
		Message:    "User fetched successfully",
		User:       pbUser,
	}, nil
}
