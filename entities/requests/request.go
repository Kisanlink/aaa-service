package requests

import (
	"fmt"

	"github.com/Kisanlink/aaa-service/pkg/errors"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) error {
	return errors.NewInvalidInputError(field, nil, message)
}

// Request defines the interface that all request DTOs must implement to support
// multiple transport protocols (HTTP, GraphQL, gRPC). This interface provides
// a protocol-agnostic way to handle incoming requests across different services.
type Request interface {
	// Validate performs validation of the request data according to business rules
	// and returns an error if validation fails. Implementations should check all
	// required fields and data constraints.
	Validate() error

	// GetType returns the request type identifier which is used for request routing
	// and handling. This should return a consistent string across all protocols
	// (e.g. "CreateUser", "UpdateRole", etc.)
	GetType() string

	// GetProtocol returns the transport protocol this request came from
	// (e.g. "http", "grpc", "graphql")
	GetProtocol() string

	// GetOperation returns the specific operation being requested
	// (e.g. "query", "mutation" for GraphQL or "get", "post" for HTTP)
	GetOperation() string

	// GetVersion returns the API version this request targets
	// (e.g. "v1", "v2")
	GetVersion() string

	// GetRequestID returns the unique identifier for tracing this request
	// through the system
	GetRequestID() string

	// GetHeaders returns a map of protocol-specific headers/metadata
	// associated with the request
	GetHeaders() map[string][]string

	// GetBody returns the main payload/body of the request as a generic
	// interface{} that can be type asserted by implementations
	GetBody() interface{}

	// GetContext returns any request-scoped context data
	GetContext() map[string]interface{}

	// ToProto converts the request to a protocol buffer message format
	// for gRPC transport if applicable
	ToProto() interface{}

	// String returns a human-readable representation of the request
	// useful for logging and debugging
	String() string
}
