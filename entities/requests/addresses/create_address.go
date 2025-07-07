package addresses

import (
	"github.com/Kisanlink/aaa-service/entities/requests"
)

// CreateAddressRequest represents a request to create an address
type CreateAddressRequest struct {
	requests.Request
	UserID      string  `json:"user_id" validate:"required"`
	House       *string `json:"house" validate:"omitempty,max:255"`
	Street      *string `json:"street" validate:"omitempty,max:255"`
	Landmark    *string `json:"landmark" validate:"omitempty,max:255"`
	Locality    *string `json:"locality" validate:"omitempty,max:255"`
	Village     *string `json:"village" validate:"omitempty,max:255"`
	SubDistrict *string `json:"sub_district" validate:"omitempty,max:255"`
	District    *string `json:"district" validate:"omitempty,max:255"`
	State       *string `json:"state" validate:"omitempty,max:255"`
	PostOffice  *string `json:"post_office" validate:"omitempty,max:255"`
	Pincode     *string `json:"pincode" validate:"omitempty,len=6"`
	Country     *string `json:"country" validate:"omitempty,max:255"`
}

// NewCreateAddressRequest creates a new CreateAddressRequest instance
func NewCreateAddressRequest(
	userID string,
	house *string,
	street *string,
	landmark *string,
	locality *string,
	village *string,
	subDistrict *string,
	district *string,
	state *string,
	postOffice *string,
	pincode *string,
	country *string,
	protocol string,
	operation string,
	version string,
	requestID string,
	headers map[string][]string,
	body interface{},
	context map[string]interface{},
) *CreateAddressRequest {
	return &CreateAddressRequest{
		Request: requests.Request{
			Protocol:  protocol,
			Operation: operation,
			Version:   version,
			RequestID: requestID,
			Headers:   headers,
			Body:      body,
			Context:   context,
		},
		UserID:      userID,
		House:       house,
		Street:      street,
		Landmark:    landmark,
		Locality:    locality,
		Village:     village,
		SubDistrict: subDistrict,
		District:    district,
		State:       state,
		PostOffice:  postOffice,
		Pincode:     pincode,
		Country:     country,
	}
}

// Validate validates the CreateAddressRequest
func (r *CreateAddressRequest) Validate() error {
	if r.UserID == "" {
		return requests.NewValidationError("user_id", "User ID is required")
	}

	if r.Pincode != nil && len(*r.Pincode) != 6 {
		return requests.NewValidationError("pincode", "Pincode must be 6 digits")
	}

	return nil
}

// GetUserID returns the user ID
func (r *CreateAddressRequest) GetUserID() string {
	return r.UserID
}

// GetHouse returns the house
func (r *CreateAddressRequest) GetHouse() *string {
	return r.House
}

// GetStreet returns the street
func (r *CreateAddressRequest) GetStreet() *string {
	return r.Street
}

// GetLandmark returns the landmark
func (r *CreateAddressRequest) GetLandmark() *string {
	return r.Landmark
}

// GetLocality returns the locality
func (r *CreateAddressRequest) GetLocality() *string {
	return r.Locality
}

// GetVillage returns the village
func (r *CreateAddressRequest) GetVillage() *string {
	return r.Village
}

// GetSubDistrict returns the sub district
func (r *CreateAddressRequest) GetSubDistrict() *string {
	return r.SubDistrict
}

// GetDistrict returns the district
func (r *CreateAddressRequest) GetDistrict() *string {
	return r.District
}

// GetState returns the state
func (r *CreateAddressRequest) GetState() *string {
	return r.State
}

// GetPostOffice returns the post office
func (r *CreateAddressRequest) GetPostOffice() *string {
	return r.PostOffice
}

// GetPincode returns the pincode
func (r *CreateAddressRequest) GetPincode() *string {
	return r.Pincode
}

// GetCountry returns the country
func (r *CreateAddressRequest) GetCountry() *string {
	return r.Country
}
