package errors

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"go.uber.org/zap"
)

// ExampleService demonstrates how to use the error handling system
type ExampleService struct {
	logger *zap.Logger
}

// NewExampleService creates a new example service
func NewExampleService(logger *zap.Logger) *ExampleService {
	return &ExampleService{
		logger: logger,
	}
}

// CreateUser demonstrates error handling in a service method
func (s *ExampleService) CreateUser(ctx context.Context, username, password string) error {
	// Using Try to handle errors
	return Try(ctx, func() error {
		// Validate input
		if username == "" {
			return NewMissingFieldError("username")
		}

		if len(password) < 8 {
			return NewInvalidInputError("password", password, "password must be at least 8 characters")
		}

		// Simulate database operation
		if username == "admin" {
			return NewUserAlreadyExistsError("username", username)
		}

		// Simulate successful operation
		s.logger.Info("User created successfully", zap.String("username", username))
		return nil
	})
}

// GetUser demonstrates error handling with results
func (s *ExampleService) GetUser(ctx context.Context, userID string) (map[string]interface{}, error) {
	// Using TryWithResult to handle errors with return values
	return TryWithResult(ctx, func() (map[string]interface{}, error) {
		if userID == "" {
			return nil, NewMissingFieldError("user_id")
		}

		// Simulate user not found
		if userID == "nonexistent" {
			return nil, NewUserNotFoundError(userID)
		}

		// Simulate successful operation
		user := map[string]interface{}{
			"id":       userID,
			"username": "example_user",
			"status":   "active",
		}

		return user, nil
	})
}

// ProcessPayment demonstrates business rule validation
func (s *ExampleService) ProcessPayment(ctx context.Context, userID string, amount int) error {
	return Try(ctx, func() error {
		// Simulate user validation
		if userID == "blocked" {
			return NewUserBlockedError(userID)
		}

		if userID == "inactive" {
			return NewUserInactiveError(userID)
		}

		// Simulate token validation
		availableTokens := 100
		if amount > availableTokens {
			return NewInsufficientTokensError(amount, availableTokens)
		}

		// Simulate successful payment
		s.logger.Info("Payment processed successfully",
			zap.String("user_id", userID),
			zap.Int("amount", amount))

		return nil
	})
}

// ExampleHTTPHandler demonstrates HTTP error handling
func (s *ExampleService) ExampleHTTPHandler() ErrorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// Extract user ID from query parameter
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			return NewMissingFieldError("user_id")
		}

		// Extract amount from query parameter
		amountStr := r.URL.Query().Get("amount")
		if amountStr == "" {
			return NewMissingFieldError("amount")
		}

		amount, err := strconv.Atoi(amountStr)
		if err != nil {
			return NewInvalidFormatError("amount", amountStr, "integer")
		}

		// Process payment
		if err := s.ProcessPayment(r.Context(), userID, amount); err != nil {
			return err
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success", "message": "Payment processed"}`))

		return nil
	}
}

// ExampleWithRecovery demonstrates panic recovery
func (s *ExampleService) ExampleWithRecovery(ctx context.Context) error {
	defer Recover(ctx)

	// Simulate a panic
	panic("something went wrong!")

	return nil
}

// ExampleWithRecoveryResult demonstrates panic recovery with default result
func (s *ExampleService) ExampleWithRecoveryResult(ctx context.Context) string {
	return RecoverWithResult(ctx, "default_value")
}

// SetupErrorHandlers demonstrates how to set up error handlers
func SetupErrorHandlers(logger *zap.Logger) {
	// Create error handler
	handler := NewErrorHandler()

	// Register handlers for specific error types
	handler.RegisterHandler(ErrorTypeValidation, func(ctx context.Context, err error) error {
		logger.Warn("Validation error occurred", zap.Error(err))
		return err
	})

	handler.RegisterHandler(ErrorTypeNotFound, func(ctx context.Context, err error) error {
		logger.Info("Resource not found", zap.Error(err))
		return err
	})

	handler.RegisterHandler(ErrorTypeDatabaseError, func(ctx context.Context, err error) error {
		logger.Error("Database error occurred", zap.Error(err))
		// Could add retry logic here
		return err
	})

	handler.RegisterHandler(ErrorTypeInsufficientTokens, func(ctx context.Context, err error) error {
		logger.Warn("Insufficient tokens", zap.Error(err))
		// Could add notification logic here
		return err
	})

	// Set custom default handler
	handler.SetDefaultHandler(func(ctx context.Context, err error) error {
		logger.Error("Unhandled error occurred", zap.Error(err))
		return err
	})

	// Set global handlers
	SetGlobalDefaultHandler(handler.defaultHandler)
	SetGlobalErrorMiddleware(logger, handler)
}

// ExampleErrorChaining demonstrates error wrapping and chaining
func (s *ExampleService) ExampleErrorChaining(ctx context.Context) error {
	// Simulate a database operation that fails
	dbErr := fmt.Errorf("connection timeout")

	// Wrap the database error with context
	wrappedErr := Wrap(dbErr, ErrorTypeDatabaseError, "failed to create user")

	// Add additional context
	wrappedErr = wrappedErr.WithField("user_id").
		WithValue("12345").
		WithDetails(map[string]interface{}{
			"operation":   "create_user",
			"retry_count": 3,
		})

	return wrappedErr
}

// ExampleErrorTypeChecking demonstrates how to check error types
func (s *ExampleService) ExampleErrorTypeChecking(err error) {
	if IsErrorType(err, ErrorTypeValidation) {
		s.logger.Warn("Validation error detected")
	} else if IsErrorType(err, ErrorTypeNotFound) {
		s.logger.Info("Not found error detected")
	} else if IsErrorType(err, ErrorTypeDatabaseError) {
		s.logger.Error("Database error detected")
	} else {
		s.logger.Error("Unknown error type", zap.Error(err))
	}
}

// ExampleHTTPMiddleware demonstrates how to use the error middleware
func ExampleHTTPMiddleware(logger *zap.Logger) http.Handler {
	// Create error handler and middleware
	handler := NewErrorHandler()
	middleware := NewErrorMiddleware(logger, handler)

	// Create a simple mux
	mux := http.NewServeMux()

	// Add routes with error handling
	mux.HandleFunc("/users", HandleWithErrorAndRecovery(func(w http.ResponseWriter, r *http.Request) error {
		// Your handler logic here
		return NewNotImplementedError("user listing")
	}))

	// Apply middleware
	return middleware.Middleware(mux)
}
