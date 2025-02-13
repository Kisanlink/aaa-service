package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Kisanlink/aaa-service/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (repo *UserRepository) CreateUser(ctx context.Context, newUser model.User) error {
	if err := repo.DB.Table("users").Create(&newUser).Error; err != nil {
		log.Printf("ERROR: Failed to create user: %v", err)
		return status.Error(codes.Internal, "Failed to create user")
	}
	return nil
}

func (repo *UserRepository) CheckIfUserExists(ctx context.Context, username string) error {
	existingUser := model.User{}
	err := repo.DB.Table("users").Where("username = ?", username).First(&existingUser).Error
	if err == nil {
		return status.Error(codes.AlreadyExists, "User Already Exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return status.Error(codes.Internal, "Database Error")
	}
	return nil
}

func (repo *UserRepository) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	err := repo.DB.Table("users").Where("id = ?", userID).First(&user).Error
	if err != nil {
		if err.Error() == "record not found" {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to fetch user: %v", err))
	}
	return &user, nil
}

func (repo *UserRepository) GetUsers(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := repo.DB.Table("users").Find(&users).Error
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to fetch users: %v", err))
	}
	return users, nil
}

func (repo *UserRepository) FindUserRoles(ctx context.Context, userID string) ([]model.UserRole, error) {
	var userRoles []model.UserRole
	err := repo.DB.Table("user_roles").Where("user_id = ?", userID).Find(&userRoles).Error
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to fetch user roles")
	}
	return userRoles, nil
}

func (repo *UserRepository) CreateUserRoles(ctx context.Context, userRoles []model.UserRole) error {
	if err := repo.DB.Table("user_roles").Create(&userRoles).Error; err != nil {
		return status.Error(codes.Internal, "Failed to create UserRole entries")
	}
	return nil
}

func (repo *UserRepository) DeleteUserRoles(ctx context.Context, id string) error {
	if err := repo.DB.Table("user_roles").Where("user_id = ?", id).Delete(&model.UserRole{}).Error; err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to delete user roles: %v", err))
	}
	return nil
}

func (repo *UserRepository) DeleteUser(ctx context.Context, id string) error {
	if err := repo.DB.Table("users").Delete(&model.User{}, "id = ?", id).Error; err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to delete user: %v", err))
	}
	return nil
}

func (repo *UserRepository) FindExistingUserByID(ctx context.Context, id string) (*model.User, error) {
	var existingUser model.User
	err := repo.DB.Table("users").Where("id = ?", id).First(&existingUser).Error
	if err != nil {
		if err.Error() == "record not found" {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to fetch user: %v", err))
	}
	return &existingUser, nil
}

func (repo *UserRepository) UpdateUser(ctx context.Context, existingUser model.User) error {
	if err := repo.DB.Table("users").Save(&existingUser).Error; err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to update user: %v", err))
	}
	return nil
}

func (repo *UserRepository) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var existingUser model.User
	err := repo.DB.Table("users").Where("username = ?", username).First(&existingUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.NotFound, "User not found")
	} else if err != nil {
		return nil, status.Error(codes.Internal, "Database error: "+err.Error())
	}
	return &existingUser, nil
}
