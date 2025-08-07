package interfaces

import (
	"context"

	"github.com/Kisanlink/aaa-service/entities/models"
	userRequests "github.com/Kisanlink/aaa-service/entities/requests/users"
	userResponses "github.com/Kisanlink/aaa-service/entities/responses/users"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger interface for structured logging
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	Named(name string) Logger
	Sync() error
}

// CacheService interface for caching operations
type CacheService interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl int) error
	Delete(key string) error
	Exists(key string) bool
	Clear() error
	Keys(pattern string) ([]string, error)
	Expire(key string, ttl int) error
	TTL(key string) (int, error)
	Close() error
}

// Validator interface for input validation
type Validator interface {
	ValidateStruct(s interface{}) error
	ValidateUserID(userID string) error
	ValidateEmail(email string) error
	ValidatePassword(password string) error
	ValidatePhoneNumber(phone string) error
	ValidateAadhaarNumber(aadhaar string) error
	ParseListFilters(c *gin.Context) (interface{}, error)
}

// Responder interface for HTTP responses
type Responder interface {
	SendSuccess(c *gin.Context, statusCode int, data interface{})
	SendError(c *gin.Context, statusCode int, message string, err error)
	SendValidationError(c *gin.Context, errors []string)
	SendInternalError(c *gin.Context, err error)
}

// UserService interface for user-related business operations
type UserService interface {
	CreateUser(ctx context.Context, req *userRequests.CreateUserRequest) (*userResponses.UserResponse, error)
	GetUserByID(ctx context.Context, userID string) (*userResponses.UserResponse, error)
	GetUserByUsername(ctx context.Context, username string) (*userResponses.UserResponse, error)
	GetUserByMobileNumber(ctx context.Context, mobileNumber uint64) (*userResponses.UserResponse, error)
	GetUserByAadhaarNumber(ctx context.Context, aadhaarNumber string) (*userResponses.UserResponse, error)
	UpdateUser(ctx context.Context, req *userRequests.UpdateUserRequest) (*userResponses.UserResponse, error)
	DeleteUser(ctx context.Context, userID string) error
	ListUsers(ctx context.Context, limit, offset int) (interface{}, error)
	ListActiveUsers(ctx context.Context, limit, offset int) (interface{}, error)
	SearchUsers(ctx context.Context, keyword string, limit, offset int) (interface{}, error)
	ValidateUser(ctx context.Context, userID string) error
	DeductTokens(ctx context.Context, userID string, amount int) error
	AddTokens(ctx context.Context, userID string, amount int) error
	GetUserWithProfile(ctx context.Context, userID string) (*userResponses.UserResponse, error)
	GetUserWithRoles(ctx context.Context, userID string) (*userResponses.UserResponse, error)
	VerifyUserPassword(ctx context.Context, username, password string) (*userResponses.UserResponse, error)
	VerifyUserPasswordByPhone(ctx context.Context, phoneNumber, countryCode, password string) (*userResponses.UserResponse, error)
	SetMPin(ctx context.Context, userID string, mPin string) error
	VerifyMPin(ctx context.Context, userID string, mPin string) error
	GetUserByPhoneNumber(ctx context.Context, phoneNumber, countryCode string) (*userResponses.UserResponse, error)
}

// AddressService interface for address-related operations
type AddressService interface {
	CreateAddress(ctx context.Context, address *models.Address) error
	GetAddressByID(ctx context.Context, addressID string) (*models.Address, error)
	UpdateAddress(ctx context.Context, address *models.Address) error
	DeleteAddress(ctx context.Context, addressID string) error
	GetAddressesByUserID(ctx context.Context, userID string) ([]*models.Address, error)
	SearchAddresses(ctx context.Context, query string, limit, offset int) ([]*models.Address, error)
}

// RoleService interface for role management operations
type RoleService interface {
	CreateRole(ctx context.Context, role *models.Role) error
	GetRoleByID(ctx context.Context, roleID string) (*models.Role, error)
	GetRoleByName(ctx context.Context, name string) (*models.Role, error)
	UpdateRole(ctx context.Context, role *models.Role) error
	DeleteRole(ctx context.Context, roleID string) error
	ListRoles(ctx context.Context, limit, offset int) ([]*models.Role, error)
	SearchRoles(ctx context.Context, query string, limit, offset int) ([]*models.Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID string) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]*models.UserRole, error)
}

// AuthService interface for authentication operations
type AuthService interface {
	Login(ctx context.Context, username, password string) (interface{}, error)
	Logout(ctx context.Context, token string) error
	RefreshToken(ctx context.Context, refreshToken string) (interface{}, error)
	ValidateToken(ctx context.Context, token string) (interface{}, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, username string) error
	VerifyEmail(ctx context.Context, userID, verificationCode string) error
	ResendVerificationEmail(ctx context.Context, userID string) error
}

// UserRepository interface for user data operations
type UserRepository interface {
	base.Repository[*models.User]
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByPhoneNumber(ctx context.Context, phoneNumber string, countryCode string) (*models.User, error)
	GetByMobileNumber(ctx context.Context, mobileNumber uint64) (*models.User, error)
	GetByAadhaarNumber(ctx context.Context, aadhaarNumber string) (*models.User, error)
	ListActive(ctx context.Context, limit, offset int) ([]*models.User, error)
	CountActive(ctx context.Context) (int64, error)
	Search(ctx context.Context, keyword string, limit, offset int) ([]*models.User, error)
	GetWithRoles(ctx context.Context, userID string) (*models.User, error)
	GetWithAddress(ctx context.Context, userID string) (*models.User, error)
	GetWithProfile(ctx context.Context, userID string) (*models.User, error)
}

// AddressRepository interface for address data operations
type AddressRepository interface {
	base.Repository[*models.Address]
	GetByUserID(ctx context.Context, userID string) ([]*models.Address, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Address, error)
}

// RoleRepository interface for role data operations
type RoleRepository interface {
	base.Repository[*models.Role]
	GetByName(ctx context.Context, name string) (*models.Role, error)
	GetActive(ctx context.Context, limit, offset int) ([]*models.Role, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Role, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
}

// UserRoleRepository interface for user-role relationship operations
type UserRoleRepository interface {
	base.Repository[*models.UserRole]
	GetByUserID(ctx context.Context, userID string) ([]*models.UserRole, error)
	GetByRoleID(ctx context.Context, roleID string) ([]*models.UserRole, error)
	GetByUserAndRole(ctx context.Context, userID, roleID string) (*models.UserRole, error)
	DeleteByUserAndRole(ctx context.Context, userID, roleID string) error
	ExistsByUserAndRole(ctx context.Context, userID, roleID string) (bool, error)
}

// TokenManager interface for token operations
type TokenManager interface {
	GenerateAccessToken(userID string, claims map[string]interface{}) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateToken(token string) (map[string]interface{}, error)
	RevokeToken(token string) error
	GetTokenClaims(token string) (map[string]interface{}, error)
	IsTokenRevoked(token string) (bool, error)
}

// PasswordManager interface for password management
type PasswordManager interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hashedPassword string) error
	GenerateRandomPassword(length int) (string, error)
	ValidatePasswordStrength(password string) error
}

// EmailService interface for email operations
type EmailService interface {
	SendWelcomeEmail(ctx context.Context, userID string) error
	SendPasswordResetEmail(ctx context.Context, userID, token string) error
	SendAccountVerificationEmail(ctx context.Context, userID, token string) error
	SendSecurityAlertEmail(ctx context.Context, userID, message string) error
	SendBulkEmail(ctx context.Context, userIDs []string, template string, data interface{}) error
}

// SMSService interface for SMS operations
type SMSService interface {
	SendOTP(ctx context.Context, phoneNumber, otp string) error
	SendSecurityAlert(ctx context.Context, phoneNumber, message string) error
	SendBulkSMS(ctx context.Context, phoneNumbers []string, message string) error
	ValidatePhoneNumber(phoneNumber string) error
}

// NotificationService interface for notifications
type NotificationService interface {
	SendNotification(ctx context.Context, userID, message string) error
	SendBulkNotification(ctx context.Context, userIDs []string, message string) error
	GetUserNotifications(ctx context.Context, userID string) (interface{}, error)
	MarkAsRead(ctx context.Context, userID, notificationID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
}

// MaintenanceService interface for maintenance mode management
type MaintenanceService interface {
	IsMaintenanceMode(ctx context.Context) (bool, interface{}, error)
	EnableMaintenanceMode(ctx context.Context, config interface{}) error
	DisableMaintenanceMode(ctx context.Context, disabledBy string) error
	GetMaintenanceStatus(ctx context.Context) (interface{}, error)
	IsUserAllowedDuringMaintenance(ctx context.Context, userID string, isAdmin bool, isReadOperation bool) (bool, error)
	UpdateMaintenanceMessage(ctx context.Context, message string, updatedBy string) error
}

// HealthService interface for health check operations
type HealthService interface {
	CheckDatabaseHealth(ctx context.Context) error
	CheckCacheHealth(ctx context.Context) error
	CheckExternalServiceHealth(ctx context.Context) error
	GetOverallHealth(ctx context.Context) (interface{}, error)
}
