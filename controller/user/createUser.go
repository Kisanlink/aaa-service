package user

import (
	"context"
	"log"
	"strings"

	// "fmt"
	// "log"
	"time"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/pb"
	"github.com/Kisanlink/aaa-service/repositories"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedUserServiceServer
	UserRepo *repositories.UserRepository
}

func NewUserServer(userRepo *repositories.UserRepository) *Server {
	return &Server{
		UserRepo: userRepo,
	}
}
func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
    if err := s.UserRepo.CheckIfUserExists(ctx, req.Username); err != nil {
        return nil, err
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

    if err := s.createUserRoles(createdUser.ID, req.UserRoleIds); err != nil {
        return nil, err
    }
    userRoles, err := s.UserRepo.FindUserRoles(ctx, createdUser.ID)
    if err != nil {
        return nil, err
    }
    newUser.Roles = userRoles
    pbUserRoles := ConvertToPBUserRoles(userRoles)
    roles, permissions, err := s.UserRepo.FindUserRolesAndPermissions(ctx, createdUser.ID)
    if err != nil {
        log.Fatalf("Failed to fetch user roles and permissions: %v", err)
    }
	updated, err := client.CreateUserRoleRelationship(
		strings.ToLower(createdUser.Username), 
		LowerCaseSlice(roles), 
		LowerCaseSlice(permissions),
	)
		if err != nil {
		log.Fatalf("Error reading schema: %v", err)
	}
	log.Printf("create Relation  Response: %+v", updated)
    // Return the response
    return &pb.CreateUserResponse{
        StatusCode: int32(codes.OK),
        Message:    "User created successfully",
        User: &pb.User{
            Id:          newUser.ID,
            CreatedAt:   newUser.CreatedAt.Format(time.RFC3339Nano),
            UpdatedAt:   newUser.UpdatedAt.Format(time.RFC3339Nano),
            Username:    newUser.Username,
            IsValidated: newUser.IsValidated,
            UserRoles:   pbUserRoles,
        },
    }, nil
}

func (s *Server) createUserRoles(userID string, rolePermissionIDs []string) error {
	if len(rolePermissionIDs) == 0 {
		return nil
	}
	var userRoles []model.UserRole
	for _, rolePermissionID := range rolePermissionIDs {
		userRole := model.UserRole{
			UserID:           userID,
			RolePermissionID: rolePermissionID,
		}
		userRoles = append(userRoles, userRole)
	}
	if err := s.UserRepo.CreateUserRoles(context.Background(), userRoles); err != nil {
		return err
	}
	return nil
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