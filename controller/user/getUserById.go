package user

import (
	"context"
	"net/http"
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
		return nil, status.Errorf(codes.Internal, "failed to fetch user: %v", err)
	}

	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Get role permissions for the user
	rolePermissions, err := s.UserRepo.FindUsageRights(ctx, user.ID)
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
		address, err := s.UserRepo.GetAddressByID(ctx, *user.AddressID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to fetch address: %v", err)
		}

		if address != nil {
			pbAddress = &pb.Address{
				Id:          address.ID,
				House:       safeString(address.House),
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
		AadhaarNumber:   safeString(user.AadhaarNumber),
		Status:          safeString(user.Status),
		Name:            safeString(user.Name),
		CareOf:          safeString(user.CareOf),
		DateOfBirth:     safeString(user.DateOfBirth),
		Photo:           safeString(user.Photo),
		EmailHash:       safeString(user.EmailHash),
		ShareCode:       safeString(user.ShareCode),
		YearOfBirth:     safeString(user.YearOfBirth),
		Message:         safeString(user.Message),
		MobileNumber:    user.MobileNumber,
		CountryCode:     safeString(user.CountryCode),
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

// Helper function to safely handle nil strings
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
