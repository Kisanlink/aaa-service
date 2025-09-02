package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/responses"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorTestingService provides utilities for testing error scenarios and response formats
type ErrorTestingService struct {
	logger *zap.Logger
}

// NewErrorTestingService creates a new error testing service
func NewErrorTestingService(logger *zap.Logger) *ErrorTestingService {
	return &ErrorTestingService{
		logger: logger,
	}
}

// ErrorTestCase represents a test case for error handling
type ErrorTestCase struct {
	Name               string
	Error              error
	ExpectedStatusCode int
	ExpectedErrorCode  string
	ExpectedMessage    string
	ShouldLogSecurity  bool
	Context            map[string]interface{}
}

// ErrorTestResult represents the result of an error test
type ErrorTestResult struct {
	TestCase       *ErrorTestCase
	ActualResponse *responses.ErrorResponse
	StatusCode     int
	Success        bool
	Issues         []string
}

// ErrorTestSuite represents a collection of error tests
type ErrorTestSuite struct {
	Name      string
	TestCases []*ErrorTestCase
	Results   []*ErrorTestResult
}

// RunErrorTests runs a comprehensive suite of error handling tests
func (s *ErrorTestingService) RunErrorTests(ctx context.Context) (*ErrorTestSuite, error) {
	suite := &ErrorTestSuite{
		Name:      "Comprehensive Error Handling Tests",
		TestCases: s.createTestCases(),
		Results:   make([]*ErrorTestResult, 0),
	}

	s.logger.Info("Starting error handling tests", zap.Int("test_count", len(suite.TestCases)))

	for _, testCase := range suite.TestCases {
		result := s.runSingleErrorTest(testCase)
		suite.Results = append(suite.Results, result)

		if result.Success {
			s.logger.Debug("Error test passed", zap.String("test_name", testCase.Name))
		} else {
			s.logger.Warn("Error test failed",
				zap.String("test_name", testCase.Name),
				zap.Strings("issues", result.Issues))
		}
	}

	// Generate summary
	passed := 0
	failed := 0
	for _, result := range suite.Results {
		if result.Success {
			passed++
		} else {
			failed++
		}
	}

	s.logger.Info("Error handling tests completed",
		zap.Int("total", len(suite.TestCases)),
		zap.Int("passed", passed),
		zap.Int("failed", failed))

	return suite, nil
}

// createTestCases creates a comprehensive set of error test cases
func (s *ErrorTestingService) createTestCases() []*ErrorTestCase {
	return []*ErrorTestCase{
		{
			Name:               "Validation Error",
			Error:              errors.NewValidationError("Invalid input", "field1 is required", "field2 must be positive"),
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedErrorCode:  "VALIDATION_ERROR",
			ExpectedMessage:    "Invalid input: field1 is required; field2 must be positive",
			ShouldLogSecurity:  false,
		},
		{
			Name:               "Bad Request Error",
			Error:              errors.NewBadRequestError("Invalid request format", "JSON parsing failed"),
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedErrorCode:  "BAD_REQUEST",
			ExpectedMessage:    "Invalid request format: JSON parsing failed",
			ShouldLogSecurity:  false,
		},
		{
			Name:               "Unauthorized Error",
			Error:              errors.NewUnauthorizedError("Invalid credentials"),
			ExpectedStatusCode: http.StatusUnauthorized,
			ExpectedErrorCode:  "UNAUTHORIZED",
			ExpectedMessage:    "Invalid credentials",
			ShouldLogSecurity:  true,
		},
		{
			Name:               "Forbidden Error",
			Error:              errors.NewForbiddenError("Access denied"),
			ExpectedStatusCode: http.StatusForbidden,
			ExpectedErrorCode:  "FORBIDDEN",
			ExpectedMessage:    "Access denied",
			ShouldLogSecurity:  true,
		},
		{
			Name:               "Not Found Error",
			Error:              errors.NewNotFoundError("User not found"),
			ExpectedStatusCode: http.StatusNotFound,
			ExpectedErrorCode:  "NOT_FOUND",
			ExpectedMessage:    "User not found",
			ShouldLogSecurity:  false,
		},
		{
			Name:               "Conflict Error",
			Error:              errors.NewConflictError("User already exists"),
			ExpectedStatusCode: http.StatusConflict,
			ExpectedErrorCode:  "CONFLICT",
			ExpectedMessage:    "User already exists",
			ShouldLogSecurity:  false,
		},
		{
			Name:               "Internal Error",
			Error:              errors.NewInternalError(fmt.Errorf("database connection failed")),
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorCode:  "INTERNAL_ERROR",
			ExpectedMessage:    "An internal server error occurred",
			ShouldLogSecurity:  false,
		},
		{
			Name:               "Authentication Failed Error",
			Error:              errors.NewAuthenticationFailedError(),
			ExpectedStatusCode: http.StatusUnauthorized,
			ExpectedErrorCode:  "UNAUTHORIZED",
			ExpectedMessage:    "Invalid credentials",
			ShouldLogSecurity:  true,
		},
		{
			Name:               "Account Locked Error",
			Error:              errors.NewAccountLockedError(),
			ExpectedStatusCode: http.StatusUnauthorized,
			ExpectedErrorCode:  "UNAUTHORIZED",
			ExpectedMessage:    "Account temporarily locked due to multiple failed attempts",
			ShouldLogSecurity:  true,
		},
		{
			Name:               "Token Expired Error",
			Error:              errors.NewTokenExpiredError(),
			ExpectedStatusCode: http.StatusUnauthorized,
			ExpectedErrorCode:  "UNAUTHORIZED",
			ExpectedMessage:    "Token has expired",
			ShouldLogSecurity:  true,
		},
		{
			Name:               "Rate Limit Error",
			Error:              errors.NewRateLimitError("60"),
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedErrorCode:  "BAD_REQUEST",
			ExpectedMessage:    "Rate limit exceeded. Please try again later.",
			ShouldLogSecurity:  false,
		},
	}
}

// runSingleErrorTest runs a single error test case
func (s *ErrorTestingService) runSingleErrorTest(testCase *ErrorTestCase) *ErrorTestResult {
	result := &ErrorTestResult{
		TestCase: testCase,
		Success:  true,
		Issues:   make([]string, 0),
	}

	// Create error response
	requestID := fmt.Sprintf("test_%d", time.Now().UnixNano())
	errorResponse := responses.NewErrorResponseFromError(testCase.Error, requestID)
	result.ActualResponse = errorResponse
	result.StatusCode = errorResponse.GetHTTPStatusCode()

	// Validate status code
	if result.StatusCode != testCase.ExpectedStatusCode {
		result.Success = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Expected status code %d, got %d", testCase.ExpectedStatusCode, result.StatusCode))
	}

	// Validate error code
	if errorResponse.Code != testCase.ExpectedErrorCode {
		result.Success = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Expected error code %s, got %s", testCase.ExpectedErrorCode, errorResponse.Code))
	}

	// Validate message (for non-internal errors)
	if testCase.ExpectedErrorCode != "INTERNAL_ERROR" {
		if !strings.Contains(errorResponse.Message, testCase.ExpectedMessage) {
			result.Success = false
			result.Issues = append(result.Issues,
				fmt.Sprintf("Expected message to contain '%s', got '%s'", testCase.ExpectedMessage, errorResponse.Message))
		}
	}

	// Validate response structure
	if errorResponse.Success != false {
		result.Success = false
		result.Issues = append(result.Issues, "Expected success field to be false")
	}

	if errorResponse.RequestID == "" {
		result.Success = false
		result.Issues = append(result.Issues, "Expected request_id to be present")
	}

	if errorResponse.Timestamp.IsZero() {
		result.Success = false
		result.Issues = append(result.Issues, "Expected timestamp to be present")
	}

	// Validate security error handling (no sensitive details)
	if testCase.ShouldLogSecurity {
		if errorResponse.Details != nil && len(errorResponse.Details) > 0 {
			result.Success = false
			result.Issues = append(result.Issues, "Security errors should not include details")
		}
	}

	return result
}

// TestErrorResponseFormat tests the format of error responses
func (s *ErrorTestingService) TestErrorResponseFormat(ctx context.Context) error {
	s.logger.Info("Testing error response format consistency")

	// Test JSON serialization
	testError := errors.NewValidationError("Test error", "field1 is required")
	errorResponse := responses.NewErrorResponseFromError(testError, "test_123")

	jsonData := errorResponse.ToJSON()

	// Validate required fields
	requiredFields := []string{"success", "error", "message", "code", "timestamp", "request_id"}
	for _, field := range requiredFields {
		if _, exists := jsonData[field]; !exists {
			return fmt.Errorf("required field '%s' missing from JSON response", field)
		}
	}

	// Validate field types
	if success, ok := jsonData["success"].(bool); !ok || success {
		return fmt.Errorf("success field should be boolean false")
	}

	if _, ok := jsonData["error"].(string); !ok {
		return fmt.Errorf("error field should be string")
	}

	if _, ok := jsonData["message"].(string); !ok {
		return fmt.Errorf("message field should be string")
	}

	if _, ok := jsonData["code"].(string); !ok {
		return fmt.Errorf("code field should be string")
	}

	if _, ok := jsonData["timestamp"].(string); !ok {
		return fmt.Errorf("timestamp field should be string")
	}

	if _, ok := jsonData["request_id"].(string); !ok {
		return fmt.Errorf("request_id field should be string")
	}

	s.logger.Info("Error response format test passed")
	return nil
}

// TestErrorMiddleware tests error middleware functionality
func (s *ErrorTestingService) TestErrorMiddleware(ctx context.Context) error {
	s.logger.Info("Testing error middleware functionality")

	// Create a test Gin engine
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add error middleware (simplified for testing)
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test_middleware_123")
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors[0].Err
			requestID := c.GetString("request_id")
			errorResponse := responses.NewErrorResponseFromError(err, requestID)
			c.JSON(errorResponse.GetHTTPStatusCode(), errorResponse.ToJSON())
		}
	})

	// Add test routes that generate errors
	router.GET("/test/validation", func(c *gin.Context) {
		c.Error(errors.NewValidationError("Test validation error", "field is required"))
	})

	router.GET("/test/unauthorized", func(c *gin.Context) {
		c.Error(errors.NewUnauthorizedError("Test unauthorized error"))
	})

	router.GET("/test/internal", func(c *gin.Context) {
		c.Error(errors.NewInternalError(fmt.Errorf("test internal error")))
	})

	// Test validation error
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test/validation", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		return fmt.Errorf("expected status 400 for validation error, got %d", w.Code)
	}

	// Test unauthorized error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test/unauthorized", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		return fmt.Errorf("expected status 401 for unauthorized error, got %d", w.Code)
	}

	// Test internal error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test/internal", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		return fmt.Errorf("expected status 500 for internal error, got %d", w.Code)
	}

	s.logger.Info("Error middleware test passed")
	return nil
}

// GenerateErrorTestReport generates a comprehensive test report
func (s *ErrorTestingService) GenerateErrorTestReport(suite *ErrorTestSuite) string {
	var report strings.Builder

	report.WriteString(fmt.Sprintf("Error Handling Test Report: %s\n", suite.Name))
	report.WriteString(fmt.Sprintf("Generated at: %s\n\n", time.Now().Format(time.RFC3339)))

	passed := 0
	failed := 0

	for _, result := range suite.Results {
		if result.Success {
			passed++
			report.WriteString(fmt.Sprintf("✓ PASS: %s\n", result.TestCase.Name))
		} else {
			failed++
			report.WriteString(fmt.Sprintf("✗ FAIL: %s\n", result.TestCase.Name))
			for _, issue := range result.Issues {
				report.WriteString(fmt.Sprintf("  - %s\n", issue))
			}
		}
	}

	report.WriteString(fmt.Sprintf("\nSummary:\n"))
	report.WriteString(fmt.Sprintf("Total Tests: %d\n", len(suite.Results)))
	report.WriteString(fmt.Sprintf("Passed: %d\n", passed))
	report.WriteString(fmt.Sprintf("Failed: %d\n", failed))
	report.WriteString(fmt.Sprintf("Success Rate: %.2f%%\n", float64(passed)/float64(len(suite.Results))*100))

	return report.String()
}

// ValidateErrorConsistency validates that all error types follow consistent patterns
func (s *ErrorTestingService) ValidateErrorConsistency(ctx context.Context) error {
	s.logger.Info("Validating error consistency across all error types")

	errorTypes := []error{
		errors.NewValidationError("test", "detail"),
		errors.NewBadRequestError("test", "detail"),
		errors.NewUnauthorizedError("test"),
		errors.NewForbiddenError("test"),
		errors.NewNotFoundError("test"),
		errors.NewConflictError("test"),
		errors.NewInternalError(fmt.Errorf("test")),
	}

	for _, err := range errorTypes {
		errorResponse := responses.NewErrorResponseFromError(err, "test_123")

		// Validate all responses have required fields
		if errorResponse.Error == "" {
			return fmt.Errorf("error type %T missing error field", err)
		}

		if errorResponse.Message == "" {
			return fmt.Errorf("error type %T missing message field", err)
		}

		if errorResponse.Code == "" {
			return fmt.Errorf("error type %T missing code field", err)
		}

		if errorResponse.RequestID == "" {
			return fmt.Errorf("error type %T missing request_id field", err)
		}

		if errorResponse.Timestamp.IsZero() {
			return fmt.Errorf("error type %T missing timestamp field", err)
		}

		// Validate success is always false
		if errorResponse.Success {
			return fmt.Errorf("error type %T has success=true, should be false", err)
		}

		// Validate HTTP status codes are appropriate
		statusCode := errorResponse.GetHTTPStatusCode()
		if statusCode < 400 || statusCode >= 600 {
			return fmt.Errorf("error type %T has invalid HTTP status code %d", err, statusCode)
		}
	}

	s.logger.Info("Error consistency validation passed")
	return nil
}
