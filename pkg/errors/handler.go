package errors

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
)

// ErrorHandler provides try-catch like functionality
type ErrorHandler struct {
	handlers map[ErrorType]ErrorHandlerFunc
	defaultHandler ErrorHandlerFunc
}

// ErrorHandlerFunc is a function that handles errors
type ErrorHandlerFunc func(ctx context.Context, err error) error

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		handlers: make(map[ErrorType]ErrorHandlerFunc),
		defaultHandler: defaultErrorHandler,
	}
}

// RegisterHandler registers a handler for a specific error type
func (eh *ErrorHandler) RegisterHandler(errorType ErrorType, handler ErrorHandlerFunc) {
	eh.handlers[errorType] = handler
}

// SetDefaultHandler sets the default error handler
func (eh *ErrorHandler) SetDefaultHandler(handler ErrorHandlerFunc) {
	eh.defaultHandler = handler
}

// Handle handles an error using the appropriate handler
func (eh *ErrorHandler) Handle(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	errorType := GetErrorType(err)
	if handler, exists := eh.handlers[errorType]; exists {
		return handler(ctx, err)
	}

	return eh.defaultHandler(ctx, err)
}

// Try executes a function and handles any errors that occur
func (eh *ErrorHandler) Try(ctx context.Context, fn func() error) error {
	defer func() {
		if r := recover(); r != nil {
			panicErr := fmt.Errorf("panic recovered: %v", r)
			eh.Handle(ctx, Wrap(panicErr, ErrorTypeInternal, "panic occurred"))
		}
	}()

	return eh.Handle(ctx, fn())
}

// TryWithResult executes a function that returns a result and handles any errors
func (eh *ErrorHandler) TryWithResult[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	defer func() {
		if r := recover(); r != nil {
			panicErr := fmt.Errorf("panic recovered: %v", r)
			eh.Handle(ctx, Wrap(panicErr, ErrorTypeInternal, "panic occurred"))
		}
	}()

	result, err := fn()
	if err != nil {
		eh.Handle(ctx, err)
		return result, err
	}

	return result, nil
}

// defaultErrorHandler is the default error handler
func defaultErrorHandler(ctx context.Context, err error) error {
	log.Printf("Unhandled error: %v", err)
	return err
}

// Global error handler instance
var globalErrorHandler = NewErrorHandler()

// RegisterGlobalHandler registers a handler for a specific error type globally
func RegisterGlobalHandler(errorType ErrorType, handler ErrorHandlerFunc) {
	globalErrorHandler.RegisterHandler(errorType, handler)
}

// SetGlobalDefaultHandler sets the global default error handler
func SetGlobalDefaultHandler(handler ErrorHandlerFunc) {
	globalErrorHandler.SetDefaultHandler(handler)
}

// Try executes a function and handles any errors using the global handler
func Try(ctx context.Context, fn func() error) error {
	return globalErrorHandler.Try(ctx, fn)
}

// TryWithResult executes a function that returns a result and handles any errors using the global handler
func TryWithResult[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	return globalErrorHandler.TryWithResult(ctx, fn)
}

// Handle handles an error using the global handler
func Handle(ctx context.Context, err error) error {
	return globalErrorHandler.Handle(ctx, err)
}

// ErrorRecovery provides recovery functionality for panics
type ErrorRecovery struct {
	handler *ErrorHandler
}

// NewErrorRecovery creates a new error recovery instance
func NewErrorRecovery(handler *ErrorHandler) *ErrorRecovery {
	return &ErrorRecovery{handler: handler}
}

// Recover recovers from a panic and handles it as an error
func (er *ErrorRecovery) Recover(ctx context.Context) {
	if r := recover(); r != nil {
		panicErr := fmt.Errorf("panic recovered: %v", r)
		stackTrace := string(debug.Stack())
		
		customErr := Wrap(panicErr, ErrorTypeInternal, "panic occurred").
			WithDetails(map[string]interface{}{
				"stack_trace": stackTrace,
			})
		
		er.handler.Handle(ctx, customErr)
	}
}

// RecoverWithResult recovers from a panic and returns a default result
func (er *ErrorRecovery) RecoverWithResult[T any](ctx context.Context, defaultValue T) T {
	if r := recover(); r != nil {
		panicErr := fmt.Errorf("panic recovered: %v", r)
		stackTrace := string(debug.Stack())
		
		customErr := Wrap(panicErr, ErrorTypeInternal, "panic occurred").
			WithDetails(map[string]interface{}{
				"stack_trace": stackTrace,
			})
		
		er.handler.Handle(ctx, customErr)
		return defaultValue
	}
	
	// This should never be reached, but Go requires a return
	var zero T
	return zero
}

// Global error recovery instance
var globalErrorRecovery = NewErrorRecovery(globalErrorHandler)

// Recover recovers from a panic using the global handler
func Recover(ctx context.Context) {
	globalErrorRecovery.Recover(ctx)
}

// RecoverWithResult recovers from a panic and returns a default result using the global handler
func RecoverWithResult[T any](ctx context.Context, defaultValue T) T {
	return globalErrorRecovery.RecoverWithResult(ctx, defaultValue)
} 