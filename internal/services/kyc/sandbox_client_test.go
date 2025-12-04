package kyc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

// Test data constants
const (
	testAadhaar       = "123456789012"
	testOTP           = "123456"
	testReferenceID   = "123456789"
	testTransactionID = "TXN1234567890"
	testAuthToken     = "Bearer test-token"
)

// mockOTPResponse creates a mock successful OTP generation response
func mockOTPResponse() SandboxOTPResponse {
	return SandboxOTPResponse{
		Timestamp:     time.Now().Unix(),
		TransactionID: testTransactionID,
		Code:          200,
		Data: OTPResponseData{
			Entity:      "in.co.sandbox.kyc.aadhaar.okyc.otp.response",
			Message:     "OTP sent successfully",
			ReferenceID: 123456789,
		},
	}
}

// mockVerifyResponse creates a mock successful OTP verification response
func mockVerifyResponse() SandboxVerifyResponse {
	return SandboxVerifyResponse{
		Timestamp:     time.Now().Unix(),
		TransactionID: testTransactionID,
		Code:          200,
		Data: KYCData{
			Entity:      "in.co.sandbox.kyc.aadhaar.okyc.response",
			Name:        "John Doe",
			Gender:      "M",
			DateOfBirth: "1990-01-01",
			YOB:         1990,
			CareOf:      "S/O Father Name",
			FullAddress: "House 123, Street ABC, City XYZ, State, 123456",
			Address: SandboxAddress{
				Entity:      "in.co.sandbox.kyc.aadhaar.address",
				House:       "123",
				Street:      "ABC Street",
				Landmark:    "Near Park",
				Locality:    "Locality",
				Vtc:         "VTC",
				Subdistrict: "Subdistrict",
				District:    "District Name",
				State:       "State Name",
				Pincode:     123456,
				PostOffice:  "Post Office",
				Country:     "India",
			},
			Photo:     "base64encodedphotodata...",
			ShareCode: "1234",
			Status:    "VALID",
			Message:   "Aadhaar Card Exists",
		},
	}
}

// mockErrorResponse creates a mock error response
func mockErrorResponse(statusCode int, message string) SandboxErrorResponse {
	return SandboxErrorResponse{
		Timestamp:     time.Now().Unix(),
		TransactionID: testTransactionID,
		Code:          statusCode,
		Message:       message,
		Error:         http.StatusText(statusCode),
	}
}

// TestGenerateOTP_Success tests successful OTP generation
func TestGenerateOTP_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/kyc/aadhaar/okyc/otp") {
			t.Errorf("Expected OTP endpoint, got %s", r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("accept") != "application/json" {
			t.Error("Missing or incorrect accept header")
		}
		if r.Header.Get("x-api-version") != "2.0" {
			t.Error("Missing or incorrect x-api-version header")
		}
		if r.Header.Get("x-api-key") != "test-api-key" {
			t.Error("Missing or incorrect x-api-key header")
		}
		if r.Header.Get("Authorization") != testAuthToken {
			t.Error("Missing or incorrect Authorization header")
		}

		// Verify request body
		var reqBody SandboxOTPRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if reqBody.AadhaarNumber != testAadhaar {
			t.Errorf("Expected Aadhaar %s, got %s", testAadhaar, reqBody.AadhaarNumber)
		}
		if reqBody.Consent != "Y" {
			t.Errorf("Expected consent Y, got %s", reqBody.Consent)
		}

		// Send success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockOTPResponse())
	}))
	defer server.Close()

	// Create client
	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	// Execute request
	ctx := context.Background()
	resp, err := client.GenerateOTP(ctx, testAadhaar, "Y", testAuthToken)

	// Verify response
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.TransactionID != testTransactionID {
		t.Errorf("Expected transaction ID %s, got %s", testTransactionID, resp.TransactionID)
	}
	if resp.Data.ReferenceID != 123456789 {
		t.Errorf("Expected reference ID 123456789, got %d", resp.Data.ReferenceID)
	}
}

// TestGenerateOTP_InvalidAadhaar tests OTP generation with invalid Aadhaar
func TestGenerateOTP_InvalidAadhaar(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(mockErrorResponse(400, "Invalid Aadhaar number"))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	ctx := context.Background()
	_, err := client.GenerateOTP(ctx, "invalid", "Y", testAuthToken)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "validation error") {
		t.Errorf("Expected validation error, got %v", err)
	}
}

// TestGenerateOTP_NetworkError tests OTP generation with network error
func TestGenerateOTP_NetworkError(t *testing.T) {
	logger := zap.NewNop()
	client := NewSandboxClient("http://invalid-url-that-does-not-exist.local", "test-api-key", "test-api-secret", logger)

	ctx := context.Background()
	_, err := client.GenerateOTP(ctx, testAadhaar, "Y", testAuthToken)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestGenerateOTP_Timeout tests OTP generation with timeout
func TestGenerateOTP_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(2 * time.Second)
		_ = json.NewEncoder(w).Encode(mockOTPResponse())
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	// Set short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.GenerateOTP(ctx, testAadhaar, "Y", testAuthToken)

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}

// TestGenerateOTP_RetrySuccess tests retry logic with eventual success
func TestGenerateOTP_RetrySuccess(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// Fail first two attempts
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(mockErrorResponse(500, "Internal server error"))
			return
		}
		// Succeed on third attempt
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockOTPResponse())
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	ctx := context.Background()
	resp, err := client.GenerateOTP(ctx, testAadhaar, "Y", testAuthToken)

	if err != nil {
		t.Fatalf("Expected success after retry, got error: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

// TestVerifyOTP_Success tests successful OTP verification
func TestVerifyOTP_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle authentication endpoint
		if strings.Contains(r.URL.Path, "/authenticate") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(SandboxAuthResponse{
				AccessToken: "test-access-token",
				ExpiresIn:   86400,
			})
			return
		}

		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/kyc/aadhaar/okyc/otp/verify") {
			t.Errorf("Expected verify endpoint, got %s", r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("accept") != "application/json" {
			t.Error("Missing or incorrect accept header")
		}
		if r.Header.Get("x-api-version") != "2.0" {
			t.Error("Missing or incorrect x-api-version header")
		}
		if r.Header.Get("x-api-key") != "test-api-key" {
			t.Error("Missing or incorrect x-api-key header")
		}

		// Verify request body
		var reqBody SandboxVerifyRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if reqBody.ReferenceID != testReferenceID {
			t.Errorf("Expected reference ID %s, got %s", testReferenceID, reqBody.ReferenceID)
		}
		if reqBody.OTP != testOTP {
			t.Errorf("Expected OTP %s, got %s", testOTP, reqBody.OTP)
		}

		// Send success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockVerifyResponse())
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	ctx := context.Background()
	resp, err := client.VerifyOTP(ctx, testReferenceID, testOTP, testAuthToken)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.Data.Name != "John Doe" {
		t.Errorf("Expected name John Doe, got %s", resp.Data.Name)
	}
	if resp.Data.Status != "VALID" {
		t.Errorf("Expected status VALID, got %s", resp.Data.Status)
	}
	if resp.Data.Address.State != "State Name" {
		t.Errorf("Expected state State Name, got %s", resp.Data.Address.State)
	}
}

// TestVerifyOTP_InvalidOTP tests OTP verification with invalid OTP
func TestVerifyOTP_InvalidOTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(mockErrorResponse(400, "Invalid OTP"))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	ctx := context.Background()
	_, err := client.VerifyOTP(ctx, testReferenceID, "wrong-otp", testAuthToken)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "validation error") {
		t.Errorf("Expected validation error, got %v", err)
	}
}

// TestVerifyOTP_ExpiredOTP tests OTP verification with expired OTP
func TestVerifyOTP_ExpiredOTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(mockErrorResponse(400, "OTP expired"))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	ctx := context.Background()
	_, err := client.VerifyOTP(ctx, testReferenceID, testOTP, testAuthToken)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "validation error") {
		t.Errorf("Expected validation error, got %v", err)
	}
}

// TestVerifyOTP_NetworkError tests OTP verification with network error
func TestVerifyOTP_NetworkError(t *testing.T) {
	logger := zap.NewNop()
	client := NewSandboxClient("http://invalid-url-that-does-not-exist.local", "test-api-key", "test-api-secret", logger)

	ctx := context.Background()
	_, err := client.VerifyOTP(ctx, testReferenceID, testOTP, testAuthToken)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestVerifyOTP_APIError tests OTP verification with API error
func TestVerifyOTP_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(mockErrorResponse(500, "Internal server error"))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	ctx := context.Background()
	_, err := client.VerifyOTP(ctx, testReferenceID, testOTP, testAuthToken)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	// Server errors (5xx) are retried 3 times, so we expect "request failed after 3 attempts"
	if !strings.Contains(err.Error(), "request failed") {
		t.Errorf("Expected request failed error, got %v", err)
	}
}

// TestVerifyOTP_RetrySuccess tests retry logic with eventual success
func TestVerifyOTP_RetrySuccess(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			// Fail first attempt
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(mockErrorResponse(500, "Internal server error"))
			return
		}
		// Succeed on second attempt
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockVerifyResponse())
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	ctx := context.Background()
	resp, err := client.VerifyOTP(ctx, testReferenceID, testOTP, testAuthToken)

	if err != nil {
		t.Fatalf("Expected success after retry, got error: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

// TestSetHeaders tests header setting
func TestSetHeaders(t *testing.T) {
	logger := zap.NewNop()
	client := NewSandboxClient("http://test.local", "test-api-key", "test-api-secret", logger)

	req, _ := http.NewRequest(http.MethodPost, "http://test.local", nil)
	client.setHeaders(req, testAuthToken)

	// Verify headers
	if req.Header.Get("accept") != "application/json" {
		t.Error("Missing or incorrect accept header")
	}
	if req.Header.Get("x-api-version") != "2.0" {
		t.Error("Missing or incorrect x-api-version header")
	}
	if req.Header.Get("content-type") != "application/json" {
		t.Error("Missing or incorrect content-type header")
	}
	if req.Header.Get("x-api-key") != "test-api-key" {
		t.Error("Missing or incorrect x-api-key header")
	}
	if req.Header.Get("Authorization") != testAuthToken {
		t.Error("Missing or incorrect Authorization header")
	}
}

// TestSetHeaders_NoAuthToken tests header setting without auth token
func TestSetHeaders_NoAuthToken(t *testing.T) {
	logger := zap.NewNop()
	client := NewSandboxClient("http://test.local", "test-api-key", "test-api-secret", logger)

	req, _ := http.NewRequest(http.MethodPost, "http://test.local", nil)
	client.setHeaders(req, "")

	// Authorization should not be set
	if req.Header.Get("Authorization") != "" {
		t.Error("Authorization header should not be set")
	}
}

// TestDoRequest_ContextCancellation tests context cancellation
func TestDoRequest_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(200 * time.Millisecond)
		json.NewEncoder(w).Encode(mockOTPResponse())
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	// Cancel context immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GenerateOTP(ctx, testAadhaar, "Y", testAuthToken)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "cancel") {
		t.Errorf("Expected cancellation error, got %v", err)
	}
}

// TestDoRequest_Unauthorized tests 401 Unauthorized response
func TestDoRequest_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(mockErrorResponse(401, "Invalid API key"))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "invalid-key", "test-api-secret", logger)

	ctx := context.Background()
	_, err := client.GenerateOTP(ctx, testAadhaar, "Y", testAuthToken)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "authentication error") {
		t.Errorf("Expected authentication error, got %v", err)
	}
}

// TestDoRequest_RateLimited tests 429 Too Many Requests response
func TestDoRequest_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(mockErrorResponse(429, "Rate limit exceeded"))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	ctx := context.Background()
	_, err := client.GenerateOTP(ctx, testAadhaar, "Y", testAuthToken)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("Expected rate limit error, got %v", err)
	}
}

// TestDoRequest_NotFound tests 404 Not Found response
func TestDoRequest_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(mockErrorResponse(404, "Resource not found"))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	ctx := context.Background()
	_, err := client.VerifyOTP(ctx, "invalid-ref-id", testOTP, testAuthToken)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected not found error, got %v", err)
	}
}

// TestMaskAadhaar tests Aadhaar masking function
func TestMaskAadhaar(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid 12-digit Aadhaar",
			input:    "123456789012",
			expected: "XXXX-XXXX-9012",
		},
		{
			name:     "Short Aadhaar",
			input:    "123",
			expected: "****",
		},
		{
			name:     "Empty Aadhaar",
			input:    "",
			expected: "****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskAadhaar(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestErrorResponse_MalformedJSON tests error handling with malformed JSON
func TestErrorResponse_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		// Send malformed JSON
		_, _ = fmt.Fprint(w, "{invalid json")
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	ctx := context.Background()
	_, err := client.GenerateOTP(ctx, testAadhaar, "Y", testAuthToken)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "API error") {
		t.Errorf("Expected API error, got %v", err)
	}
}

// TestNewSandboxClient tests client initialization
func TestNewSandboxClient(t *testing.T) {
	logger := zap.NewNop()
	client := NewSandboxClient("https://api.sandbox.co.in", "test-key", "test-secret", logger)

	if client == nil {
		t.Fatal("Expected client, got nil")
	}
	if client.baseURL != "https://api.sandbox.co.in" {
		t.Errorf("Expected base URL https://api.sandbox.co.in, got %s", client.baseURL)
	}
	if client.apiKey != "test-key" {
		t.Errorf("Expected API key test-key, got %s", client.apiKey)
	}
	if client.client == nil {
		t.Fatal("Expected HTTP client, got nil")
	}
}

// TestNewSandboxClient_TrailingSlash tests URL normalization
func TestNewSandboxClient_TrailingSlash(t *testing.T) {
	logger := zap.NewNop()
	client := NewSandboxClient("https://api.sandbox.co.in/", "test-key", "test-secret", logger)

	if client.baseURL != "https://api.sandbox.co.in" {
		t.Errorf("Expected base URL without trailing slash, got %s", client.baseURL)
	}
}

// TestDoRequest_Forbidden tests 403 Forbidden response
func TestDoRequest_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(mockErrorResponse(403, "Access forbidden"))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	ctx := context.Background()
	_, err := client.GenerateOTP(ctx, testAadhaar, "Y", testAuthToken)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "authorization error") {
		t.Errorf("Expected authorization error, got %v", err)
	}
}

// TestGenerateOTP_JSONMarshalError tests marshaling error
func TestGenerateOTP_ResponseParseError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Send invalid JSON
		fmt.Fprint(w, "{invalid json")
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	ctx := context.Background()
	_, err := client.GenerateOTP(ctx, testAadhaar, "Y", testAuthToken)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse response") {
		t.Errorf("Expected parse error, got %v", err)
	}
}

// TestVerifyOTP_ResponseParseError tests response parsing error
func TestVerifyOTP_ResponseParseError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Send invalid JSON
		fmt.Fprint(w, "{invalid json")
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	ctx := context.Background()
	_, err := client.VerifyOTP(ctx, testReferenceID, testOTP, testAuthToken)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse response") {
		t.Errorf("Expected parse error, got %v", err)
	}
}

// TestDoRequest_UnknownError tests unknown HTTP status code
func TestDoRequest_UnknownError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(418) // I'm a teapot - uncommon status code
		json.NewEncoder(w).Encode(mockErrorResponse(418, "I'm a teapot"))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	ctx := context.Background()
	_, err := client.GenerateOTP(ctx, testAadhaar, "Y", testAuthToken)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "API error") {
		t.Errorf("Expected API error, got %v", err)
	}
}

// TestDoRequest_ServerErrorRetry tests that 5xx errors are retried
func TestDoRequest_ServerErrorRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway) // 502 is a server error
		json.NewEncoder(w).Encode(mockErrorResponse(502, "Bad Gateway"))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSandboxClient(server.URL, "test-api-key", "test-api-secret", logger)

	ctx := context.Background()
	_, err := client.VerifyOTP(ctx, testReferenceID, testOTP, testAuthToken)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if attempts != 3 {
		t.Errorf("Expected 3 retry attempts, got %d", attempts)
	}
	if !strings.Contains(err.Error(), "request failed") {
		t.Errorf("Expected request failed error, got %v", err)
	}
}
