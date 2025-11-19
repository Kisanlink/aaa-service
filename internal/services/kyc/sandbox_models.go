package kyc

// SandboxOTPRequest represents the request payload for OTP generation
type SandboxOTPRequest struct {
	Entity        string `json:"@entity"`
	AadhaarNumber string `json:"aadhaar_number"`
	Consent       string `json:"consent"`
	Reason        string `json:"reason"` // Purpose of verification (optional but recommended)
}

// SandboxOTPResponse represents the response from OTP generation
type SandboxOTPResponse struct {
	Timestamp     int64           `json:"timestamp"`
	TransactionID string          `json:"transaction_id"`
	Code          int             `json:"code"`
	Data          OTPResponseData `json:"data"`
}

// OTPResponseData contains the OTP generation response data
type OTPResponseData struct {
	Entity      string `json:"@entity"`
	Message     string `json:"message"`
	ReferenceID int    `json:"reference_id"`
}

// SandboxVerifyRequest represents the request payload for OTP verification
type SandboxVerifyRequest struct {
	Entity      string `json:"@entity"`
	ReferenceID string `json:"reference_id"`
	OTP         string `json:"otp"`
}

// SandboxVerifyResponse represents the response from OTP verification
type SandboxVerifyResponse struct {
	Timestamp     int64   `json:"timestamp"`
	TransactionID string  `json:"transaction_id"`
	Code          int     `json:"code"`
	Data          KYCData `json:"data"`
}

// KYCData contains the Aadhaar verification data
type KYCData struct {
	Entity      string         `json:"@entity"`
	Name        string         `json:"name"`
	Gender      string         `json:"gender"`
	DateOfBirth string         `json:"date_of_birth"`
	YOB         int            `json:"year_of_birth"`
	CareOf      string         `json:"care_of"`
	FullAddress string         `json:"full_address"`
	Address     SandboxAddress `json:"address"`
	Photo       string         `json:"photo"` // base64 encoded
	ShareCode   string         `json:"share_code"`
	Status      string         `json:"status"`
	Message     string         `json:"message"`
}

// SandboxAddress represents the structured address from Aadhaar
type SandboxAddress struct {
	Entity      string `json:"@entity"`
	House       string `json:"house"`
	Street      string `json:"street"`
	Landmark    string `json:"landmark"`
	Locality    string `json:"locality"`
	Vtc         string `json:"vtc"`
	Subdistrict string `json:"subdistrict"`
	District    string `json:"district"`
	State       string `json:"state"`
	Pincode     int    `json:"pincode"`
	PostOffice  string `json:"post_office"`
	Country     string `json:"country"`
}

// SandboxErrorResponse represents error responses from Sandbox API
type SandboxErrorResponse struct {
	Timestamp     int64  `json:"timestamp"`
	TransactionID string `json:"transaction_id"`
	Code          int    `json:"code"`
	Message       string `json:"message"`
	Error         string `json:"error"`
}

// SandboxAuthResponse represents the authentication response from Sandbox
type SandboxAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"` // seconds (24 hours = 86400)
}
