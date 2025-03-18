package user

import (
	"context"
	"time"

	"github.com/Kisanlink/aaa-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	// Validate the request
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	// Fetch the user from the repository
	user, err := s.UserRepo.FindExistingUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Fetch user roles, permissions, and actions
	roles, permissions, actions, err := s.UserRepo.FindUserRolesAndPermissions(ctx, user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user roles and permissions: %v", err)
	}

	// Fetch address details for the user
	address, err := s.UserRepo.GetAddressByID(ctx, *user.AddressID)
	if err != nil {
		return nil, err
	}

	// Construct the userRole object in the desired format
	userRole := &pb.UserRoleResponse{
		Roles:       roles,
		Permissions: permissions,
		Actions:     actions,
	}

	// Populate all fields of the User message
	pbUser := &pb.User{
		Id:           user.ID,
		Username:     user.Username,
		Password:     user.Password,
		IsValidated:  user.IsValidated,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:    user.UpdatedAt.Format(time.RFC3339Nano),
		UsageRight:   userRole, // Use the constructed userRole object
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
		MobileNumber: *user.MobileNumber,
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

	// Return the response
	return &pb.GetUserByIdResponse{
		StatusCode: int32(codes.OK),
		Message:    "User fetched successfully",
		User:       pbUser,
	}, nil
}
