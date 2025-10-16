package catalog

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/resources"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"go.uber.org/zap"
)

// ResourceManager handles resource-related operations for the catalog service
type ResourceManager struct {
	resourceRepo *resources.ResourceRepository
	logger       *zap.Logger
}

// NewResourceManager creates a new resource manager
func NewResourceManager(resourceRepo *resources.ResourceRepository, logger *zap.Logger) *ResourceManager {
	return &ResourceManager{
		resourceRepo: resourceRepo,
		logger:       logger,
	}
}

// UpsertResources creates or updates resources using UPSERT pattern
// Returns the count of resources created/updated
func (rm *ResourceManager) UpsertResources(ctx context.Context, resourceDefs []ResourceDefinition, force bool) (int32, error) {
	var count int32

	for _, resourceDef := range resourceDefs {
		// Check if resource already exists
		existing, err := rm.getResourceByName(ctx, resourceDef.Name)

		if err == nil && existing != nil {
			// Resource exists
			if !force {
				// Skip if not forcing update
				rm.logger.Debug("Resource already exists, skipping",
					zap.String("resource", resourceDef.Name))
				continue
			}

			// Update existing resource
			existing.Type = resourceDef.Type
			existing.Description = resourceDef.Description

			if err := rm.resourceRepo.Update(ctx, existing); err != nil {
				return count, fmt.Errorf("failed to update resource %s: %w", resourceDef.Name, err)
			}

			rm.logger.Debug("Resource updated",
				zap.String("resource", resourceDef.Name))
			count++
		} else {
			// Create new resource
			resource := models.NewResource(
				resourceDef.Name,
				resourceDef.Type,
				resourceDef.Description,
			)

			if err := rm.resourceRepo.Create(ctx, resource); err != nil {
				return count, fmt.Errorf("failed to create resource %s: %w", resourceDef.Name, err)
			}

			rm.logger.Debug("Resource created",
				zap.String("resource", resourceDef.Name),
				zap.String("id", resource.ID))
			count++
		}
	}

	return count, nil
}

// GetResourceByName retrieves a resource by name
func (rm *ResourceManager) GetResourceByName(ctx context.Context, name string) (*models.Resource, error) {
	return rm.getResourceByName(ctx, name)
}

// getResourceByName is a private helper to retrieve resource by name
func (rm *ResourceManager) getResourceByName(ctx context.Context, name string) (*models.Resource, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Build()

	resources, err := rm.resourceRepo.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find resource by name: %w", err)
	}

	if len(resources) == 0 {
		return nil, nil
	}

	return resources[0], nil
}

// GetAllResources retrieves all resources
func (rm *ResourceManager) GetAllResources(ctx context.Context) ([]*models.Resource, error) {
	filter := base.NewFilterBuilder().Build()
	return rm.resourceRepo.Find(ctx, filter)
}
