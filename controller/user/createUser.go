package user

import (
	"context"
	"errors"
	"time"

	"github.com/Kisanlink/aaa-service/entities/models"
	pb "github.com/Kisanlink/aaa-service/proto"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Server struct {
	pb.UnimplementedUserServiceServer
	DB *gorm.DB
}

var GlobalUserHandler *UserHandler

type UserHandler struct {
	client pb.UserServiceClient
}

func NewUserHandler(client pb.UserServiceClient) *UserHandler {
	GlobalUserHandler = &UserHandler{client: client}
	return GlobalUserHandler
}
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	existingUser := models.User{}
	if err := s.DB.Table("users").Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, status.Error(codes.AlreadyExists, "User Already Exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.Internal, "Database Error")
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Failed to hash password")
	}
	newUser := models.User{
		Username:    req.Username,
		Password:    hashedPassword,
		IsValidated: false,
	}
	newUser.Password = hashedPassword
	if err := s.DB.Table("users").Create(&newUser).Error; err != nil {
		return nil, status.Error(codes.Internal, "Failed to create user")
	}
	if err := s.createUserRoles(newUser.ID, req.UserRoleIds); err != nil {
		return nil, err
	}
	var userRoles []models.UserRole
	if err := s.DB.Table("user_roles").Where("user_id = ?", newUser.ID).Find(&userRoles).Error; err != nil {
		return nil, status.Error(codes.Internal, "Failed to fetch user roles")
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

	var userRoles []models.UserRole
	for _, rolePermissionID := range rolePermissionIDs {
		userRole := models.UserRole{
			UserID:           userID,
			RolePermissionID: rolePermissionID,
		}
		userRoles = append(userRoles, userRole)
	}
	if err := s.DB.Table("user_roles").Create(&userRoles).Error; err != nil {
		return status.Error(codes.Internal, "Failed to create UserRole entries")
	}

	return nil
}

func ConvertToPBUserRoles(userRoles []models.UserRole) []*pb.UserRole {
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
