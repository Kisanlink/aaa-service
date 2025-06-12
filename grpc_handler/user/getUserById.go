package user

import (
	"context"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	user, err := s.userService.FindExistingUserByID(id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user: %v", err)
	}

	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	var pbAddress *pb.Address
	if user.AddressID != nil {
		address, err := s.userService.GetAddressByID(*user.AddressID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to fetch address: %v", err)
		}

		if address != nil {
			pbAddress = &pb.Address{
				Id:          address.ID,
				House:       helper.SafeString(address.House),
				Street:      helper.SafeString(address.Street),
				Landmark:    helper.SafeString(address.Landmark),
				PostOffice:  helper.SafeString(address.PostOffice),
				Subdistrict: helper.SafeString(address.Subdistrict),
				District:    helper.SafeString(address.District),
				Vtc:         helper.SafeString(address.VTC),
				State:       helper.SafeString(address.State),
				Country:     helper.SafeString(address.Country),
				Pincode:     helper.SafeString(address.Pincode),
				FullAddress: helper.SafeString(address.FullAddress),
				CreatedAt:   address.CreatedAt.Format(time.RFC3339),
				UpdatedAt:   address.UpdatedAt.Format(time.RFC3339),
			}
		}
	}
	// Get updated user details with roles and permissions
	rolesResponse, err := s.userService.GetUserRolesWithPermissions(user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch user roles: %v", err)
	}
	// Prepare user response with role permissions
	pbUser := &pb.User{
		Id:            user.ID,
		Username:      user.Username,
		Password:      "", // Explicitly empty for security
		IsValidated:   user.IsValidated,
		CreatedAt:     user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     user.UpdatedAt.Format(time.RFC3339),
		AadhaarNumber: helper.SafeString(user.AadhaarNumber),
		Status:        helper.SafeString(user.Status),
		Name:          helper.SafeString(user.Name),
		CareOf:        helper.SafeString(user.CareOf),
		DateOfBirth:   helper.SafeString(user.DateOfBirth),
		Photo:         helper.SafeString(user.Photo),
		EmailHash:     helper.SafeString(user.EmailHash),
		ShareCode:     helper.SafeString(user.ShareCode),
		YearOfBirth:   helper.SafeString(user.YearOfBirth),
		Message:       helper.SafeString(user.Message),
		MobileNumber:  user.MobileNumber,
		CountryCode:   helper.SafeString(user.CountryCode),
		Address:       pbAddress,
		Roles:         convertRoleResponseToPB(rolesResponse),
	}

	return &pb.GetUserByIdResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "User fetched successfully",
		Data:          pbUser,
		DataTimeStamp: time.Now().Format(time.RFC3339),
	}, nil
}
