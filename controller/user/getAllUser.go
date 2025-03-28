package user

import (
	"context"
	"net/http"
	"time"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	users, err := s.UserRepo.GetUsers(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch users: %v", err)
	}

	var pbUsers []*pb.User
	for _, user := range users {
		roles, permissions, err := s.UserRepo.FindUsageRights(ctx, user.ID)
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
		var pbAddress *pb.Address
		if user.AddressID != nil {
			address, err := s.UserRepo.GetAddressByID(ctx, *user.AddressID)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to fetch address: %v", err)
			}

			pbAddress = &pb.Address{
				Id:          address.ID,
				House:        safeString(address.House),
				Street:      safeString(address.Street),
				Landmark:    safeString(address.Landmark),
				PostOffice:  safeString(address.PostOffice),
				Subdistrict: safeString(address.Subdistrict),
				District:    safeString(address.District),
				Vtc:         safeString(address.VTC),
				State:       safeString(address.State),
				Country:     safeString(address.Country),
				Pincode:     safeString(address.Pincode),
				FullAddress: safeString(address.FullAddress),
			}
		}
		pbUser := &pb.User{
			Id:            user.ID,
			Username:      user.Username,
			Password:      "", // Don't return password in response
			IsValidated:   user.IsValidated,
			CreatedAt:     user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     user.UpdatedAt.Format(time.RFC3339),
			UsageRight:    userRoleResponse,
			AadhaarNumber: safeString(user.AadhaarNumber),
			Status:        safeString(user.Status),
			Name:          safeString(user.Name),
			CareOf:        safeString(user.CareOf),
			DateOfBirth:   safeString(user.DateOfBirth),
			Photo:         safeString(user.Photo),
			EmailHash:     safeString(user.EmailHash),
			ShareCode:     safeString(user.ShareCode),
			YearOfBirth:   safeString(user.YearOfBirth),
			Message:       safeString(user.Message),
			MobileNumber:  user.MobileNumber,
			CountryCode:   safeString(user.CountryCode),
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
