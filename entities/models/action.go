package models

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// Action represents an action in the AAA service that corresponds to SpiceDB permissions
type Action struct {
	*base.BaseModel
	Name        string `json:"name" gorm:"size:100;not null;unique"`
	Description string `json:"description" gorm:"type:text"`
	Category    string `json:"category" gorm:"size:50;not null"` // e.g., "user", "role", "system", "api", "database"
	IsActive    bool   `json:"is_active" gorm:"default:true"`
}

// Action constants matching SpiceDB schema permissions
const (
	// Basic CRUD actions
	ActionView   = "view"
	ActionEdit   = "edit"
	ActionDelete = "delete"
	ActionManage = "manage"
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"

	// User-specific actions
	ActionReadProfile     = "read_profile"
	ActionUpdateProfile   = "update_profile"
	ActionReadContacts    = "read_contacts"
	ActionUpdateContacts  = "update_contacts"
	ActionReadAddresses   = "read_addresses"
	ActionUpdateAddresses = "update_addresses"
	ActionManageTokens    = "manage_tokens"
	ActionValidateUser    = "validate_user"
	ActionSuspendUser     = "suspend_user"
	ActionBlockUser       = "block_user"

	// Role-specific actions
	ActionAssign            = "assign"
	ActionAssignPermissions = "assign_permissions"
	ActionRemovePermissions = "remove_permissions"
	ActionAssignUsers       = "assign_users"
	ActionRemoveUsers       = "remove_users"

	// Permission-specific actions
	ActionCreatePermission = "create_permission"
	ActionAssignToRoles    = "assign_to_roles"
	ActionRemoveFromRoles  = "remove_from_roles"

	// System-level actions
	ActionManageUsers       = "manage_users"
	ActionManageRoles       = "manage_roles"
	ActionManagePermissions = "manage_permissions"
	ActionViewAuditLogs     = "view_audit_logs"
	ActionSystemConfig      = "system_config"
	ActionBackupRestore     = "backup_restore"

	// Audit actions
	ActionExport  = "export"
	ActionReadAll = "read_all"

	// Temporary permission actions
	ActionExtend = "extend"
	ActionRevoke = "revoke"

	// Hierarchical actions
	ActionInheritFromParent   = "inherit_from_parent"
	ActionPropagateToChildren = "propagate_to_children"

	// API endpoint actions
	ActionGet     = "get"
	ActionPost    = "post"
	ActionPut     = "put"
	ActionPatch   = "patch"
	ActionHead    = "head"
	ActionOptions = "options"

	// Database operation actions
	ActionSelect      = "select"
	ActionInsert      = "insert"
	ActionCreateTable = "create_table"
	ActionDropTable   = "drop_table"
	ActionAlterTable  = "alter_table"
	ActionCreateIndex = "create_index"
	ActionDropIndex   = "drop_index"

	// Table-specific actions
	ActionReadAllRows = "read_all_rows"
	ActionReadOwnRows = "read_own_rows"
	ActionInsertRows  = "insert_rows"
	ActionUpdateRows  = "update_rows"
	ActionDeleteRows  = "delete_rows"

	// Database-specific actions
	ActionBackup  = "backup"
	ActionRestore = "restore"
	ActionMigrate = "migrate"

	// Special action for resource deletion
	ActionDeleteResource = "delete_resource"

	// Execution action
	ActionExecute = "execute"
)

// Action categories
const (
	CategoryUser     = "user"
	CategoryRole     = "role"
	CategorySystem   = "system"
	CategoryAPI      = "api"
	CategoryDatabase = "database"
	CategoryAudit    = "audit"
	CategoryGeneral  = "general"
)

// NewAction creates a new Action instance
func NewAction(name, description, category string) *Action {
	return &Action{
		BaseModel:   base.NewBaseModel("act", hash.Small),
		Name:        name,
		Description: description,
		Category:    category,
		IsActive:    true,
	}
}

// BeforeCreate is called before creating a new action
func (a *Action) BeforeCreate() error {
	return a.BaseModel.BeforeCreate()
}

// BeforeUpdate is called before updating an action
func (a *Action) BeforeUpdate() error {
	return a.BaseModel.BeforeUpdate()
}

// BeforeDelete is called before deleting an action
func (a *Action) BeforeDelete() error {
	return a.BaseModel.BeforeDelete()
}

// BeforeSoftDelete is called before soft deleting an action
func (a *Action) BeforeSoftDelete() error {
	return a.BaseModel.BeforeSoftDelete()
}

// GORM Hooks - These are for GORM compatibility
// BeforeCreateGORM is called by GORM before creating a new record
func (a *Action) BeforeCreateGORM(tx *gorm.DB) error {
	return a.BeforeCreate()
}

// BeforeUpdateGORM is called by GORM before updating an existing record
func (a *Action) BeforeUpdateGORM(tx *gorm.DB) error {
	return a.BeforeUpdate()
}

// BeforeDeleteGORM is called by GORM before hard deleting a record
func (a *Action) BeforeDeleteGORM(tx *gorm.DB) error {
	return a.BeforeDelete()
}

// GetTableIdentifier returns the table identifier for Action
func (a *Action) GetTableIdentifier() string {
	return "act"
}

// GetTableSize returns the table size for Action
func (a *Action) GetTableSize() hash.TableSize {
	return hash.Small
}

// IsCategory checks if the action belongs to a specific category
func (a *Action) IsCategory(category string) bool {
	return a.Category == category
}

// GetSpiceDBPermission returns the SpiceDB permission name for this action
func (a *Action) GetSpiceDBPermission() string {
	return a.Name
}
