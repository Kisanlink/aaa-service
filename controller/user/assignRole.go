package user

import (
	"context"
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
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
	roleName := req.GetRole()
	role, err := s.RoleRepo.GetRoleByName(ctx, roleName)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Role with name %s not found", roleName)
	}
	userRole := model.UserRole{
		UserID:   req.UserId,
		RoleID:   role.ID,
		IsActive: true,
	}

	if err := s.UserRepo.CreateUserRoles(ctx, userRole); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create user-role connection: %v", err)
	}
	updatedUser, err := s.UserRepo.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch user details: %v", err)
	}

	roles, permissions, actions, err := s.UserRepo.FindUserRolesAndPermissions(ctx, updatedUser.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user roles and permissions: %v", err)
	}
	response, err := client.DeleteUserRoleRelationship(
		updatedUser.Username,
		helper.LowerCaseSlice(roles),
		helper.LowerCaseSlice(permissions),
		helper.LowerCaseSlice(actions),
	)
	if err != nil {
		log.Printf("Failed to delete relationships: %v", err)
	}
	log.Printf("User roles and permission deleted successfully: %s", response)

	results, err := client.CreateUserRoleRelationship(
		updatedUser.Username,
		helper.LowerCaseSlice(roles),
		helper.LowerCaseSlice(permissions),
		helper.LowerCaseSlice(actions),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create relation: %v", err)
	}
	log.Printf("relationship created successfully: %v", results)
	userRoleResponse := &pb.UserRoleResponse{
		Roles:       roles,
		Permissions: permissions,
		Actions:     actions,
	}
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
		MobileNumber:  updatedUser.MobileNumber,
		CountryCode:   *updatedUser.CountryCode,
		UsageRight:    userRoleResponse,
	}

	return &pb.AssignRoleToUserResponse{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Role assigned to user successfully",
		User:       connUser,
	}, nil
}
func safeDereferenceString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}