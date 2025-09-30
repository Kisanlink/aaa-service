package utils

import (
	"testing"

	v10 "github.com/go-playground/validator/v10"
)

func TestValidateOrganizationID(t *testing.T) {
	tests := []struct {
		name    string
		orgID   string
		wantErr bool
	}{
		{
			name:    "Valid organization ID - ORGN00000001",
			orgID:   "ORGN00000001",
			wantErr: false,
		},
		{
			name:    "Valid organization ID - ORGN00000002",
			orgID:   "ORGN00000002",
			wantErr: false,
		},
		{
			name:    "Valid organization ID - ORGN99999999",
			orgID:   "ORGN99999999",
			wantErr: false,
		},
		{
			name:    "Empty string (should be valid with omitempty)",
			orgID:   "",
			wantErr: false,
		},
		{
			name:    "Invalid - wrong prefix (ORG)",
			orgID:   "ORG000000001",
			wantErr: true,
		},
		{
			name:    "Invalid - too few digits",
			orgID:   "ORGN0000001",
			wantErr: true,
		},
		{
			name:    "Invalid - too many digits",
			orgID:   "ORGN000000001",
			wantErr: true,
		},
		{
			name:    "Invalid - UUID format",
			orgID:   "550e8400-e29b-41d4-a716-446655440000",
			wantErr: true,
		},
		{
			name:    "Invalid - contains letters in numeric part",
			orgID:   "ORGNabcd1234",
			wantErr: true,
		},
		{
			name:    "Invalid - lowercase prefix",
			orgID:   "orgn00000001",
			wantErr: true,
		},
	}

	validate := v10.New()
	if err := validate.RegisterValidation("org_id", validateOrganizationID); err != nil {
		t.Fatalf("Failed to register org_id validator: %v", err)
	}

	type testStruct struct {
		OrgID string `validate:"org_id"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{OrgID: tt.orgID}
			err := validate.Struct(s)

			if tt.wantErr && err == nil {
				t.Errorf("Expected validation error for orgID=%q, but got none", tt.orgID)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no validation error for orgID=%q, but got: %v", tt.orgID, err)
			}
		})
	}
}

func TestValidateOrganizationIDWithOmitEmpty(t *testing.T) {
	validate := v10.New()
	if err := validate.RegisterValidation("org_id", validateOrganizationID); err != nil {
		t.Fatalf("Failed to register org_id validator: %v", err)
	}

	type testStruct struct {
		OrgID *string `validate:"omitempty,org_id"`
	}

	tests := []struct {
		name    string
		orgID   *string
		wantErr bool
	}{
		{
			name:    "Nil pointer (should be valid with omitempty)",
			orgID:   nil,
			wantErr: false,
		},
		{
			name:    "Valid organization ID pointer",
			orgID:   stringPtr("ORGN00000001"),
			wantErr: false,
		},
		{
			name:    "Invalid organization ID pointer",
			orgID:   stringPtr("INVALID123"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{OrgID: tt.orgID}
			err := validate.Struct(s)

			if tt.wantErr && err == nil {
				orgIDVal := "nil"
				if tt.orgID != nil {
					orgIDVal = *tt.orgID
				}
				t.Errorf("Expected validation error for orgID=%q, but got none", orgIDVal)
			}
			if !tt.wantErr && err != nil {
				orgIDVal := "nil"
				if tt.orgID != nil {
					orgIDVal = *tt.orgID
				}
				t.Errorf("Expected no validation error for orgID=%q, but got: %v", orgIDVal, err)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
