package kyc

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// ConsentMetadata represents detailed consent information for audit purposes
type ConsentMetadata struct {
	Purpose   string `json:"purpose"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// Consent represents user consent for Aadhaar verification
// It can accept multiple formats:
// 1. Boolean: true
// 2. String: "Y"
// 3. Object: {"purpose": "...", "timestamp": "...", "version": "..."}
type Consent string

// UnmarshalJSON implements custom JSON unmarshaling for Consent
// Accepts: true (bool), "Y" (string), or consent object with metadata
// Rejects: false (bool), "N" (string), or invalid formats
func (c *Consent) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as boolean first
	var boolValue bool
	if err := json.Unmarshal(data, &boolValue); err == nil {
		if boolValue {
			*c = "Y"
			return nil
		}
		return fmt.Errorf("consent must be true, 'Y', or a consent object with purpose/timestamp/version")
	}

	// Try to unmarshal as string
	var strValue string
	if err := json.Unmarshal(data, &strValue); err == nil {
		if strValue == "Y" || strValue == "true" {
			*c = "Y"
			return nil
		}
		return fmt.Errorf("consent must be true, 'Y', or a consent object with purpose/timestamp/version")
	}

	// Try to unmarshal as consent object
	var consentObj ConsentMetadata
	if err := json.Unmarshal(data, &consentObj); err == nil {
		// Validate that required fields are present
		if consentObj.Purpose == "" {
			return fmt.Errorf("consent object must include 'purpose' field")
		}
		if consentObj.Timestamp == "" {
			return fmt.Errorf("consent object must include 'timestamp' field")
		}
		if consentObj.Version == "" {
			return fmt.Errorf("consent object must include 'version' field")
		}

		// Accept the consent object and convert to "Y" for Sandbox API
		// The metadata is captured in the request and can be stored for audit
		*c = "Y"
		return nil
	}

	return fmt.Errorf("consent must be true, 'Y', or a consent object with purpose/timestamp/version")
}

// String returns the string representation of Consent
func (c Consent) String() string {
	return string(c)
}

// GenerateOTPRequest represents the request to generate OTP for Aadhaar verification
type GenerateOTPRequest struct {
	AadhaarNumber string  `json:"aadhaar_number" validate:"required,len=12,numeric"`
	Consent       Consent `json:"consent" validate:"required,eq=Y"`
}

// Validate validates the GenerateOTPRequest
func (r *GenerateOTPRequest) Validate() error {
	// Validate Aadhaar number
	if r.AadhaarNumber == "" {
		return fmt.Errorf("aadhaar_number is required")
	}
	if len(r.AadhaarNumber) != 12 {
		return fmt.Errorf("aadhaar_number must be exactly 12 digits")
	}
	aadhaarRegex := regexp.MustCompile(`^[0-9]{12}$`)
	if !aadhaarRegex.MatchString(r.AadhaarNumber) {
		return fmt.Errorf("aadhaar_number must contain only numeric digits")
	}

	// Validate consent
	if r.Consent == "" {
		return fmt.Errorf("consent is required")
	}
	if r.Consent != "Y" {
		return fmt.Errorf("consent must be true, 'Y', or a consent object with purpose/timestamp/version")
	}

	return nil
}

// GetType returns the type of request
func (r *GenerateOTPRequest) GetType() string {
	return "generate_aadhaar_otp"
}
