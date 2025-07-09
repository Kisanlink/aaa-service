package interfaces

import (
	"context"

	"github.com/Kisanlink/aaa-service/entities/models"
	userRequests "github.com/Kisanlink/aaa-service/entities/requests/users"
	userResponses "github.com/Kisanlink/aaa-service/entities/responses/users"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
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

// DatabaseManager interface for database operations
type DatabaseManager interface {
	Connect(ctx context.Context) error
	Close() error
	HealthCheck(ctx context.Context) error
	GetPostgresManager() *db.PostgresManager
	GetDynamoManager() *db.DynamoManager
	GetSpiceManager() *db.SpiceManager
	GetManager(backend db.BackendType) db.DBManager
	GetAllManagers() []db.DBManager
	IsConnected(backend db.BackendType) bool
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

// Repository interfaces

// UserRepository interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByMobileNumber(ctx context.Context, mobileNumber uint64) (*models.User, error)
	GetByAadhaarNumber(ctx context.Context, aadhaarNumber string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
	ListActive(ctx context.Context, limit, offset int) ([]*models.User, error)
	Count(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)
	GetWithRoles(ctx context.Context, userID string) (*models.User, error)
	GetWithAddress(ctx context.Context, userID string) (*models.User, error)
	GetWithProfile(ctx context.Context, userID string) (*models.User, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.User, error)
}

// AddressRepository interface for address data operations
type AddressRepository interface {
	Create(ctx context.Context, address *models.Address) error
	GetByID(ctx context.Context, id string) (*models.Address, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Address, error)
	Update(ctx context.Context, address *models.Address) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.Address, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Address, error)
}

// RoleRepository interface for role data operations
type RoleRepository interface {
	Create(ctx context.Context, role *models.Role) error
	GetByID(ctx context.Context, id string) (*models.Role, error)
	GetByName(ctx context.Context, name string) (*models.Role, error)
	Update(ctx context.Context, role *models.Role) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.Role, error)
	GetActive(ctx context.Context, limit, offset int) ([]*models.Role, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Role, error)
}

// UserRoleRepository interface for user-role relationship operations
type UserRoleRepository interface {
	Create(ctx context.Context, userRole *models.UserRole) error
	GetByID(ctx context.Context, id string) (*models.UserRole, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.UserRole, error)
	GetByRoleID(ctx context.Context, roleID string) ([]*models.UserRole, error)
	GetByUserAndRole(ctx context.Context, userID, roleID string) (*models.UserRole, error)
	Update(ctx context.Context, userRole *models.UserRole) error
	Delete(ctx context.Context, id string) error
	DeleteByUserAndRole(ctx context.Context, userID, roleID string) error
	List(ctx context.Context, limit, offset int) ([]*models.UserRole, error)
	GetActiveByUserID(ctx context.Context, userID string) ([]*models.UserRole, error)
}

// Service interfaces

// UserService interface for user business logic
type UserService interface {
	CreateUser(ctx context.Context, req *userRequests.CreateUserRequest) (*userResponses.UserResponse, error)
	GetUserByID(ctx context.Context, userID string) (*userResponses.UserResponse, error)
	GetUserByUsername(ctx context.Context, username string) (*userResponses.UserResponse, error)
	GetUserByMobileNumber(ctx context.Context, mobileNumber uint64) (*userResponses.UserResponse, error)
	UpdateUser(ctx context.Context, req *userRequests.UpdateUserRequest) (*userResponses.UserResponse, error)
	DeleteUser(ctx context.Context, userID string) (*userResponses.UserResponse, error)
	ListUsers(ctx context.Context, filters interface{}) (interface{}, error)
	SearchUsers(ctx context.Context, query string, limit, offset int) (interface{}, error)
	ValidateUser(ctx context.Context, userID string) (*userResponses.UserResponse, error)
	AssignRole(ctx context.Context, userID, roleID string) (*userResponses.UserResponse, error)
	RemoveRole(ctx context.Context, userID, roleID string) (*userResponses.UserResponse, error)
	GetUserRoles(ctx context.Context, userID string) (interface{}, error)
	GetUserProfile(ctx context.Context, userID string) (interface{}, error)
	UpdateUserProfile(ctx context.Context, userID string, req interface{}) (interface{}, error)
	LockAccount(ctx context.Context, userID string) (*userResponses.UserResponse, error)
	UnlockAccount(ctx context.Context, userID string) (*userResponses.UserResponse, error)
	GetUserActivity(ctx context.Context, userID string) (interface{}, error)
	GetUserAuditTrail(ctx context.Context, userID string) (interface{}, error)
	BulkOperations(ctx context.Context, req interface{}) (interface{}, error)
}

// AddressService interface for address business logic
type AddressService interface {
	CreateAddress(ctx context.Context, req interface{}) (interface{}, error)
	GetAddressByID(ctx context.Context, addressID string) (interface{}, error)
	GetAddressByUserID(ctx context.Context, userID string) (interface{}, error)
	UpdateAddress(ctx context.Context, req interface{}) (interface{}, error)
	DeleteAddress(ctx context.Context, addressID string) error
	ListAddresses(ctx context.Context, filters interface{}) (interface{}, error)
	SearchAddresses(ctx context.Context, query string, limit, offset int) (interface{}, error)
	ValidateAddress(ctx context.Context, address interface{}) error
	GeocodingAddress(ctx context.Context, address interface{}) (interface{}, error)
}

// RoleService interface for role business logic
type RoleService interface {
	CreateRole(ctx context.Context, req interface{}) (interface{}, error)
	GetRoleByID(ctx context.Context, roleID string) (interface{}, error)
	GetRoleByName(ctx context.Context, name string) (interface{}, error)
	UpdateRole(ctx context.Context, req interface{}) (interface{}, error)
	DeleteRole(ctx context.Context, roleID string) error
	ListRoles(ctx context.Context, filters interface{}) (interface{}, error)
	SearchRoles(ctx context.Context, query string, limit, offset int) (interface{}, error)
	GetActiveRoles(ctx context.Context, limit, offset int) (interface{}, error)
	AssignPermission(ctx context.Context, roleID, permissionID string) (interface{}, error)
	RemovePermission(ctx context.Context, roleID, permissionID string) error
	GetRolePermissions(ctx context.Context, roleID string) (interface{}, error)
	GetRoleHierarchy(ctx context.Context) (interface{}, error)
	AddChildRole(ctx context.Context, parentRoleID, childRoleID string) (interface{}, error)
	ValidateRoleHierarchy(ctx context.Context, roleID string) error
}

// AuthService interface for authentication operations
type AuthService interface {
	Login(ctx context.Context, req interface{}) (interface{}, error)
	Register(ctx context.Context, req interface{}) (interface{}, error)
	RefreshToken(ctx context.Context, req interface{}) (interface{}, error)
	Logout(ctx context.Context, req interface{}) error
	ForgotPassword(ctx context.Context, req interface{}) (interface{}, error)
	ResetPassword(ctx context.Context, req interface{}) (interface{}, error)
	ValidateToken(ctx context.Context, token string) (interface{}, error)
	GenerateTokens(ctx context.Context, userID string) (interface{}, error)
	RevokeToken(ctx context.Context, token string) error
	GetCurrentUser(ctx context.Context, token string) (interface{}, error)
}

// MFAService interface for multi-factor authentication
type MFAService interface {
	SetupMFA(ctx context.Context, userID string, req interface{}) (interface{}, error)
	VerifyMFA(ctx context.Context, userID string, req interface{}) (interface{}, error)
	DisableMFA(ctx context.Context, userID string) error
	GetMFAStatus(ctx context.Context, userID string) (interface{}, error)
	GenerateBackupCodes(ctx context.Context, userID string) (interface{}, error)
	ValidateBackupCode(ctx context.Context, userID, code string) (interface{}, error)
}

// PermissionService interface for permission management
type PermissionService interface {
	CreatePermission(ctx context.Context, req interface{}) (interface{}, error)
	GetPermissionByID(ctx context.Context, permissionID string) (interface{}, error)
	UpdatePermission(ctx context.Context, req interface{}) (interface{}, error)
	DeletePermission(ctx context.Context, permissionID string) error
	ListPermissions(ctx context.Context, filters interface{}) (interface{}, error)
	EvaluatePermission(ctx context.Context, userID, resource, action string) (bool, error)
	GrantTemporaryPermission(ctx context.Context, req interface{}) (interface{}, error)
	RevokeTemporaryPermission(ctx context.Context, permissionID string) error
	GetUserPermissions(ctx context.Context, userID string) (interface{}, error)
	CheckPermission(ctx context.Context, userID, resource, action string) (bool, error)
}

// AuditService interface for audit logging
type AuditService interface {
	LogEvent(ctx context.Context, event interface{}) error
	GetAuditTrail(ctx context.Context, userID string) (interface{}, error)
	GetSystemAuditTrail(ctx context.Context, filters interface{}) (interface{}, error)
	GetUserActivity(ctx context.Context, userID string) (interface{}, error)
	LogSecurityEvent(ctx context.Context, event interface{}) error
	GenerateAuditReport(ctx context.Context, req interface{}) (interface{}, error)
}

// HealthService interface for health checking
type HealthService interface {
	CheckHealth(ctx context.Context) error
	CheckReadiness(ctx context.Context) error
	GetDetailedHealth(ctx context.Context) (interface{}, error)
	GetMetrics(ctx context.Context) (interface{}, error)
	GetSystemInfo(ctx context.Context) (interface{}, error)
}

// Handler interfaces

// UserHandler interface for user HTTP handlers
type UserHandler interface {
	CreateUser(c *gin.Context)
	GetUserByID(c *gin.Context)
	UpdateUser(c *gin.Context)
	DeleteUser(c *gin.Context)
	ListUsers(c *gin.Context)
	SearchUsers(c *gin.Context)
	ValidateUser(c *gin.Context)
	AssignRole(c *gin.Context)
	RemoveRole(c *gin.Context)
	GetUserRoles(c *gin.Context)
	GetUserProfile(c *gin.Context)
	UpdateUserProfile(c *gin.Context)
	LockAccount(c *gin.Context)
	UnlockAccount(c *gin.Context)
	GetUserActivity(c *gin.Context)
	GetUserAuditTrail(c *gin.Context)
	BulkOperations(c *gin.Context)
}

// Middleware interfaces

// AuthMiddleware interface for authentication middleware
type AuthMiddleware interface {
	Authenticate() gin.HandlerFunc
	RequireRole(roles ...string) gin.HandlerFunc
	RequirePermission(resource, action string) gin.HandlerFunc
	OptionalAuth() gin.HandlerFunc
}

// LoggingMiddleware interface for logging middleware
type LoggingMiddleware interface {
	RequestLogger() gin.HandlerFunc
	ErrorLogger() gin.HandlerFunc
}

// RateLimitMiddleware interface for rate limiting
type RateLimitMiddleware interface {
	RateLimit(requests int, duration int) gin.HandlerFunc
	UserRateLimit(requests int, duration int) gin.HandlerFunc
	IPRateLimit(requests int, duration int) gin.HandlerFunc
}

// ErrorHandler interface for error handling
type ErrorHandler interface {
	HandleError(c *gin.Context, err error)
	HandleValidationError(c *gin.Context, err error)
	HandleAuthError(c *gin.Context, err error)
	HandleNotFoundError(c *gin.Context)
	HandleInternalError(c *gin.Context, err error)
}

// Utility interfaces

// TokenManager interface for token management
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

// FileService interface for file operations
type FileService interface {
	UploadFile(ctx context.Context, file interface{}) (string, error)
	DownloadFile(ctx context.Context, fileID string) ([]byte, error)
	DeleteFile(ctx context.Context, fileID string) error
	GetFileMetadata(ctx context.Context, fileID string) (interface{}, error)
	GeneratePresignedURL(ctx context.Context, fileID string) (string, error)
}

// ConfigService interface for configuration management
type ConfigService interface {
	GetConfig(key string) (interface{}, error)
	SetConfig(key string, value interface{}) error
	GetAllConfigs() (map[string]interface{}, error)
	ReloadConfig() error
}

// MetricsService interface for metrics collection
type MetricsService interface {
	IncrementCounter(metric string, tags map[string]string)
	RecordHistogram(metric string, value float64, tags map[string]string)
	RecordGauge(metric string, value float64, tags map[string]string)
	GetMetrics() (interface{}, error)
}

// Server interfaces

// HTTPServer interface for HTTP server operations
type HTTPServer interface {
	Start() error
	Stop(ctx context.Context) error
	GetRouter() *gin.Engine
	RegisterRoutes()
	RegisterMiddleware()
}

// GRPCServer interface for gRPC server operations
type GRPCServer interface {
	Start(addr string) error
	Stop() error
	RegisterServices()
}

// WebSocketServer interface for WebSocket server operations
type WebSocketServer interface {
	Start() error
	Stop() error
	HandleConnection(c *gin.Context)
	BroadcastMessage(message interface{}) error
	SendMessageToUser(userID string, message interface{}) error
}
