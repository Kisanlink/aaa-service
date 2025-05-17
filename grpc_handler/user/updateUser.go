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

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	existingUser, err := s.userService.FindExistingUserByID(id)
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
	if err := s.userService.UpdateUser(*existingUser); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	// Get updated user's role permissions
	rolePermissions, err := s.userService.FindUsageRights(existingUser.ID)
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
			RoleName:    role, // Set the role name here
			Permissions: pbPermissions,
		})
	}

	// Handle address if needed
	var pbAddress *pb.Address
	if existingUser.AddressID != nil {
		address, err := s.userService.GetAddressByID(*existingUser.AddressID)
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
		RolePermissions: pbRolePermissions, // Now this is a slice of RolePermissions
		AadhaarNumber:   helper.SafeString(existingUser.AadhaarNumber),
		Status:          helper.SafeString(existingUser.Status),
		Name:            helper.SafeString(existingUser.Name),
		CareOf:          helper.SafeString(existingUser.CareOf),
		DateOfBirth:     helper.SafeString(existingUser.DateOfBirth),
		Photo:           helper.SafeString(existingUser.Photo),
		EmailHash:       helper.SafeString(existingUser.EmailHash),
		ShareCode:       helper.SafeString(existingUser.ShareCode),
		YearOfBirth:     helper.SafeString(existingUser.YearOfBirth),
		Message:         helper.SafeString(existingUser.Message),
		MobileNumber:    existingUser.MobileNumber,
		CountryCode:     helper.SafeString(existingUser.CountryCode),
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
