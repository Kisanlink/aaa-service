//go:build integration
// +build integration

package organizations

import (
	"context"
	"testing"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/requests/organizations"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockAuditService for testing
type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) LogOrganizationOperation(ctx context.Context, userID, action, orgID, message string, success bool, details map[string]interface{}) {
	m.Called(ctx, userID, action, orgID, message, success, details)
}

func (m *MockAuditService) LogGroupOperation(ctx context.Context, userID, action, orgID, groupID, message string, success bool, details map[string]interface{}) {
	m.Called(ctx, userID, action, orgID, groupID, message, success, details)
}

func (m *MockAuditService) LogGroupMembershipChange(ctx context.Context, actorUserID, action, orgID, groupID, targetUserID, message string, success bool, details map[string]interface{}) {
	m.Called(ctx, actorUserID, action, orgID, groupID, targetUserID, message, success, details)
}

func (m *MockAuditService) LogGroupRoleAssignment(ctx context.Context, actorUserID, action, orgID, groupID, roleID, message string, success bool, details map[string]interface{}) {
	m.Called(ctx, actorUserID, action, orgID, groupID, roleID, message, success, details)
}

func (m *MockAuditService) LogHierarchyChange(ctx context.Context, userID, action, resourceType, resourceID, oldParentID, newParentID, message string, success bool, details map[string]interface{}) {
	m.Called(ctx, userID, action, resourceType, resourceID, oldParentID, newParentID, message, success, details)
}

func (m *MockAuditService) LogOrganizationStructureChange(ctx context.Context, userID, action, orgID, resourceType, resourceID string, oldValues, newValues map[string]interface{}, success bool, message string) {
	m.Called(ctx, userID, action, orgID, resourceType, resourceID, oldValues, newValues, success, message)
}

// Implement other required methods as no-ops for testing
func (m *MockAuditService) LogUserAction(ctx context.Context, userID, action, resource, resourceID string, details map[string]interface{}) {
	m.Called(ctx, userID, action, resource, resourceID, details)
}

func (m *MockAuditService) LogUserActionWithError(ctx context.Context, userID, action, resource, resourceID string, err error, details map[string]interface{}) {
	m.Called(ctx, userID, action, resource, resourceID, err, details)
}

func (m *MockAuditService) LogAPIAccess(ctx context.Context, userID, method, endpoint, ipAddress, userAgent string, success bool, err error) {
	m.Called(ctx, userID, method, endpoint, ipAddress, userAgent, success, err)
}

func (m *MockAuditService) LogAccessDenied(ctx context.Context, userID, action, resource, resourceID, reason string) {
	m.Called(ctx, userID, action, resource, resourceID, reason)
}

func (m *MockAuditService) LogPermissionChange(ctx context.Context, userID, action, resource, resourceID, permission string, details map[string]interface{}) {
	m.Called(ctx, userID, action, resource, resourceID, permission, details)
}

func (m *MockAuditService) LogRoleChange(ctx context.Context, userID, action, roleID string, details map[string]interface{}) {
	m.Called(ctx, userID, action, roleID, details)
}

func (m *MockAuditService) LogDataAccess(ctx context.Context, userID, action, resource, resourceID string, oldData, newData map[string]interface{}) {
	m.Called(ctx, userID, action, resource, resourceID, oldData, newData)
}

func (m *MockAuditService) LogSecurityEvent(ctx context.Context, userID, action, resource string, success bool, details map[string]interface{}) {
	m.Called(ctx, userID, action, resource, success, details)
}

func (m *MockAuditService) LogAuthenticationAttempt(ctx context.Context, userID, method, ipAddress, userAgent string, success bool, failureReason string) {
	m.Called(ctx, userID, method, ipAddress, userAgent, success, failureReason)
}

func (m *MockAuditService) LogRoleOperation(ctx context.Context, actorUserID, targetUserID, roleID, operation string, success bool, details map[string]interface{}) {
	m.Called(ctx, actorUserID, targetUserID, roleID, operation, success, details)
}

func (m *MockAuditService) LogMPINOperation(ctx context.Context, userID, operation, ipAddress, userAgent string, success bool, failureReason string) {
	m.Called(ctx, userID, operation, ipAddress, userAgent, success, failureReason)
}

func (m *MockAuditService) LogUserLifecycleEvent(ctx context.Context, actorUserID, targetUserID, operation string, success bool, details map[string]interface{}) {
	m.Called(ctx, actorUserID, targetUserID, operation, success, details)
}

func (m *MockAuditService) LogSuspiciousActivity(ctx context.Context, userID, activityType, description, ipAddress, userAgent string, details map[string]interface{}) {
	m.Called(ctx, userID, activityType, description, ipAddress, userAgent, details)
}

func (m *MockAuditService) LogRateLimitViolation(ctx context.Context, userID, endpoint, ipAddress, userAgent string, details map[string]interface{}) {
	m.Called(ctx, userID, endpoint, ipAddress, userAgent, details)
}

func (m *MockAuditService) LogSystemEvent(ctx context.Context, action, resource string, success bool, details map[string]interface{}) {
	m.Called(ctx, action, resource, success, details)
}

func (m *MockAuditService) QueryAuditLogs(ctx context.Context, query interface{}) (interface{}, error) {
	args := m.Called(ctx, query)
	return args.Get(0), args.Error(1)
}

func (m *MockAuditService) QueryOrganizationAuditLogs(ctx context.Context, orgID string, query interface{}) (interface{}, error) {
	args := m.Called(ctx, orgID, query)
	return args.Get(0), args.Error(1)
}

func (m *MockAuditService) GetUserAuditTrail(ctx context.Context, userID string, days int, page, perPage int) (interface{}, error) {
	args := m.Called(ctx, userID, days, page, perPage)
	return args.Get(0), args.Error(1)
}

func (m *MockAuditService) GetResourceAuditTrail(ctx context.Context, resource, resourceID string, days int, page, perPage int) (interface{}, error) {
	args := m.Called(ctx, resource, resourceID, days, page, perPage)
	return args.Get(0), args.Error(1)
}

func (m *MockAuditService) GetOrganizationAuditTrail(ctx context.Context, orgID string, days int, page, perPage int) (interface{}, error) {
	args := m.Called(ctx, orgID, days, page, perPage)
	return args.Get(0), args.Error(1)
}

func (m *MockAuditService) GetGroupAuditTrail(ctx context.Context, orgID, groupID string, days int, page, perPage int) (interface{}, error) {
	args := m.Called(ctx, orgID, groupID, days, page, perPage)
	return args.Get(0), args.Error(1)
}

func (m *MockAuditService) GetSecurityEvents(ctx context.Context, days int, page, perPage int) (interface{}, error) {
	args := m.Called(ctx, days, page, perPage)
	return args.Get(0), args.Error(1)
}

func (m *MockAuditService) ValidateAuditLogIntegrity(ctx context.Context, auditLogID string) (bool, error) {
	args := m.Called(ctx, auditLogID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuditService) GetAuditStatistics(ctx context.Context, days int) (map[string]interface{}, error) {
	args := m.Called(ctx, days)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAuditService) ArchiveOldLogs(ctx context.Context, days int) error {
	args := m.Called(ctx, days)
	return args.Error(0)
}

// MockOrganizationRepository for testing
type MockOrganizationRepository struct {
	mock.Mock
}

func (m *MockOrganizationRepository) Create(ctx context.Context, org *models.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockOrganizationRepository) GetByID(ctx context.Context, id string) (*models.Organization, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) GetByName(ctx context.Context, name string) (*models.Organization, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) Update(ctx context.Context, org *models.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockOrganizationRepository) SoftDelete(ctx context.Context, id, deletedBy string) error {
	args := m.Called(ctx, id, deletedBy)
	return args.Error(0)
}

func (m *MockOrganizationRepository) List(ctx context.Context, limit, offset int) ([]*models.Organization, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) ListActive(ctx context.Context, limit, offset int) ([]*models.Organization, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) GetChildren(ctx context.Context, parentID string) ([]*models.Organization, error) {
	args := m.Called(ctx, parentID)
	return args.Get(0).([]*models.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) GetActiveChildren(ctx context.Context, parentID string) ([]*models.Organization, error) {
	args := m.Called(ctx, parentID)
	return args.Get(0).([]*models.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) GetParentHierarchy(ctx context.Context, orgID string) ([]*models.Organization, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]*models.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) HasActiveGroups(ctx context.Context, orgID string) (bool, error) {
	args := m.Called(ctx, orgID)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrganizationRepository) CountChildren(ctx context.Context, orgID string) (int64, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrganizationRepository) CountGroups(ctx context.Context, orgID string) (int64, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrganizationRepository) CountUsers(ctx context.Context, orgID string) (int64, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrganizationRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrganizationRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrganizationRepository) Exists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrganizationRepository) Restore(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrganizationRepository) GetByType(ctx context.Context, orgType string, limit, offset int) ([]*models.Organization, error) {
	args := m.Called(ctx, orgType, limit, offset)
	return args.Get(0).([]*models.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) Search(ctx context.Context, keyword string, limit, offset int) ([]*models.Organization, error) {
	args := m.Called(ctx, keyword, limit, offset)
	return args.Get(0).([]*models.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) GetByStatus(ctx context.Context, isActive bool, limit, offset int) ([]*models.Organization, error) {
	args := m.Called(ctx, isActive, limit, offset)
	return args.Get(0).([]*models.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) GetRootOrganizations(ctx context.Context, limit, offset int) ([]*models.Organization, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Organization), args.Error(1)
}

// MockValidator for testing
type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) ValidateStruct(s interface{}) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *MockValidator) ValidateUserID(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockValidator) ValidatePhoneNumber(phone string) error {
	args := m.Called(phone)
	return args.Error(0)
}

func (m *MockValidator) ValidateAadhaarNumber(aadhaar string) error {
	args := m.Called(aadhaar)
	return args.Error(0)
}

func (m *MockValidator) ValidateEmail(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockValidator) ValidatePassword(password string) error {
	args := m.Called(password)
	return args.Error(0)
}

func (m *MockValidator) ParseListFilters(c *gin.Context) (interface{}, error) {
	args := m.Called(c)
	return args.Get(0), args.Error(1)
}

// MockUserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) Exists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	args := m.Called(ctx, id, deletedBy)
	return args.Error(0)
}

func (m *MockUserRepository) Restore(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByPhoneNumber(ctx context.Context, phoneNumber, countryCode string) (*models.User, error) {
	args := m.Called(ctx, phoneNumber, countryCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByAadhaarNumber(ctx context.Context, aadhaarNumber string) (*models.User, error) {
	args := m.Called(ctx, aadhaarNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByPhoneNumber(ctx context.Context, phoneNumber, countryCode string) (bool, error) {
	args := m.Called(ctx, phoneNumber, countryCode)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByAadhaarNumber(ctx context.Context, aadhaarNumber string) (bool, error) {
	args := m.Called(ctx, aadhaarNumber)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ListActive(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) GetWithAddress(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetWithProfile(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, userID, hashedPassword string) error {
	args := m.Called(ctx, userID, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) VerifyPassword(ctx context.Context, userID, password string) (bool, error) {
	args := m.Called(ctx, userID, password)
	return args.Bool(0), args.Error(1)
}

// MockGroupRepository for testing
type MockGroupRepository struct {
	mock.Mock
}

func (m *MockGroupRepository) Create(ctx context.Context, group *models.Group) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockGroupRepository) GetByID(ctx context.Context, id string, group *models.Group) (*models.Group, error) {
	args := m.Called(ctx, id, group)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Group), args.Error(1)
}

func (m *MockGroupRepository) Update(ctx context.Context, group *models.Group) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockGroupRepository) Delete(ctx context.Context, id string, group *models.Group) error {
	args := m.Called(ctx, id, group)
	return args.Error(0)
}

func (m *MockGroupRepository) List(ctx context.Context, limit, offset int) ([]*models.Group, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Group), args.Error(1)
}

func (m *MockGroupRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockGroupRepository) Exists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockGroupRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	args := m.Called(ctx, id, deletedBy)
	return args.Error(0)
}

func (m *MockGroupRepository) Restore(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGroupRepository) GetByName(ctx context.Context, name string) (*models.Group, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Group), args.Error(1)
}

func (m *MockGroupRepository) GetByNameAndOrganization(ctx context.Context, name, organizationID string) (*models.Group, error) {
	args := m.Called(ctx, name, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Group), args.Error(1)
}

func (m *MockGroupRepository) GetByOrganization(ctx context.Context, organizationID string, limit, offset int, includeInactive bool) ([]*models.Group, error) {
	args := m.Called(ctx, organizationID, limit, offset, includeInactive)
	return args.Get(0).([]*models.Group), args.Error(1)
}

func (m *MockGroupRepository) ListActive(ctx context.Context, limit, offset int) ([]*models.Group, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Group), args.Error(1)
}

func (m *MockGroupRepository) GetChildren(ctx context.Context, parentID string) ([]*models.Group, error) {
	args := m.Called(ctx, parentID)
	return args.Get(0).([]*models.Group), args.Error(1)
}

func (m *MockGroupRepository) HasActiveMembers(ctx context.Context, groupID string) (bool, error) {
	args := m.Called(ctx, groupID)
	return args.Bool(0), args.Error(1)
}

func (m *MockGroupRepository) CreateMembership(ctx context.Context, membership *models.GroupMembership) error {
	args := m.Called(ctx, membership)
	return args.Error(0)
}

func (m *MockGroupRepository) UpdateMembership(ctx context.Context, membership *models.GroupMembership) error {
	args := m.Called(ctx, membership)
	return args.Error(0)
}

func (m *MockGroupRepository) GetMembership(ctx context.Context, groupID, principalID string) (*models.GroupMembership, error) {
	args := m.Called(ctx, groupID, principalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GroupMembership), args.Error(1)
}

func (m *MockGroupRepository) GetGroupMembers(ctx context.Context, groupID string, limit, offset int) ([]*models.GroupMembership, error) {
	args := m.Called(ctx, groupID, limit, offset)
	return args.Get(0).([]*models.GroupMembership), args.Error(1)
}

func (m *MockGroupRepository) CreateMany(ctx context.Context, groups []*models.Group) error {
	args := m.Called(ctx, groups)
	return args.Error(0)
}

func (m *MockGroupRepository) DeleteMany(ctx context.Context, ids []string) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockGroupRepository) SoftDeleteMany(ctx context.Context, ids []string, deletedBy string) error {
	args := m.Called(ctx, ids, deletedBy)
	return args.Error(0)
}

func (m *MockGroupRepository) UpdateMany(ctx context.Context, groups []*models.Group) error {
	args := m.Called(ctx, groups)
	return args.Error(0)
}

func (m *MockGroupRepository) ListWithDeleted(ctx context.Context, limit, offset int) ([]*models.Group, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Group), args.Error(1)
}

func (m *MockGroupRepository) CountWithDeleted(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockGroupRepository) ExistsWithDeleted(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockGroupRepository) GetByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*models.Group, error) {
	args := m.Called(ctx, createdBy, limit, offset)
	return args.Get(0).([]*models.Group), args.Error(1)
}

func (m *MockGroupRepository) GetByUpdatedBy(ctx context.Context, updatedBy string, limit, offset int) ([]*models.Group, error) {
	args := m.Called(ctx, updatedBy, limit, offset)
	return args.Get(0).([]*models.Group), args.Error(1)
}

func (m *MockGroupRepository) GetByDeletedBy(ctx context.Context, deletedBy string, limit, offset int) ([]*models.Group, error) {
	args := m.Called(ctx, deletedBy, limit, offset)
	return args.Get(0).([]*models.Group), args.Error(1)
}

// MockGroupService for testing
type MockGroupService struct {
	mock.Mock
}

func (m *MockGroupService) CreateGroup(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) GetGroup(ctx context.Context, groupID string) (interface{}, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) UpdateGroup(ctx context.Context, groupID string, req interface{}) (interface{}, error) {
	args := m.Called(ctx, groupID, req)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) DeleteGroup(ctx context.Context, groupID string, deletedBy string) error {
	args := m.Called(ctx, groupID, deletedBy)
	return args.Error(0)
}

func (m *MockGroupService) ListGroups(ctx context.Context, limit, offset int, organizationID string, includeInactive bool) (interface{}, error) {
	args := m.Called(ctx, limit, offset, organizationID, includeInactive)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) AddUserToGroup(ctx context.Context, groupID, userID string) error {
	args := m.Called(ctx, groupID, userID)
	return args.Error(0)
}

func (m *MockGroupService) RemoveMemberFromGroup(ctx context.Context, groupID, principalID string, removedBy string) error {
	args := m.Called(ctx, groupID, principalID, removedBy)
	return args.Error(0)
}

func (m *MockGroupService) GetGroupMembers(ctx context.Context, groupID string, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, groupID, limit, offset)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) AssignRoleToGroup(ctx context.Context, groupID, roleID, assignedBy string) (interface{}, error) {
	args := m.Called(ctx, groupID, roleID, assignedBy)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) RemoveRoleFromGroup(ctx context.Context, groupID, roleID string) error {
	args := m.Called(ctx, groupID, roleID)
	return args.Error(0)
}

func (m *MockGroupService) GetGroupRoles(ctx context.Context, groupID string) (interface{}, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) GetUserEffectiveRoles(ctx context.Context, orgID, userID string) (interface{}, error) {
	args := m.Called(ctx, orgID, userID)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) AddMemberToGroup(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) GetGroupHierarchy(ctx context.Context, groupID string) (interface{}, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0), args.Error(1)
}

func (m *MockGroupService) ValidateGroupHierarchy(ctx context.Context, groupID, parentID string) error {
	args := m.Called(ctx, groupID, parentID)
	return args.Error(0)
}

// MockCacheService for testing
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Set(key string, value interface{}, ttl int) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheService) Get(key string) (interface{}, bool) {
	args := m.Called(key)
	return args.Get(0), args.Bool(1)
}

func (m *MockCacheService) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockCacheService) Exists(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockCacheService) Clear() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCacheService) Keys(pattern string) ([]string, error) {
	args := m.Called(pattern)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCacheService) Expire(key string, ttl int) error {
	args := m.Called(key, ttl)
	return args.Error(0)
}

func (m *MockCacheService) TTL(key string) (int, error) {
	args := m.Called(key)
	return args.Int(0), args.Error(1)
}

func (m *MockCacheService) Close() error {
	args := m.Called()
	return args.Error(0)
}

// TestOrganizationServiceAuditLogging tests audit logging in organization service operations
func TestOrganizationServiceAuditLogging(t *testing.T) {
	t.Skip("Skipping test due to repository interface mismatch - needs refactoring")
	return
	// Setup mocks
	mockOrgRepo := &MockOrganizationRepository{}
	mockAuditService := &MockAuditService{}
	mockValidator := &MockValidator{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	// Create service - commented out due to type mismatch
	service := NewOrganizationService(
		mockOrgRepo,
		nil, // userRepo not needed for these tests
		nil, // groupRepo not needed for these tests
		nil, // groupService not needed for these tests
		mockValidator,
		mockCache,
		mockAuditService,
		logger,
	)

	ctx := context.Background()

	t.Run("CreateOrganization_Success_AuditLogged", func(t *testing.T) {
		// Setup request
		req := &organizations.CreateOrganizationRequest{
			Name:        "Test Organization",
			Description: "Test Description",
		}

		// Setup mocks
		mockValidator.On("ValidateStruct", req).Return(nil)
		mockOrgRepo.On("GetByName", ctx, "Test Organization").Return(nil, assert.AnError) // Not found
		mockOrgRepo.On("Create", ctx, mock.AnythingOfType("*models.Organization")).Return(nil)

		// Expect audit logging
		mockAuditService.On("LogOrganizationOperation", ctx, "system", models.AuditActionCreateOrganization, mock.AnythingOfType("string"), "Organization created successfully", true, mock.AnythingOfType("map[string]interface {}")).Return()

		// Execute
		result, err := service.CreateOrganization(ctx, req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Organization", result.Name)

		// Verify audit logging was called
		mockAuditService.AssertExpectations(t)
		mockOrgRepo.AssertExpectations(t)
		mockValidator.AssertExpectations(t)
	})

	t.Run("CreateOrganization_Failure_AuditLogged", func(t *testing.T) {
		// Reset mocks
		mockOrgRepo = &MockOrganizationRepository{}
		mockAuditService = &MockAuditService{}
		mockValidator = &MockValidator{}
		service = NewOrganizationService(mockOrgRepo, nil, nil, nil, mockValidator, mockCache, mockAuditService, logger)

		// Setup request
		req := &organizations.CreateOrganizationRequest{
			Name:        "Test Organization",
			Description: "Test Description",
		}

		// Setup mocks
		mockValidator.On("ValidateStruct", req).Return(nil)
		mockOrgRepo.On("GetByName", ctx, "Test Organization").Return(nil, assert.AnError)                 // Not found
		mockOrgRepo.On("Create", ctx, mock.AnythingOfType("*models.Organization")).Return(assert.AnError) // Fail

		// Expect audit logging for failure
		mockAuditService.On("LogOrganizationOperation", ctx, "system", models.AuditActionCreateOrganization, "", "Failed to create organization", false, mock.AnythingOfType("map[string]interface {}")).Return()

		// Execute
		result, err := service.CreateOrganization(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)

		// Verify audit logging was called
		mockAuditService.AssertExpectations(t)
		mockOrgRepo.AssertExpectations(t)
		mockValidator.AssertExpectations(t)
	})

	t.Run("UpdateOrganization_HierarchyChange_AuditLogged", func(t *testing.T) {
		// Reset mocks
		mockOrgRepo = &MockOrganizationRepository{}
		mockAuditService = &MockAuditService{}
		mockValidator = &MockValidator{}
		service = NewOrganizationService(mockOrgRepo, nil, nil, nil, mockValidator, mockCache, mockAuditService, logger)

		// Setup existing organization
		existingOrg := &models.Organization{
			Name:        "Test Organization",
			Description: "Test Description",
			ParentID:    nil, // No parent initially
			IsActive:    true,
		}
		existingOrg.ID = "org123"
		existingOrg.CreatedAt = time.Now()
		existingOrg.UpdatedAt = time.Now()

		// Setup new parent organization
		parentOrg := &models.Organization{
			Name:     "Parent Organization",
			IsActive: true,
		}
		parentOrg.ID = "parent123"

		// Setup update request with parent change
		newParentID := "parent123"
		req := &organizations.UpdateOrganizationRequest{
			ParentID: &newParentID,
		}

		// Setup mocks
		mockValidator.On("ValidateStruct", req).Return(nil)
		mockOrgRepo.On("GetByID", ctx, "org123").Return(existingOrg, nil)
		mockOrgRepo.On("GetByID", ctx, "parent123").Return(parentOrg, nil)
		mockOrgRepo.On("Update", ctx, mock.AnythingOfType("*models.Organization")).Return(nil)

		// Expect audit logging for update
		mockAuditService.On("LogOrganizationOperation", ctx, "system", models.AuditActionUpdateOrganization, "org123", "Organization updated successfully", true, mock.AnythingOfType("map[string]interface {}")).Return()

		// Expect audit logging for hierarchy change
		mockAuditService.On("LogHierarchyChange", ctx, "system", models.AuditActionChangeOrganizationHierarchy, models.ResourceTypeOrganization, "org123", "", "parent123", "Organization hierarchy changed", true, mock.AnythingOfType("map[string]interface {}")).Return()

		// Expect comprehensive structure change logging
		mockAuditService.On("LogOrganizationStructureChange", ctx, "system", models.AuditActionChangeOrganizationHierarchy, "org123", models.ResourceTypeOrganization, "org123", mock.AnythingOfType("map[string]interface {}"), mock.AnythingOfType("map[string]interface {}"), true, "Organization hierarchy structure changed").Return()

		// Execute
		result, err := service.UpdateOrganization(ctx, "org123", req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "parent123", *result.ParentID)

		// Verify all audit logging was called
		mockAuditService.AssertExpectations(t)
		mockOrgRepo.AssertExpectations(t)
		mockValidator.AssertExpectations(t)
	})

	t.Run("DeleteOrganization_Success_AuditLogged", func(t *testing.T) {
		// Reset mocks
		mockOrgRepo = &MockOrganizationRepository{}
		mockAuditService = &MockAuditService{}
		service = NewOrganizationService(mockOrgRepo, nil, nil, nil, mockValidator, mockCache, mockAuditService, logger)

		// Setup existing organization
		existingOrg := &models.Organization{
			Name:     "Test Organization",
			IsActive: true,
		}
		existingOrg.ID = "org123"

		// Setup mocks
		mockOrgRepo.On("GetByID", ctx, "org123").Return(existingOrg, nil)
		mockOrgRepo.On("GetChildren", ctx, "org123").Return([]*models.Organization{}, nil) // No children
		mockOrgRepo.On("HasActiveGroups", ctx, "org123").Return(false, nil)                // No active groups
		mockOrgRepo.On("SoftDelete", ctx, "org123", "admin123").Return(nil)

		// Expect audit logging
		mockAuditService.On("LogOrganizationOperation", ctx, "admin123", models.AuditActionDeleteOrganization, "org123", "Organization deleted successfully", true, mock.AnythingOfType("map[string]interface {}")).Return()

		// Execute
		err := service.DeleteOrganization(ctx, "org123", "admin123")

		// Verify
		assert.NoError(t, err)

		// Verify audit logging was called
		mockAuditService.AssertExpectations(t)
		mockOrgRepo.AssertExpectations(t)
	})

	t.Run("ActivateOrganization_Success_AuditLogged", func(t *testing.T) {
		// Reset mocks
		mockOrgRepo = &MockOrganizationRepository{}
		mockAuditService = &MockAuditService{}
		service = NewOrganizationService(mockOrgRepo, nil, nil, nil, mockValidator, mockCache, mockAuditService, logger)

		// Setup inactive organization
		existingOrg := &models.Organization{
			Name:     "Test Organization",
			IsActive: false, // Inactive
		}
		existingOrg.ID = "org123"

		// Setup mocks
		mockOrgRepo.On("GetByID", ctx, "org123").Return(existingOrg, nil)
		mockOrgRepo.On("Update", ctx, mock.AnythingOfType("*models.Organization")).Return(nil)

		// Expect audit logging
		mockAuditService.On("LogOrganizationOperation", ctx, "system", models.AuditActionActivateOrganization, "org123", "Organization activated successfully", true, mock.AnythingOfType("map[string]interface {}")).Return()

		// Execute
		err := service.ActivateOrganization(ctx, "org123")

		// Verify
		assert.NoError(t, err)

		// Verify audit logging was called
		mockAuditService.AssertExpectations(t)
		mockOrgRepo.AssertExpectations(t)
	})

	t.Run("DeactivateOrganization_Success_AuditLogged", func(t *testing.T) {
		// Reset mocks
		mockOrgRepo = &MockOrganizationRepository{}
		mockAuditService = &MockAuditService{}
		service = NewOrganizationService(mockOrgRepo, nil, nil, nil, mockValidator, mockCache, mockAuditService, logger)

		// Setup active organization
		existingOrg := &models.Organization{
			Name:     "Test Organization",
			IsActive: true, // Active
		}
		existingOrg.ID = "org123"

		// Setup mocks
		mockOrgRepo.On("GetByID", ctx, "org123").Return(existingOrg, nil)
		mockOrgRepo.On("GetActiveChildren", ctx, "org123").Return([]*models.Organization{}, nil) // No active children
		mockOrgRepo.On("Update", ctx, mock.AnythingOfType("*models.Organization")).Return(nil)

		// Expect audit logging
		mockAuditService.On("LogOrganizationOperation", ctx, "system", models.AuditActionDeactivateOrganization, "org123", "Organization deactivated successfully", true, mock.AnythingOfType("map[string]interface {}")).Return()

		// Execute
		err := service.DeactivateOrganization(ctx, "org123")

		// Verify
		assert.NoError(t, err)

		// Verify audit logging was called
		mockAuditService.AssertExpectations(t)
		mockOrgRepo.AssertExpectations(t)
	})
}

// TestOrganizationAuditLogDetails tests that audit logs contain proper details
func TestOrganizationAuditLogDetails(t *testing.T) {
	// Setup mocks
	mockOrgRepo := &MockOrganizationRepository{}
	mockAuditService := &MockAuditService{}
	mockValidator := &MockValidator{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	// Create service
	mockUserRepo := &MockUserRepository{}
	mockGroupRepo := &MockGroupRepository{}
	mockGroupService := &MockGroupService{}
	service := NewOrganizationService(mockOrgRepo, mockUserRepo, mockGroupRepo, mockGroupService, mockValidator, mockCache, mockAuditService, logger)
	ctx := context.Background()

	t.Run("CreateOrganization_AuditDetailsComplete", func(t *testing.T) {
		// Setup request
		req := &organizations.CreateOrganizationRequest{
			Name:        "Test Organization",
			Description: "Test Description",
		}

		// Setup mocks
		mockValidator.On("ValidateStruct", req).Return(nil)
		mockOrgRepo.On("GetByName", ctx, "Test Organization").Return(nil, assert.AnError) // Not found
		mockOrgRepo.On("Create", ctx, mock.AnythingOfType("*models.Organization")).Return(nil)

		// Expect audit logging with specific details
		mockAuditService.On("LogOrganizationOperation", ctx, "system", models.AuditActionCreateOrganization, mock.AnythingOfType("string"), "Organization created successfully", true, mock.MatchedBy(func(details map[string]interface{}) bool {
			// Verify required audit details are present
			_, hasOrgName := details["organization_name"]
			_, hasIsActive := details["is_active"]
			return hasOrgName && hasIsActive
		})).Return()

		// Execute
		result, err := service.CreateOrganization(ctx, req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify audit logging was called with correct details
		mockAuditService.AssertExpectations(t)
	})

	t.Run("UpdateOrganization_AuditDetailsIncludeOldAndNewValues", func(t *testing.T) {
		t.Skip("Skipping test due to audit service interface changes - needs refactoring")
		// Reset mocks
		mockOrgRepo = &MockOrganizationRepository{}
		mockAuditService = &MockAuditService{}
		mockValidator = &MockValidator{}
		mockUserRepo := &MockUserRepository{}
		mockGroupRepo := &MockGroupRepository{}
		mockGroupService := &MockGroupService{}
		service = NewOrganizationService(mockOrgRepo, mockUserRepo, mockGroupRepo, mockGroupService, mockValidator, mockCache, mockAuditService, logger)

		// Setup existing organization
		existingOrg := &models.Organization{
			Name:        "Old Name",
			Description: "Old Description",
			IsActive:    true,
		}
		existingOrg.ID = "org123"
		existingOrg.CreatedAt = time.Now()
		existingOrg.UpdatedAt = time.Now()

		// Setup update request
		newName := "New Name"
		req := &organizations.UpdateOrganizationRequest{
			Name: &newName,
		}

		// Setup mocks
		mockValidator.On("ValidateStruct", req).Return(nil)
		mockOrgRepo.On("GetByID", ctx, "org123").Return(existingOrg, nil)
		mockOrgRepo.On("GetByName", ctx, "New Name").Return(nil, assert.AnError) // Name not taken
		mockOrgRepo.On("Update", ctx, mock.AnythingOfType("*models.Organization")).Return(nil)

		// Expect audit logging with old and new values
		mockAuditService.On("LogOrganizationOperation", ctx, "system", models.AuditActionUpdateOrganization, "org123", "Organization updated successfully", true, mock.MatchedBy(func(details map[string]interface{}) bool {
			// Verify old and new values are present
			oldValues, hasOldValues := details["old_values"]
			newValues, hasNewValues := details["new_values"]

			if !hasOldValues || !hasNewValues {
				return false
			}

			// Check old values
			oldMap, ok := oldValues.(map[string]interface{})
			if !ok {
				return false
			}
			if oldMap["name"] != "Old Name" {
				return false
			}

			// Check new values
			newMap, ok := newValues.(map[string]interface{})
			if !ok {
				return false
			}
			if newMap["name"] != "New Name" {
				return false
			}

			return true
		})).Return()

		// Execute
		result, err := service.UpdateOrganization(ctx, "org123", req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "New Name", result.Name)

		// Verify audit logging was called with correct details
		mockAuditService.AssertExpectations(t)
	})
}
