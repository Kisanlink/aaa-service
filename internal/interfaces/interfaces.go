package interfaces

import (
	"context"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	userRequests "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/users"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/responses"
	userResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/users"
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
	SendPaginatedResponse(c *gin.Context, data interface{}, total, limit, offset int)
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
	ListUsers(ctx context.Context, limit, offset int) (*responses.PaginatedResult, error)
	ListActiveUsers(ctx context.Context, limit, offset int) (interface{}, error)
	SearchUsers(ctx context.Context, keyword string, limit, offset int) (*responses.PaginatedResult, error)
	ValidateUser(ctx context.Context, userID string) error
	DeductTokens(ctx context.Context, userID string, amount int) error
	AddTokens(ctx context.Context, userID string, amount int) error
	GetUserWithProfile(ctx context.Context, userID string) (*userResponses.UserResponse, error)
	GetUserWithRoles(ctx context.Context, userID string) (*userResponses.UserResponse, error)
	VerifyUserPassword(ctx context.Context, username, password string) (*userResponses.UserResponse, error)
	VerifyUserPasswordByPhone(ctx context.Context, phoneNumber, countryCode, password string) (*userResponses.UserResponse, error)
	SetMPin(ctx context.Context, userID string, mPin string, currentPassword string) error
	VerifyMPin(ctx context.Context, userID string, mPin string) error
	UpdateMPin(ctx context.Context, userID, currentMPin, newMPin string) error
	VerifyUserCredentials(ctx context.Context, phone, countryCode string, password, mpin *string) (*userResponses.UserResponse, error)
	GetUserByPhoneNumber(ctx context.Context, phoneNumber, countryCode string) (*userResponses.UserResponse, error)
	SoftDeleteUserWithCascade(ctx context.Context, userID, deletedBy string) error
	GetUserOrganizations(ctx context.Context, userID string) ([]map[string]interface{}, error)
	GetUserGroups(ctx context.Context, userID string) ([]map[string]interface{}, error)
	InitiatePasswordReset(ctx context.Context, phoneNumber, countryCode, username, email *string) (string, error)
	ResetPassword(ctx context.Context, token, newPassword string) error
	GetUserByEmail(ctx context.Context, email string) (*userResponses.UserResponse, error)
}

// AddressService interface for address-related operations
type AddressService interface {
	CreateAddress(ctx context.Context, address *models.Address) error
	FindOrCreateAddress(ctx context.Context, address *models.Address) (string, bool, error)
	GetAddressByID(ctx context.Context, addressID string) (*models.Address, error)
	UpdateAddress(ctx context.Context, address *models.Address) error
	DeleteAddress(ctx context.Context, addressID string) error
	GetAddressesByUserID(ctx context.Context, userID string) ([]*models.Address, error)
	SearchAddresses(ctx context.Context, query string, limit, offset int) ([]*models.Address, error)
	SearchAddressesCount(ctx context.Context, query string) (int64, error)
}

// RoleService interface for role management operations
type RoleService interface {
	CreateRole(ctx context.Context, role *models.Role) error
	GetRoleByID(ctx context.Context, roleID string) (*models.Role, error)
	GetRoleByName(ctx context.Context, name string) (*models.Role, error)
	UpdateRole(ctx context.Context, role *models.Role) error
	DeleteRole(ctx context.Context, roleID string) error
	ListRoles(ctx context.Context, limit, offset int) ([]*models.Role, error)
	CountRoles(ctx context.Context) (int64, error)
	SearchRoles(ctx context.Context, query string, limit, offset int) ([]*models.Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID string) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]*models.UserRole, error)
	GetRoleHierarchy(ctx context.Context) ([]*models.Role, error)
	AddChildRole(ctx context.Context, parentRoleID, childRoleID string) error
	RemoveChildRole(ctx context.Context, parentRoleID, childRoleID string) error
	GetRoleWithChildren(ctx context.Context, roleID string) (*models.Role, error)
}

// GroupService interface for group management operations
type GroupService interface {
	CreateGroup(ctx context.Context, req interface{}) (interface{}, error)
	GetGroup(ctx context.Context, groupID string) (interface{}, error)
	UpdateGroup(ctx context.Context, groupID string, req interface{}) (interface{}, error)
	DeleteGroup(ctx context.Context, groupID string, deletedBy string) error
	ListGroups(ctx context.Context, limit, offset int, organizationID string, includeInactive bool) (interface{}, error)
	CountGroups(ctx context.Context, organizationID string, includeInactive bool) (int64, error)
	AddMemberToGroup(ctx context.Context, req interface{}) (interface{}, error)
	RemoveMemberFromGroup(ctx context.Context, groupID, principalID string, removedBy string) error
	GetGroupMembers(ctx context.Context, groupID string, limit, offset int) (interface{}, error)
	CountGroupMembers(ctx context.Context, groupID string) (int64, error)
	GetUserGroupsInOrganization(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error)

	// Role assignment methods for organization-scoped group operations
	AssignRoleToGroup(ctx context.Context, groupID, roleID, assignedBy string) (interface{}, error)
	RemoveRoleFromGroup(ctx context.Context, groupID, roleID string) error
	GetGroupRoles(ctx context.Context, groupID string) (interface{}, error)
	GetUserEffectiveRoles(ctx context.Context, orgID, userID string) (interface{}, error)
}

// OrganizationService interface for organization management operations
type OrganizationService interface {
	CreateOrganization(ctx context.Context, req interface{}) (interface{}, error)
	GetOrganization(ctx context.Context, orgID string) (interface{}, error)
	UpdateOrganization(ctx context.Context, orgID string, req interface{}) (interface{}, error)
	DeleteOrganization(ctx context.Context, orgID string, deletedBy string) error
	ListOrganizations(ctx context.Context, limit, offset int, includeInactive bool, orgType string) ([]interface{}, error)
	CountOrganizations(ctx context.Context, includeInactive bool, orgType string) (int64, error)
	GetOrganizationHierarchy(ctx context.Context, orgID string) (interface{}, error)
	ActivateOrganization(ctx context.Context, orgID string) error
	DeactivateOrganization(ctx context.Context, orgID string) error
	GetOrganizationStats(ctx context.Context, orgID string) (interface{}, error)

	// New group management methods within organization context
	GetOrganizationGroups(ctx context.Context, orgID string, limit, offset int, includeInactive bool) (interface{}, error)
	CreateGroupInOrganization(ctx context.Context, orgID string, req interface{}) (interface{}, error)
	GetGroupInOrganization(ctx context.Context, orgID, groupID string) (interface{}, error)
	UpdateGroupInOrganization(ctx context.Context, orgID, groupID string, req interface{}) (interface{}, error)
	DeleteGroupInOrganization(ctx context.Context, orgID, groupID string, deletedBy string) error
	GetGroupHierarchyInOrganization(ctx context.Context, orgID, groupID string) (interface{}, error)

	// User-group management within organization context
	AddUserToGroupInOrganization(ctx context.Context, orgID, groupID, userID string, req interface{}) (interface{}, error)
	RemoveUserFromGroupInOrganization(ctx context.Context, orgID, groupID, userID string, removedBy string) error
	GetGroupUsersInOrganization(ctx context.Context, orgID, groupID string, limit, offset int) (interface{}, error)
	GetUserGroupsInOrganization(ctx context.Context, orgID, userID string, limit, offset int) (interface{}, error)

	// Role-group management within organization context
	AssignRoleToGroupInOrganization(ctx context.Context, orgID, groupID, roleID string, req interface{}) (interface{}, error)
	RemoveRoleFromGroupInOrganization(ctx context.Context, orgID, groupID, roleID string, removedBy string) error
	GetGroupRolesInOrganization(ctx context.Context, orgID, groupID string, limit, offset int) (interface{}, error)
	GetUserEffectiveRolesInOrganization(ctx context.Context, orgID, userID string) (interface{}, error)
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
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	ListActive(ctx context.Context, limit, offset int) ([]*models.User, error)
	CountActive(ctx context.Context) (int64, error)
	Search(ctx context.Context, keyword string, limit, offset int) ([]*models.User, error)
	SearchCount(ctx context.Context, keyword string) (int64, error)
	ListAll(ctx context.Context) ([]*models.User, error)
	GetWithAddress(ctx context.Context, userID string) (*models.User, error)
	GetWithProfile(ctx context.Context, userID string) (*models.User, error)
}

// ServiceRepository interface for service data operations (for service-to-service auth)
type ServiceRepository interface {
	GetByAPIKey(ctx context.Context, apiKeyHash string) (*models.Service, error)
}

// AddressRepository interface for address data operations
type AddressRepository interface {
	base.Repository[*models.Address]
	GetByUserID(ctx context.Context, userID string) ([]*models.Address, error)
	GetByFullAddress(ctx context.Context, fullAddress string) (*models.Address, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Address, error)
	SearchCount(ctx context.Context, query string) (int64, error)
}

// RoleRepository interface for role data operations
type RoleRepository interface {
	base.Repository[*models.Role]
	GetByName(ctx context.Context, name string) (*models.Role, error)
	GetActive(ctx context.Context, limit, offset int) ([]*models.Role, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Role, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	GetChildRoles(ctx context.Context, parentRoleID string) ([]*models.Role, error)
	GetAll(ctx context.Context) ([]*models.Role, error)
}

// UserRoleRepository interface for user-role relationship operations
type UserRoleRepository interface {
	base.Repository[*models.UserRole]
	GetByUserID(ctx context.Context, userID string) ([]*models.UserRole, error)
	GetByRoleID(ctx context.Context, roleID string) ([]*models.UserRole, error)
	GetByUserAndRole(ctx context.Context, userID, roleID string) (*models.UserRole, error)
	DeleteByUserAndRole(ctx context.Context, userID, roleID string) error
	DeleteBySourceGroup(ctx context.Context, userID, groupID string) (int, error) // For cleaning up inherited roles
	ExistsByUserAndRole(ctx context.Context, userID, roleID string) (bool, error)
	// Enhanced methods for role management operations
	GetActiveRolesByUserID(ctx context.Context, userID string) ([]*models.UserRole, error)
	AssignRole(ctx context.Context, userID, roleID string) error
	RemoveRole(ctx context.Context, userID, roleID string) error
	IsRoleAssigned(ctx context.Context, userID, roleID string) (bool, error)
}

// ContactRepository interface for contact data operations
type ContactRepository interface {
	base.Repository[*models.Contact]
	GetByUserID(ctx context.Context, userID string) ([]*models.Contact, error)
	GetByType(ctx context.Context, contactType string) ([]*models.Contact, error)
	GetByValue(ctx context.Context, value string) (*models.Contact, error)
	GetPrimaryContact(ctx context.Context, userID string) (*models.Contact, error)
}

// GroupRepository interface for group data operations
type GroupRepository interface {
	base.Repository[*models.Group]
	GetByName(ctx context.Context, name string) (*models.Group, error)
	GetByNameAndOrganization(ctx context.Context, name, organizationID string) (*models.Group, error)
	GetByOrganization(ctx context.Context, organizationID string, limit, offset int, includeInactive bool) ([]*models.Group, error)
	ListActive(ctx context.Context, limit, offset int) ([]*models.Group, error)
	GetChildren(ctx context.Context, parentID string) ([]*models.Group, error)
	HasActiveMembers(ctx context.Context, groupID string) (bool, error)
	CreateMembership(ctx context.Context, membership *models.GroupMembership) error
	UpdateMembership(ctx context.Context, membership *models.GroupMembership) error
	GetMembership(ctx context.Context, groupID, principalID string) (*models.GroupMembership, error)
	GetGroupMembers(ctx context.Context, groupID string, limit, offset int) ([]*models.GroupMembership, error)
}

// GroupRoleRepository interface for group-role relationship operations
type GroupRoleRepository interface {
	base.Repository[*models.GroupRole]
	GetByGroupID(ctx context.Context, groupID string) ([]*models.GroupRole, error)
	GetByRoleID(ctx context.Context, roleID string) ([]*models.GroupRole, error)
	GetByGroupAndRole(ctx context.Context, groupID, roleID string) (*models.GroupRole, error)
	GetByOrganizationID(ctx context.Context, organizationID string, limit, offset int) ([]*models.GroupRole, error)
	GetByOrganizationAndGroupID(ctx context.Context, organizationID, groupID string, limit, offset int) ([]*models.GroupRole, error)
	ExistsByGroupAndRole(ctx context.Context, groupID, roleID string) (bool, error)
	DeactivateByGroupAndRole(ctx context.Context, groupID, roleID string) error
	GetEffectiveRolesForUser(ctx context.Context, organizationID, userID string) ([]*models.GroupRole, error)
	GetByGroupIDWithRoles(ctx context.Context, groupID string) ([]*models.GroupRole, error)
}

// GroupMembershipRepository interface for group membership operations
type GroupMembershipRepository interface {
	base.Repository[*models.GroupMembership]
	GetByGroupID(ctx context.Context, groupID string, limit, offset int) ([]*models.GroupMembership, error)
	GetByPrincipalID(ctx context.Context, principalID string, limit, offset int) ([]*models.GroupMembership, error)
	GetByGroupAndPrincipal(ctx context.Context, groupID, principalID string) (*models.GroupMembership, error)
	GetUserDirectGroups(ctx context.Context, orgID, userID string) ([]*models.Group, error)
	GetEffectiveMemberships(ctx context.Context, groupID string, limit, offset int) ([]*models.GroupMembership, error)
	GetUserGroupsInOrganization(ctx context.Context, orgID, userID string, limit, offset int) ([]*models.Group, error)
	CountUserGroupsInOrganization(ctx context.Context, orgID, userID string) (int64, error)
	ExistsByGroupAndPrincipal(ctx context.Context, groupID, principalID string) (bool, error)
	DeactivateMembership(ctx context.Context, groupID, principalID string) error
	ActivateMembership(ctx context.Context, groupID, principalID string) error
	GetGroupMembersWithDetails(ctx context.Context, groupID string, limit, offset int) ([]map[string]interface{}, error)
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

// TransformOptions defines options for response transformation
type TransformOptions struct {
	// Include flags for nested objects
	IncludeProfile  bool
	IncludeContacts bool
	IncludeRole     bool
	IncludeUser     bool
	IncludeAddress  bool

	// Exclusion flags
	ExcludeDeleted  bool
	ExcludeInactive bool
	OnlyActiveRoles bool

	// Field control
	MaskSensitiveData bool
	IncludeTimestamps bool
}

// ResponseTransformer interface for transforming models to standardized responses
type ResponseTransformer interface {
	// User transformations
	TransformUser(user *models.User, options TransformOptions) interface{}
	TransformUsers(users []models.User, options TransformOptions) []interface{}

	// Role transformations
	TransformRole(role *models.Role, options TransformOptions) interface{}
	TransformRoles(roles []models.Role, options TransformOptions) []interface{}

	// UserRole transformations
	TransformUserRole(userRole *models.UserRole, options TransformOptions) interface{}
	TransformUserRoles(userRoles []models.UserRole, options TransformOptions) []interface{}

	// Other entity transformations
	TransformOrganization(org *models.Organization, options TransformOptions) interface{}
	TransformOrganizations(orgs []models.Organization, options TransformOptions) []interface{}
	TransformGroup(group *models.Group, options TransformOptions) interface{}
	TransformGroups(groups []models.Group, options TransformOptions) []interface{}
	TransformContact(contact *models.Contact, options TransformOptions) interface{}
	TransformContacts(contacts []models.Contact, options TransformOptions) []interface{}
	TransformAddress(address *models.Address, options TransformOptions) interface{}
	TransformAddresses(addresses []models.Address, options TransformOptions) []interface{}
	TransformPermission(permission *models.Permission, options TransformOptions) interface{}
	TransformPermissions(permissions []models.Permission, options TransformOptions) []interface{}
}

// QueryParameterHandler interface for parsing query parameters
type QueryParameterHandler interface {
	ParseTransformOptions(c *gin.Context) TransformOptions
	ValidateQueryParameters(c *gin.Context) error
	GetDefaultOptions() TransformOptions
	GetPaginationParams(c *gin.Context) (limit, offset int, err error)
	GetSortParams(c *gin.Context) (sortBy, order string)
	GetSearchParam(c *gin.Context) string
	GetFilterParams(c *gin.Context) map[string]string
}

// ResponseValidator interface for validating response structures
type ResponseValidator interface {
	ValidateUserResponse(response interface{}) error
	ValidateRoleResponse(response interface{}) error
	ValidateUserRoleResponse(response interface{}) error
	ValidateResponseConsistency(responses []interface{}) error
	ValidateNoSensitiveData(response interface{}) error
}

// AuditService interface for audit logging operations
type AuditService interface {
	// Basic audit logging
	LogUserAction(ctx context.Context, userID, action, resource, resourceID string, details map[string]interface{})
	LogUserActionWithError(ctx context.Context, userID, action, resource, resourceID string, err error, details map[string]interface{})
	LogAPIAccess(ctx context.Context, userID, method, endpoint, ipAddress, userAgent string, success bool, err error)
	LogAccessDenied(ctx context.Context, userID, action, resource, resourceID, reason string)
	LogPermissionChange(ctx context.Context, userID, action, resource, resourceID, permission string, details map[string]interface{})
	LogRoleChange(ctx context.Context, userID, action, roleID string, details map[string]interface{})
	LogDataAccess(ctx context.Context, userID, action, resource, resourceID string, oldData, newData map[string]interface{})
	LogSecurityEvent(ctx context.Context, userID, action, resource string, success bool, details map[string]interface{})
	LogAuthenticationAttempt(ctx context.Context, userID, method, ipAddress, userAgent string, success bool, failureReason string)
	LogRoleOperation(ctx context.Context, actorUserID, targetUserID, roleID, operation string, success bool, details map[string]interface{})
	LogMPINOperation(ctx context.Context, userID, operation, ipAddress, userAgent string, success bool, failureReason string)
	LogUserLifecycleEvent(ctx context.Context, actorUserID, targetUserID, operation string, success bool, details map[string]interface{})
	LogSuspiciousActivity(ctx context.Context, userID, activityType, description, ipAddress, userAgent string, details map[string]interface{})
	LogRateLimitViolation(ctx context.Context, userID, endpoint, ipAddress, userAgent string, details map[string]interface{})
	LogSystemEvent(ctx context.Context, action, resource string, success bool, details map[string]interface{})

	// Organization-specific audit logging
	LogOrganizationOperation(ctx context.Context, userID, action, orgID, message string, success bool, details map[string]interface{})
	LogGroupOperation(ctx context.Context, userID, action, orgID, groupID, message string, success bool, details map[string]interface{})
	LogGroupMembershipChange(ctx context.Context, actorUserID, action, orgID, groupID, targetUserID, message string, success bool, details map[string]interface{})
	LogGroupRoleAssignment(ctx context.Context, actorUserID, action, orgID, groupID, roleID, message string, success bool, details map[string]interface{})
	LogHierarchyChange(ctx context.Context, userID, action, resourceType, resourceID, oldParentID, newParentID, message string, success bool, details map[string]interface{})
	LogOrganizationStructureChange(ctx context.Context, userID, action, orgID, resourceType, resourceID string, oldValues, newValues map[string]interface{}, success bool, message string)

	// Audit query operations
	QueryAuditLogs(ctx context.Context, query interface{}) (interface{}, error)
	QueryOrganizationAuditLogs(ctx context.Context, orgID string, query interface{}) (interface{}, error)
	GetUserAuditTrail(ctx context.Context, userID string, days int, page, perPage int) (interface{}, error)
	GetResourceAuditTrail(ctx context.Context, resource, resourceID string, days int, page, perPage int) (interface{}, error)
	GetOrganizationAuditTrail(ctx context.Context, orgID string, days int, page, perPage int) (interface{}, error)
	GetGroupAuditTrail(ctx context.Context, orgID, groupID string, days int, page, perPage int) (interface{}, error)
	GetSecurityEvents(ctx context.Context, days int, page, perPage int) (interface{}, error)

	// Audit integrity and management
	ValidateAuditLogIntegrity(ctx context.Context, auditLogID string) (bool, error)
	GetAuditStatistics(ctx context.Context, days int) (map[string]interface{}, error)
	ArchiveOldLogs(ctx context.Context, days int) error
}

// AuditRepository defines the interface for audit log repository operations
type AuditRepository interface {
	Create(ctx context.Context, auditLog *models.AuditLog) error
	GetByID(ctx context.Context, id string) (*models.AuditLog, error)
	Update(ctx context.Context, auditLog *models.AuditLog) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.AuditLog, error)
	ListByOrganization(ctx context.Context, orgID string, limit, offset int) ([]*models.AuditLog, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*models.AuditLog, error)
	ListByAction(ctx context.Context, action string, limit, offset int) ([]*models.AuditLog, error)
	ListByResourceType(ctx context.Context, resourceType string, limit, offset int) ([]*models.AuditLog, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*models.AuditLog, error)
	ListByTimeRange(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*models.AuditLog, error)
	ListByOrganizationAndTimeRange(ctx context.Context, orgID string, startTime, endTime time.Time, limit, offset int) ([]*models.AuditLog, error)
	ListByUserAndTimeRange(ctx context.Context, userID string, startTime, endTime time.Time, limit, offset int) ([]*models.AuditLog, error)
	ListByGroupAndTimeRange(ctx context.Context, orgID, groupID string, startTime, endTime time.Time, limit, offset int) ([]*models.AuditLog, error)
	CountByOrganization(ctx context.Context, orgID string) (int64, error)
	CountByUser(ctx context.Context, userID string) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	CountByTimeRange(ctx context.Context, startTime, endTime time.Time) (int64, error)
	CountByOrganizationAndTimeRange(ctx context.Context, orgID string, startTime, endTime time.Time) (int64, error)
	GetSecurityEvents(ctx context.Context, days int, limit, offset int) ([]*models.AuditLog, error)
	GetFailedOperations(ctx context.Context, days int, limit, offset int) ([]*models.AuditLog, error)
	ArchiveOldLogs(ctx context.Context, cutoffDate time.Time) (int64, error)
	ValidateIntegrity(ctx context.Context, auditLogID string) (*models.AuditLog, error)
	ListByResourceAndActions(ctx context.Context, resourceType, resourceID string, actions []string, limit, offset int) ([]*models.AuditLog, error)
}

// OrganizationRepository interface for organization data operations
type OrganizationRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, org *models.Organization) error
	GetByID(ctx context.Context, id string) (*models.Organization, error)
	Update(ctx context.Context, org *models.Organization) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.Organization, error)
	Count(ctx context.Context) (int64, error)
	Exists(ctx context.Context, id string) (bool, error)
	SoftDelete(ctx context.Context, id string, deletedBy string) error
	Restore(ctx context.Context, id string) error
	GetByName(ctx context.Context, name string) (*models.Organization, error)
	GetByType(ctx context.Context, orgType string, limit, offset int) ([]*models.Organization, error)
	ListActive(ctx context.Context, limit, offset int) ([]*models.Organization, error)
	CountActive(ctx context.Context) (int64, error)
	CountByType(ctx context.Context, orgType string) (int64, error)
	GetChildren(ctx context.Context, parentID string) ([]*models.Organization, error)
	GetActiveChildren(ctx context.Context, parentID string) ([]*models.Organization, error)
	GetParentHierarchy(ctx context.Context, orgID string) ([]*models.Organization, error)
	CountChildren(ctx context.Context, parentID string) (int64, error)
	CountGroups(ctx context.Context, orgID string) (int64, error)
	CountUsers(ctx context.Context, orgID string) (int64, error)
	HasActiveGroups(ctx context.Context, orgID string) (bool, error)
	Search(ctx context.Context, keyword string, limit, offset int) ([]*models.Organization, error)
	GetByStatus(ctx context.Context, isActive bool, limit, offset int) ([]*models.Organization, error)
	GetRootOrganizations(ctx context.Context, limit, offset int) ([]*models.Organization, error)
}

// UserRepositoryInterface interface for user data operations (renamed to avoid conflict)
type UserRepositoryInterface interface {
	// Basic CRUD operations
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
	Count(ctx context.Context) (int64, error)
	Exists(ctx context.Context, id string) (bool, error)
	SoftDelete(ctx context.Context, id string, deletedBy string) error
	Restore(ctx context.Context, id string) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByPhoneNumber(ctx context.Context, phoneNumber, countryCode string) (*models.User, error)
	GetByAadhaarNumber(ctx context.Context, aadhaarNumber string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByPhoneNumber(ctx context.Context, phoneNumber, countryCode string) (bool, error)
	ExistsByAadhaarNumber(ctx context.Context, aadhaarNumber string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ListActive(ctx context.Context, limit, offset int) ([]*models.User, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.User, error)
	GetWithAddress(ctx context.Context, userID string) (*models.User, error)
	GetWithProfile(ctx context.Context, userID string) (*models.User, error)
	UpdatePassword(ctx context.Context, userID, hashedPassword string) error
	UpdateLastLogin(ctx context.Context, userID string) error
	VerifyPassword(ctx context.Context, userID, password string) (bool, error)
}
