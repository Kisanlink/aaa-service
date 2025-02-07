package user

import (
	"context"
	"errors"
	"time"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/pb"
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

// func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
// 	existingUser := model.User{}
// 	if err := s.DB.Table("users").Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
// 		return nil, status.Error(codes.AlreadyExists, "User Already Exists")
// 	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
// 		return nil, status.Error(codes.Internal, "Database Error")
// 	}

// 	hashedPassword, err := HashPassword(req.Password)
// 	if err != nil {
// 		return nil, status.Error(codes.InvalidArgument, "Failed to hash password")
// 	}
// 	newUser := model.User{
// 		Username:    req.Username,
// 		Password:    hashedPassword,
// 		IsValidated: false,
// 	}
// 	newUser.Password = hashedPassword
// 	if err := s.DB.Table("users").Create(&newUser).Error; err != nil {
// 		return nil, status.Error(codes.Internal, "Failed to create user")
// 	}

// 	return &pb.CreateUserResponse{
// 		User: &pb.User{
// 			Id:          newUser.ID.String(),
// 			CreatedAt:   newUser.CreatedAt.Format(time.RFC3339Nano),
// 			UpdatedAt:   newUser.UpdatedAt.Format(time.RFC3339Nano),
// 			Username:    newUser.Username,
// 			IsValidated: newUser.IsValidated,
// 		},
// 	}, nil
// }

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	existingUser := model.User{}
	if err := s.DB.Table("users").Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, status.Error(codes.AlreadyExists, "User Already Exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.Internal, "Database Error")
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
	if err := s.DB.Table("users").Create(&newUser).Error; err != nil {
		return nil, status.Error(codes.Internal, "Failed to create user")
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, roleID := range req.RoleIds {
		for _, permissionID := range req.PermissionIds {
			userRole := model.UserRole{
				UserID:       newUser.ID,
				RoleID:       roleID,
				PermissionID: permissionID,
			}
			if err := tx.Table("user_roles").Create(&userRole).Error; err != nil {
				tx.Rollback()
				return nil, status.Error(codes.Internal, "Failed to create user-role-permission connection")
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, status.Error(codes.Internal, "Transaction commit failed")
	}

	var createdUser model.User
	if err := s.DB.Table("users").
		Preload("Roles.Role").
		Preload("Roles.Permission").
		First(&createdUser, "id = ?", newUser.ID).
		Error; err != nil {
		return nil, status.Error(codes.Internal, "Failed to fetch user details")
	}

	response := &pb.CreateUserResponse{
		User: MapUserToProto(createdUser),
	}
	return response, nil
}

func MapUserToProto(user model.User) *pb.User {
	roles := make([]*pb.RoleConn, len(user.Roles))
	for i, userRole := range user.Roles {
		roles[i] = &pb.RoleConn{
			Id:          userRole.Role.ID, // Convert UUID to string
			Name:        userRole.Role.Name,
			Description: userRole.Role.Description,
			Permissions: []*pb.PermissionConn{
				{
					Id:          userRole.Permission.ID, // Convert UUID to string
					Name:        userRole.Permission.Name,
					Description: userRole.Permission.Description,
				},
			},
		}
	}

	return &pb.User{
		Id:          user.ID, // Convert UUID to string
		Username:    user.Username,
		IsValidated: user.IsValidated,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339Nano),
		Roles:       roles,
	}
}
