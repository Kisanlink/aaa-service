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
	// Validate request
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	authInterceptor := middleware.AuthInterceptor()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		userID, ok := ctx.Value("user_id").(string)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "user_id not found in context")
		}

		log.Printf("Authenticated user with ID: %s", userID)

		// Type assertion with check
		validateReq, ok := req.(*pb.ValidateUserRequest)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "invalid request type")
		}

		existingUser, err := s.userService.FindExistingUserByID(validateReq.UserId)
		if err != nil {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		if existingUser == nil {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		var address *model.Address
		if validateReq.Address != nil {
			newAddress := &model.Address{
				House:       &validateReq.Address.House,
				Street:      &validateReq.Address.Street,
				Landmark:    &validateReq.Address.Landmark,
				PostOffice:  &validateReq.Address.PostOffice,
				Subdistrict: &validateReq.Address.Subdistrict,
				District:    &validateReq.Address.District,
				VTC:         &validateReq.Address.Vtc,
				State:       &validateReq.Address.State,
				Country:     &validateReq.Address.Country,
				Pincode:     &validateReq.Address.Pincode,
				FullAddress: &validateReq.Address.FullAddress,
			}

			address, err = s.userService.CreateAddress(newAddress)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to create address")
			}
		}

		// Update user fields with nil checks
		existingUser.IsValidated = validateReq.IsValidated
		if validateReq.Status != "" {
			existingUser.Status = &validateReq.Status
		}
		if validateReq.Name != "" {
			existingUser.Name = &validateReq.Name
		}
		if validateReq.ShareCode != "" {
			existingUser.ShareCode = &validateReq.ShareCode
		}
		if validateReq.CareOf != "" {
			existingUser.CareOf = &validateReq.CareOf
		}
		if validateReq.Photo != "" {
			existingUser.Photo = &validateReq.Photo
		}
		if validateReq.DateOfBirth != "" {
			existingUser.DateOfBirth = &validateReq.DateOfBirth
		}
		if validateReq.EmailHash != "" {
			existingUser.EmailHash = &validateReq.EmailHash
		}
		if validateReq.YearOfBirth != "" {
			existingUser.YearOfBirth = &validateReq.YearOfBirth
		}
		if validateReq.Message != "" {
			existingUser.Message = &validateReq.Message
		}
		if address != nil {
			existingUser.AddressID = &address.ID
		}

		if err := s.userService.UpdateUser(*existingUser); err != nil {
			return nil, status.Error(codes.Internal, "failed to update user")
		}

		// Prepare response with nil checks
		pbUser := &pb.User{
			Id:           existingUser.ID,
			Username:     existingUser.Username,
			IsValidated:  existingUser.IsValidated,
			MobileNumber: existingUser.MobileNumber,
		}

		if existingUser.AadhaarNumber != nil {
			pbUser.AadhaarNumber = *existingUser.AadhaarNumber
		}
		if existingUser.Status != nil {
			pbUser.Status = *existingUser.Status
		}
		if existingUser.Name != nil {
			pbUser.Name = *existingUser.Name
		}
		if existingUser.CareOf != nil {
			pbUser.CareOf = *existingUser.CareOf
		}
		if existingUser.DateOfBirth != nil {
			pbUser.DateOfBirth = *existingUser.DateOfBirth
		}
		if existingUser.Photo != nil {
			pbUser.Photo = *existingUser.Photo
		}
		if existingUser.EmailHash != nil {
			pbUser.EmailHash = *existingUser.EmailHash
		}
		if existingUser.ShareCode != nil {
			pbUser.ShareCode = *existingUser.ShareCode
		}
		if existingUser.YearOfBirth != nil {
			pbUser.YearOfBirth = *existingUser.YearOfBirth
		}
		if existingUser.Message != nil {
			pbUser.Message = *existingUser.Message
		}
		if existingUser.CountryCode != nil {
			pbUser.CountryCode = *existingUser.CountryCode
		}

		pbUser.CreatedAt = existingUser.CreatedAt.Format(time.RFC3339Nano)
		pbUser.UpdatedAt = existingUser.UpdatedAt.Format(time.RFC3339Nano)

		if address != nil {
			pbUser.Address = &pb.Address{
				Id:          address.ID,
				House:       safeDerefString(address.House),
				Street:      safeDerefString(address.Street),
				Landmark:    safeDerefString(address.Landmark),
				PostOffice:  safeDerefString(address.PostOffice),
				Subdistrict: safeDerefString(address.Subdistrict),
				District:    safeDerefString(address.District),
				Vtc:         safeDerefString(address.VTC),
				State:       safeDerefString(address.State),
				Country:     safeDerefString(address.Country),
				Pincode:     safeDerefString(address.Pincode),
				FullAddress: safeDerefString(address.FullAddress),
				CreatedAt:   address.CreatedAt.Format(time.RFC3339Nano),
				UpdatedAt:   address.UpdatedAt.Format(time.RFC3339Nano),
			}
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

// Helper function to safely dereference string pointers
func safeDerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
