package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	// Test setup
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		origin         string
		envOrigins     string
		expectedOrigin string
		expectedStatus int
	}{
		{
			name:           "OPTIONS request with default origins echoes back origin when credentials enabled",
			method:         "OPTIONS",
			origin:         "http://localhost:3000",
			envOrigins:     "",
			expectedOrigin: "http://localhost:3000", // With credentials=true, wildcard echoes origin
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "OPTIONS request with specific origins returns matched origin",
			method:         "OPTIONS",
			origin:         "http://localhost:3000",
			envOrigins:     "http://localhost:3000,http://localhost:3001",
			expectedOrigin: "http://localhost:3000", // Only the matched origin is returned
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "GET request with default origins echoes back origin when credentials enabled",
			method:         "GET",
			origin:         "http://localhost:3000",
			envOrigins:     "",
			expectedOrigin: "http://localhost:3000", // With credentials=true, wildcard echoes origin
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST request with specific origins returns matched origin",
			method:         "POST",
			origin:         "http://localhost:3000",
			envOrigins:     "http://localhost:3000",
			expectedOrigin: "http://localhost:3000",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Request from non-allowed origin returns no origin header",
			method:         "GET",
			origin:         "http://malicious.com",
			envOrigins:     "http://localhost:3000",
			expectedOrigin: "", // Non-matching origin should not be allowed
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Wildcard pattern matches valid subdomain",
			method:         "GET",
			origin:         "https://admin.kisanlink.in",
			envOrigins:     "https://*.kisanlink.in",
			expectedOrigin: "https://admin.kisanlink.in",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Wildcard pattern matches another subdomain",
			method:         "GET",
			origin:         "https://farmers.kisanlink.in",
			envOrigins:     "https://*.kisanlink.in,https://*.agroslink.in",
			expectedOrigin: "https://farmers.kisanlink.in",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Wildcard pattern matches second domain",
			method:         "GET",
			origin:         "https://www.agroslink.in",
			envOrigins:     "https://*.kisanlink.in,https://*.agroslink.in",
			expectedOrigin: "https://www.agroslink.in",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Wildcard pattern rejects non-matching domain",
			method:         "GET",
			origin:         "https://evil.com",
			envOrigins:     "https://*.kisanlink.in",
			expectedOrigin: "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Wildcard pattern rejects nested subdomain",
			method:         "GET",
			origin:         "https://a.b.kisanlink.in",
			envOrigins:     "https://*.kisanlink.in",
			expectedOrigin: "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Wildcard pattern rejects wrong scheme",
			method:         "GET",
			origin:         "http://admin.kisanlink.in",
			envOrigins:     "https://*.kisanlink.in",
			expectedOrigin: "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Wildcard with exact origin mix - exact match works",
			method:         "GET",
			origin:         "http://localhost:3000",
			envOrigins:     "https://*.kisanlink.in,http://localhost:3000",
			expectedOrigin: "http://localhost:3000",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OPTIONS preflight with wildcard pattern",
			method:         "OPTIONS",
			origin:         "https://admin.kisanlink.in",
			envOrigins:     "https://*.kisanlink.in",
			expectedOrigin: "https://admin.kisanlink.in",
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable for test
			if tt.envOrigins != "" {
				_ = os.Setenv("AAA_CORS_ALLOWED_ORIGINS", tt.envOrigins)
				defer func() { _ = os.Unsetenv("AAA_CORS_ALLOWED_ORIGINS") }()
			} else {
				_ = os.Unsetenv("AAA_CORS_ALLOWED_ORIGINS")
			}

			// Create router with CORS middleware
			router := gin.New()
			router.Use(CORS())
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})
			router.POST("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})
			router.OPTIONS("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			// Create request
			req, _ := http.NewRequest(tt.method, "/test", nil)
			req.Header.Set("Origin", tt.origin)
			if tt.method == "OPTIONS" {
				req.Header.Set("Access-Control-Request-Method", "POST")
				req.Header.Set("Access-Control-Request-Headers", "Content-Type")
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve request
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedOrigin, w.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))

			if tt.method == "OPTIONS" {
				assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "OPTIONS")
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
			}

			// Verify Vary: Origin header is always present
			assert.Equal(t, "Origin", w.Header().Get("Vary"))
		})
	}
}

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		pattern string
		want    bool
	}{
		{"exact match", "https://admin.kisanlink.in", "https://admin.kisanlink.in", true},
		{"exact mismatch", "https://admin.kisanlink.in", "https://other.kisanlink.in", false},
		{"wildcard match", "https://admin.kisanlink.in", "https://*.kisanlink.in", true},
		{"wildcard match www", "https://www.kisanlink.in", "https://*.kisanlink.in", true},
		{"wildcard rejects nested", "https://a.b.kisanlink.in", "https://*.kisanlink.in", false},
		{"wildcard rejects wrong scheme", "http://admin.kisanlink.in", "https://*.kisanlink.in", false},
		{"wildcard rejects bare domain", "https://kisanlink.in", "https://*.kisanlink.in", false},
		{"wildcard rejects empty subdomain", "https://.kisanlink.in", "https://*.kisanlink.in", false},
		{"wildcard rejects different domain", "https://admin.evil.com", "https://*.kisanlink.in", false},
		{"wildcard rejects slash in subdomain", "https://admin/x.kisanlink.in", "https://*.kisanlink.in", false},
		{"no wildcard exact match", "https://localhost:3000", "https://localhost:3000", true},
		{"no wildcard mismatch", "https://localhost:3001", "https://localhost:3000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isOriginAllowed(tt.origin, tt.pattern)
			assert.Equal(t, tt.want, got)
		})
	}
}
