package kyc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"go.uber.org/zap"
)

const (
	// API endpoints
	authenticateEndpoint = "/authenticate"
	otpGenerateEndpoint  = "/kyc/aadhaar/okyc/otp"
	otpVerifyEndpoint    = "/kyc/aadhaar/okyc/otp/verify"

	// Entity types as per Sandbox API
	otpRequestEntity = "in.co.sandbox.kyc.aadhaar.okyc.otp.request"
	otpVerifyEntity  = "in.co.sandbox.kyc.aadhaar.okyc.request"

	// Retry configuration
	maxRetries     = 3
	initialBackoff = 1 * time.Second
	requestTimeout = 10 * time.Second

	// Token validity (24 hours minus 5 minutes buffer for refresh)
	tokenValiditySeconds = 86400 - 300
)

// SandboxClient handles communication with Sandbox.co.in API
type SandboxClient struct {
	baseURL     string
	apiKey      string
	apiSecret   string
	client      *http.Client
	logger      *zap.Logger
	accessToken string    // Cached access token
	tokenExpiry time.Time // When the access token expires
	tokenMutex  sync.RWMutex
}

// NewSandboxClient creates a new Sandbox API client
func NewSandboxClient(baseURL, apiKey, apiSecret string, logger *zap.Logger) *SandboxClient {
	// Log initialization (mask sensitive values for security)
	maskedKey := "NOT_SET"
	maskedSecret := "NOT_SET"
	if apiKey != "" && len(apiKey) > 8 {
		maskedKey = apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
	} else if apiKey != "" {
		maskedKey = "SET_BUT_TOO_SHORT"
	}
	if apiSecret != "" && len(apiSecret) > 8 {
		maskedSecret = apiSecret[:4] + "..." + apiSecret[len(apiSecret)-4:]
	} else if apiSecret != "" {
		maskedSecret = "SET_BUT_TOO_SHORT"
	}

	logger.Info("Initializing Sandbox API client",
		zap.String("base_url", baseURL),
		zap.String("api_key_masked", maskedKey),
		zap.String("api_secret_masked", maskedSecret),
		zap.Bool("api_key_set", apiKey != ""),
		zap.Bool("api_secret_set", apiSecret != ""))

	if baseURL == "" || apiKey == "" {
		logger.Warn("Sandbox API client initialized with missing credentials",
			zap.Bool("base_url_set", baseURL != ""),
			zap.Bool("api_key_set", apiKey != ""),
			zap.Bool("api_secret_set", apiSecret != ""))
	}

	client := &SandboxClient{
		baseURL:   strings.TrimSuffix(baseURL, "/"),
		apiKey:    apiKey,
		apiSecret: apiSecret,
		client:    &http.Client{Timeout: requestTimeout},
		logger:    logger,
	}

	// Authenticate immediately to validate credentials and get access token
	if baseURL != "" && apiKey != "" && apiSecret != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.authenticate(ctx); err != nil {
			logger.Warn("Failed to authenticate with Sandbox API on initialization",
				zap.Error(err),
				zap.String("hint", "Will retry on first API call"))
		} else {
			logger.Info("Successfully authenticated with Sandbox API")
		}
	}

	return client
}

// authenticate authenticates with Sandbox API to get an access token
func (s *SandboxClient) authenticate(ctx context.Context) error {
	s.logger.Info("Authenticating with Sandbox API")

	// Create request
	url := s.baseURL + authenticateEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	// Set auth headers (x-api-key and x-api-secret)
	req.Header.Set("accept", "application/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("x-api-secret", s.apiSecret)
	req.Header.Set("x-api-version", "2.0")

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("Failed to authenticate with Sandbox API",
			zap.Error(err))
		return fmt.Errorf("authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Sandbox API authentication failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)))
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var authResp SandboxAuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		s.logger.Error("Failed to parse auth response",
			zap.Error(err),
			zap.String("response", string(body)))
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	// Store access token
	s.tokenMutex.Lock()
	s.accessToken = authResp.AccessToken
	s.tokenExpiry = time.Now().Add(time.Duration(tokenValiditySeconds) * time.Second)
	s.tokenMutex.Unlock()

	s.logger.Info("Sandbox API authentication successful",
		zap.Time("token_expiry", s.tokenExpiry))

	return nil
}

// getAccessToken returns a valid access token, refreshing if needed
func (s *SandboxClient) getAccessToken(ctx context.Context) (string, error) {
	s.tokenMutex.RLock()
	token := s.accessToken
	expiry := s.tokenExpiry
	s.tokenMutex.RUnlock()

	// Check if token is expired or will expire soon
	if token == "" || time.Now().After(expiry) {
		s.logger.Info("Access token expired or missing, refreshing")
		if err := s.authenticate(ctx); err != nil {
			return "", err
		}

		s.tokenMutex.RLock()
		token = s.accessToken
		s.tokenMutex.RUnlock()
	}

	return token, nil
}

// GenerateOTP sends OTP to Aadhaar-linked mobile number
func (s *SandboxClient) GenerateOTP(ctx context.Context, aadhaarNumber, consent, authToken string) (*SandboxOTPResponse, error) {
	startTime := time.Now()

	// Get valid access token
	accessToken, err := s.getAccessToken(ctx)
	if err != nil {
		s.logger.Error("Failed to get Sandbox access token",
			zap.Error(err))
		return nil, errors.NewUnauthorizedError("Sandbox API authentication failed")
	}

	// Create request payload
	reqBody := SandboxOTPRequest{
		Entity:        otpRequestEntity,
		AadhaarNumber: aadhaarNumber,
		Consent:       consent,
		Reason:        "User KYC verification", // Required by Sandbox API
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

	// Execute request with retry logic using Sandbox access token
	url := s.baseURL + otpGenerateEndpoint
	resp, err := s.doRequest(ctx, http.MethodPost, url, jsonData, accessToken)
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

	// Get valid access token
	accessToken, err := s.getAccessToken(ctx)
	if err != nil {
		s.logger.Error("Failed to get Sandbox access token",
			zap.Error(err))
		return nil, errors.NewUnauthorizedError("Sandbox API authentication failed")
	}

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

	// Execute request with retry logic using Sandbox access token
	url := s.baseURL + otpVerifyEndpoint
	resp, err := s.doRequest(ctx, http.MethodPost, url, jsonData, accessToken)
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
// Per official docs: https://developer.sandbox.co.in/reference/aadhaar-okyc-generate-otp-api
func (s *SandboxClient) setHeaders(req *http.Request, sandboxAccessToken string) {
	req.Header.Set("accept", "application/json")
	req.Header.Set("x-api-version", "2.0")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)

	// Authorization header: Sandbox access token (NOT our JWT!)
	// Per docs: "Authorization: {{access_token}}" (no "Bearer" prefix)
	if sandboxAccessToken != "" {
		req.Header.Set("Authorization", sandboxAccessToken)
	}

	// Log headers being sent (excluding sensitive values)
	s.logger.Debug("Setting Sandbox API request headers",
		zap.String("url", req.URL.String()),
		zap.Bool("x-api-key_set", s.apiKey != ""),
		zap.Bool("authorization_set", sandboxAccessToken != ""))
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
		return errors.NewInternalError(fmt.Errorf("sandbox API error (status %d): %s", statusCode, string(body)))
	}

	// Log the error with helpful context
	s.logger.Error("Sandbox API error",
		zap.Int("status_code", statusCode),
		zap.String("error", errResp.Error),
		zap.String("message", errResp.Message),
		zap.String("transaction_id", errResp.TransactionID))

	// Return typed errors based on status code for proper error handling
	switch statusCode {
	case http.StatusBadRequest:
		return errors.NewBadRequestError(
			fmt.Sprintf("Sandbox API validation error: %s", errResp.Message),
		)
	case http.StatusUnauthorized:
		// API key or credentials are invalid
		s.logger.Error("Sandbox API authentication failed - check AADHAAR_SANDBOX_API_KEY and AADHAAR_SANDBOX_API_SECRET",
			zap.String("hint", "Verify your Sandbox.co.in API credentials are correct"))
		return errors.NewUnauthorizedError(
			"Sandbox API authentication failed - invalid API credentials. Please verify AADHAAR_SANDBOX_API_KEY and AADHAAR_SANDBOX_API_SECRET environment variables.",
		)
	case http.StatusForbidden:
		// API key is valid but doesn't have permission for this operation
		s.logger.Error("Sandbox API authorization failed - insufficient privileges",
			zap.String("message", errResp.Message),
			zap.String("hint", "Your API key may not have permission to access Aadhaar OTP endpoints. Contact Sandbox.co.in support or check your API key permissions."))
		return errors.NewForbiddenError(
			fmt.Sprintf("Sandbox API access denied: %s. Your API key may not have permission for Aadhaar OTP operations. Please verify your Sandbox.co.in account permissions.", errResp.Message),
		)
	case http.StatusNotFound:
		return errors.NewNotFoundError(
			fmt.Sprintf("Sandbox API resource not found: %s", errResp.Message),
		)
	case http.StatusTooManyRequests:
		return errors.NewBadRequestError(
			fmt.Sprintf("Sandbox API rate limit exceeded: %s", errResp.Message),
		)
	case http.StatusInternalServerError:
		return errors.NewInternalError(
			fmt.Errorf("sandbox API server error: %s", errResp.Message),
		)
	default:
		return errors.NewInternalError(
			fmt.Errorf("sandbox API error (%d): %s", statusCode, errResp.Message),
		)
	}
}

// maskAadhaar masks Aadhaar number showing only last 4 digits
func maskAadhaar(aadhaar string) string {
	if len(aadhaar) < 4 {
		return "****"
	}
	return "XXXX-XXXX-" + aadhaar[len(aadhaar)-4:]
}
