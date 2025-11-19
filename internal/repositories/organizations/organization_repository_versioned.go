package organizations

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	pkgErrors "github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"gorm.io/gorm"
)

// UpdateWithVersion updates an organization with optimistic locking
// Returns OptimisticLockError if version mismatch occurs
func (r *OrganizationRepository) UpdateWithVersion(ctx context.Context, org *models.Organization, expectedVersion int) error {
	db, err := r.getDB(ctx, false) // Write operation
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// First, get the current version
	var current models.Organization
	if err := db.Where("id = ?", org.ID).First(&current).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkgErrors.NewNotFoundError(fmt.Sprintf("organization not found with id: %s", org.ID))
		}
		return fmt.Errorf("failed to get current organization: %w", err)
	}

	// Check if version matches
	if current.Version != expectedVersion {
		return pkgErrors.NewOptimisticLockError("organization", org.ID, expectedVersion, current.Version)
	}

	// Perform the update with version check and increment
	result := db.Model(&models.Organization{}).
		Where("id = ? AND version = ?", org.ID, expectedVersion).
		Updates(map[string]interface{}{
			"name":        org.Name,
			"type":        org.Type,
			"description": org.Description,
			"parent_id":   org.ParentID,
			"is_active":   org.IsActive,
			"metadata":    org.Metadata,
			"version":     gorm.Expr("version + 1"),
			"updated_at":  gorm.Expr("NOW()"),
			"updated_by":  org.UpdatedBy,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update organization: %w", result.Error)
	}

	// Double-check that exactly one row was affected
	if result.RowsAffected == 0 {
		// Re-fetch to get current version for accurate error reporting
		if err := db.Where("id = ?", org.ID).First(&current).Error; err == nil {
			return pkgErrors.NewOptimisticLockError("organization", org.ID, expectedVersion, current.Version)
		}
		return fmt.Errorf("no rows updated for organization: %s", org.ID)
	}

	// Update the model's version to reflect the new state
	org.Version = expectedVersion + 1

	return nil
}

// getDB is a helper method to get the GORM database connection from the database manager
func (r *OrganizationRepository) getDB(ctx context.Context, readOnly bool) (*gorm.DB, error) {
	if postgresMgr, ok := r.dbManager.(interface {
		GetDB(context.Context, bool) (*gorm.DB, error)
	}); ok {
		return postgresMgr.GetDB(ctx, readOnly)
	}
	return nil, fmt.Errorf("database manager does not support GetDB method")
}
