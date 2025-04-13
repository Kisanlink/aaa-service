package user

import (
	"context"
	"net/http"
	"time"

	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	existingUser, err := s.UserRepo.FindExistingUserByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user: %v", err)
	}
	if existingUser == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Update user fields
	if req.Username != "" {
		existingUser.Username = req.Username
	}
	if req.AadhaarNumber != "" {
		existingUser.AadhaarNumber = &req.AadhaarNumber
	}
	if req.Status != "" {
		existingUser.Status = &req.Status
	}
	if req.Name != "" {
		existingUser.Name = &req.Name
	}
	if req.CareOf != "" {
		existingUser.CareOf = &req.CareOf
	}
	if req.DateOfBirth != "" {
		existingUser.DateOfBirth = &req.DateOfBirth
	}
	if req.Photo != "" {
		existingUser.Photo = &req.Photo
	}
	if req.EmailHash != "" {
		existingUser.EmailHash = &req.EmailHash
	}
	if req.ShareCode != "" {
		existingUser.ShareCode = &req.ShareCode
	}
	if req.YearOfBirth != "" {
		existingUser.YearOfBirth = &req.YearOfBirth
	}
	if req.Message != "" {
		existingUser.Message = &req.Message
	}
	if req.MobileNumber != 0 {
		existingUser.MobileNumber = req.MobileNumber
	}

	// Update the user in database
	if err := s.UserRepo.UpdateUser(ctx, *existingUser); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	// Get updated user's role permissions
	rolePermissions, err := s.UserRepo.FindUsageRights(ctx, existingUser.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch role permissions: %v", err)
	}

	// Convert role permissions to protobuf format
	pbRolePermissions := make(map[string]*pb.RolePermissions)
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

		pbRolePermissions[role] = &pb.RolePermissions{
			Permissions: pbPermissions,
		}
	}

	// Handle address if needed
	var pbAddress *pb.Address
	if existingUser.AddressID != nil {
		address, err := s.UserRepo.GetAddressByID(ctx, *existingUser.AddressID)
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
				CreatedAt:   address.CreatedAt.Format(time.RFC3339Nano),
				UpdatedAt:   address.UpdatedAt.Format(time.RFC3339Nano),
			}
		}
	}

	// Prepare response with role permissions
	pbUser := &pb.User{
		Id:              existingUser.ID,
		Username:        existingUser.Username,
		Password:        "", // Explicitly empty for security
		IsValidated:     existingUser.IsValidated,
		CreatedAt:       existingUser.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:       existingUser.UpdatedAt.Format(time.RFC3339Nano),
		RolePermissions: pbRolePermissions,
		AadhaarNumber:   safeString(existingUser.AadhaarNumber),
		Status:          safeString(existingUser.Status),
		Name:            safeString(existingUser.Name),
		CareOf:          safeString(existingUser.CareOf),
		DateOfBirth:     safeString(existingUser.DateOfBirth),
		Photo:           safeString(existingUser.Photo),
		EmailHash:       safeString(existingUser.EmailHash),
		ShareCode:       safeString(existingUser.ShareCode),
		YearOfBirth:     safeString(existingUser.YearOfBirth),
		Message:         safeString(existingUser.Message),
		MobileNumber:    existingUser.MobileNumber,
		CountryCode:     safeString(existingUser.CountryCode),
		Address:         pbAddress,
	}

	return &pb.UpdateUserResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "User updated successfully",
		Data:          pbUser,
		DataTimeStamp: time.Now().Format(time.RFC3339),
	}, nil
}
