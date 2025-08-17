package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/pkg/models"
)

// Client represents the AAA service client
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

// Config represents client configuration
type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// NewClient creates a new AAA service client
func NewClient(config *Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Login authenticates a user
func (c *Client) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	var response models.LoginResponse
	err := c.makeRequest(ctx, "POST", "/api/v2/auth/login", req, &response)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}
	return &response, nil
}

// Register creates a new user account
func (c *Client) Register(ctx context.Context, req *models.RegisterRequest) (*models.RegisterResponse, error) {
	var response models.RegisterResponse
	err := c.makeRequest(ctx, "POST", "/api/v2/auth/register", req, &response)
	if err != nil {
		return nil, fmt.Errorf("registration failed: %w", err)
	}
	return &response, nil
}

// RefreshToken refreshes an access token
func (c *Client) RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (*models.RefreshTokenResponse, error) {
	var response models.RefreshTokenResponse
	err := c.makeRequest(ctx, "POST", "/api/v2/auth/refresh", req, &response)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}
	return &response, nil
}

// GetUser retrieves user information by ID
func (c *Client) GetUser(ctx context.Context, userID string, token string) (*models.User, error) {
	var response models.User
	err := c.makeAuthenticatedRequest(ctx, "GET", fmt.Sprintf("/api/v2/users/%s", userID), nil, &response, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &response, nil
}

// CreateUser creates a new user (admin only)
func (c *Client) CreateUser(ctx context.Context, req *models.CreateUserRequest, token string) (*models.User, error) {
	var response models.User
	err := c.makeAuthenticatedRequest(ctx, "POST", "/api/v2/users", req, &response, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return &response, nil
}

// UpdateUser updates user information
func (c *Client) UpdateUser(ctx context.Context, userID string, req *models.UpdateUserRequest, token string) (*models.User, error) {
	var response models.User
	err := c.makeAuthenticatedRequest(ctx, "PUT", fmt.Sprintf("/api/v2/users/%s", userID), req, &response, token)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return &response, nil
}

// HealthCheck performs a health check
func (c *Client) HealthCheck(ctx context.Context) (map[string]interface{}, error) {
	var response map[string]interface{}
	err := c.makeRequest(ctx, "GET", "/api/v2/health", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}
	return response, nil
}

// makeRequest makes an HTTP request without authentication
func (c *Client) makeRequest(ctx context.Context, method, path string, reqBody, respBody interface{}) error {
	return c.makeAuthenticatedRequest(ctx, method, path, reqBody, respBody, "")
}

// makeAuthenticatedRequest makes an HTTP request with authentication
func (c *Client) makeAuthenticatedRequest(ctx context.Context, method, path string, reqBody, respBody interface{}, token string) error {
	var body []byte
	var err error

	if reqBody != nil {
		body, err = json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't fail the request
		}
	}()

	if resp.StatusCode >= 400 {
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			// If we can't decode the error response, create a generic one
			errorResp = map[string]interface{}{
				"error":  "request failed",
				"status": resp.StatusCode,
			}
		}
		return fmt.Errorf("request failed with status %d: %v", resp.StatusCode, errorResp)
	}

	if respBody != nil {
		err = json.NewDecoder(resp.Body).Decode(respBody)
		if err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
