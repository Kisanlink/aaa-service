package user

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/middleware"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ValidateUser(ctx context.Context, req *pb.ValidateUserRequest) (*pb.ValidateUserResponse, error) {
	authInterceptor := middleware.AuthInterceptor()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		userID, ok := ctx.Value("user_id").(string)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "user_id not found in context")
		}

		log.Printf("Authenticated user with ID: %s", userID)

		existingUser, err := s.userService.FindExistingUserByID(userID)
		if err != nil {
			return nil, err
		}

		newAddress := &model.Address{
			House:       &req.(*pb.ValidateUserRequest).Address.House,
			Street:      &req.(*pb.ValidateUserRequest).Address.Street,
			Landmark:    &req.(*pb.ValidateUserRequest).Address.Landmark,
			PostOffice:  &req.(*pb.ValidateUserRequest).Address.PostOffice,
			Subdistrict: &req.(*pb.ValidateUserRequest).Address.Subdistrict,
			District:    &req.(*pb.ValidateUserRequest).Address.District,
			VTC:         &req.(*pb.ValidateUserRequest).Address.Vtc,
			State:       &req.(*pb.ValidateUserRequest).Address.State,
			Country:     &req.(*pb.ValidateUserRequest).Address.Country,
			Pincode:     &req.(*pb.ValidateUserRequest).Address.Pincode,
			FullAddress: &req.(*pb.ValidateUserRequest).Address.FullAddress,
		}

		address, err := s.userService.CreateAddress(newAddress)
		if err != nil {
			return nil, err
		}

		existingUser.IsValidated = req.(*pb.ValidateUserRequest).IsValidated
		existingUser.AadhaarNumber = &req.(*pb.ValidateUserRequest).AadhaarNumber
		existingUser.Status = &req.(*pb.ValidateUserRequest).Status
		existingUser.Name = &req.(*pb.ValidateUserRequest).Name
		existingUser.ShareCode = &req.(*pb.ValidateUserRequest).ShareCode
		existingUser.CareOf = &req.(*pb.ValidateUserRequest).CareOf
		existingUser.Photo = &req.(*pb.ValidateUserRequest).Photo
		existingUser.DateOfBirth = &req.(*pb.ValidateUserRequest).DateOfBirth
		existingUser.YearOfBirth = &req.(*pb.ValidateUserRequest).YearOfBirth
		existingUser.Message = &req.(*pb.ValidateUserRequest).Message
		existingUser.AddressID = &address.ID
		existingUser.CountryCode = &req.(*pb.ValidateUserRequest).CountryCode

		if err := s.userService.UpdateUser(*existingUser); err != nil {
			return nil, err
		}
		rolesResponse, err := s.userService.GetUserRolesWithPermissions(existingUser.ID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to fetch user roles: %v", err)
		}
		pbUser := &pb.User{
			Id:            existingUser.ID,
			Username:      existingUser.Username,
			IsValidated:   existingUser.IsValidated,
			AadhaarNumber: *existingUser.AadhaarNumber,
			Status:        *existingUser.Status,
			Name:          *existingUser.Name,
			CareOf:        *existingUser.CareOf,
			DateOfBirth:   *existingUser.DateOfBirth,
			Photo:         *existingUser.Photo,
			EmailHash:     *existingUser.EmailHash,
			ShareCode:     *existingUser.ShareCode,
			YearOfBirth:   *existingUser.YearOfBirth,
			Message:       *existingUser.Message,
			MobileNumber:  existingUser.MobileNumber,
			CountryCode:   *existingUser.CountryCode,
			CreatedAt:     existingUser.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:     existingUser.UpdatedAt.Format(time.RFC3339Nano),
			Address: &pb.Address{
				Id:          address.ID,
				House:       *address.House,
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
				CreatedAt:   existingUser.CreatedAt.Format(time.RFC3339Nano),
				UpdatedAt:   existingUser.UpdatedAt.Format(time.RFC3339Nano),
			},
			Roles: convertRoleResponseToPB(rolesResponse),
		}

		return &pb.ValidateUserResponse{
			StatusCode: http.StatusOK,
			Success:    true,
			Message:    "User updated successfully",
			Data:       pbUser,
		}, nil
	}

	resp, err := authInterceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "ValidateUser"}, handler)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ValidateUserResponse), nil
}
