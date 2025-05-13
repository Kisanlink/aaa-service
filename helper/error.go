package helper

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AppError represents an application error with status code
type AppError struct {
	StatusCode int
	Err        error
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Err.Error()
}

// NewAppError creates a new AppError
func NewAppError(statusCode int, err error) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Err:        err,
	}
}

// Common error types
var (
	ErrNotFound     = NewAppError(http.StatusNotFound, errors.New("resource not found"))
	ErrBadRequest   = NewAppError(http.StatusBadRequest, errors.New("bad request"))
	ErrUnauthorized = NewAppError(http.StatusUnauthorized, errors.New("unauthorized"))
	ErrForbidden    = NewAppError(http.StatusForbidden, errors.New("forbidden"))
	ErrInternal     = NewAppError(http.StatusInternalServerError, errors.New("internal server error"))
)

// SilentRecoveryMiddleware is a custom recovery middleware that doesn't log errors
func SilentRecoveryMiddleware(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			// Handle AppError specifically
			if appErr, ok := err.(*AppError); ok {
				SendErrorResponse(c.Writer, appErr.StatusCode, []string{appErr.Error()})
				c.Abort()
				return
			}

			// Handle other error types
			if e, ok := err.(error); ok {
				SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{e.Error()})
			} else {
				SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"internal server error"})
			}
			c.Abort()
		}
	}()

	c.Next()
}

type GRPCError struct {
	Code          codes.Code
	Message       string
	Errors        []string
	Data          interface{}
	DataTimeStamp string
}

func (e *GRPCError) Error() string {
	return e.Message
}

func (e *GRPCError) GRPCStatus() *status.Status {
	st := status.New(e.Code, e.Message)
	// You can add details here if needed
	return st
}

func NewGRPCError(code codes.Code, message string, errors []string) error {
	return &GRPCError{
		Code:          code,
		Message:       message,
		Errors:        errors,
		Data:          nil,
		DataTimeStamp: time.Now().Format(time.RFC3339),
	}

}
