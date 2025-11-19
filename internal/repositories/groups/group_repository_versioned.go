package groups

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	pkgErrors "github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"gorm.io/gorm"
)

// UpdateWithVersion updates a group with optimistic locking
// Returns OptimisticLockError if version mismatch occurs
func (r *GroupRepository) UpdateWithVersion(ctx context.Context, group *models.Group, expectedVersion int) error {
	db, err := r.getDB(ctx, false) // Write operation
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// First, get the current version
	var current models.Group
	if err := db.Where("id = ?", group.ID).First(&current).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkgErrors.NewNotFoundError(fmt.Sprintf("group not found with id: %s", group.ID))
		}
		return fmt.Errorf("failed to get current group: %w", err)
	}

	// Check if version matches
	if current.Version != expectedVersion {
		return pkgErrors.NewOptimisticLockError("group", group.ID, expectedVersion, current.Version)
	}

	// Perform the update with version check and increment
	result := db.Model(&models.Group{}).
		Where("id = ? AND version = ?", group.ID, expectedVersion).
		Updates(map[string]interface{}{
			"name":            group.Name,
			"description":     group.Description,
			"organization_id": group.OrganizationID,
			"parent_id":       group.ParentID,
			"is_active":       group.IsActive,
			"metadata":        group.Metadata,
			"version":         gorm.Expr("version + 1"),
			"updated_at":      gorm.Expr("NOW()"),
			"updated_by":      group.UpdatedBy,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update group: %w", result.Error)
	}

	// Double-check that exactly one row was affected
	if result.RowsAffected == 0 {
		// Re-fetch to get current version for accurate error reporting
		if err := db.Where("id = ?", group.ID).First(&current).Error; err == nil {
			return pkgErrors.NewOptimisticLockError("group", group.ID, expectedVersion, current.Version)
		}
		return fmt.Errorf("no rows updated for group: %s", group.ID)
	}

	// Update the model's version to reflect the new state
	group.Version = expectedVersion + 1

	return nil
}

// getDB is a helper method to get the GORM database connection from the database manager
func (r *GroupRepository) getDB(ctx context.Context, readOnly bool) (*gorm.DB, error) {
	if postgresMgr, ok := r.dbManager.(interface {
		GetDB(context.Context, bool) (*gorm.DB, error)
	}); ok {
		return postgresMgr.GetDB(ctx, readOnly)
	}
	return nil, fmt.Errorf("database manager does not support GetDB method")
}
