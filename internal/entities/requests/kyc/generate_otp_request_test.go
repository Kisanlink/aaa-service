package kyc

import (
	"encoding/json"
	"testing"
)

func TestConsentUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Consent
		shouldError bool
	}{
		{
			name:        "boolean true",
			input:       `{"consent": true}`,
			expected:    "Y",
			shouldError: false,
		},
		{
			name:        "string Y",
			input:       `{"consent": "Y"}`,
			expected:    "Y",
			shouldError: false,
		},
		{
			name:        "string true",
			input:       `{"consent": "true"}`,
			expected:    "Y",
			shouldError: false,
		},
		{
			name:        "boolean false should error",
			input:       `{"consent": false}`,
			expected:    "",
			shouldError: true,
		},
		{
			name:        "string N should error",
			input:       `{"consent": "N"}`,
			expected:    "",
			shouldError: true,
		},
		{
			name:        "string false should error",
			input:       `{"consent": "false"}`,
			expected:    "",
			shouldError: true,
		},
		{
			name:        "invalid type should error",
			input:       `{"consent": 123}`,
			expected:    "",
			shouldError: true,
		},
		{
			name:        "consent object with all fields",
			input:       `{"consent": {"purpose": "User verification", "timestamp": "2025-11-19T18:18:39.206Z", "version": "1.0"}}`,
			expected:    "Y",
			shouldError: false,
		},
		{
			name:        "consent object missing purpose",
			input:       `{"consent": {"timestamp": "2025-11-19T18:18:39.206Z", "version": "1.0"}}`,
			expected:    "",
			shouldError: true,
		},
		{
			name:        "consent object missing timestamp",
			input:       `{"consent": {"purpose": "User verification", "version": "1.0"}}`,
			expected:    "",
			shouldError: true,
		},
		{
			name:        "consent object missing version",
			input:       `{"consent": {"purpose": "User verification", "timestamp": "2025-11-19T18:18:39.206Z"}}`,
			expected:    "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result struct {
				Consent Consent `json:"consent"`
			}

			err := json.Unmarshal([]byte(tt.input), &result)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.Consent != tt.expected {
					t.Errorf("Expected consent to be %q, got %q", tt.expected, result.Consent)
				}
			}
		})
	}
}

func TestGenerateOTPRequest_UnmarshalJSON_BooleanConsent(t *testing.T) {
	// Test with boolean consent = true
	jsonData := `{
		"aadhaar_number": "123456789012",
		"consent": true
	}`

	var req GenerateOTPRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	if err != nil {
		t.Fatalf("Failed to unmarshal request with boolean consent: %v", err)
	}

	if req.AadhaarNumber != "123456789012" {
		t.Errorf("Expected aadhaar_number to be '123456789012', got %q", req.AadhaarNumber)
	}

	if req.Consent != "Y" {
		t.Errorf("Expected consent to be 'Y', got %q", req.Consent)
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		t.Errorf("Validation failed for valid request: %v", err)
	}
}

func TestGenerateOTPRequest_UnmarshalJSON_StringConsent(t *testing.T) {
	// Test with string consent = "Y"
	jsonData := `{
		"aadhaar_number": "123456789012",
		"consent": "Y"
	}`

	var req GenerateOTPRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	if err != nil {
		t.Fatalf("Failed to unmarshal request with string consent: %v", err)
	}

	if req.AadhaarNumber != "123456789012" {
		t.Errorf("Expected aadhaar_number to be '123456789012', got %q", req.AadhaarNumber)
	}

	if req.Consent != "Y" {
		t.Errorf("Expected consent to be 'Y', got %q", req.Consent)
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		t.Errorf("Validation failed for valid request: %v", err)
	}
}

func TestGenerateOTPRequest_UnmarshalJSON_ConsentObject(t *testing.T) {
	// Test with consent object (the new format)
	jsonData := `{
		"aadhaar_number": "123456789012",
		"consent": {
			"purpose": "User verification for admin panel",
			"timestamp": "2025-11-19T18:18:39.206Z",
			"version": "1.0"
		}
	}`

	var req GenerateOTPRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	if err != nil {
		t.Fatalf("Failed to unmarshal request with consent object: %v", err)
	}

	if req.AadhaarNumber != "123456789012" {
		t.Errorf("Expected aadhaar_number to be '123456789012', got %q", req.AadhaarNumber)
	}

	if req.Consent != "Y" {
		t.Errorf("Expected consent to be 'Y', got %q", req.Consent)
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		t.Errorf("Validation failed for valid request: %v", err)
	}
}

func TestGenerateOTPRequest_UnmarshalJSON_InvalidConsent(t *testing.T) {
	tests := []struct {
		name      string
		jsonData  string
		shouldErr bool
	}{
		{
			name: "boolean false",
			jsonData: `{
				"aadhaar_number": "123456789012",
				"consent": false
			}`,
			shouldErr: true,
		},
		{
			name: "string N",
			jsonData: `{
				"aadhaar_number": "123456789012",
				"consent": "N"
			}`,
			shouldErr: true,
		},
		{
			name: "number",
			jsonData: `{
				"aadhaar_number": "123456789012",
				"consent": 1
			}`,
			shouldErr: true,
		},
		{
			name: "consent object missing purpose",
			jsonData: `{
				"aadhaar_number": "123456789012",
				"consent": {"timestamp": "2025-11-19T18:18:39.206Z", "version": "1.0"}
			}`,
			shouldErr: true,
		},
		{
			name: "consent object missing timestamp",
			jsonData: `{
				"aadhaar_number": "123456789012",
				"consent": {"purpose": "User verification", "version": "1.0"}
			}`,
			shouldErr: true,
		},
		{
			name: "consent object missing version",
			jsonData: `{
				"aadhaar_number": "123456789012",
				"consent": {"purpose": "User verification", "timestamp": "2025-11-19T18:18:39.206Z"}
			}`,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req GenerateOTPRequest
			err := json.Unmarshal([]byte(tt.jsonData), &req)

			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGenerateOTPRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     GenerateOTPRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with Y consent",
			req: GenerateOTPRequest{
				AadhaarNumber: "123456789012",
				Consent:       "Y",
			},
			wantErr: false,
		},
		{
			name: "empty aadhaar number",
			req: GenerateOTPRequest{
				AadhaarNumber: "",
				Consent:       "Y",
			},
			wantErr: true,
			errMsg:  "aadhaar_number is required",
		},
		{
			name: "invalid aadhaar length",
			req: GenerateOTPRequest{
				AadhaarNumber: "12345",
				Consent:       "Y",
			},
			wantErr: true,
			errMsg:  "aadhaar_number must be exactly 12 digits",
		},
		{
			name: "non-numeric aadhaar",
			req: GenerateOTPRequest{
				AadhaarNumber: "12345678901a",
				Consent:       "Y",
			},
			wantErr: true,
			errMsg:  "aadhaar_number must contain only numeric digits",
		},
		{
			name: "empty consent",
			req: GenerateOTPRequest{
				AadhaarNumber: "123456789012",
				Consent:       "",
			},
			wantErr: true,
			errMsg:  "consent is required",
		},
		{
			name: "invalid consent value",
			req: GenerateOTPRequest{
				AadhaarNumber: "123456789012",
				Consent:       "N",
			},
			wantErr: true,
			errMsg:  "consent must be true, 'Y', or a consent object with purpose/timestamp/version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errMsg {
					t.Errorf("Expected error message %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
