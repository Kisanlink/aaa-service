package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

func (repo *UserRepository) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
    if err := repo.DB.Table("users").Create(user).Error; err != nil {
        return nil, status.Error(codes.Internal, "Failed to create user")
    }
    return user, nil 
}
func (repo *UserRepository) CreateAddress(ctx context.Context, address *model.Address) (*model.Address, error) {
    if err := repo.DB.Table("addresses").Create(address).Error; err != nil {
        return nil, status.Error(codes.Internal, "Failed to create adress")
    }
    return address, nil 
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
func (repo *UserRepository) GetAddressByID(ctx context.Context, addressId string) (*model.Address, error) {
	var address model.Address
	err := repo.DB.Table("addresses").Where("id = ?", addressId).First(&address).Error
	if err != nil {
		if err.Error() == "record not found" {
			return nil, status.Error(codes.NotFound, "Address not found")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to fetch address: %v", err))
	}
	return &address, nil
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
func (repo *UserRepository) FindUserRolesAndPermissions(ctx context.Context, userID string) ([]string, []string, []string, error) {
    var userRoles []model.UserRole
    err := repo.DB.Table("user_roles").Where("user_id = ?", userID).Find(&userRoles).Error
    if err != nil {
        return nil, nil, nil, status.Error(codes.Internal, "Failed to fetch user roles")
    }

    var roles []string
    var permissions []string
    var actions []string
    permissionSet := make(map[string]struct{})
    actionSet := make(map[string]struct{})

    for _, userRole := range userRoles {
        var rolePermission model.RolePermission
        err := repo.DB.Table("role_permissions").
            Where("id = ?", userRole.RolePermissionID).
            Preload("Role").
            Preload("Permission").
            First(&rolePermission).Error
        if err != nil {
            return nil, nil, nil, status.Error(codes.Internal, "Failed to fetch role permissions")
        }

        if rolePermission.Role.ID != "" { 
            roles = append(roles, rolePermission.Role.Name)
        }

        if rolePermission.Permission.ID != "" {
            permissionSet[rolePermission.Permission.Name] = struct{}{}
            actionSet[rolePermission.Permission.Action] = struct{}{}
        }
    }
    for permission := range permissionSet {
        permissions = append(permissions, permission)
    }
    for action := range actionSet {
        actions = append(actions, action)
    }
    for i, role := range roles {
        roles[i] = strings.ToLower(role)
    }
    for i, permission := range permissions {
        permissions[i] = strings.ToLower(permission)
    }
    for i, action := range actions {
        actions[i] = strings.ToLower(action)
    }

    return roles, permissions, actions, nil
}
func (repo *UserRepository) CreateUserRoles(ctx context.Context, userRoles []model.UserRole) error {
        if err := repo.DB.Table("user_roles").Create(&userRoles).Error; err != nil {
            return status.Error(codes.Internal, "Failed to create UserRole entries")
        }
    
    return nil
}
func (repo *UserRepository) GetUserRoleByID(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	if err := repo.DB.Preload("UserRoles").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
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
