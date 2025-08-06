package models

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// AuditLog represents an audit log entry in the AAA service
// This aligns with the SpiceDB audit_log resource type
type AuditLog struct {
	*base.BaseModel
	UserID       *string                `json:"user_id" gorm:"type:varchar(255);default:null"`
	Action       string                 `json:"action" gorm:"size:100;not null"`
	ResourceType string                 `json:"resource_type" gorm:"size:100;not null"` // e.g., "aaa/user", "aaa/role"
	ResourceID   *string                `json:"resource_id" gorm:"type:varchar(255);default:null"`
	IPAddress    string                 `json:"ip_address" gorm:"size:45"`
	UserAgent    string                 `json:"user_agent" gorm:"type:text"`
	Status       string                 `json:"status" gorm:"size:20;not null"` // success, failure, warning
	Message      string                 `json:"message" gorm:"type:text"`
	Details      map[string]interface{} `json:"details" gorm:"type:jsonb"`
	Timestamp    time.Time              `json:"timestamp" gorm:"not null;default:CURRENT_TIMESTAMP"`

	// Relationships
	User     *User     `json:"user" gorm:"foreignKey:UserID;references:ID"`
	Resource *Resource `json:"resource" gorm:"foreignKey:ResourceID;references:ID"`
}

// Audit log status constants
const (
	AuditStatusSuccess = "success"
	AuditStatusFailure = "failure"
	AuditStatusWarning = "warning"
)

// Audit log action constants matching SpiceDB schema and our actions
const (
	AuditActionLogin             = "login"
	AuditActionLogout            = "logout"
	AuditActionRegister          = "register"
	AuditActionCreateUser        = "create_user"
	AuditActionUpdateUser        = "update_user"
	AuditActionDeleteUser        = "delete_user"
	AuditActionValidateUser      = "validate_user"
	AuditActionSuspendUser       = "suspend_user"
	AuditActionBlockUser         = "block_user"
	AuditActionCreateRole        = "create_role"
	AuditActionUpdateRole        = "update_role"
	AuditActionDeleteRole        = "delete_role"
	AuditActionAssignRole        = "assign_role"
	AuditActionRemoveRole        = "remove_role"
	AuditActionCreatePermission  = "create_permission"
	AuditActionUpdatePermission  = "update_permission"
	AuditActionDeletePermission  = "delete_permission"
	AuditActionGrantPermission   = "grant_permission"
	AuditActionRevokePermission  = "revoke_permission"
	AuditActionCheckPermission   = "check_permission"
	AuditActionAccessDenied      = "access_denied"
	AuditActionDataAccess        = "data_access"
	AuditActionSecurityEvent     = "security_event"
	AuditActionSystemConfig      = "system_config"
	AuditActionBackup            = "backup"
	AuditActionRestore           = "restore"
	AuditActionAPICall           = "api_call"
	AuditActionDatabaseOperation = "database_operation"
)

// NewAuditLog creates a new AuditLog instance
func NewAuditLog(action, resourceType, status, message string) *AuditLog {
	return &AuditLog{
		BaseModel:    base.NewBaseModel("audit", hash.Medium),
		Action:       action,
		ResourceType: resourceType,
		Status:       status,
		Message:      message,
		Details:      make(map[string]interface{}),
		Timestamp:    time.Now(),
	}
}

// NewAuditLogWithUser creates a new AuditLog instance with a user
func NewAuditLogWithUser(userID, action, resourceType, status, message string) *AuditLog {
	auditLog := NewAuditLog(action, resourceType, status, message)
	auditLog.UserID = &userID
	return auditLog
}

// NewAuditLogWithResource creates a new AuditLog instance with a resource
func NewAuditLogWithResource(action, resourceType, resourceID, status, message string) *AuditLog {
	auditLog := NewAuditLog(action, resourceType, status, message)
	auditLog.ResourceID = &resourceID
	return auditLog
}

// NewAuditLogWithUserAndResource creates a new AuditLog instance with both user and resource
func NewAuditLogWithUserAndResource(userID, action, resourceType, resourceID, status, message string) *AuditLog {
	auditLog := NewAuditLog(action, resourceType, status, message)
	auditLog.UserID = &userID
	auditLog.ResourceID = &resourceID
	return auditLog
}

// BeforeCreate is called before creating a new audit log
func (al *AuditLog) BeforeCreate() error {
	if al.Timestamp.IsZero() {
		al.Timestamp = time.Now()
	}
	return al.BaseModel.BeforeCreate()
}

// BeforeUpdate is called before updating an audit log
func (al *AuditLog) BeforeUpdate() error {
	return al.BaseModel.BeforeUpdate()
}

// BeforeDelete is called before deleting an audit log
func (al *AuditLog) BeforeDelete() error {
	return al.BaseModel.BeforeDelete()
}

// BeforeSoftDelete is called before soft deleting an audit log
func (al *AuditLog) BeforeSoftDelete() error {
	return al.BaseModel.BeforeSoftDelete()
}

// GORM Hooks - These are for GORM compatibility
// BeforeCreateGORM is called by GORM before creating a new record
func (al *AuditLog) BeforeCreateGORM(tx *gorm.DB) error {
	return al.BeforeCreate()
}

// BeforeUpdateGORM is called by GORM before updating an existing record
func (al *AuditLog) BeforeUpdateGORM(tx *gorm.DB) error {
	return al.BeforeUpdate()
}

// BeforeDeleteGORM is called by GORM before hard deleting a record
func (al *AuditLog) BeforeDeleteGORM(tx *gorm.DB) error {
	return al.BeforeDelete()
}

// GetTableIdentifier returns the table identifier for AuditLog
func (al *AuditLog) GetTableIdentifier() string {
	return "audit"
}

// GetTableSize returns the table size for AuditLog
func (al *AuditLog) GetTableSize() hash.TableSize {
	return hash.Medium
}

// TableName specifies the table name for the AuditLog model
func (al *AuditLog) TableName() string {
	return "audit_logs"
}

// IsSuccess checks if the audit log status is success
func (al *AuditLog) IsSuccess() bool {
	return al.Status == AuditStatusSuccess
}

// IsFailure checks if the audit log status is failure
func (al *AuditLog) IsFailure() bool {
	return al.Status == AuditStatusFailure
}

// IsWarning checks if the audit log status is warning
func (al *AuditLog) IsWarning() bool {
	return al.Status == AuditStatusWarning
}

// AddDetail adds a key-value pair to the audit log details
func (al *AuditLog) AddDetail(key string, value interface{}) {
	if al.Details == nil {
		al.Details = make(map[string]interface{})
	}
	al.Details[key] = value
}

// GetDetail retrieves a value from the audit log details
func (al *AuditLog) GetDetail(key string) (interface{}, bool) {
	if al.Details == nil {
		return nil, false
	}
	value, exists := al.Details[key]
	return value, exists
}

// SetRequestDetails sets common HTTP request details
func (al *AuditLog) SetRequestDetails(method, path, ipAddress, userAgent string) {
	al.IPAddress = ipAddress
	al.UserAgent = userAgent
	al.AddDetail("http_method", method)
	al.AddDetail("http_path", path)
}

// SetErrorDetails sets error-related details
func (al *AuditLog) SetErrorDetails(errorMessage, errorCode string) {
	al.AddDetail("error_message", errorMessage)
	al.AddDetail("error_code", errorCode)
}

// GetSpiceDBResourceType returns the SpiceDB resource type for audit logs
func (al *AuditLog) GetSpiceDBResourceType() string {
	return ResourceTypeAuditLog
}

// GetSpiceDBObjectID returns the SpiceDB object ID for this audit log
func (al *AuditLog) GetSpiceDBObjectID() string {
	return al.ID
}
