package adapters

import (
	"context"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	groupRepo "github.com/Kisanlink/aaa-service/internal/repositories/groups"
)

// GroupRepositoryAdapter adapts the concrete group repository to the GroupRepository interface
type GroupRepositoryAdapter struct {
	repo *groupRepo.GroupRepository
}

// NewGroupRepositoryAdapter creates a new group repository adapter
func NewGroupRepositoryAdapter(repo *groupRepo.GroupRepository) interfaces.GroupRepository {
	return &GroupRepositoryAdapter{repo: repo}
}

// Create implements GroupRepository.Create
func (a *GroupRepositoryAdapter) Create(ctx context.Context, group *models.Group) error {
	return a.repo.Create(ctx, group)
}

// GetByID implements GroupRepository.GetByID
func (a *GroupRepositoryAdapter) GetByID(ctx context.Context, id string, group *models.Group) (*models.Group, error) {
	return a.repo.GetByID(ctx, id)
}

// Update implements GroupRepository.Update
func (a *GroupRepositoryAdapter) Update(ctx context.Context, group *models.Group) error {
	return a.repo.Update(ctx, group)
}

// Delete implements GroupRepository.Delete (adapter method)
func (a *GroupRepositoryAdapter) Delete(ctx context.Context, id string, group *models.Group) error {
	return a.repo.Delete(ctx, id)
}

// List implements GroupRepository.List
func (a *GroupRepositoryAdapter) List(ctx context.Context, limit, offset int) ([]*models.Group, error) {
	return a.repo.List(ctx, limit, offset)
}

// Count implements GroupRepository.Count
func (a *GroupRepositoryAdapter) Count(ctx context.Context) (int64, error) {
	return a.repo.Count(ctx)
}

// Exists implements GroupRepository.Exists
func (a *GroupRepositoryAdapter) Exists(ctx context.Context, id string) (bool, error) {
	return a.repo.Exists(ctx, id)
}

// SoftDelete implements GroupRepository.SoftDelete
func (a *GroupRepositoryAdapter) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	return a.repo.SoftDelete(ctx, id, deletedBy)
}

// Restore implements GroupRepository.Restore
func (a *GroupRepositoryAdapter) Restore(ctx context.Context, id string) error {
	return a.repo.Restore(ctx, id)
}

// GetByName implements GroupRepository.GetByName
func (a *GroupRepositoryAdapter) GetByName(ctx context.Context, name string) (*models.Group, error) {
	return a.repo.GetByName(ctx, name)
}

// GetByNameAndOrganization implements GroupRepository.GetByNameAndOrganization
func (a *GroupRepositoryAdapter) GetByNameAndOrganization(ctx context.Context, name, organizationID string) (*models.Group, error) {
	return a.repo.GetByNameAndOrganization(ctx, name, organizationID)
}

// GetByOrganization implements GroupRepository.GetByOrganization
func (a *GroupRepositoryAdapter) GetByOrganization(ctx context.Context, organizationID string, limit, offset int, includeInactive bool) ([]*models.Group, error) {
	return a.repo.GetByOrganization(ctx, organizationID, limit, offset, includeInactive)
}

// ListActive implements GroupRepository.ListActive
func (a *GroupRepositoryAdapter) ListActive(ctx context.Context, limit, offset int) ([]*models.Group, error) {
	return a.repo.ListActive(ctx, limit, offset)
}

// GetChildren implements GroupRepository.GetChildren
func (a *GroupRepositoryAdapter) GetChildren(ctx context.Context, parentID string) ([]*models.Group, error) {
	return a.repo.GetChildren(ctx, parentID)
}

// HasActiveMembers implements GroupRepository.HasActiveMembers
func (a *GroupRepositoryAdapter) HasActiveMembers(ctx context.Context, groupID string) (bool, error) {
	return a.repo.HasActiveMembers(ctx, groupID)
}

// CreateMembership implements GroupRepository.CreateMembership
func (a *GroupRepositoryAdapter) CreateMembership(ctx context.Context, membership *models.GroupMembership) error {
	return a.repo.CreateMembership(ctx, membership)
}

// UpdateMembership implements GroupRepository.UpdateMembership
func (a *GroupRepositoryAdapter) UpdateMembership(ctx context.Context, membership *models.GroupMembership) error {
	return a.repo.UpdateMembership(ctx, membership)
}

// GetMembership implements GroupRepository.GetMembership
func (a *GroupRepositoryAdapter) GetMembership(ctx context.Context, groupID, principalID string) (*models.GroupMembership, error) {
	return a.repo.GetMembership(ctx, groupID, principalID)
}

// GetGroupMembers implements GroupRepository.GetGroupMembers
func (a *GroupRepositoryAdapter) GetGroupMembers(ctx context.Context, groupID string, limit, offset int) ([]*models.GroupMembership, error) {
	return a.repo.GetGroupMembers(ctx, groupID, limit, offset)
}

// ListWithDeleted implements GroupRepository.ListWithDeleted
func (a *GroupRepositoryAdapter) ListWithDeleted(ctx context.Context, limit, offset int) ([]*models.Group, error) {
	return a.repo.ListWithDeleted(ctx, limit, offset)
}

// CountWithDeleted implements GroupRepository.CountWithDeleted
func (a *GroupRepositoryAdapter) CountWithDeleted(ctx context.Context) (int64, error) {
	return a.repo.CountWithDeleted(ctx)
}

// ExistsWithDeleted implements GroupRepository.ExistsWithDeleted
func (a *GroupRepositoryAdapter) ExistsWithDeleted(ctx context.Context, id string) (bool, error) {
	return a.repo.ExistsWithDeleted(ctx, id)
}

// GetByCreatedBy implements GroupRepository.GetByCreatedBy
func (a *GroupRepositoryAdapter) GetByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*models.Group, error) {
	return a.repo.GetByCreatedBy(ctx, createdBy, limit, offset)
}

// GetByUpdatedBy implements GroupRepository.GetByUpdatedBy
func (a *GroupRepositoryAdapter) GetByUpdatedBy(ctx context.Context, updatedBy string, limit, offset int) ([]*models.Group, error) {
	return a.repo.GetByUpdatedBy(ctx, updatedBy, limit, offset)
}

// GetByDeletedBy implements GroupRepository.GetByDeletedBy
func (a *GroupRepositoryAdapter) GetByDeletedBy(ctx context.Context, deletedBy string, limit, offset int) ([]*models.Group, error) {
	return a.repo.GetByDeletedBy(ctx, deletedBy, limit, offset)
}

// CreateMany implements GroupRepository.CreateMany
func (a *GroupRepositoryAdapter) CreateMany(ctx context.Context, groups []*models.Group) error {
	return a.repo.CreateMany(ctx, groups)
}

// UpdateMany implements GroupRepository.UpdateMany
func (a *GroupRepositoryAdapter) UpdateMany(ctx context.Context, groups []*models.Group) error {
	return a.repo.UpdateMany(ctx, groups)
}

// DeleteMany implements GroupRepository.DeleteMany
func (a *GroupRepositoryAdapter) DeleteMany(ctx context.Context, ids []string) error {
	return a.repo.DeleteMany(ctx, ids)
}

// SoftDeleteMany implements GroupRepository.SoftDeleteMany
func (a *GroupRepositoryAdapter) SoftDeleteMany(ctx context.Context, ids []string, deletedBy string) error {
	return a.repo.SoftDeleteMany(ctx, ids, deletedBy)
}
