package middleware

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ResponseTransformationMiddleware handles query parameter validation and transformation options
type ResponseTransformationMiddleware struct {
	queryHandler interfaces.QueryParameterHandler
	logger       interfaces.Logger
}

// NewResponseTransformationMiddleware creates a new response transformation middleware
func NewResponseTransformationMiddleware(
	queryHandler interfaces.QueryParameterHandler,
	logger interfaces.Logger,
) *ResponseTransformationMiddleware {
	return &ResponseTransformationMiddleware{
		queryHandler: queryHandler,
		logger:       logger,
	}
}

// ValidateQueryParameters validates query parameters before processing
func (rtm *ResponseTransformationMiddleware) ValidateQueryParameters() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate query parameters
		if err := rtm.queryHandler.ValidateQueryParameters(c); err != nil {
			rtm.logger.Warn("Invalid query parameters",
				zap.String("error", err.Error()),
				zap.String("path", c.Request.URL.Path),
				zap.String("query", c.Request.URL.RawQuery),
			)

			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid query parameters",
				"message": err.Error(),
				"success": false,
			})
			c.Abort()
			return
		}

		// Parse and store transformation options in context
		options := rtm.queryHandler.ParseTransformOptions(c)
		c.Set("transform_options", options)

		// Parse and store pagination parameters
		limit, offset, err := rtm.queryHandler.GetPaginationParams(c)
		if err != nil {
			rtm.logger.Warn("Invalid pagination parameters",
				zap.String("error", err.Error()),
				zap.String("path", c.Request.URL.Path),
			)

			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid pagination parameters",
				"message": err.Error(),
				"success": false,
			})
			c.Abort()
			return
		}

		c.Set("pagination_limit", limit)
		c.Set("pagination_offset", offset)

		// Parse and store sorting parameters
		sortBy, order := rtm.queryHandler.GetSortParams(c)
		c.Set("sort_by", sortBy)
		c.Set("sort_order", order)

		// Parse and store search parameter
		search := rtm.queryHandler.GetSearchParam(c)
		if search != "" {
			c.Set("search_query", search)
		}

		// Parse and store filter parameters
		filters := rtm.queryHandler.GetFilterParams(c)
		if len(filters) > 0 {
			c.Set("filters", filters)
		}

		c.Next()
	}
}

// GetTransformOptions retrieves transformation options from context
func GetTransformOptions(c *gin.Context) interfaces.TransformOptions {
	if options, exists := c.Get("transform_options"); exists {
		if transformOptions, ok := options.(interfaces.TransformOptions); ok {
			return transformOptions
		}
	}

	// Return default options if not found
	handler := &QueryParameterHandlerImpl{}
	return handler.GetDefaultOptions()
}

// GetPaginationParams retrieves pagination parameters from context
func GetPaginationParams(c *gin.Context) (limit, offset int) {
	limit = 20 // default
	offset = 0 // default

	if l, exists := c.Get("pagination_limit"); exists {
		if limitVal, ok := l.(int); ok {
			limit = limitVal
		}
	}

	if o, exists := c.Get("pagination_offset"); exists {
		if offsetVal, ok := o.(int); ok {
			offset = offsetVal
		}
	}

	return limit, offset
}

// GetSortParams retrieves sorting parameters from context
func GetSortParams(c *gin.Context) (sortBy, order string) {
	sortBy = "created_at" // default
	order = "desc"        // default

	if s, exists := c.Get("sort_by"); exists {
		if sortVal, ok := s.(string); ok {
			sortBy = sortVal
		}
	}

	if o, exists := c.Get("sort_order"); exists {
		if orderVal, ok := o.(string); ok {
			order = orderVal
		}
	}

	return sortBy, order
}

// GetSearchQuery retrieves search query from context
func GetSearchQuery(c *gin.Context) string {
	if search, exists := c.Get("search_query"); exists {
		if searchVal, ok := search.(string); ok {
			return searchVal
		}
	}
	return ""
}

// GetFilters retrieves filter parameters from context
func GetFilters(c *gin.Context) map[string]string {
	if filters, exists := c.Get("filters"); exists {
		if filterMap, ok := filters.(map[string]string); ok {
			return filterMap
		}
	}
	return make(map[string]string)
}

// QueryParameterHandlerImpl is a local implementation for default options
type QueryParameterHandlerImpl struct{}

func (qph *QueryParameterHandlerImpl) GetDefaultOptions() interfaces.TransformOptions {
	return interfaces.TransformOptions{
		IncludeProfile:    false,
		IncludeContacts:   false,
		IncludeRole:       false,
		IncludeUser:       false,
		IncludeAddress:    false,
		ExcludeDeleted:    true,
		ExcludeInactive:   false,
		OnlyActiveRoles:   false,
		MaskSensitiveData: true,
		IncludeTimestamps: true,
	}
}
