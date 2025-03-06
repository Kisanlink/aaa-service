package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
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
        log.Printf("Failed to create user: %v", err)
        return nil, status.Error(codes.Internal, "Failed to create user")
    }
    return user, nil 
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
func (repo *UserRepository) FindUserRolesAndPermissions(ctx context.Context, userID string) ([]string, []string, error) {
    var userRoles []model.UserRole
    err := repo.DB.Table("user_roles").Where("user_id = ?", userID).Find(&userRoles).Error
    if err != nil {
        return nil, nil, status.Error(codes.Internal, "Failed to fetch user roles")
    }

    var roles []string
    var permissions []string

    // Use a map to avoid duplicate permissions
    permissionSet := make(map[string]struct{})

    for _, userRole := range userRoles {
        // Fetch the RolePermission associated with the UserRole
        var rolePermission model.RolePermission
        err := repo.DB.Table("role_permissions").
            Where("id = ?", userRole.RolePermissionID).
            Preload("Role").
            Preload("PermissionOnRoles.Permission").
            First(&rolePermission).Error
        if err != nil {
            return nil, nil, status.Error(codes.Internal, "Failed to fetch role permissions")
        }

        // Extract Role Name
        if rolePermission.Role != nil {
            roles = append(roles, rolePermission.Role.Name)
        }

        // Extract Permissions
        for _, permissionOnRole := range rolePermission.PermissionOnRoles {
            if permissionOnRole.Permission != nil {
                // Add permission to the set to avoid duplicates
                permissionSet[permissionOnRole.Permission.Name] = struct{}{}
            }
        }
    }

    // Convert the permission set to a slice
    for permission := range permissionSet {
        permissions = append(permissions, permission)
    }
	for i, role := range roles {
		roles[i] = strings.ToLower(role)
	}
	for i, permission := range permissions {
		permissions[i] = strings.ToLower(permission)
	}
    return roles, permissions, nil
}
func (repo *UserRepository) CreateUserRoles(ctx context.Context, userRoles []model.UserRole) error {
    if err := repo.DB.Table("user_roles").Create(&userRoles).Error; err != nil {
        log.Printf("Failed to create UserRole entries: %v", err) // Log the actual error
        return status.Error(codes.Internal, "Failed to create UserRole entries")
    }
    return nil
}
func (repo *UserRepository) ValidateUserAndRolePermission(ctx context.Context, userID string, rolePermissionID string) error {
    // Check if the user exists
    var user model.User
    if err := repo.DB.Table("users").Where("id = ?", userID).First(&user).Error; err != nil {
        return fmt.Errorf("user with ID %s does not exist", userID)
    }

    // Check if the role permission exists
    var rolePermission model.RolePermission
    if err := repo.DB.Table("role_permissions").Where("id = ?", rolePermissionID).First(&rolePermission).Error; err != nil {
        return fmt.Errorf("role permission with ID %s does not exist", rolePermissionID)
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
