package responses

import (
	"testing"
	"time"
)

func TestUserInfo_HelperMethods(t *testing.T) {
	userInfo := &UserInfo{
		ID:          "user123",
		PhoneNumber: "1234567890",
		CountryCode: "+91",
		Username:    stringPtr("testuser"),
		IsValidated: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tokens:      1000,
		HasMPin:     true,
		Roles: []UserRoleDetail{
			{
				ID:     "ur1",
				UserID: "user123",
				RoleID: "role1",
				Role: RoleDetail{
					ID:          "role1",
					Name:        "admin",
					Description: "Administrator role",
					Scope:       "GLOBAL",
					IsActive:    true,
					Version:     1,
				},
				IsActive: true,
			},
			{
				ID:     "ur2",
				UserID: "user123",
				RoleID: "role2",
				Role: RoleDetail{
					ID:          "role2",
					Name:        "user",
					Description: "Regular user role",
					Scope:       "GLOBAL",
					IsActive:    false, // Inactive role
					Version:     1,
				},
				IsActive: true,
			},
		},
		Profile: &UserProfileInfo{
			ID:   "profile123",
			Name: stringPtr("Test User"),
		},
		Contacts: []ContactInfo{
			{
				ID:         "contact1",
				Type:       "email",
				Value:      "test@example.com",
				IsPrimary:  true,
				IsVerified: true,
			},
		},
	}

	// Test HasRoles
	if !userInfo.HasRoles() {
		t.Error("Expected HasRoles() to return true")
	}

	// Test HasProfile
	if !userInfo.HasProfile() {
		t.Error("Expected HasProfile() to return true")
	}

	// Test HasContacts
	if !userInfo.HasContacts() {
		t.Error("Expected HasContacts() to return true")
	}

	// Test GetActiveRoles (should only return the admin role since user role is inactive)
	activeRoles := userInfo.GetActiveRoles()
	if len(activeRoles) != 1 {
		t.Errorf("Expected 1 active role, got %d", len(activeRoles))
	}
	if activeRoles[0].Role.Name != "admin" {
		t.Errorf("Expected active role to be 'admin', got '%s'", activeRoles[0].Role.Name)
	}

	// Test HasRole
	if !userInfo.HasRole("admin") {
		t.Error("Expected HasRole('admin') to return true")
	}
	if userInfo.HasRole("user") {
		t.Error("Expected HasRole('user') to return false (inactive role)")
	}
	if userInfo.HasRole("nonexistent") {
		t.Error("Expected HasRole('nonexistent') to return false")
	}

	// Test GetRoleNames
	roleNames := userInfo.GetRoleNames()
	if len(roleNames) != 1 {
		t.Errorf("Expected 1 role name, got %d", len(roleNames))
	}
	if roleNames[0] != "admin" {
		t.Errorf("Expected role name to be 'admin', got '%s'", roleNames[0])
	}
}

func TestUserInfo_EmptyData(t *testing.T) {
	userInfo := &UserInfo{
		ID:          "user123",
		PhoneNumber: "1234567890",
		CountryCode: "+91",
		IsValidated: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tokens:      0,
		HasMPin:     false,
		Roles:       []UserRoleDetail{},
		Profile:     nil,
		Contacts:    []ContactInfo{},
	}

	// Test with empty data
	if userInfo.HasRoles() {
		t.Error("Expected HasRoles() to return false for empty roles")
	}

	if userInfo.HasProfile() {
		t.Error("Expected HasProfile() to return false for nil profile")
	}

	if userInfo.HasContacts() {
		t.Error("Expected HasContacts() to return false for empty contacts")
	}

	activeRoles := userInfo.GetActiveRoles()
	if len(activeRoles) != 0 {
		t.Errorf("Expected 0 active roles, got %d", len(activeRoles))
	}

	if userInfo.HasRole("admin") {
		t.Error("Expected HasRole('admin') to return false for empty roles")
	}

	roleNames := userInfo.GetRoleNames()
	if len(roleNames) != 0 {
		t.Errorf("Expected 0 role names, got %d", len(roleNames))
	}
}

func TestLoginResponse_Methods(t *testing.T) {
	response := &LoginResponse{
		AccessToken:  "token123",
		RefreshToken: "refresh123",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		User: &UserInfo{
			ID:          "user123",
			PhoneNumber: "1234567890",
		},
		Message: "Login successful",
	}

	if response.GetType() != "login" {
		t.Errorf("Expected GetType() to return 'login', got '%s'", response.GetType())
	}

	if !response.IsSuccess() {
		t.Error("Expected IsSuccess() to return true when access token is present")
	}

	// Test with empty access token
	response.AccessToken = ""
	if response.IsSuccess() {
		t.Error("Expected IsSuccess() to return false when access token is empty")
	}
}

func TestAssignRoleResponse_Methods(t *testing.T) {
	response := &AssignRoleResponse{
		Message: "Role assigned successfully",
		UserID:  "user123",
		Role: RoleDetail{
			ID:          "role1",
			Name:        "admin",
			Description: "Administrator role",
			Scope:       "GLOBAL",
			IsActive:    true,
			Version:     1,
		},
		Success: true,
	}

	if response.GetType() != "assign_role" {
		t.Errorf("Expected GetType() to return 'assign_role', got '%s'", response.GetType())
	}

	if !response.IsSuccess() {
		t.Error("Expected IsSuccess() to return true when success is true")
	}

	response.Success = false
	if response.IsSuccess() {
		t.Error("Expected IsSuccess() to return false when success is false")
	}
}

func TestSetMPinResponse_Methods(t *testing.T) {
	response := &SetMPinResponse{
		Message: "MPIN set successfully",
		Success: true,
	}

	if response.GetType() != "set_mpin" {
		t.Errorf("Expected GetType() to return 'set_mpin', got '%s'", response.GetType())
	}

	if !response.IsSuccess() {
		t.Error("Expected IsSuccess() to return true when success is true")
	}
}

func TestUpdateMPinResponse_Methods(t *testing.T) {
	response := &UpdateMPinResponse{
		Message: "MPIN updated successfully",
		Success: true,
	}

	if response.GetType() != "update_mpin" {
		t.Errorf("Expected GetType() to return 'update_mpin', got '%s'", response.GetType())
	}

	if !response.IsSuccess() {
		t.Error("Expected IsSuccess() to return true when success is true")
	}
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}
