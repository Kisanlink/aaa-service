package resources

import (
	"context"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	resourceRepo "github.com/Kisanlink/aaa-service/internal/repositories/resources"
	"go.uber.org/zap"
)

// ResourceService defines the interface for resource management operations
type ResourceService interface {
	// Create operations
	Create(ctx context.Context, name, resourceType, description string) (*models.Resource, error)
	CreateWithParent(ctx context.Context, name, resourceType, description, parentID string) (*models.Resource, error)
	CreateWithOwner(ctx context.Context, name, resourceType, description, ownerID string) (*models.Resource, error)
	CreateBatch(ctx context.Context, resources []*models.Resource) error

	// Read operations
	GetByID(ctx context.Context, id string) (*models.Resource, error)
	GetByName(ctx context.Context, name string) (*models.Resource, error)
	GetByType(ctx context.Context, resourceType string, limit, offset int) ([]*models.Resource, error)
	List(ctx context.Context, limit, offset int) ([]*models.Resource, error)
	Count(ctx context.Context) (int64, error)
	Exists(ctx context.Context, id string) (bool, error)

	// Update operations
	Update(ctx context.Context, resource *models.Resource) error
	UpdateName(ctx context.Context, id, name string) error
	UpdateDescription(ctx context.Context, id, description string) error
	Activate(ctx context.Context, id string) error
	Deactivate(ctx context.Context, id string) error
	SetParent(ctx context.Context, id, parentID string) error
	SetOwner(ctx context.Context, id, ownerID string) error

	// Delete operations
	Delete(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id, deletedBy string) error
	DeleteCascade(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error

	// Hierarchy operations
	GetChildren(ctx context.Context, parentID string) ([]*models.Resource, error)
	GetParent(ctx context.Context, id string) (*models.Resource, error)
	GetHierarchy(ctx context.Context, rootID string) (*ResourceTree, error)
	GetAncestors(ctx context.Context, id string) ([]*models.Resource, error)
	GetDescendants(ctx context.Context, id string) ([]*models.Resource, error)
	HasChildren(ctx context.Context, id string) (bool, error)
	ValidateHierarchy(ctx context.Context, id, parentID string) error
}

// Service implements the ResourceService interface
type Service struct {
	resourceRepo *resourceRepo.ResourceRepository
	cacheService interfaces.CacheService
	logger       *zap.Logger
}

// ResourceTree represents a hierarchical tree structure of resources
type ResourceTree struct {
	Resource *models.Resource `json:"resource"`
	Children []*ResourceTree  `json:"children,omitempty"`
}

// NewService creates a new ResourceService instance
func NewService(
	resourceRepo *resourceRepo.ResourceRepository,
	cacheService interfaces.CacheService,
	logger *zap.Logger,
) ResourceService {
	return &Service{
		resourceRepo: resourceRepo,
		cacheService: cacheService,
		logger:       logger,
	}
}
