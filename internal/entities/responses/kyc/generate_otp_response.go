package kyc

// GenerateOTPResponse represents the response for OTP generation request
type GenerateOTPResponse struct {
	StatusCode    int    `json:"status_code"`
	Message       string `json:"message"`
	ReferenceID   string `json:"reference_id"`
	TransactionID string `json:"transaction_id"`
	Timestamp     int64  `json:"timestamp"`
	ExpiresAt     int64  `json:"expires_at"`
}

// GetType returns the type of response
func (r *GenerateOTPResponse) GetType() string {
	return "generate_aadhaar_otp"
}

// IsSuccess returns whether the response indicates success
func (r *GenerateOTPResponse) IsSuccess() bool {
	return r.StatusCode == 200 && r.ReferenceID != ""
}
