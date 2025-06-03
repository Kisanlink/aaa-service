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
	// Fetch all user roles for the given userID
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
		var role model.Role
		err := repo.DB.Table("roles").
			Where("id = ?", userRole.RoleID).
			Preload("RolePermissions").
			Preload("RolePermissions.Permission").
			First(&role).Error
		if err != nil {
			return nil, nil, nil, status.Error(codes.Internal, "Failed to fetch role details")
		}

		if role.ID != "" {
			roles = append(roles, role.Name)
		}

		// Iterate through all role permissions of this role
		for _, rolePermission := range role.RolePermissions {
			if rolePermission.Permission.ID != "" {
				permissionSet[rolePermission.Permission.Name] = struct{}{}
				actionSet[rolePermission.Permission.Action] = struct{}{}
			}
		}
	}

	// Convert sets to slices
	for permission := range permissionSet {
		permissions = append(permissions, permission)
	}
	for action := range actionSet {
		actions = append(actions, action)
	}

	// Convert all strings to lowercase
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

func (repo *UserRepository) FindRoleUsersAndPermissionsByRoleId(ctx context.Context, roleID string) ([]string, []string, []string, []string, error) {
	// First verify the role exists
	var role model.Role
	err := repo.DB.Table("roles").
		Where("id = ?", roleID).
		Preload("RolePermissions").
		Preload("RolePermissions.Permission").
		First(&role).Error
	if err != nil {
		return nil, nil, nil, nil, status.Error(codes.NotFound, "Role not found")
	}

	// Initialize data structures
	var roles []string
	var permissions []string
	var actions []string
	var connectedUsernames []string
	permissionSet := make(map[string]struct{})
	actionSet := make(map[string]struct{})
	usernameSet := make(map[string]struct{})

	// Add the main role name
	if role.ID != "" {
		roles = append(roles, role.Name)
	}

	// Get all users connected to this role with proper preloading
	var roleUsers []model.UserRole
	err = repo.DB.Table("user_roles").
		Where("role_id = ?", roleID).
		Preload("User"). // Preload the User relationship
		Find(&roleUsers).Error
	if err != nil {
		return nil, nil, nil, nil, status.Error(codes.Internal, "Failed to fetch connected users")
	}

	// Collect all usernames
	for _, ru := range roleUsers {
		if ru.User.Username != "" { // Directly access the User's Username field
			usernameSet[ru.User.Username] = struct{}{}
		}
	}

	// Get all permissions and actions from this role
	for _, rolePermission := range role.RolePermissions {
		if rolePermission.Permission.ID != "" {
			permissionSet[rolePermission.Permission.Name] = struct{}{}
			actionSet[rolePermission.Permission.Action] = struct{}{}
		}
	}

	// Convert sets to slices
	for permission := range permissionSet {
		permissions = append(permissions, permission)
	}
	for action := range actionSet {
		actions = append(actions, action)
	}
	for username := range usernameSet {
		connectedUsernames = append(connectedUsernames, username)
	}

	// Convert all strings to lowercase
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

// func (repo *UserRepository) CreateUserRoles(ctx context.Context, userRoles model.UserRole) error {
// 	if err := repo.DB.Table("user_roles").Create(&userRoles).Error; err != nil {
// 		return status.Error(codes.Internal, "Failed to create UserRole entries")
// 	}

//		return nil
//	}

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
		return nil, status.Errorf(codes.Internal, "failed to fetch users: %w", err)
	}

	return users, nil
}
func (repo *UserRepository) CreateUserRoles(ctx context.Context, userRole model.UserRole) error {
	// First check if this user-role assignment already exists
	var count int64
	err := repo.DB.Table("user_roles").
		Where("user_id = ? AND role_id = ?", userRole.UserID, userRole.RoleID).
		Count(&count).Error

	if err != nil {
		return status.Errorf(codes.Internal, "failed to check existing role assignment: %v", err)
	}

	if count > 0 {
		return status.Errorf(codes.AlreadyExists, "user already has this role assigned")
	}

	// If we get here, the assignment doesn't exist, so create it
	if err := repo.DB.Table("user_roles").Create(&userRole).Error; err != nil {
		return status.Errorf(codes.Internal, "failed to create user role assignment: %v", err)
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

// added for password reset
func (repo *UserRepository) UpdatePassword(ctx context.Context, userID string, newPassword string) error {
	err := repo.DB.Table("users").Where("id = ?", userID).Update("password", newPassword).Error
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed to update password: %v", err))
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

func (repo *UserRepository) FindUserByMobile(ctx context.Context, mobileNumber uint64) (*model.User, error) {
	var existingUser model.User

	query := repo.DB.Table("users").Where("mobile_number = ?", mobileNumber)
	err := query.First(&existingUser).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "database error: "+err.Error())
	}
	return &existingUser, nil
}

func (repo *UserRepository) FindUserByAadhaar(ctx context.Context, aadhaarNumber string) (*model.User, error) {
	var existingUser model.User

	query := repo.DB.Table("users").Where("aadhaar_number = ?", aadhaarNumber)
	err := query.First(&existingUser).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "database error: "+err.Error())
	}
	return &existingUser, nil
}

func (repo *UserRepository) FindUsageRights(ctx context.Context, userID string) (map[string][]model.Permission, error) {
	// Fetch user roles
	var userRoles []model.UserRole
	if err := repo.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&userRoles).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch user roles")
	}

	// Create a map to store permissions by role
	rolePermissions := make(map[string][]model.Permission)

	// Process each user role
	for _, userRole := range userRoles {
		var role model.Role
		if err := repo.DB.WithContext(ctx).
			Where("id = ?", userRole.RoleID).
			Preload("RolePermissions.Permission").
			First(&role).Error; err != nil {
			return nil, status.Error(codes.Internal, "failed to fetch role details")
		}

		// Skip if role name is empty
		if role.Name == "" {
			continue
		}

		// Initialize the slice if this role hasn't been seen before
		if _, exists := rolePermissions[role.Name]; !exists {
			rolePermissions[role.Name] = make([]model.Permission, 0)
		}

		// Collect permissions for this role
		for _, rp := range role.RolePermissions {
			if rp.Permission.ID != "" {
				// Copy only needed fields to avoid including RolePermissions
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

// Get user's current token count
func (repo *UserRepository) GetTokensByUserID(ctx context.Context, userID string) (int, error) {
	var user model.User
	err := repo.DB.WithContext(ctx).
		Where("id = ?", userID).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, status.Error(codes.NotFound, "User not found")
		}
		return 0, status.Error(codes.Internal, fmt.Sprintf("Failed to fetch tokens: %v", err))
	}

	return user.Tokens, nil
}

// Credit tokens to a user
func (repo *UserRepository) CreditUserByID(ctx context.Context, userID string, tokens int) (*model.User, error) {
	var user model.User
	err := repo.DB.WithContext(ctx).
		Where("id = ?", userID).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to fetch user: %v", err))
	}

	user.Tokens += tokens

	if err := repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}
	return &user, nil
}

// Debit tokens from a user
func (repo *UserRepository) DebitUserByID(ctx context.Context, userID string, tokens int) (*model.User, error) {
	var user model.User
	err := repo.DB.WithContext(ctx).
		Where("id = ?", userID).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to fetch user: %v", err))
	}

	if user.Tokens < tokens {
		return nil, errors.New("insufficient tokens")
	}

	user.Tokens -= tokens

	if err := repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}
	return &user, nil
}
