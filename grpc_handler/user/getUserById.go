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

	// Get role permissions for the user
	rolePermissions, err := s.userService.FindUsageRights(user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch role permissions: %v", err)
	}

	// Convert role permissions to protobuf format
	var pbRolePermissions []*pb.RolePermissions
	for role, permissions := range rolePermissions {
		// Deduplicate permissions
		uniquePerms := make(map[string]*pb.PermissionResponse)
		for _, perm := range permissions {
			key := perm.Name + ":" + perm.Action + ":" + perm.Resource
			if _, exists := uniquePerms[key]; !exists {
				uniquePerms[key] = &pb.PermissionResponse{
					Name:        perm.Name,
					Description: perm.Description,
					Action:      perm.Action,
					Source:      perm.Source,
					Resource:    perm.Resource,
				}
			}
		}

		// Convert to slice
		var pbPermissions []*pb.PermissionResponse
		for _, perm := range uniquePerms {
			pbPermissions = append(pbPermissions, perm)
		}

		pbRolePermissions = append(pbRolePermissions, &pb.RolePermissions{
			RoleName:    role,
			Permissions: pbPermissions,
		})
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

	// Prepare user response with role permissions
	pbUser := &pb.User{
		Id:              user.ID,
		Username:        user.Username,
		Password:        "", // Explicitly empty for security
		IsValidated:     user.IsValidated,
		CreatedAt:       user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       user.UpdatedAt.Format(time.RFC3339),
		RolePermissions: pbRolePermissions,
		AadhaarNumber:   helper.SafeString(user.AadhaarNumber),
		Status:          helper.SafeString(user.Status),
		Name:            helper.SafeString(user.Name),
		CareOf:          helper.SafeString(user.CareOf),
		DateOfBirth:     helper.SafeString(user.DateOfBirth),
		Photo:           helper.SafeString(user.Photo),
		EmailHash:       helper.SafeString(user.EmailHash),
		ShareCode:       helper.SafeString(user.ShareCode),
		YearOfBirth:     helper.SafeString(user.YearOfBirth),
		Message:         helper.SafeString(user.Message),
		MobileNumber:    user.MobileNumber,
		CountryCode:     helper.SafeString(user.CountryCode),
		Address:         pbAddress,
	}

	return &pb.GetUserByIdResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "User fetched successfully",
		Data:          pbUser,
		DataTimeStamp: time.Now().Format(time.RFC3339),
	}, nil
}
