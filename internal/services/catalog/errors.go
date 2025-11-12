package catalog

import "errors"

// Provider registration errors
var (
	ErrProviderAlreadyRegistered = errors.New("seed provider already registered")
	ErrProviderNotFound          = errors.New("seed provider not found")
	ErrInvalidServiceID          = errors.New("invalid service ID")
	ErrInvalidServiceName        = errors.New("invalid service name")
)

// Validation errors
var (
	ErrInvalidRoleName       = errors.New("invalid role name")
	ErrInvalidRoleScope      = errors.New("invalid role scope")
	ErrEmptyPermissions      = errors.New("role must have at least one permission")
	ErrInvalidResourceName   = errors.New("invalid resource name")
	ErrInvalidResourceType   = errors.New("invalid resource type")
	ErrInvalidActionName     = errors.New("invalid action name")
	ErrInvalidActionCategory = errors.New("invalid action category")
)

// Seeding errors
var (
	ErrSeedingFailed         = errors.New("seeding operation failed")
	ErrPartialSeedingSuccess = errors.New("partial seeding success")
	ErrRollbackFailed        = errors.New("rollback operation failed")
)
