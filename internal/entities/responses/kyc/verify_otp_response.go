package kyc

// VerifyOTPResponse represents the response for OTP verification request
type VerifyOTPResponse struct {
	StatusCode  int          `json:"status_code"`
	Message     string       `json:"message"`
	AadhaarData *AadhaarData `json:"aadhaar_data,omitempty"`
	ProfileID   string       `json:"profile_id,omitempty"`
	AddressID   string       `json:"address_id,omitempty"`
}

// AadhaarData represents the Aadhaar verification data returned in the response
type AadhaarData struct {
	Name        string       `json:"name"`
	Gender      string       `json:"gender"`
	DateOfBirth string       `json:"date_of_birth"`
	YearOfBirth int          `json:"year_of_birth"`
	CareOf      string       `json:"care_of"`
	FullAddress string       `json:"full_address"`
	Address     *AadhaarAddr `json:"address"`
	PhotoURL    string       `json:"photo_url"`
	ShareCode   string       `json:"share_code"`
	Status      string       `json:"status"`
}

// AadhaarAddr represents the address structure in Aadhaar data
type AadhaarAddr struct {
	House    string `json:"house"`
	Street   string `json:"street"`
	Landmark string `json:"landmark"`
	District string `json:"district"`
	State    string `json:"state"`
	Pincode  int    `json:"pincode"`
	Country  string `json:"country"`
}

// GetType returns the type of response
func (r *VerifyOTPResponse) GetType() string {
	return "verify_aadhaar_otp"
}

// IsSuccess returns whether the response indicates success
func (r *VerifyOTPResponse) IsSuccess() bool {
	return r.StatusCode == 200 && r.AadhaarData != nil
}
