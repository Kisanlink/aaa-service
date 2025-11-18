package kyc

import (
	"fmt"
	"regexp"
)

// GenerateOTPRequest represents the request to generate OTP for Aadhaar verification
type GenerateOTPRequest struct {
	AadhaarNumber string `json:"aadhaar_number" validate:"required,len=12,numeric"`
	Consent       string `json:"consent" validate:"required,eq=Y"`
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
		return fmt.Errorf("consent must be 'Y'")
	}

	return nil
}

// GetType returns the type of request
func (r *GenerateOTPRequest) GetType() string {
	return "generate_aadhaar_otp"
}
