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

func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	users, err := s.userService.GetUsers()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch users: %v", err)
	}

	var pbUsers []*pb.User
	for _, user := range users {

		var pbAddress *pb.Address
		if user.AddressID != nil {
			address, err := s.userService.GetAddressByID(*user.AddressID)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to fetch address: %v", err)
			}

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
			}
		}

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
		}

		pbUsers = append(pbUsers, pbUser)
	}

	return &pb.GetUserResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "Users fetched successfully",
		Data:          pbUsers,
		DataTimeStamp: time.Now().Format(time.RFC3339),
	}, nil
}
