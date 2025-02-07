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
	newUser.Password = hashedPassword
	if err := s.DB.Table("users").Create(&newUser).Error; err != nil {
		return nil, status.Error(codes.Internal, "Failed to create user")
	}

	return &pb.CreateUserResponse{
		User: &pb.User{
			Id:          newUser.ID.String(),
			CreatedAt:   newUser.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:   newUser.UpdatedAt.Format(time.RFC3339Nano),
			Username:    newUser.Username,
			IsValidated: newUser.IsValidated,
		},
	}, nil
}
