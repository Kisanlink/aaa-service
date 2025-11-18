package kyc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	// API endpoints
	otpGenerateEndpoint = "/kyc/aadhaar/okyc/otp"
	otpVerifyEndpoint   = "/kyc/aadhaar/okyc/otp/verify"

	// Entity types as per Sandbox API
	otpRequestEntity = "in.co.sandbox.kyc.aadhaar.okyc.otp.request"
	otpVerifyEntity  = "in.co.sandbox.kyc.aadhaar.okyc.request"

	// Retry configuration
	maxRetries     = 3
	initialBackoff = 1 * time.Second
	requestTimeout = 10 * time.Second
)

// SandboxClient handles communication with Sandbox.co.in API
type SandboxClient struct {
	baseURL   string
	apiKey    string
	apiSecret string
	client    *http.Client
	logger    *zap.Logger
}

// NewSandboxClient creates a new Sandbox API client
func NewSandboxClient(baseURL, apiKey, apiSecret string, logger *zap.Logger) *SandboxClient {
	return &SandboxClient{
		baseURL:   strings.TrimSuffix(baseURL, "/"),
		apiKey:    apiKey,
		apiSecret: apiSecret,
		client:    &http.Client{Timeout: requestTimeout},
		logger:    logger,
	}
}

// GenerateOTP sends OTP to Aadhaar-linked mobile number
func (s *SandboxClient) GenerateOTP(ctx context.Context, aadhaarNumber, consent, authToken string) (*SandboxOTPResponse, error) {
	startTime := time.Now()

	// Create request payload
	reqBody := SandboxOTPRequest{
		Entity:        otpRequestEntity,
		AadhaarNumber: aadhaarNumber,
		Consent:       consent,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		s.logger.Error("Failed to marshal OTP request",
			zap.Error(err),
			zap.String("aadhaar_masked", maskAadhaar(aadhaarNumber)))
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Log request (mask sensitive data)
	s.logger.Info("Sending OTP generation request to Sandbox API",
		zap.String("endpoint", otpGenerateEndpoint),
		zap.String("method", http.MethodPost),
		zap.String("aadhaar_masked", maskAadhaar(aadhaarNumber)))

	// Execute request with retry logic
	url := s.baseURL + otpGenerateEndpoint
	resp, err := s.doRequest(ctx, http.MethodPost, url, jsonData, authToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Log response time
	responseTime := time.Since(startTime).Milliseconds()
	s.logger.Info("Received response from Sandbox API",
		zap.Int("status_code", resp.StatusCode),
		zap.Int64("response_time_ms", responseTime),
		zap.String("endpoint", otpGenerateEndpoint))

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Failed to read response body",
			zap.Error(err))
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		return nil, s.handleErrorResponse(resp.StatusCode, body)
	}

	// Parse success response
	var otpResp SandboxOTPResponse
	if err := json.Unmarshal(body, &otpResp); err != nil {
		s.logger.Error("Failed to parse OTP response",
			zap.Error(err),
			zap.String("response", string(body)))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	s.logger.Info("OTP generated successfully",
		zap.String("transaction_id", otpResp.TransactionID),
		zap.Int("reference_id", otpResp.Data.ReferenceID))

	return &otpResp, nil
}

// VerifyOTP verifies the OTP and retrieves KYC data
func (s *SandboxClient) VerifyOTP(ctx context.Context, referenceID, otp, authToken string) (*SandboxVerifyResponse, error) {
	startTime := time.Now()

	// Create request payload
	reqBody := SandboxVerifyRequest{
		Entity:      otpVerifyEntity,
		ReferenceID: referenceID,
		OTP:         otp,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		s.logger.Error("Failed to marshal verify request",
			zap.Error(err),
			zap.String("reference_id", referenceID))
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Log request (never log OTP)
	s.logger.Info("Sending OTP verification request to Sandbox API",
		zap.String("endpoint", otpVerifyEndpoint),
		zap.String("method", http.MethodPost),
		zap.String("reference_id", referenceID))

	// Execute request with retry logic
	url := s.baseURL + otpVerifyEndpoint
	resp, err := s.doRequest(ctx, http.MethodPost, url, jsonData, authToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Log response time
	responseTime := time.Since(startTime).Milliseconds()
	s.logger.Info("Received response from Sandbox API",
		zap.Int("status_code", resp.StatusCode),
		zap.Int64("response_time_ms", responseTime),
		zap.String("endpoint", otpVerifyEndpoint))

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Failed to read response body",
			zap.Error(err))
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		return nil, s.handleErrorResponse(resp.StatusCode, body)
	}

	// Parse success response
	var verifyResp SandboxVerifyResponse
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		s.logger.Error("Failed to parse verify response",
			zap.Error(err),
			zap.String("response", string(body)))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	s.logger.Info("OTP verified successfully",
		zap.String("transaction_id", verifyResp.TransactionID),
		zap.String("name", verifyResp.Data.Name),
		zap.String("status", verifyResp.Data.Status))

	return &verifyResp, nil
}

// setHeaders sets common headers for Sandbox API requests
func (s *SandboxClient) setHeaders(req *http.Request, authToken string) {
	req.Header.Set("accept", "application/json")
	req.Header.Set("x-api-version", "2.0")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)

	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}
}

// doRequest executes HTTP request with retry logic and exponential backoff
func (s *SandboxClient) doRequest(ctx context.Context, method, url string, body []byte, authToken string) (*http.Response, error) {
	var resp *http.Response
	var err error

	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Wait for backoff period before retry (skip on first attempt)
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("request cancelled: %w", ctx.Err())
			case <-time.After(backoff):
				backoff *= 2 // Exponential backoff
			}

			s.logger.Warn("Retrying request",
				zap.Int("attempt", attempt+1),
				zap.Int("max_retries", maxRetries),
				zap.String("url", url))
		}

		// Create fresh request for each attempt (body needs to be recreated)
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Enable request body to be re-read for retries (required for connection reuse)
		if len(body) > 0 {
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(body)), nil
			}
		}

		// Prevent connection reuse for retries to avoid body consumption issues
		if attempt > 0 {
			req.Close = true
		}

		// Set headers
		s.setHeaders(req, authToken)

		// Execute request
		resp, err = s.client.Do(req)

		// Success or client error (4xx) - don't retry
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		// Log failure
		if err != nil {
			s.logger.Warn("Request failed",
				zap.Int("attempt", attempt+1),
				zap.Int("max_retries", maxRetries),
				zap.Error(err))
		} else {
			s.logger.Warn("Request failed with server error",
				zap.Int("attempt", attempt+1),
				zap.Int("max_retries", maxRetries),
				zap.Int("status_code", resp.StatusCode))
			resp.Body.Close()
		}
	}

	// All retries exhausted
	if err != nil {
		return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, err)
	}

	if resp != nil {
		return nil, fmt.Errorf("request failed after %d attempts with status %d", maxRetries, resp.StatusCode)
	}

	return nil, fmt.Errorf("request failed after %d attempts", maxRetries)
}

// handleErrorResponse processes error responses from Sandbox API
func (s *SandboxClient) handleErrorResponse(statusCode int, body []byte) error {
	// Try to parse as error response
	var errResp SandboxErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		// If parsing fails, return raw error
		s.logger.Error("Failed to parse error response",
			zap.Int("status_code", statusCode),
			zap.String("response", string(body)))
		return fmt.Errorf("API error (status %d): %s", statusCode, string(body))
	}

	// Log the error
	s.logger.Error("Sandbox API error",
		zap.Int("status_code", statusCode),
		zap.String("error", errResp.Error),
		zap.String("message", errResp.Message),
		zap.String("transaction_id", errResp.TransactionID))

	// Return descriptive error based on status code
	switch statusCode {
	case http.StatusBadRequest:
		return fmt.Errorf("validation error: %s", errResp.Message)
	case http.StatusUnauthorized:
		return fmt.Errorf("authentication error: invalid API credentials")
	case http.StatusForbidden:
		return fmt.Errorf("authorization error: %s", errResp.Message)
	case http.StatusNotFound:
		return fmt.Errorf("resource not found: %s", errResp.Message)
	case http.StatusTooManyRequests:
		return fmt.Errorf("rate limit exceeded: %s", errResp.Message)
	case http.StatusInternalServerError:
		return fmt.Errorf("server error: %s", errResp.Message)
	default:
		return fmt.Errorf("API error (%d): %s", statusCode, errResp.Message)
	}
}

// maskAadhaar masks Aadhaar number showing only last 4 digits
func maskAadhaar(aadhaar string) string {
	if len(aadhaar) < 4 {
		return "****"
	}
	return "XXXX-XXXX-" + aadhaar[len(aadhaar)-4:]
}
