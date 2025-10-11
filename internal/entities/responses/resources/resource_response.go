package resources

import (
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
)

// ResourceResponse represents a single resource in API responses
// @Description Response structure for a single resource
type ResourceResponse struct {
	ID          string     `json:"id" example:"RES_abc123"`
	Name        string     `json:"name" example:"User Management"`
	Type        string     `json:"type" example:"aaa/user"`
	Description string     `json:"description" example:"Resource for managing users"`
	IsActive    bool       `json:"is_active" example:"true"`
	ParentID    *string    `json:"parent_id,omitempty" example:"RES_parent123"`
	OwnerID     *string    `json:"owner_id,omitempty" example:"USR_owner123"`
	CreatedAt   time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" example:"2024-01-01T00:00:00Z"`
}

// NewResourceResponse creates a new ResourceResponse from a Resource model
func NewResourceResponse(resource *models.Resource) *ResourceResponse {
	if resource == nil {
		return nil
	}

	return &ResourceResponse{
		ID:          resource.ID,
		Name:        resource.Name,
		Type:        resource.Type,
		Description: resource.Description,
		IsActive:    resource.IsActive,
		ParentID:    resource.ParentID,
		OwnerID:     resource.OwnerID,
		CreatedAt:   resource.CreatedAt,
		UpdatedAt:   resource.UpdatedAt,
		DeletedAt:   resource.DeletedAt,
	}
}

// ResourceWithChildrenResponse represents a resource with its children
// @Description Response structure for a resource with children
type ResourceWithChildrenResponse struct {
	*ResourceResponse
	Children []*ResourceResponse `json:"children,omitempty"`
}

// NewResourceWithChildrenResponse creates a new ResourceWithChildrenResponse
func NewResourceWithChildrenResponse(resource *models.Resource, children []*models.Resource) *ResourceWithChildrenResponse {
	response := &ResourceWithChildrenResponse{
		ResourceResponse: NewResourceResponse(resource),
		Children:         make([]*ResourceResponse, 0, len(children)),
	}

	for _, child := range children {
		response.Children = append(response.Children, NewResourceResponse(child))
	}

	return response
}

// ResourceHierarchyResponse represents a hierarchical tree of resources
// @Description Response structure for hierarchical resource tree
type ResourceHierarchyResponse struct {
	*ResourceResponse
	Children []*ResourceHierarchyResponse `json:"children,omitempty"`
}

// NewResourceHierarchyResponse creates a new ResourceHierarchyResponse from a ResourceTree
func NewResourceHierarchyResponse(tree *ResourceTree) *ResourceHierarchyResponse {
	if tree == nil || tree.Resource == nil {
		return nil
	}

	response := &ResourceHierarchyResponse{
		ResourceResponse: NewResourceResponse(tree.Resource),
		Children:         make([]*ResourceHierarchyResponse, 0, len(tree.Children)),
	}

	for _, childTree := range tree.Children {
		childResponse := NewResourceHierarchyResponse(childTree)
		if childResponse != nil {
			response.Children = append(response.Children, childResponse)
		}
	}

	return response
}

// ResourceTree represents a hierarchical tree structure of resources
// This is imported from the service layer for convenience
type ResourceTree struct {
	Resource *models.Resource
	Children []*ResourceTree
}
