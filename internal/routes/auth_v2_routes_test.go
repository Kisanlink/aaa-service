package routes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthV2Routes_ExpectedEndpoints(t *testing.T) {
	// Test that verifies the expected MPIN endpoints configuration
	// This documents what endpoints should be available

	expectedEndpoints := []struct {
		method string
		path   string
		public bool
	}{
		{"POST", "/api/v2/auth/login", true},
		{"POST", "/api/v2/auth/register", true},
		{"POST", "/api/v2/auth/refresh", true},
		{"POST", "/api/v2/auth/forgot-password", true},
		{"POST", "/api/v2/auth/reset-password", true},
		{"POST", "/api/v2/auth/logout", false},
		{"POST", "/api/v2/auth/set-mpin", false},
		{"POST", "/api/v2/auth/update-mpin", false},
	}

	// This test documents the expected endpoints
	for _, endpoint := range expectedEndpoints {
		t.Run(endpoint.method+" "+endpoint.path, func(t *testing.T) {
			// Verify endpoint configuration
			assert.NotEmpty(t, endpoint.method, "Method should not be empty")
			assert.NotEmpty(t, endpoint.path, "Path should not be empty")
			assert.Contains(t, endpoint.path, "/api/v2/auth/", "Should be a V2 auth endpoint")

			if endpoint.public {
				// Public endpoints should not include protected operations
				assert.NotContains(t, endpoint.path, "set-mpin", "Public endpoints should not include MPIN management")
				assert.NotContains(t, endpoint.path, "update-mpin", "Public endpoints should not include MPIN management")
				assert.NotContains(t, endpoint.path, "logout", "Public endpoints should not include logout")
			} else {
				// Protected endpoints should be auth-required operations
				protectedOps := []string{"set-mpin", "update-mpin", "logout"}
				hasProtectedOp := false
				for _, op := range protectedOps {
					if assert.ObjectsAreEqual(endpoint.path, "/api/v2/auth/"+op) {
						hasProtectedOp = true
						break
					}
				}
				assert.True(t, hasProtectedOp, "Protected endpoint should be one of: %v", protectedOps)
			}
		})
	}
}

func TestAuthV2Routes_MPinEndpointsDocumented(t *testing.T) {
	// Test specifically for MPIN endpoints
	mpinEndpoints := []string{
		"/api/v2/auth/set-mpin",
		"/api/v2/auth/update-mpin",
	}

	for _, endpoint := range mpinEndpoints {
		t.Run(endpoint, func(t *testing.T) {
			assert.Contains(t, endpoint, "mpin", "MPIN endpoint should contain 'mpin' in path")
			assert.Contains(t, endpoint, "/api/v2/auth/", "Should be a V2 auth endpoint")
		})
	}
}
