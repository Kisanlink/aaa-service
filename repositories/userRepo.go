package repositories

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

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
	GetUsers() ([]model.User, error)
	FindUserRoles(userID string) ([]model.UserRole, error)
	FindUserRolesAndPermissions(userID string) ([]string, []string, []string, error)
	FindRoleUsersAndPermissionsByRoleId(roleID string) ([]string, []string, []string, []string, error)
	CreateUserRoles(userRole model.UserRole) error
	GetUserRoleByID(userID string) (*model.User, error)
	DeleteUserRoles(id string) error
	DeleteUser(id string) error
	FindExistingUserByID(id string) (*model.User, error)
	UpdateUser(existingUser model.User) error
	UpdatePassword(userID string, newPassword string) error
	FindUserByUsername(username string) (*model.User, error)
	FindUserByMobile(mobileNumber uint64) (*model.User, error)
	FindUserByAadhaar(aadhaarNumber string) (*model.User, error)
	FindUsageRights(userID string) (map[string][]model.Permission, error)
	GetTokensByUserID(userID string) (int, error)
	CreditUserByID(userID string, tokens int) (*model.User, error)
	DebitUserByID(userID string, tokens int) (*model.User, error)
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

func (repo *UserRepository) GetUsers() ([]model.User, error) {
	var users []model.User
	err := repo.DB.Table("users").Find(&users).Error
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch users: %w", err))
	}
	return users, nil
}

func (repo *UserRepository) FindUserRoles(userID string) ([]model.UserRole, error) {
	var userRoles []model.UserRole
	err := repo.DB.Table("user_roles").Where("user_id = ?", userID).Find(&userRoles).Error
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch user roles: %w", err))
	}
	return userRoles, nil
}

func (repo *UserRepository) FindUserRolesAndPermissions(userID string) ([]string, []string, []string, error) {
	var userRoles []model.UserRole
	err := repo.DB.Table("user_roles").Where("user_id = ?", userID).Find(&userRoles).Error
	if err != nil {
		return nil, nil, nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch user roles: %w", err))
	}

	var roles []string
	var permissions []string
	var actions []string
	permissionSet := make(map[string]struct{})
	actionSet := make(map[string]struct{})

	for _, userRole := range userRoles {
		var role model.Role
		err := repo.DB.Table("roles").
			Where("id = ?", userRole.RoleID).
			Preload("RolePermissions").
			Preload("RolePermissions.Permission").
			First(&role).Error
		if err != nil {
			return nil, nil, nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch role details: %w", err))
		}

		if role.ID != "" {
			roles = append(roles, role.Name)
		}

		for _, rolePermission := range role.RolePermissions {
			if rolePermission.Permission.ID != "" {
				permissionSet[rolePermission.Permission.Name] = struct{}{}
				actionSet[rolePermission.Permission.Action] = struct{}{}
			}
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

func (repo *UserRepository) FindRoleUsersAndPermissionsByRoleId(roleID string) ([]string, []string, []string, []string, error) {
	var role model.Role
	err := repo.DB.Table("roles").
		Where("id = ?", roleID).
		Preload("RolePermissions").
		Preload("RolePermissions.Permission").
		First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, nil, helper.NewAppError(http.StatusNotFound, errors.New("role not found"))
		}
		return nil, nil, nil, nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch role: %w", err))
	}

	var roles []string
	var permissions []string
	var actions []string
	var connectedUsernames []string
	permissionSet := make(map[string]struct{})
	actionSet := make(map[string]struct{})
	usernameSet := make(map[string]struct{})

	if role.ID != "" {
		roles = append(roles, role.Name)
	}

	var roleUsers []model.UserRole
	err = repo.DB.Table("user_roles").
		Where("role_id = ?", roleID).
		Preload("User").
		Find(&roleUsers).Error
	if err != nil {
		return nil, nil, nil, nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch connected users: %w", err))
	}

	for _, ru := range roleUsers {
		if ru.User.Username != "" {
			usernameSet[ru.User.Username] = struct{}{}
		}
	}

	for _, rolePermission := range role.RolePermissions {
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
	for username := range usernameSet {
		connectedUsernames = append(connectedUsernames, username)
	}

	for i := range roles {
		roles[i] = strings.ToLower(roles[i])
	}
	for i := range permissions {
		permissions[i] = strings.ToLower(permissions[i])
	}
	for i := range actions {
		actions[i] = strings.ToLower(actions[i])
	}
	for i := range connectedUsernames {
		connectedUsernames[i] = strings.ToLower(connectedUsernames[i])
	}

	return roles, permissions, actions, connectedUsernames, nil
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

func (repo *UserRepository) DeleteUserRoles(id string) error {
	if err := repo.DB.Table("user_roles").Where("user_id = ?", id).Delete(&model.UserRole{}).Error; err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete user roles: %w", err))
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

func (repo *UserRepository) FindUsageRights(userID string) (map[string][]model.Permission, error) {
	var userRoles []model.UserRole
	if err := repo.DB.
		Where("user_id = ?", userID).
		Find(&userRoles).Error; err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch user roles: %w", err))
	}

	rolePermissions := make(map[string][]model.Permission)

	for _, userRole := range userRoles {
		var role model.Role
		if err := repo.DB.
			Where("id = ?", userRole.RoleID).
			Preload("RolePermissions.Permission").
			First(&role).Error; err != nil {
			return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to fetch role details: %w", err))
		}

		if role.Name == "" {
			continue
		}

		if _, exists := rolePermissions[role.Name]; !exists {
			rolePermissions[role.Name] = make([]model.Permission, 0)
		}

		for _, rp := range role.RolePermissions {
			if rp.Permission.ID != "" {
				perm := model.Permission{
					Base:           rp.Permission.Base,
					Name:           rp.Permission.Name,
					Description:    rp.Permission.Description,
					Action:         rp.Permission.Action,
					Resource:       rp.Permission.Resource,
					Source:         rp.Permission.Source,
					ValidStartTime: rp.Permission.ValidStartTime,
					ValidEndTime:   rp.Permission.ValidEndTime,
				}
				rolePermissions[role.Name] = append(rolePermissions[role.Name], perm)
			}
		}
	}

	return rolePermissions, nil
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
