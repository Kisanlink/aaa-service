package user

import (
	"context"
	"time"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
)

func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	users, err := s.UserRepo.GetUsers(ctx)
	if err != nil {
		return nil, err
	}
	var pbUsers []*pb.User
	for _, user := range users {
		roles, permissions, actions, err := s.UserRepo.FindUserRolesAndPermissions(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		userRole := &pb.UserRoleResponse{
			Roles:       roles,
			Permissions: permissions,
			Actions:     actions,
		}

		// Fetch address details for the user
		address, err := s.UserRepo.GetAddressByID(ctx, *user.AddressID)
		if err != nil {
			return nil, err
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

		// Append the user to the response
		pbUsers = append(pbUsers, pbUser)
	}

	// Return the response
	return &pb.GetUserResponse{
		StatusCode: int32(codes.OK),
		Message:    "Users fetched successfully",
		Users:      pbUsers,
	}, nil
}
