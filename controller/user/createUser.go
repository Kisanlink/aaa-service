package user

import (
	"context"
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/repositories"
	"github.com/kisanlink/protobuf/pb-aaa"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedUserServiceServer
	UserRepo *repositories.UserRepository
	RoleRepo *repositories.RoleRepository
	PermRepo *repositories.PermissionRepository
	RolePermRepo *repositories.RolePermissionRepository
}

func NewUserServer(userRepo *repositories.UserRepository,roleRepo *repositories.RoleRepository,permRepo *repositories.PermissionRepository,rolePermRepo *repositories.RolePermissionRepository) *Server {
	return &Server{
		UserRepo: userRepo,
		RoleRepo: roleRepo,
		PermRepo: permRepo,
		RolePermRepo: rolePermRepo,
	}
}
func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
    if err := s.UserRepo.CheckIfUserExists(ctx, req.Username); err != nil {
        return nil, err
    }
	if req.Password == "" {
		return nil, status.Error(codes.NotFound,"Password is required")
	}
    hashedPassword, err := HashPassword(req.Password)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "Failed to hash password")
    }
    newUser := &model.User{ 
        Username:    req.Username,
        Password:    hashedPassword,
        IsValidated: false,
    }

	createdUser, err := s.UserRepo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, err
	}
	roles, permissions, actions, err := s.UserRepo.FindUserRolesAndPermissions(ctx, createdUser.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user roles and permissions: %v", err)
	}
	userRole := &pb.UserRoleResponse{
		Roles:       roles,
		Permissions: permissions,
		Actions:     actions,
	}
    return &pb.CreateUserResponse{
        StatusCode: int32(codes.OK),
        Message:    "User created successfully",
        User: &pb.User{
            Id:          createdUser.ID,
            CreatedAt:   createdUser.CreatedAt.Format(time.RFC3339Nano),
            UpdatedAt:   createdUser.UpdatedAt.Format(time.RFC3339Nano),
            Username:    createdUser.Username,
            IsValidated: createdUser.IsValidated,
            UsageRight:   userRole,
        },
    }, nil
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func ConvertToPBUserRoles(userRoles []model.UserRole) []*pb.UserRole {
	var pbUserRoles []*pb.UserRole
	for _, userRole := range userRoles {
		pbUserRole := &pb.UserRole{
			Id:               userRole.ID,
			UserId:           userRole.UserID,
			RolePermissionId: userRole.RolePermissionID,
			CreatedAt:        userRole.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:        userRole.UpdatedAt.Format(time.RFC3339Nano),
		}
		pbUserRoles = append(pbUserRoles, pbUserRole)
	}
	return pbUserRoles
}
func LowerCaseSlice(input []string) []string {
    for i, val := range input {
        input[i] = strings.ToLower(val)
    }
    return input
}