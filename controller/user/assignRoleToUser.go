// package user

// import (
// 	"context"
// 	"log"
// 	"net/http"

// 	"github.com/Kisanlink/aaa-service/client"
// 	"github.com/Kisanlink/aaa-service/model"
// 	"github.com/kisanlink/protobuf/pb-aaa"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )

// func (s *Server) AssignRoleToUser(ctx context.Context, req *pb.AssignRoleToUserRequest) (*pb.AssignRoleToUserResponse, error) {
// 	_, err := s.UserRepo.GetUserByID(ctx, req.UserId)
// 	if err != nil {
// 		return nil, status.Errorf(codes.NotFound, "User with ID %s not found", req.UserId)
// 	}

// 	roleIDs := make([]string, 0)
// 	for _, roleName := range req.GetRoles() {
// 		role, err := s.RoleRepo.GetRoleByName(ctx, roleName)
// 		if err != nil {
// 			return nil, status.Errorf(codes.NotFound, "Role with name %s not found", roleName)
// 		}
// 		roleIDs = append(roleIDs, role.ID)
// 	}
// 	rolePermissions, err := s.RolePermRepo.GetRolePermissionsByRoleIDs(ctx, roleIDs)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "Failed to fetch role-permission connections: %v", err)
// 	}

// 	var userRoles []model.UserRole
// 	for _, rp := range rolePermissions {
// 		userRole := model.UserRole{
// 			UserID:           req.UserId,
// 			RolePermissionID: rp.ID,
// 			IsActive:         true,
// 		}
// 		userRoles = append(userRoles, userRole)
// 	}

// 	if err := s.UserRepo.CreateUserRoles(ctx, userRoles); err != nil {
// 		return nil, status.Errorf(codes.Internal, "Failed to create user-role connections: %v", err)
// 	}

// 	updatedUser, err := s.UserRepo.GetUserByID(ctx, req.UserId)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "Failed to fetch user details: %v", err)
// 	}
// 	roles, permissions, action, err := s.UserRepo.FindUserRolesAndPermissions(ctx, updatedUser.ID)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "failed to fetch user roles and permissions")
// 	}
// 	response, err := client.DeleteUserRoleRelationship(updatedUser.Username,
// 		LowerCaseSlice(roles),
// 		LowerCaseSlice(permissions),
// 		LowerCaseSlice(action),)
// 	if err != nil {
// 		log.Printf("Failed to delete relationships: %v", err)
// 	}
// 	log.Printf("User roles and permission deleted successfully: %s", response)
// 	results, err := client.CreateUserRoleRelationship(
// 		updatedUser.Username,
// 		LowerCaseSlice(roles),
// 		LowerCaseSlice(permissions),
// 		LowerCaseSlice(action),
// 	)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "failed to create relation")
// 	}
// 	log.Printf("relationship created successfully: %v", results)

// 	connUser := &pb.User{
// 		Id:            updatedUser.ID,
// 		Username:      updatedUser.Username,
// 		Password:      updatedUser.Password,
// 		IsValidated:   updatedUser.IsValidated,
// 		CreatedAt:     updatedUser.CreatedAt.String(),
// 		UpdatedAt:     updatedUser.UpdatedAt.String(),
// 		AadhaarNumber: safeDereferenceString(updatedUser.AadhaarNumber),
// 		Status:        safeDereferenceString(updatedUser.Status),
// 		Name:          safeDereferenceString(updatedUser.Name),
// 		CareOf:        safeDereferenceString(updatedUser.CareOf),
// 		DateOfBirth:   safeDereferenceString(updatedUser.DateOfBirth),
// 		Photo:         safeDereferenceString(updatedUser.Photo),
// 		EmailHash:     safeDereferenceString(updatedUser.EmailHash),
// 		Message:       safeDereferenceString(updatedUser.Message),
// 		ShareCode:     safeDereferenceString(updatedUser.ShareCode),
// 		YearOfBirth:   safeDereferenceString(updatedUser.YearOfBirth),
// 		MobileNumber:  safeDereferenceString(updatedUser.MobileNumber),
// 	}

// 	for _, ur := range updatedUser.Roles {
// 		connUser.UserRoles = append(connUser.UserRoles, &pb.UserRole{
// 			Id:               ur.ID,
// 			UserId:           ur.UserID,
// 			RolePermissionId: ur.RolePermissionID,
// 			CreatedAt:        ur.CreatedAt.String(),
// 			UpdatedAt:        ur.UpdatedAt.String(),
// 		})
// 	}

// 	return &pb.AssignRoleToUserResponse{
// 		StatusCode: http.StatusOK,
// 		Message:    "Roles assigned to user successfully",
// 		User:       connUser,
// 	}, nil
// }

// // Helper function to safely dereference string pointers
//
//	func safeDereferenceString(s *string) string {
//		if s == nil {
//			return ""
//		}
//		return *s
//	}
package user

import (
	"context"
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/kisanlink/protobuf/pb-aaa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) AssignRole(ctx context.Context, req *pb.AssignRoleToUserRequest) (*pb.AssignRoleToUserResponse, error) {
	_, err := s.UserRepo.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "User with ID %s not found", req.UserId)
	}

	// Changed to handle single role string instead of array
	roleName := req.GetRole() // Changed from GetRoles() to GetRole()
	role, err := s.RoleRepo.GetRoleByName(ctx, roleName)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Role with name %s not found", roleName)
	}
	roleIDs := []string{role.ID} // Create a slice with single role ID

	rolePermissions, err := s.RolePermRepo.GetRolePermissionsByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch role-permission connections: %v", err)
	}

	var userRoles []model.UserRole
	for _, rp := range rolePermissions {
		userRole := model.UserRole{
			UserID:           req.UserId,
			RolePermissionID: rp.ID,
			IsActive:         true,
		}
		userRoles = append(userRoles, userRole)
	}

	if err := s.UserRepo.CreateUserRoles(ctx, userRoles); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create user-role connections: %v", err)
	}

	updatedUser, err := s.UserRepo.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch user details: %v", err)
	}
	roles, permissions, action, err := s.UserRepo.FindUserRolesAndPermissions(ctx, updatedUser.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user roles and permissions")
	}
	response, err := client.DeleteUserRoleRelationship(updatedUser.Username,
		LowerCaseSlice(roles),
		LowerCaseSlice(permissions),
		LowerCaseSlice(action))
	if err != nil {
		log.Printf("Failed to delete relationships: %v", err)
	}
	log.Printf("User roles and permission deleted successfully: %s", response)
	results, err := client.CreateUserRoleRelationship(
		updatedUser.Username,
		LowerCaseSlice(roles),
		LowerCaseSlice(permissions),
		LowerCaseSlice(action),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create relation")
	}
	log.Printf("relationship created successfully: %v", results)

	connUser := &pb.User{
		Id:            updatedUser.ID,
		Username:      updatedUser.Username,
		Password:      updatedUser.Password,
		IsValidated:   updatedUser.IsValidated,
		CreatedAt:     updatedUser.CreatedAt.String(),
		UpdatedAt:     updatedUser.UpdatedAt.String(),
		AadhaarNumber: safeDereferenceString(updatedUser.AadhaarNumber),
		Status:        safeDereferenceString(updatedUser.Status),
		Name:          safeDereferenceString(updatedUser.Name),
		CareOf:        safeDereferenceString(updatedUser.CareOf),
		DateOfBirth:   safeDereferenceString(updatedUser.DateOfBirth),
		Photo:         safeDereferenceString(updatedUser.Photo),
		EmailHash:     safeDereferenceString(updatedUser.EmailHash),
		Message:       safeDereferenceString(updatedUser.Message),
		ShareCode:     safeDereferenceString(updatedUser.ShareCode),
		YearOfBirth:   safeDereferenceString(updatedUser.YearOfBirth),
		MobileNumber:  (updatedUser.MobileNumber),
	}

	for _, ur := range updatedUser.Roles {
		connUser.UserRoles = append(connUser.UserRoles, &pb.UserRole{
			Id:               ur.ID,
			UserId:           ur.UserID,
			RolePermissionId: ur.RolePermissionID,
			CreatedAt:        ur.CreatedAt.String(),
			UpdatedAt:        ur.UpdatedAt.String(),
		})
	}

	return &pb.AssignRoleToUserResponse{
		StatusCode: http.StatusOK,
		Message:    "Role assigned to user successfully",
		User:       connUser,
	}, nil
}

// Helper function to safely dereference string pointers
func safeDereferenceString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}