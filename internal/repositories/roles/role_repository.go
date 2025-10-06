package roles

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// RoleRepository handles role-related database operations
type RoleRepository struct {
	*base.BaseFilterableRepository[*models.Role]
	dbManager db.DBManager
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(dbManager db.DBManager) *RoleRepository {
	repo := &RoleRepository{
		BaseFilterableRepository: base.NewBaseFilterableRepository[*models.Role](),
		dbManager:                dbManager,
	}

	// Set the database manager on the BaseFilterableRepository so it can use database-level operations
	repo.BaseFilterableRepository.SetDBManager(dbManager)

	return repo
}

// Create creates a new role in the database
func (r *RoleRepository) Create(ctx context.Context, role *models.Role) error {
	return r.BaseFilterableRepository.Create(ctx, role)
}

// GetByID retrieves a role by ID
func (r *RoleRepository) GetByID(ctx context.Context, id string, role *models.Role) (*models.Role, error) {
	return r.BaseFilterableRepository.GetByID(ctx, id, role)
}

// Update updates an existing role
func (r *RoleRepository) Update(ctx context.Context, role *models.Role) error {
	return r.BaseFilterableRepository.Update(ctx, role)
}

// Delete deletes a role by ID (hard delete)
func (r *RoleRepository) Delete(ctx context.Context, id string, role *models.Role) error {
	// Use the BaseFilterableRepository directly - it now properly handles table names
	return r.BaseFilterableRepository.Delete(ctx, id, role)
}

// SoftDelete soft deletes a role by setting deleted_at and deleted_by fields
func (r *RoleRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	// Use the BaseFilterableRepository directly - it now properly handles soft delete
	return r.BaseFilterableRepository.SoftDelete(ctx, id, deletedBy)
}

// Restore restores a soft-deleted role
func (r *RoleRepository) Restore(ctx context.Context, id string) error {
	// Use the BaseFilterableRepository directly - it now properly handles restore
	return r.BaseFilterableRepository.Restore(ctx, id)
}

// List retrieves all roles with pagination using database-level filtering
func (r *RoleRepository) List(ctx context.Context, limit, offset int) ([]*models.Role, error) {
	// Use base filterable repository for optimized database-level filtering
	// Only return active roles (not deleted)
	filter := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Count returns the total number of roles using database-level counting
func (r *RoleRepository) Count(ctx context.Context) (int64, error) {
	filter := base.NewFilter()
	return r.BaseFilterableRepository.Count(ctx, filter, models.Role{})
}

// CountWithDeleted returns count including soft-deleted roles
func (r *RoleRepository) CountWithDeleted(ctx context.Context) (int64, error) {
	return r.BaseFilterableRepository.CountWithDeleted(ctx, &models.Role{})
}

// GetByName retrieves a role by name using base filterable repository
func (r *RoleRepository) GetByName(ctx context.Context, name string) (*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Build()

	roles, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("role not found")
	}

	return roles[0], nil
}

// GetActive retrieves all active roles with pagination using database-level filtering
func (r *RoleRepository) GetActive(ctx context.Context, limit, offset int) ([]*models.Role, error) {
	// For now, we'll consider all non-deleted roles as active
	// In the future, you might want to add an "active" field to the Role model
	filter := base.NewFilterBuilder().
		Where("is_deleted", base.OpEqual, false).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Search searches roles by keyword using database-level filtering
func (r *RoleRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpContains, query).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// ExistsByName checks if a role exists with the given name using database-level filtering
func (r *RoleRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Build()

	roles, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return false, err
	}

	return len(roles) > 0, nil
}

// GetByDescription retrieves roles by description using database-level filtering
func (r *RoleRepository) GetByDescription(ctx context.Context, description string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("description", base.OpContains, description).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByPermission retrieves roles by permission using database-level filtering
func (r *RoleRepository) GetByPermission(ctx context.Context, permission string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("permissions", base.OpContains, permission).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByCreatedBy retrieves roles by creator using database-level filtering
func (r *RoleRepository) GetByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("created_by", base.OpEqual, createdBy).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUpdatedBy retrieves roles by updater using database-level filtering
func (r *RoleRepository) GetByUpdatedBy(ctx context.Context, updatedBy string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("updated_by", base.OpEqual, updatedBy).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByDeletedBy retrieves roles by deleter using database-level filtering
func (r *RoleRepository) GetByDeletedBy(ctx context.Context, deletedBy string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("deleted_by", base.OpEqual, deletedBy).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByDateRange retrieves roles created within a date range using database-level filtering
func (r *RoleRepository) GetByDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUpdatedDateRange retrieves roles updated within a date range using database-level filtering
func (r *RoleRepository) GetByUpdatedDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		WhereBetween("updated_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByDeletedDateRange retrieves roles deleted within a date range using database-level filtering
func (r *RoleRepository) GetByDeletedDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		WhereBetween("deleted_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByNameAndDescription retrieves roles by name and description using database-level filtering
func (r *RoleRepository) GetByNameAndDescription(ctx context.Context, name, description string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Where("description", base.OpContains, description).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByNameAndPermission retrieves roles by name and permission using database-level filtering
func (r *RoleRepository) GetByNameAndPermission(ctx context.Context, name, permission string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Where("permissions", base.OpContains, permission).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByNameAndCreatedBy retrieves roles by name and creator using database-level filtering
func (r *RoleRepository) GetByNameAndCreatedBy(ctx context.Context, name, createdBy string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Where("created_by", base.OpEqual, createdBy).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByNameAndUpdatedBy retrieves roles by name and updater using database-level filtering
func (r *RoleRepository) GetByNameAndUpdatedBy(ctx context.Context, name, updatedBy string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Where("updated_by", base.OpEqual, updatedBy).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByNameAndDeletedBy retrieves roles by name and deleter using database-level filtering
func (r *RoleRepository) GetByNameAndDeletedBy(ctx context.Context, name, deletedBy string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Where("deleted_by", base.OpEqual, deletedBy).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByNameAndDateRange retrieves roles by name and date range using database-level filtering
func (r *RoleRepository) GetByNameAndDateRange(ctx context.Context, name, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByNameAndUpdatedDateRange retrieves roles by name and updated date range using database-level filtering
func (r *RoleRepository) GetByNameAndUpdatedDateRange(ctx context.Context, name, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		WhereBetween("updated_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByNameAndDeletedDateRange retrieves roles by name and deleted date range using database-level filtering
func (r *RoleRepository) GetByNameAndDeletedDateRange(ctx context.Context, name, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		WhereBetween("deleted_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByDescriptionAndPermission retrieves roles by description and permission using database-level filtering
func (r *RoleRepository) GetByDescriptionAndPermission(ctx context.Context, description, permission string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("description", base.OpContains, description).
		Where("permissions", base.OpContains, permission).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByDescriptionAndCreatedBy retrieves roles by description and creator using database-level filtering
func (r *RoleRepository) GetByDescriptionAndCreatedBy(ctx context.Context, description, createdBy string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("description", base.OpContains, description).
		Where("created_by", base.OpEqual, createdBy).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByDescriptionAndUpdatedBy retrieves roles by description and updater using database-level filtering
func (r *RoleRepository) GetByDescriptionAndUpdatedBy(ctx context.Context, description, updatedBy string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("description", base.OpContains, description).
		Where("updated_by", base.OpEqual, updatedBy).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByDescriptionAndDeletedBy retrieves roles by description and deleter using database-level filtering
func (r *RoleRepository) GetByDescriptionAndDeletedBy(ctx context.Context, description, deletedBy string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("description", base.OpContains, description).
		Where("deleted_by", base.OpEqual, deletedBy).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByDescriptionAndDateRange retrieves roles by description and date range using database-level filtering
func (r *RoleRepository) GetByDescriptionAndDateRange(ctx context.Context, description, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("description", base.OpContains, description).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByDescriptionAndUpdatedDateRange retrieves roles by description and updated date range using database-level filtering
func (r *RoleRepository) GetByDescriptionAndUpdatedDateRange(ctx context.Context, description, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("description", base.OpContains, description).
		WhereBetween("updated_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByDescriptionAndDeletedDateRange retrieves roles by description and deleted date range using database-level filtering
func (r *RoleRepository) GetByDescriptionAndDeletedDateRange(ctx context.Context, description, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("description", base.OpContains, description).
		WhereBetween("deleted_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByPermissionAndCreatedBy retrieves roles by permission and creator using database-level filtering
func (r *RoleRepository) GetByPermissionAndCreatedBy(ctx context.Context, permission, createdBy string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("permissions", base.OpContains, permission).
		Where("created_by", base.OpEqual, createdBy).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByPermissionAndUpdatedBy retrieves roles by permission and updater using database-level filtering
func (r *RoleRepository) GetByPermissionAndUpdatedBy(ctx context.Context, permission, updatedBy string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("permissions", base.OpContains, permission).
		Where("updated_by", base.OpEqual, updatedBy).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByPermissionAndDeletedBy retrieves roles by permission and deleter using database-level filtering
func (r *RoleRepository) GetByPermissionAndDeletedBy(ctx context.Context, permission, deletedBy string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("permissions", base.OpContains, permission).
		Where("deleted_by", base.OpEqual, deletedBy).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByPermissionAndDateRange retrieves roles by permission and date range using database-level filtering
func (r *RoleRepository) GetByPermissionAndDateRange(ctx context.Context, permission, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("permissions", base.OpContains, permission).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByPermissionAndUpdatedDateRange retrieves roles by permission and updated date range using database-level filtering
func (r *RoleRepository) GetByPermissionAndUpdatedDateRange(ctx context.Context, permission, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("permissions", base.OpContains, permission).
		WhereBetween("updated_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByPermissionAndDeletedDateRange retrieves roles by permission and deleted date range using database-level filtering
func (r *RoleRepository) GetByPermissionAndDeletedDateRange(ctx context.Context, permission, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("permissions", base.OpContains, permission).
		WhereBetween("deleted_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByCreatedByAndUpdatedBy retrieves roles by creator and updater using database-level filtering
func (r *RoleRepository) GetByCreatedByAndUpdatedBy(ctx context.Context, createdBy, updatedBy string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("created_by", base.OpEqual, createdBy).
		Where("updated_by", base.OpEqual, updatedBy).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByCreatedByAndDeletedBy retrieves roles by creator and deleter using database-level filtering
func (r *RoleRepository) GetByCreatedByAndDeletedBy(ctx context.Context, createdBy, deletedBy string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("created_by", base.OpEqual, createdBy).
		Where("deleted_by", base.OpEqual, deletedBy).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByCreatedByAndDateRange retrieves roles by creator and date range using database-level filtering
func (r *RoleRepository) GetByCreatedByAndDateRange(ctx context.Context, createdBy, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("created_by", base.OpEqual, createdBy).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByCreatedByAndUpdatedDateRange retrieves roles by creator and updated date range using database-level filtering
func (r *RoleRepository) GetByCreatedByAndUpdatedDateRange(ctx context.Context, createdBy, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("created_by", base.OpEqual, createdBy).
		WhereBetween("updated_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByCreatedByAndDeletedDateRange retrieves roles by creator and deleted date range using database-level filtering
func (r *RoleRepository) GetByCreatedByAndDeletedDateRange(ctx context.Context, createdBy, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("created_by", base.OpEqual, createdBy).
		WhereBetween("deleted_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUpdatedByAndDeletedBy retrieves roles by updater and deleter using database-level filtering
func (r *RoleRepository) GetByUpdatedByAndDeletedBy(ctx context.Context, updatedBy, deletedBy string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("updated_by", base.OpEqual, updatedBy).
		Where("deleted_by", base.OpEqual, deletedBy).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUpdatedByAndDateRange retrieves roles by updater and date range using database-level filtering
func (r *RoleRepository) GetByUpdatedByAndDateRange(ctx context.Context, updatedBy, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("updated_by", base.OpEqual, updatedBy).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUpdatedByAndUpdatedDateRange retrieves roles by updater and updated date range using database-level filtering
func (r *RoleRepository) GetByUpdatedByAndUpdatedDateRange(ctx context.Context, updatedBy, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("updated_by", base.OpEqual, updatedBy).
		WhereBetween("updated_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUpdatedByAndDeletedDateRange retrieves roles by updater and deleted date range using database-level filtering
func (r *RoleRepository) GetByUpdatedByAndDeletedDateRange(ctx context.Context, updatedBy, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("updated_by", base.OpEqual, updatedBy).
		WhereBetween("deleted_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByDeletedByAndDateRange retrieves roles by deleter and date range using database-level filtering
func (r *RoleRepository) GetByDeletedByAndDateRange(ctx context.Context, deletedBy, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("deleted_by", base.OpEqual, deletedBy).
		WhereBetween("created_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByDeletedByAndUpdatedDateRange retrieves roles by deleter and updated date range using database-level filtering
func (r *RoleRepository) GetByDeletedByAndUpdatedDateRange(ctx context.Context, deletedBy, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("deleted_by", base.OpEqual, deletedBy).
		WhereBetween("updated_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByDeletedByAndDeletedDateRange retrieves roles by deleter and deleted date range using database-level filtering
func (r *RoleRepository) GetByDeletedByAndDeletedDateRange(ctx context.Context, deletedBy, startDate, endDate string, limit, offset int) ([]*models.Role, error) {
	filter := base.NewFilterBuilder().
		Where("deleted_by", base.OpEqual, deletedBy).
		WhereBetween("deleted_at", startDate, endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}
