//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestHTTPServer_HealthEndpoint(t *testing.T) {
	baseURL := os.Getenv("INTEGRATION_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Failed to GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}
}

func TestHTTPServer_ReadyEndpoint(t *testing.T) {
	baseURL := os.Getenv("INTEGRATION_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/ready")
	if err != nil {
		t.Fatalf("Failed to GET /ready: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestHTTPServer_LiveEndpoint(t *testing.T) {
	baseURL := os.Getenv("INTEGRATION_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/live")
	if err != nil {
		t.Fatalf("Failed to GET /live: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestHTTPServer_CreateUser(t *testing.T) {
	baseURL := os.Getenv("INTEGRATION_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	payload := map[string]interface{}{
		"username": "integration_user",
		"password": "integration_pass",
		"email":    "integration@test.com",
	}
	body, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(baseURL+"/api/v2/users", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to POST /api/v1/users: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["username"] != "integration_user" {
		t.Errorf("Expected username 'integration_user', got %v", response["username"])
	}
}

func TestHTTPServer_ListUsers(t *testing.T) {
	baseURL := os.Getenv("INTEGRATION_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/api/v2/users")
	if err != nil {
		t.Fatalf("Failed to GET /api/v1/users: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}
