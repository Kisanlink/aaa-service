package user

import (
	"context"
	"time"

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
	newUser := model.User{
		Username:    req.Username,
		Password:    hashedPassword,
		IsValidated: false,
	}
	if err := s.UserRepo.CreateUser(ctx, newUser); err != nil {
		return nil, err
	}
	if err := s.createUserRoles(newUser.ID, req.UserRoleIds); err != nil {
		return nil, err
	}
	userRoles, err := s.UserRepo.FindUserRoles(ctx, newUser.ID)
	if err != nil {
		return nil, err
	}
	newUser.Roles = userRoles
	pbUserRoles := ConvertToPBUserRoles(userRoles)
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
