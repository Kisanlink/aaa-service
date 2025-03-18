package user

import (
	"context"
	"log"

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

		existingUser, err := s.UserRepo.FindExistingUserByID(ctx, userID)
		if err != nil {
			return nil, err
		}

		newAddress := &model.Address{
			Plot:        &req.(*pb.ValidateUserRequest).Address.Plot,
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

		address, err := s.UserRepo.CreateAddress(ctx, newAddress)
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
		existingUser.EmailHash = &req.(*pb.ValidateUserRequest).EmailHash
		existingUser.YearOfBirth = &req.(*pb.ValidateUserRequest).YearOfBirth
		existingUser.Message = &req.(*pb.ValidateUserRequest).Message
		existingUser.AddressID = &address.ID

		if err := s.UserRepo.UpdateUser(ctx, *existingUser); err != nil {
			return nil, err
		}

		userRoles, err := s.UserRepo.FindUserRoles(ctx, existingUser.ID)
		if err != nil {
			return nil, err
		}

		pbRoles := ConvertToPBUserRoles(userRoles)
		pbUser := &pb.User{
			Id:           existingUser.ID,
			IsValidated:  existingUser.IsValidated,
			UserRoles:    pbRoles,
			AadhaarNumber: *existingUser.AadhaarNumber,
			Status:       *existingUser.Status,
			Name:         *existingUser.Name,
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

		return &pb.ValidateUserResponse{
			StatusCode: int32(codes.OK),
			Message:    "User updated successfully",
			User:       pbUser,
		}, nil
	}

	resp, err := authInterceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "ValidateUser"}, handler)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ValidateUserResponse), nil
}
