package migrations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// StaticAction represents a built-in action that all services support
type StaticAction struct {
	Name        string
	Description string
	Metadata    map[string]string
}

// GetStaticActions returns all built-in actions
func GetStaticActions() []StaticAction {
	return []StaticAction{
		// Basic CRUD operations
		{
			Name:        "create",
			Description: "Create a new resource",
			Metadata:    map[string]string{"category": "crud", "requires_write": "true"},
		},
		{
			Name:        "read",
			Description: "Read/view a resource",
			Metadata:    map[string]string{"category": "crud", "requires_write": "false"},
		},
		{
			Name:        "view",
			Description: "View a resource (alias for read)",
			Metadata:    map[string]string{"category": "crud", "requires_write": "false"},
		},
		{
			Name:        "update",
			Description: "Update/modify a resource",
			Metadata:    map[string]string{"category": "crud", "requires_write": "true"},
		},
		{
			Name:        "edit",
			Description: "Edit a resource (alias for update)",
			Metadata:    map[string]string{"category": "crud", "requires_write": "true"},
		},
		{
			Name:        "delete",
			Description: "Delete a resource",
			Metadata:    map[string]string{"category": "crud", "requires_write": "true"},
		},
		{
			Name:        "list",
			Description: "List multiple resources",
			Metadata:    map[string]string{"category": "crud", "requires_write": "false"},
		},

		// Administrative actions
		{
			Name:        "manage",
			Description: "Full management access to a resource",
			Metadata:    map[string]string{"category": "admin", "requires_admin": "true"},
		},
		{
			Name:        "admin",
			Description: "Administrative access to a resource",
			Metadata:    map[string]string{"category": "admin", "requires_admin": "true"},
		},
		{
			Name:        "assign",
			Description: "Assign roles or permissions",
			Metadata:    map[string]string{"category": "admin", "requires_admin": "true"},
		},
		{
			Name:        "unassign",
			Description: "Remove roles or permissions",
			Metadata:    map[string]string{"category": "admin", "requires_admin": "true"},
		},
		{
			Name:        "grant",
			Description: "Grant access to a resource",
			Metadata:    map[string]string{"category": "admin", "requires_admin": "true"},
		},
		{
			Name:        "revoke",
			Description: "Revoke access to a resource",
			Metadata:    map[string]string{"category": "admin", "requires_admin": "true"},
		},

		// Ownership actions
		{
			Name:        "own",
			Description: "Own a resource",
			Metadata:    map[string]string{"category": "ownership", "requires_admin": "true"},
		},
		{
			Name:        "transfer",
			Description: "Transfer ownership of a resource",
			Metadata:    map[string]string{"category": "ownership", "requires_admin": "true"},
		},
		{
			Name:        "share",
			Description: "Share a resource with others",
			Metadata:    map[string]string{"category": "ownership", "requires_write": "true"},
		},

		// Data operations
		{
			Name:        "export",
			Description: "Export data from a resource",
			Metadata:    map[string]string{"category": "data", "requires_write": "false"},
		},
		{
			Name:        "import",
			Description: "Import data into a resource",
			Metadata:    map[string]string{"category": "data", "requires_write": "true"},
		},
		{
			Name:        "backup",
			Description: "Backup a resource",
			Metadata:    map[string]string{"category": "data", "requires_admin": "true"},
		},
		{
			Name:        "restore",
			Description: "Restore a resource from backup",
			Metadata:    map[string]string{"category": "data", "requires_admin": "true"},
		},

		// API/Service actions
		{
			Name:        "execute",
			Description: "Execute an operation",
			Metadata:    map[string]string{"category": "api", "requires_write": "true"},
		},
		{
			Name:        "invoke",
			Description: "Invoke a service or function",
			Metadata:    map[string]string{"category": "api", "requires_write": "true"},
		},
		{
			Name:        "call",
			Description: "Call an API endpoint",
			Metadata:    map[string]string{"category": "api", "requires_write": "false"},
		},

		// Database-specific actions
		{
			Name:        "select",
			Description: "Select data from database",
			Metadata:    map[string]string{"category": "database", "requires_write": "false"},
		},
		{
			Name:        "insert",
			Description: "Insert data into database",
			Metadata:    map[string]string{"category": "database", "requires_write": "true"},
		},
		{
			Name:        "update_rows",
			Description: "Update rows in database",
			Metadata:    map[string]string{"category": "database", "requires_write": "true"},
		},
		{
			Name:        "delete_rows",
			Description: "Delete rows from database",
			Metadata:    map[string]string{"category": "database", "requires_write": "true"},
		},
		{
			Name:        "truncate",
			Description: "Truncate a database table",
			Metadata:    map[string]string{"category": "database", "requires_admin": "true"},
		},

		// Audit/Monitoring actions
		{
			Name:        "audit",
			Description: "View audit logs",
			Metadata:    map[string]string{"category": "audit", "requires_admin": "true"},
		},
		{
			Name:        "monitor",
			Description: "Monitor resource activity",
			Metadata:    map[string]string{"category": "audit", "requires_admin": "true"},
		},
		{
			Name:        "inspect",
			Description: "Inspect resource details",
			Metadata:    map[string]string{"category": "audit", "requires_write": "false"},
		},

		// Workflow actions
		{
			Name:        "approve",
			Description: "Approve a request or change",
			Metadata:    map[string]string{"category": "workflow", "requires_write": "true"},
		},
		{
			Name:        "reject",
			Description: "Reject a request or change",
			Metadata:    map[string]string{"category": "workflow", "requires_write": "true"},
		},
		{
			Name:        "submit",
			Description: "Submit for approval",
			Metadata:    map[string]string{"category": "workflow", "requires_write": "true"},
		},
		{
			Name:        "cancel",
			Description: "Cancel an operation",
			Metadata:    map[string]string{"category": "workflow", "requires_write": "true"},
		},

		// Special actions
		{
			Name:        "impersonate",
			Description: "Impersonate another user",
			Metadata:    map[string]string{"category": "special", "requires_admin": "true", "sensitive": "true"},
		},
		{
			Name:        "bypass",
			Description: "Bypass normal restrictions",
			Metadata:    map[string]string{"category": "special", "requires_admin": "true", "sensitive": "true"},
		},
		{
			Name:        "override",
			Description: "Override settings or decisions",
			Metadata:    map[string]string{"category": "special", "requires_admin": "true", "sensitive": "true"},
		},
	}
}

// SeedStaticActions seeds the database with static (built-in) actions
func SeedStaticActions(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	staticActions := GetStaticActions()

	for _, sa := range staticActions {
		action := &models.Action{}

		// Check if action already exists
		err := db.WithContext(ctx).Where("name = ?", sa.Name).First(action).Error
		if err == nil {
			// Action already exists, update if needed
			if logger != nil {
				logger.Info("Static action already exists", zap.String("action", sa.Name))
			}
			continue
		}

		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("error checking action %s: %w", sa.Name, err)
		}

		// Create new action
		// Extract category from metadata if available
		category := models.CategoryGeneral
		if cat, ok := sa.Metadata["category"]; ok {
			category = cat
		}

		action = models.NewStaticAction(sa.Name, sa.Description, category)

		// Convert metadata to JSON string if needed
		if len(sa.Metadata) > 0 {
			metadataJSON, err := json.Marshal(sa.Metadata)
			if err != nil {
				return fmt.Errorf("error marshaling metadata for action %s: %w", sa.Name, err)
			}
			metadataStr := string(metadataJSON)
			action.Metadata = &metadataStr
		}

		if err := db.WithContext(ctx).Create(action).Error; err != nil {
			return fmt.Errorf("error creating action %s: %w", sa.Name, err)
		}

		if logger != nil {
			logger.Info("Created static action", zap.String("action", sa.Name))
		}
	}

	if logger != nil {
		logger.Info("Successfully seeded static actions", zap.Int("count", len(staticActions)))
	}

	return nil
}

// RunSeedStaticActions is deprecated - use SeedStaticActionsWithDBManager instead
func RunSeedStaticActions() error {
	return fmt.Errorf("RunSeedStaticActions is deprecated, use SeedStaticActionsWithDBManager instead")
}

// SeedStaticActionsWithDBManager is a convenience wrapper that uses kisanlink-db DatabaseManager
// to obtain the primary GORM DB and then seeds static actions.
func SeedStaticActionsWithDBManager(ctx context.Context, dm *db.DatabaseManager, logger *zap.Logger) error {
	if dm == nil {
		return fmt.Errorf("database manager is nil")
	}
	pm := dm.GetPostgresManager()
	if pm == nil {
		return fmt.Errorf("postgres manager not available")
	}
	gormDB, err := pm.GetDB(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get postgres DB: %w", err)
	}
	return SeedStaticActions(ctx, gormDB, logger)
}
