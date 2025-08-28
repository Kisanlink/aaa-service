package requests

import (
	"testing"
)

func TestLoginRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request LoginRequest
		wantErr bool
	}{
		{
			name: "valid password login",
			request: LoginRequest{
				PhoneNumber: "1234567890",
				CountryCode: "+91",
				Password:    stringPtr("password123"),
			},
			wantErr: false,
		},
		{
			name: "valid mpin login",
			request: LoginRequest{
				PhoneNumber: "1234567890",
				CountryCode: "+91",
				MPin:        stringPtr("1234"),
			},
			wantErr: false,
		},
		{
			name: "valid both password and mpin",
			request: LoginRequest{
				PhoneNumber: "1234567890",
				CountryCode: "+91",
				Password:    stringPtr("password123"),
				MPin:        stringPtr("1234"),
			},
			wantErr: false,
		},
		{
			name: "missing phone number",
			request: LoginRequest{
				CountryCode: "+91",
				Password:    stringPtr("password123"),
			},
			wantErr: true,
		},
		{
			name: "missing country code",
			request: LoginRequest{
				PhoneNumber: "1234567890",
				Password:    stringPtr("password123"),
			},
			wantErr: true,
		},
		{
			name: "missing both password and mpin",
			request: LoginRequest{
				PhoneNumber: "1234567890",
				CountryCode: "+91",
			},
			wantErr: true,
		},
		{
			name: "invalid password too short",
			request: LoginRequest{
				PhoneNumber: "1234567890",
				CountryCode: "+91",
				Password:    stringPtr("123"),
			},
			wantErr: true,
		},
		{
			name: "invalid mpin too short",
			request: LoginRequest{
				PhoneNumber: "1234567890",
				CountryCode: "+91",
				MPin:        stringPtr("12"),
			},
			wantErr: true,
		},
		{
			name: "invalid mpin non-numeric",
			request: LoginRequest{
				PhoneNumber: "1234567890",
				CountryCode: "+91",
				MPin:        stringPtr("12ab"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoginRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoginRequest_HelperMethods(t *testing.T) {
	request := LoginRequest{
		PhoneNumber:     "1234567890",
		CountryCode:     "+91",
		Password:        stringPtr("password123"),
		MPin:            stringPtr("1234"),
		IncludeProfile:  boolPtr(true),
		IncludeRoles:    boolPtr(false),
		IncludeContacts: boolPtr(true),
	}

	if !request.HasPassword() {
		t.Error("Expected HasPassword() to return true")
	}

	if !request.HasMPin() {
		t.Error("Expected HasMPin() to return true")
	}

	if request.GetPassword() != "password123" {
		t.Errorf("Expected GetPassword() to return 'password123', got '%s'", request.GetPassword())
	}

	if request.GetMPin() != "1234" {
		t.Errorf("Expected GetMPin() to return '1234', got '%s'", request.GetMPin())
	}

	if !request.ShouldIncludeProfile() {
		t.Error("Expected ShouldIncludeProfile() to return true")
	}

	if request.ShouldIncludeRoles() {
		t.Error("Expected ShouldIncludeRoles() to return false")
	}

	if !request.ShouldIncludeContacts() {
		t.Error("Expected ShouldIncludeContacts() to return true")
	}
}

func TestUpdateMPinRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request UpdateMPinRequest
		wantErr bool
	}{
		{
			name: "valid update",
			request: UpdateMPinRequest{
				CurrentMPin: "1234",
				NewMPin:     "5678",
			},
			wantErr: false,
		},
		{
			name: "same mpin",
			request: UpdateMPinRequest{
				CurrentMPin: "1234",
				NewMPin:     "1234",
			},
			wantErr: true,
		},
		{
			name: "invalid current mpin",
			request: UpdateMPinRequest{
				CurrentMPin: "12ab",
				NewMPin:     "5678",
			},
			wantErr: true,
		},
		{
			name: "invalid new mpin",
			request: UpdateMPinRequest{
				CurrentMPin: "1234",
				NewMPin:     "56cd",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateMPinRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
