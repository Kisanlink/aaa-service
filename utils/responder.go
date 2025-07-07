package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
}

// Response represents a standardized API response
type Response struct {
	Success    bool            `json:"success"`
	Message    string          `json:"message,omitempty"`
	Data       interface{}     `json:"data,omitempty"`
	Error      string          `json:"error,omitempty"`
	Pagination *PaginationMeta `json:"pagination,omitempty"`
	Timestamp  string          `json:"timestamp"`
	RequestID  string          `json:"request_id,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success   bool     `json:"success"`
	Error     string   `json:"error"`
	Message   string   `json:"message,omitempty"`
	Details   []string `json:"details,omitempty"`
	Timestamp string   `json:"timestamp"`
	RequestID string   `json:"request_id,omitempty"`
}

// Responder provides utilities for sending standardized HTTP responses
type Responder struct {
	includeRequestID bool
}

// NewResponder creates a new Responder instance
func NewResponder(includeRequestID bool) *Responder {
	return &Responder{
		includeRequestID: includeRequestID,
	}
}

// SendSuccess sends a successful response
func (r *Responder) SendSuccess(c *gin.Context, statusCode int, data interface{}) {
	response := Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if r.includeRequestID {
		response.RequestID = c.GetString("request_id")
	}

	c.JSON(statusCode, response)
}

// SendError sends an error response
func (r *Responder) SendError(c *gin.Context, statusCode int, message string, err error) {
	errorResponse := ErrorResponse{
		Success:   false,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if err != nil {
		errorResponse.Error = err.Error()
	}

	if r.includeRequestID {
		errorResponse.RequestID = c.GetString("request_id")
	}

	c.JSON(statusCode, errorResponse)
}

// SendValidationError sends a validation error response
func (r *Responder) SendValidationError(c *gin.Context, message string, details []string) {
	errorResponse := ErrorResponse{
		Success:   false,
		Message:   message,
		Details:   details,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if r.includeRequestID {
		errorResponse.RequestID = c.GetString("request_id")
	}

	c.JSON(http.StatusBadRequest, errorResponse)
}

// SendNotFound sends a not found response
func (r *Responder) SendNotFound(c *gin.Context, resource string) {
	errorResponse := ErrorResponse{
		Success:   false,
		Message:   resource + " not found",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if r.includeRequestID {
		errorResponse.RequestID = c.GetString("request_id")
	}

	c.JSON(http.StatusNotFound, errorResponse)
}

// SendConflict sends a conflict response
func (r *Responder) SendConflict(c *gin.Context, message string) {
	errorResponse := ErrorResponse{
		Success:   false,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if r.includeRequestID {
		errorResponse.RequestID = c.GetString("request_id")
	}

	c.JSON(http.StatusConflict, errorResponse)
}

// SendUnauthorized sends an unauthorized response
func (r *Responder) SendUnauthorized(c *gin.Context, message string) {
	errorResponse := ErrorResponse{
		Success:   false,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if r.includeRequestID {
		errorResponse.RequestID = c.GetString("request_id")
	}

	c.JSON(http.StatusUnauthorized, errorResponse)
}

// SendForbidden sends a forbidden response
func (r *Responder) SendForbidden(c *gin.Context, message string) {
	errorResponse := ErrorResponse{
		Success:   false,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if r.includeRequestID {
		errorResponse.RequestID = c.GetString("request_id")
	}

	c.JSON(http.StatusForbidden, errorResponse)
}

// SendInternalError sends an internal server error response
func (r *Responder) SendInternalError(c *gin.Context, message string, err error) {
	errorResponse := ErrorResponse{
		Success:   false,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if err != nil {
		errorResponse.Error = err.Error()
	}

	if r.includeRequestID {
		errorResponse.RequestID = c.GetString("request_id")
	}

	c.JSON(http.StatusInternalServerError, errorResponse)
}

// SendPaginatedResponse sends a paginated response
func (r *Responder) SendPaginatedResponse(c *gin.Context, data interface{}, total int, page, limit int) {
	totalPages := (total + limit - 1) / limit

	response := Response{
		Success: true,
		Data:    data,
		Pagination: &PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if r.includeRequestID {
		response.RequestID = c.GetString("request_id")
	}

	c.JSON(http.StatusOK, response)
}

// SendRawJSON sends raw JSON response
func (r *Responder) SendRawJSON(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

// SendFile sends a file response
func (r *Responder) SendFile(c *gin.Context, filepath, filename string) {
	c.FileAttachment(filepath, filename)
}

// SendRedirect sends a redirect response
func (r *Responder) SendRedirect(c *gin.Context, url string) {
	c.Redirect(http.StatusFound, url)
}

// SendNoContent sends a no content response
func (r *Responder) SendNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// SendCreated sends a created response
func (r *Responder) SendCreated(c *gin.Context, data interface{}) {
	r.SendSuccess(c, http.StatusCreated, data)
}

// SendAccepted sends an accepted response
func (r *Responder) SendAccepted(c *gin.Context, data interface{}) {
	r.SendSuccess(c, http.StatusAccepted, data)
}
