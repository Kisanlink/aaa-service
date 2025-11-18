package kyc

import (
	"fmt"
	"regexp"
)

// VerifyOTPRequest represents the request to verify Aadhaar OTP
type VerifyOTPRequest struct {
	ReferenceID string `json:"reference_id" validate:"required"`
	OTP         string `json:"otp" validate:"required,len=6,numeric"`
}

// Validate validates the VerifyOTPRequest
func (r *VerifyOTPRequest) Validate() error {
	// Validate reference ID
	if r.ReferenceID == "" {
		return fmt.Errorf("reference_id is required")
	}

	// Validate OTP
	if r.OTP == "" {
		return fmt.Errorf("otp is required")
	}
	if len(r.OTP) != 6 {
		return fmt.Errorf("otp must be exactly 6 digits")
	}
	otpRegex := regexp.MustCompile(`^[0-9]{6}$`)
	if !otpRegex.MatchString(r.OTP) {
		return fmt.Errorf("otp must contain only numeric digits")
	}

	return nil
}

// GetType returns the type of request
func (r *VerifyOTPRequest) GetType() string {
	return "verify_aadhaar_otp"
}
