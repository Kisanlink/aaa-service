package services

import (
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/repositories"
)

type UserServiceInterface interface {
	CreateUser(user model.User) (*model.User, error)
	CreateAddress(address *model.Address) (*model.Address, error)
	CheckIfUserExists(username string) error
	GetUserByID(id string) (*model.User, error)
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

type UserService struct {
	repo repositories.UserRepositoryInterface
}

func NewUserService(repo repositories.UserRepositoryInterface) UserServiceInterface {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) CreateUser(user model.User) (*model.User, error) {
	result, err := s.repo.CreateUser(&user)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create user: %w", err))
	}
	return result, nil
}

func (s *UserService) CreateAddress(address *model.Address) (*model.Address, error) {
	result, err := s.repo.CreateAddress(address)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create address: %w", err))
	}
	return result, nil
}

func (s *UserService) CheckIfUserExists(username string) error {
	err := s.repo.CheckIfUserExists(username)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to check if user exists: %w", err))
	}
	return nil
}
func (s *UserService) GetUserByID(id string) (*model.User, error) {
	result, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get user by ID: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("user not found"))
	}
	return result, nil
}
func (s *UserService) GetAddressByID(addressId string) (*model.Address, error) {
	result, err := s.repo.GetAddressByID(addressId)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get address by ID: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("address not found"))
	}
	return result, nil
}

func (s *UserService) GetUsers() ([]model.User, error) {
	result, err := s.repo.GetUsers()
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get users: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("users not found"))
	}
	return result, nil
}

func (s *UserService) FindUserRoles(userID string) ([]model.UserRole, error) {
	result, err := s.repo.FindUserRoles(userID)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to find user roles: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("user roles not found"))
	}
	return result, nil
}
func (s *UserService) FindUserRolesAndPermissions(userID string) ([]string, []string, []string, error) {
	role, permission, action, err := s.repo.FindUserRolesAndPermissions(userID)
	if err != nil {
		return nil, nil, nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to find user roles and permissions: %w", err))
	}

	return role, permission, action, nil
}

func (s *UserService) FindRoleUsersAndPermissionsByRoleId(roleID string) ([]string, []string, []string, []string, error) {
	role, permission, action, users, err := s.repo.FindRoleUsersAndPermissionsByRoleId(roleID)
	if err != nil {
		return nil, nil, nil, nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to find role users and permissions: %w", err))
	}

	return role, permission, action, users, nil
}

func (s *UserService) CreateUserRoles(userRole model.UserRole) error {
	err := s.repo.CreateUserRoles(userRole)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to create user roles: %w", err))
	}
	return nil

}
func (s *UserService) GetUserRoleByID(userID string) (*model.User, error) {
	result, err := s.repo.GetUserRoleByID(userID)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get user role by ID: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("user role not found"))
	}
	return result, nil
}

func (s *UserService) DeleteUserRoles(id string) error {

	err := s.repo.DeleteUserRoles(id)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete user roles: %w", err))
	}
	return nil
}

func (s *UserService) DeleteUser(id string) error {
	err := s.repo.DeleteUser(id)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to delete user: %w", err))
	}
	return nil
}

func (s *UserService) FindExistingUserByID(id string) (*model.User, error) {
	result, err := s.repo.FindExistingUserByID(id)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to find existing user by ID: %w", err))
	}
	return result, nil
}

func (s *UserService) UpdateUser(existingUser model.User) error {
	err := s.repo.UpdateUser(existingUser)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to update user: %w", err))
	}
	return nil
}

func (s *UserService) UpdatePassword(userID string, newPassword string) error {
	err := s.repo.UpdatePassword(userID, newPassword)
	if err != nil {
		return helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to update password: %w", err))
	}
	return nil

}

func (s *UserService) FindUserByUsername(username string) (*model.User, error) {
	result, err := s.repo.FindUserByUsername(username)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to find user by username: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("user not found"))
	}
	return result, nil
}

func (s *UserService) FindUserByMobile(mobileNumber uint64) (*model.User, error) {
	result, err := s.repo.FindUserByMobile(mobileNumber)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to find user by mobile number: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("user not found"))
	}
	return result, nil
}

func (s *UserService) FindUserByAadhaar(aadhaarNumber string) (*model.User, error) {
	result, err := s.repo.FindUserByAadhaar(aadhaarNumber)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to find user by aadhaar number: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("user not found"))
	}
	return result, nil
}

func (s *UserService) FindUsageRights(userID string) (map[string][]model.Permission, error) {
	result, err := s.repo.FindUsageRights(userID)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to find usage rights: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("usage rights not found"))
	}
	return result, nil
}
func (s *UserService) GetTokensByUserID(userID string) (int, error) {
	result, err := s.repo.GetTokensByUserID(userID)
	if err != nil {
		return 0, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to get tokens by user ID: %w", err))
	}
	if result == 0 {
		return 0, helper.NewAppError(http.StatusNotFound, fmt.Errorf("tokens not found"))
	}
	return result, nil
}

func (s *UserService) CreditUserByID(userID string, tokens int) (*model.User, error) {
	result, err := s.repo.CreditUserByID(userID, tokens)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to credit user by ID: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("user not found"))
	}
	return result, nil
}

func (s *UserService) DebitUserByID(userID string, tokens int) (*model.User, error) {
	result, err := s.repo.DebitUserByID(userID, tokens)
	if err != nil {
		return nil, helper.NewAppError(http.StatusInternalServerError, fmt.Errorf("failed to debit user by ID: %w", err))
	}
	if result == nil {
		return nil, helper.NewAppError(http.StatusNotFound, fmt.Errorf("user not found"))
	}
	return result, nil
}
