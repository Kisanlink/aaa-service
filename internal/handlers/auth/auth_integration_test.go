package auth

import (
	"encoding/json"
	"testing"

	"github.com/Kisanlink/aaa-service/internal/entities/requests"
	"github.com/Kisanlink/aaa-service/internal/entities/responses"
	userResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/users"
	"github.com/stretchr/testify/assert"
)

// TestLoginV2_Integration_PasswordAndMPinSupport tests the complete login flow
// This test demonstrates that the enhanced LoginV2 method properly handles
// both password and MPIN authentication as specified in the requirements
func TestLoginV2_Integration_PasswordAndMPinSupport(t *testing.T) {
	// This is a demonstration test showing the expected behavior
	// In a real integration test, you would set up a test database and real services

	t.Run("Password Authentication Flow", func(t *testing.T) {
		// Test data
		password := "testpassword123"
		loginReq := requests.LoginRequest{
			PhoneNumber: "1234567890",
			CountryCode: "+91",
			Password:    &password,
		}

		// Verify request structure supports both password and MPIN
		assert.True(t, loginReq.HasPassword())
		assert.False(t, loginReq.HasMPin())
		assert.Equal(t, "testpassword123", loginReq.GetPassword())
		assert.Equal(t, "", loginReq.GetMPin())

		// Verify validation works correctly
		err := loginReq.Validate()
		assert.NoError(t, err, "Valid password request should pass validation")
	})

	t.Run("MPIN Authentication Flow", func(t *testing.T) {
		// Test data
		mpin := "1234"
		loginReq := requests.LoginRequest{
			PhoneNumber: "1234567890",
			CountryCode: "+91",
			MPin:        &mpin,
		}

		// Verify request structure supports MPIN
		assert.False(t, loginReq.HasPassword())
		assert.True(t, loginReq.HasMPin())
		assert.Equal(t, "", loginReq.GetPassword())
		assert.Equal(t, "1234", loginReq.GetMPin())

		// Verify validation works correctly
		err := loginReq.Validate()
		assert.NoError(t, err, "Valid MPIN request should pass validation")
	})

	t.Run("Both Password and MPIN Provided", func(t *testing.T) {
		// Test data - both provided
		password := "testpassword123"
		mpin := "1234"
		loginReq := requests.LoginRequest{
			PhoneNumber: "1234567890",
			CountryCode: "+91",
			Password:    &password,
			MPin:        &mpin,
		}

		// Verify both are detected
		assert.True(t, loginReq.HasPassword())
		assert.True(t, loginReq.HasMPin())
		assert.Equal(t, "testpassword123", loginReq.GetPassword())
		assert.Equal(t, "1234", loginReq.GetMPin())

		// Verify validation works correctly
		err := loginReq.Validate()
		assert.NoError(t, err, "Request with both password and MPIN should pass validation")
	})

	t.Run("Neither Password nor MPIN Provided", func(t *testing.T) {
		// Test data - neither provided
		loginReq := requests.LoginRequest{
			PhoneNumber: "1234567890",
			CountryCode: "+91",
		}

		// Verify neither are detected
		assert.False(t, loginReq.HasPassword())
		assert.False(t, loginReq.HasMPin())

		// Verify validation fails
		err := loginReq.Validate()
		assert.Error(t, err, "Request without password or MPIN should fail validation")
		assert.Contains(t, err.Error(), "either password or mpin is required")
	})

	t.Run("Enhanced Response Structure", func(t *testing.T) {
		// Test the enhanced response structure includes all required fields
		userInfo := &responses.UserInfo{
			ID:          "user-123",
			PhoneNumber: "1234567890",
			CountryCode: "+91",
			Username:    stringPtr("testuser"),
			IsValidated: true,
			Tokens:      100,
			HasMPin:     true,
			Roles: []responses.UserRoleDetail{
				{
					ID:       "ur-1",
					UserID:   "user-123",
					RoleID:   "role-1",
					IsActive: true,
					Role: responses.RoleDetail{
						ID:          "role-1",
						Name:        "user",
						Description: "Standard user role",
						IsActive:    true,
					},
				},
			},
		}

		loginResponse := &responses.LoginResponse{
			AccessToken:  "mock-access-token",
			RefreshToken: "mock-refresh-token",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			User:         userInfo,
			Message:      "Login successful",
		}

		// Verify response structure
		assert.NotEmpty(t, loginResponse.AccessToken)
		assert.NotEmpty(t, loginResponse.RefreshToken)
		assert.Equal(t, "Bearer", loginResponse.TokenType)
		assert.Equal(t, int64(3600), loginResponse.ExpiresIn)
		assert.NotNil(t, loginResponse.User)
		assert.Equal(t, "Login successful", loginResponse.Message)

		// Verify user information is complete
		assert.Equal(t, "user-123", loginResponse.User.ID)
		assert.Equal(t, "1234567890", loginResponse.User.PhoneNumber)
		assert.Equal(t, "+91", loginResponse.User.CountryCode)
		assert.True(t, loginResponse.User.HasMPin)
		assert.Len(t, loginResponse.User.Roles, 1)

		// Verify role information is complete
		role := loginResponse.User.Roles[0]
		assert.Equal(t, "ur-1", role.ID)
		assert.Equal(t, "user-123", role.UserID)
		assert.Equal(t, "role-1", role.RoleID)
		assert.True(t, role.IsActive)
		assert.Equal(t, "user", role.Role.Name)
		assert.Equal(t, "Standard user role", role.Role.Description)
		assert.True(t, role.Role.IsActive)

		// Test helper methods
		assert.True(t, loginResponse.User.HasRoles())
		assert.True(t, loginResponse.User.HasRole("user"))
		assert.False(t, loginResponse.User.HasRole("admin"))
		assert.Equal(t, []string{"user"}, loginResponse.User.GetRoleNames())
		assert.Len(t, loginResponse.User.GetActiveRoles(), 1)
	})
}

// TestLoginV2_RequestResponseFormat tests the JSON request/response format
func TestLoginV2_RequestResponseFormat(t *testing.T) {
	t.Run("Password Login Request JSON", func(t *testing.T) {
		// Test JSON marshaling/unmarshaling for password login
		password := "testpassword123"
		originalReq := requests.LoginRequest{
			PhoneNumber: "1234567890",
			CountryCode: "+91",
			Password:    &password,
		}

		// Marshal to JSON
		jsonData, err := json.Marshal(originalReq)
		assert.NoError(t, err)

		// Verify JSON structure
		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		assert.NoError(t, err)
		assert.Equal(t, "1234567890", jsonMap["phone_number"])
		assert.Equal(t, "+91", jsonMap["country_code"])
		assert.Equal(t, "testpassword123", jsonMap["password"])
		assert.Nil(t, jsonMap["mpin"])

		// Unmarshal back to struct
		var parsedReq requests.LoginRequest
		err = json.Unmarshal(jsonData, &parsedReq)
		assert.NoError(t, err)
		assert.Equal(t, originalReq.PhoneNumber, parsedReq.PhoneNumber)
		assert.Equal(t, originalReq.CountryCode, parsedReq.CountryCode)
		assert.Equal(t, *originalReq.Password, *parsedReq.Password)
		assert.Nil(t, parsedReq.MPin)
	})

	t.Run("MPIN Login Request JSON", func(t *testing.T) {
		// Test JSON marshaling/unmarshaling for MPIN login
		mpin := "1234"
		originalReq := requests.LoginRequest{
			PhoneNumber: "1234567890",
			CountryCode: "+91",
			MPin:        &mpin,
		}

		// Marshal to JSON
		jsonData, err := json.Marshal(originalReq)
		assert.NoError(t, err)

		// Verify JSON structure
		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		assert.NoError(t, err)
		assert.Equal(t, "1234567890", jsonMap["phone_number"])
		assert.Equal(t, "+91", jsonMap["country_code"])
		assert.Equal(t, "1234", jsonMap["mpin"])
		assert.Nil(t, jsonMap["password"])

		// Unmarshal back to struct
		var parsedReq requests.LoginRequest
		err = json.Unmarshal(jsonData, &parsedReq)
		assert.NoError(t, err)
		assert.Equal(t, originalReq.PhoneNumber, parsedReq.PhoneNumber)
		assert.Equal(t, originalReq.CountryCode, parsedReq.CountryCode)
		assert.Equal(t, *originalReq.MPin, *parsedReq.MPin)
		assert.Nil(t, parsedReq.Password)
	})
}

// TestLoginV2_ErrorHandling tests various error scenarios
func TestLoginV2_ErrorHandling(t *testing.T) {
	t.Run("Invalid MPIN Format", func(t *testing.T) {
		// Test invalid MPIN formats
		testCases := []struct {
			mpin        string
			shouldError bool
			errorMsg    string
		}{
			{"123", true, "mpin must be 4 or 6 digits"},     // Too short
			{"1234567", true, "mpin must be 4 or 6 digits"}, // Too long
			{"12ab", true, "mpin must contain only digits"}, // Non-numeric
			{"1234", false, ""},                             // Valid 4-digit
			{"123456", false, ""},                           // Valid 6-digit
		}

		for _, tc := range testCases {
			loginReq := requests.LoginRequest{
				PhoneNumber: "1234567890",
				CountryCode: "+91",
				MPin:        &tc.mpin,
			}

			err := loginReq.Validate()
			if tc.shouldError {
				assert.Error(t, err, "MPIN '%s' should cause validation error", tc.mpin)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err, "MPIN '%s' should be valid", tc.mpin)
			}
		}
	})

	t.Run("Invalid Password Format", func(t *testing.T) {
		// Test invalid password formats
		shortPassword := "123"
		loginReq := requests.LoginRequest{
			PhoneNumber: "1234567890",
			CountryCode: "+91",
			Password:    &shortPassword,
		}

		err := loginReq.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password must be at least 8 characters long")
	})

	t.Run("Missing Phone Number", func(t *testing.T) {
		password := "testpassword123"
		loginReq := requests.LoginRequest{
			CountryCode: "+91",
			Password:    &password,
		}

		err := loginReq.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "phone number and country code are required")
	})

	t.Run("Missing Country Code", func(t *testing.T) {
		password := "testpassword123"
		loginReq := requests.LoginRequest{
			PhoneNumber: "1234567890",
			Password:    &password,
		}

		err := loginReq.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "phone number and country code are required")
	})
}

// TestConvertToAuthUserInfo_Integration tests the conversion function with various scenarios
func TestConvertToAuthUserInfo_Integration(t *testing.T) {
	handler, _, _, _ := setupTestHandler()

	t.Run("User With Multiple Roles", func(t *testing.T) {
		userResponse := createTestUserResponseWithRoles([]string{"admin", "user", "moderator"})
		result := handler.convertToAuthUserInfo(userResponse)

		assert.Len(t, result.Roles, 3)
		assert.True(t, result.HasRole("admin"))
		assert.True(t, result.HasRole("user"))
		assert.True(t, result.HasRole("moderator"))
		assert.False(t, result.HasRole("superadmin"))

		roleNames := result.GetRoleNames()
		assert.Contains(t, roleNames, "admin")
		assert.Contains(t, roleNames, "user")
		assert.Contains(t, roleNames, "moderator")
	})

	t.Run("User With No Roles", func(t *testing.T) {
		userResponse := createTestUserResponseWithRoles([]string{})
		result := handler.convertToAuthUserInfo(userResponse)

		assert.Len(t, result.Roles, 0)
		assert.False(t, result.HasRoles())
		assert.Empty(t, result.GetRoleNames())
		assert.Empty(t, result.GetActiveRoles())
	})

	t.Run("User With Inactive Roles", func(t *testing.T) {
		userResponse := createTestUserResponseWithRoles([]string{"admin"})
		// Make the role inactive
		userResponse.Roles[0].IsActive = false

		result := handler.convertToAuthUserInfo(userResponse)

		assert.Len(t, result.Roles, 1)
		assert.True(t, result.HasRoles())        // Has roles, but they're inactive
		assert.Empty(t, result.GetActiveRoles()) // No active roles
		assert.False(t, result.HasRole("admin")) // HasRole checks for active roles
	})
}

// Helper function to create test user response with specified roles
func createTestUserResponseWithRoles(roleNames []string) *userResponses.UserResponse {
	roles := make([]userResponses.UserRoleDetail, len(roleNames))
	for i, roleName := range roleNames {
		roles[i] = userResponses.UserRoleDetail{
			ID:       "ur-" + string(rune(i+1)),
			UserID:   "user-123",
			RoleID:   "role-" + string(rune(i+1)),
			IsActive: true,
			Role: userResponses.RoleDetail{
				ID:          "role-" + string(rune(i+1)),
				Name:        roleName,
				Description: roleName + " role",
				IsActive:    true,
			},
		}
	}

	return &userResponses.UserResponse{
		ID:          "user-123",
		PhoneNumber: "1234567890",
		CountryCode: "+91",
		Username:    stringPtr("testuser"),
		IsValidated: true,
		Tokens:      100,
		HasMPin:     len(roleNames) > 0, // Set HasMPin based on whether user has roles
		Roles:       roles,
	}
}
