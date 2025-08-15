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
	return errors.NewValidationError(fmt.Sprintf("validation error for field '%s': %s", field, message))
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
	GetBody() any

	// GetContext returns any request-scoped context data
	GetContext() map[string]any

	// ToProto converts the request to a protocol buffer message format
	// for gRPC transport if applicable
	ToProto() any

	// String returns a human-readable representation of the request
	// useful for logging and debugging
	String() string
}

// BaseRequest provides a concrete implementation of the Request interface
// that other request types can embed
type BaseRequest struct {
	Protocol  string              `json:"protocol"`
	Operation string              `json:"operation"`
	Version   string              `json:"version"`
	RequestID string              `json:"request_id"`
	Headers   map[string][]string `json:"headers"`
	Body      any                 `json:"body"`
	Context   map[string]any      `json:"context"`
	Type      string              `json:"type"`
}

// NewBaseRequest creates a new BaseRequest
func NewBaseRequest(
	protocol, operation, version, requestID, requestType string,
	headers map[string][]string,
	body any,
	context map[string]any,
) *BaseRequest {
	return &BaseRequest{
		Protocol:  protocol,
		Operation: operation,
		Version:   version,
		RequestID: requestID,
		Headers:   headers,
		Body:      body,
		Context:   context,
		Type:      requestType,
	}
}

// Validate implements the Request interface
func (br *BaseRequest) Validate() error {
	return nil // Base validation is empty, specific requests override this
}

// GetType implements the Request interface
func (br *BaseRequest) GetType() string {
	return br.Type
}

// GetProtocol implements the Request interface
func (br *BaseRequest) GetProtocol() string {
	return br.Protocol
}

// GetOperation implements the Request interface
func (br *BaseRequest) GetOperation() string {
	return br.Operation
}

// GetVersion implements the Request interface
func (br *BaseRequest) GetVersion() string {
	return br.Version
}

// GetRequestID implements the Request interface
func (br *BaseRequest) GetRequestID() string {
	return br.RequestID
}

// GetHeaders implements the Request interface
func (br *BaseRequest) GetHeaders() map[string][]string {
	return br.Headers
}

// GetBody implements the Request interface
func (br *BaseRequest) GetBody() any {
	return br.Body
}

// GetContext implements the Request interface
func (br *BaseRequest) GetContext() map[string]any {
	return br.Context
}

// ToProto implements the Request interface
func (br *BaseRequest) ToProto() any {
	return nil // Default implementation returns nil, specific requests can override
}

// String implements the Request interface
func (br *BaseRequest) String() string {
	return fmt.Sprintf("Request{Type: %s, Protocol: %s, Operation: %s, Version: %s, RequestID: %s}",
		br.Type, br.Protocol, br.Operation, br.Version, br.RequestID)
}
