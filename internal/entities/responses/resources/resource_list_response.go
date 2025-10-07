package resources

import (
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
)

// ResourceListResponse represents a paginated list of resources
// @Description Response structure for a list of resources with pagination
type ResourceListResponse struct {
	Success    bool              `json:"success" example:"true"`
	Message    string            `json:"message" example:"Resources retrieved successfully"`
	Data       *ResourceListData `json:"data"`
	Pagination *PaginationInfo   `json:"pagination"`
	Timestamp  time.Time         `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID  string            `json:"request_id" example:"req_abc123"`
}

// ResourceListData contains the actual resource data
type ResourceListData struct {
	Resources []*ResourceResponse `json:"resources"`
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Page       int `json:"page" example:"1"`
	Limit      int `json:"limit" example:"10"`
	Total      int `json:"total" example:"100"`
	TotalPages int `json:"total_pages" example:"10"`
}

// NewResourceListResponse creates a new ResourceListResponse
func NewResourceListResponse(
	resources []*models.Resource,
	page, limit, total int,
	requestID string,
) *ResourceListResponse {
	resourceResponses := make([]*ResourceResponse, 0, len(resources))
	for _, resource := range resources {
		resourceResponses = append(resourceResponses, NewResourceResponse(resource))
	}

	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	return &ResourceListResponse{
		Success: true,
		Message: "Resources retrieved successfully",
		Data: &ResourceListData{
			Resources: resourceResponses,
		},
		Pagination: &PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
		Timestamp: time.Now(),
		RequestID: requestID,
	}
}
