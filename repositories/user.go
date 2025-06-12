package repositories

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	CreateUser(user *model.User) (*model.User, error)
	CreateAddress(address *model.Address) (*model.Address, error)
	CheckIfUserExists(username string) error
	GetUserByID(userID string) (*model.User, error)
	GetAddressByID(addressId string) (*model.Address, error)
	GetUsers(roleId, roleName string, page, limit int) ([]model.User, error)
	FindUserRoles(userID string) ([]model.UserRole, error)
	CreateUserRoles(userRole model.UserRole) error
	GetUserRoleByID(userID string) (*model.User, error)
	DeleteUserRoles(userID string, roleID string) error
	DeleteUser(id string) error
	FindExistingUserByID(id string) (*model.User, error)
	UpdateUser(existingUser model.User) error
	UpdatePassword(userID string, newPassword string) error
	FindUserByUsername(username string) (*model.User, error)
	FindUserByMobile(mobileNumber uint64) (*model.User, error)
	FindUserByAadhaar(aadhaarNumber string) (*model.User, error)
	GetTokensByUserID(userID string) (int, error)
	CreditUserByID(userID string, tokens int) (*model.User, error)
	DebitUserByID(userID string, tokens int) (*model.User, error)
	GetUsersByRole(roleId string, page, limit int) ([]model.User, error)
}

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &UserRepository{
		DB: db,
	}
}

func (repo *UserRepository) CreateUser(user *model.User) (*model.User, error) {
	if err := repo.DB.Table("users").Create(user).Error; err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create user: %w", err))
	}
	return user, nil
}

func (repo *UserRepository) CreateAddress(address *model.Address) (*model.Address, error) {
	if err := repo.DB.Table("addresses").Create(address).Error; err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create address: %w", err))
	}
	return address, nil
}

func (repo *UserRepository) CheckIfUserExists(username string) error {
	existingUser := model.User{}
	err := repo.DB.Table("users").Where("username = ?", username).First(&existingUser).Error
	if err == nil {
		return helper.NewAppError(http.StatusConflict, errors.New("user already exists"))
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("database error: %w", err))
	}
	return nil
}

func (repo *UserRepository) GetUserByID(userID string) (*model.User, error) {
	var user model.User
	err := repo.DB.Table("users").Where("id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.NewAppError(http.StatusNotFound, errors.New("user not found"))
		}
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch user: %w", err))
	}
	return &user, nil
}

func (repo *UserRepository) GetAddressByID(addressId string) (*model.Address, error) {
	var address model.Address
	err := repo.DB.Table("addresses").Where("id = ?", addressId).First(&address).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.NewAppError(http.StatusNotFound, errors.New("address not found"))
		}
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch address: %w", err))
	}
	return &address, nil
}

// func (repo *UserRepository) GetUsers(page, limit int) ([]model.User, error) {
// 	var users []model.User
// 	query := repo.DB.Table("users")

// 	// Apply pagination if both page and limit are provided and valid
// 	if page > 0 && limit > 0 {
// 		offset := (page - 1) * limit
// 		query = query.Offset(offset).Limit(limit)
// 	}

//		err := query.Find(&users).Error
//		if err != nil {
//			return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch users: %w", err))
//		}
//		return users, nil
//	}
func (repo *UserRepository) GetUsers(roleId, roleName string, page, limit int) ([]model.User, error) {
	var users []model.User
	query := repo.DB.Table("users")

	// Handle role filtering if either roleId or roleName is provided
	if roleId != "" || roleName != "" {
		// If roleName is provided, we need to get the roleId first
		if roleName != "" {
			var role model.Role
			if err := repo.DB.Table("roles").Where("name = ?", roleName).First(&role).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("role not found: %w", err))
				}
				return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to find role: %w", err))
			}
			roleId = role.ID
		}

		// Join with user_roles table to filter by role
		query = query.
			Joins("JOIN user_roles ON user_roles.user_id = users.id AND user_roles.role_id = ?", roleId).
			Preload("Roles")
	}

	// Apply pagination if parameters are provided
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	// Execute the query
	if err := query.Find(&users).Error; err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch users: %w", err))

	}

	return users, nil
}
func (repo *UserRepository) FindUserRoles(userID string) ([]model.UserRole, error) {
	var userRoles []model.UserRole
	err := repo.DB.
		// Preload("Role"). // This ensures Role data is loaded
		Where("user_id = ?", userID).
		Find(&userRoles).Error

	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError,
			fmt.Errorf("failed to fetch user roles: %w", err))
	}
	return userRoles, nil
}

func (repo *UserRepository) CreateUserRoles(userRole model.UserRole) error {
	var count int64
	err := repo.DB.Table("user_roles").
		Where("user_id = ? AND role_id = ?", userRole.UserID, userRole.RoleID).
		Count(&count).Error

	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to check existing role assignment: %w", err))
	}

	if count > 0 {
		return helper.NewAppError(http.StatusConflict, errors.New("user already has this role assigned"))
	}

	if err := repo.DB.Table("user_roles").Create(&userRole).Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create user role assignment: %w", err))
	}

	return nil
}

func (repo *UserRepository) GetUserRoleByID(userID string) (*model.User, error) {
	var user model.User
	if err := repo.DB.Preload("UserRoles").Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.NewAppError(http.StatusNotFound, errors.New("user not found"))
		}
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch user roles: %w", err))
	}
	return &user, nil
}

func (repo *UserRepository) GetUsersByRole(roleId string, page, limit int) ([]model.User, error) {
	var users []model.User

	query := repo.DB.Table("users").
		Joins("JOIN user_roles ON user_roles.user_id = users.id AND user_roles.role_id = ?", roleId).
		Preload("Roles")

	// Apply pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	err := query.Find(&users).Error
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch users: %w", err))
	}

	return users, nil
}
func (repo *UserRepository) DeleteUserRoles(userID string, roleID string) error {
	if err := repo.DB.Table("user_roles").
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&model.UserRole{}).Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("Failed to delete user role: %w", err))

	}
	return nil
}

func (repo *UserRepository) DeleteUser(id string) error {
	if err := repo.DB.Table("users").Delete(&model.User{}, "id = ?", id).Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete user: %w", err))
	}
	return nil
}

func (repo *UserRepository) FindExistingUserByID(id string) (*model.User, error) {
	var existingUser model.User
	err := repo.DB.Table("users").Where("id = ?", id).First(&existingUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.NewAppError(http.StatusNotFound, errors.New("user not found"))
		}
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch user: %w", err))
	}
	return &existingUser, nil
}

func (repo *UserRepository) UpdateUser(existingUser model.User) error {
	if err := repo.DB.Table("users").Save(&existingUser).Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to update user: %w", err))
	}
	return nil
}

func (repo *UserRepository) UpdatePassword(userID string, newPassword string) error {
	err := repo.DB.Table("users").Where("id = ?", userID).Update("password", newPassword).Error
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to update password: %w", err))
	}
	return nil
}

func (repo *UserRepository) FindUserByUsername(username string) (*model.User, error) {
	var existingUser model.User
	err := repo.DB.Table("users").Where("username = ?", username).First(&existingUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, helper.NewAppError(http.StatusNotFound, errors.New("user not found"))
	} else if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("database error: %w", err))
	}
	return &existingUser, nil
}

func (repo *UserRepository) FindUserByMobile(mobileNumber uint64) (*model.User, error) {
	var existingUser model.User
	err := repo.DB.Table("users").Where("mobile_number = ?", mobileNumber).First(&existingUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("database error: %w", err))
	}
	return &existingUser, nil
}

func (repo *UserRepository) FindUserByAadhaar(aadhaarNumber string) (*model.User, error) {
	var existingUser model.User
	err := repo.DB.Table("users").Where("aadhaar_number = ?", aadhaarNumber).First(&existingUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("database error: %w", err))
	}
	return &existingUser, nil
}

func (repo *UserRepository) GetTokensByUserID(userID string) (int, error) {
	var user model.User
	err := repo.DB.
		Where("id = ?", userID).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, helper.NewAppError(http.StatusNotFound, errors.New("user not found"))
		}
		return 0, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch tokens: %w", err))
	}

	return user.Tokens, nil
}

func (repo *UserRepository) CreditUserByID(userID string, tokens int) (*model.User, error) {
	var user model.User
	err := repo.DB.
		Where("id = ?", userID).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.NewAppError(http.StatusNotFound, errors.New("user not found"))
		}
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch user: %w", err))
	}

	user.Tokens += tokens

	if err := repo.UpdateUser(user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *UserRepository) DebitUserByID(userID string, tokens int) (*model.User, error) {
	var user model.User
	err := repo.DB.
		Where("id = ?", userID).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.NewAppError(http.StatusNotFound, errors.New("user not found"))
		}
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch user: %w", err))
	}

	if user.Tokens < tokens {
		return nil, helper.NewAppError(http.StatusBadRequest, errors.New("insufficient tokens"))
	}

	user.Tokens -= tokens

	if err := repo.UpdateUser(user); err != nil {
		return nil, err
	}
	return &user, nil
}
