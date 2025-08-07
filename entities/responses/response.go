package responses

// Response defines the interface that all response DTOs must implement to support
// multiple transport protocols (HTTP, GraphQL, gRPC). This interface provides
// a protocol-agnostic way to handle outgoing responses across different services.
type Response interface {
	// GetType returns the response type identifier which is used for response routing
	// and handling. This should return a consistent string across all protocols
	// (e.g. "UserCreated", "RoleUpdated", etc.)
	GetType() string

	// IsSuccess returns whether the response indicates success
	IsSuccess() bool

	// GetProtocol returns the transport protocol this response will use
	// (e.g. "http", "grpc", "graphql")
	GetProtocol() string

	// GetOperation returns the specific operation being responded to
	// (e.g. "query", "mutation" for GraphQL or "get", "post" for HTTP)
	GetOperation() string

	// GetVersion returns the API version this response targets
	// (e.g. "v1", "v2")
	GetVersion() string

	// GetResponseID returns the unique identifier for tracing this response
	// through the system
	GetResponseID() string

	// GetHeaders returns a map of protocol-specific headers/metadata
	// associated with the response
	GetHeaders() map[string][]string

	// GetBody returns the main payload/body of the response as a generic
	// interface{} that can be type asserted by implementations
	GetBody() interface{}

	// GetContext returns any response-scoped context data
	GetContext() map[string]interface{}

	// ToProto converts the response to a protocol buffer message format
	// for gRPC transport if applicable
	ToProto() interface{}

	// String returns a human-readable representation of the response
	// useful for logging and debugging
	String() string
}
