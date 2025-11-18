package kyc

import "time"

// KYCStatusResponse represents the response for KYC status request
type KYCStatusResponse struct {
	StatusCode              int        `json:"status_code"`
	UserID                  string     `json:"user_id"`
	KYCStatus               string     `json:"kyc_status"`
	AadhaarVerified         bool       `json:"aadhaar_verified"`
	AadhaarVerifiedAt       *time.Time `json:"aadhaar_verified_at,omitempty"`
	VerificationAttempts    int        `json:"verification_attempts"`
	LastVerificationAttempt *time.Time `json:"last_verification_attempt,omitempty"`
}

// GetType returns the type of response
func (r *KYCStatusResponse) GetType() string {
	return "kyc_status"
}

// IsSuccess returns whether the response indicates success
func (r *KYCStatusResponse) IsSuccess() bool {
	return r.StatusCode == 200 && r.UserID != ""
}
