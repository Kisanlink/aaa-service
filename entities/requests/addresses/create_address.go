package addresses

import (
	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/aaa-service/entities/requests"
)

// CreateAddressRequest represents a request to create an address
type CreateAddressRequest struct {
	*requests.BaseRequest
	House       *string `json:"house,omitempty" validate:"omitempty,max=255"`
	Street      *string `json:"street,omitempty" validate:"omitempty,max=255"`
	Landmark    *string `json:"landmark,omitempty" validate:"omitempty,max=255"`
	PostOffice  *string `json:"post_office,omitempty" validate:"omitempty,max=255"`
	Subdistrict *string `json:"subdistrict,omitempty" validate:"omitempty,max=255"`
	District    *string `json:"district,omitempty" validate:"omitempty,max=255"`
	VTC         *string `json:"vtc,omitempty" validate:"omitempty,max=255"`
	State       *string `json:"state,omitempty" validate:"omitempty,max=255"`
	Country     *string `json:"country,omitempty" validate:"omitempty,max=255"`
	Pincode     *string `json:"pincode,omitempty" validate:"omitempty,len=6"`
	FullAddress *string `json:"full_address,omitempty" validate:"omitempty,max=1000"`
	UserID      string  `json:"user_id" validate:"required,min=1"`
}

// NewCreateAddressRequest creates a new CreateAddressRequest instance
func NewCreateAddressRequest(
	userID string,
	house, street, landmark, postOffice, subdistrict, district, vtc, state, country, pincode, fullAddress *string,
	protocol, operation, version, requestID string,
	headers map[string][]string,
	body interface{},
	context map[string]interface{},
) *CreateAddressRequest {
	return &CreateAddressRequest{
		BaseRequest: requests.NewBaseRequest(
			protocol,
			operation,
			version,
			requestID,
			"CreateAddress",
			headers,
			body,
			context,
		),
		UserID:      userID,
		House:       house,
		Street:      street,
		Landmark:    landmark,
		PostOffice:  postOffice,
		Subdistrict: subdistrict,
		District:    district,
		VTC:         vtc,
		State:       state,
		Country:     country,
		Pincode:     pincode,
		FullAddress: fullAddress,
	}
}

// Validate validates the CreateAddressRequest
func (r *CreateAddressRequest) Validate() error {
	if r.UserID == "" {
		return requests.NewValidationError("user_id", "User ID is required")
	}

	// At least one address field should be provided
	if (r.House == nil || *r.House == "") &&
		(r.Street == nil || *r.Street == "") &&
		(r.FullAddress == nil || *r.FullAddress == "") &&
		(r.District == nil || *r.District == "") {
		return requests.NewValidationError("address", "At least one address field is required")
	}

	// Validate pincode format if provided
	if r.Pincode != nil && len(*r.Pincode) != 6 {
		return requests.NewValidationError("pincode", "Pincode must be 6 digits")
	}

	return nil
}

// ToModel converts the request to an Address model
func (r *CreateAddressRequest) ToModel() *models.Address {
	address := models.NewAddress()
	address.House = r.House
	address.Street = r.Street
	address.Landmark = r.Landmark
	address.PostOffice = r.PostOffice
	address.Subdistrict = r.Subdistrict
	address.District = r.District
	address.VTC = r.VTC
	address.State = r.State
	address.Country = r.Country
	address.Pincode = r.Pincode
	address.FullAddress = r.FullAddress

	return address
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

// GetPostOffice returns the post office
func (r *CreateAddressRequest) GetPostOffice() *string {
	return r.PostOffice
}

// GetSubdistrict returns the subdistrict
func (r *CreateAddressRequest) GetSubdistrict() *string {
	return r.Subdistrict
}

// GetDistrict returns the district
func (r *CreateAddressRequest) GetDistrict() *string {
	return r.District
}

// GetVTC returns the VTC
func (r *CreateAddressRequest) GetVTC() *string {
	return r.VTC
}

// GetState returns the state
func (r *CreateAddressRequest) GetState() *string {
	return r.State
}

// GetCountry returns the country
func (r *CreateAddressRequest) GetCountry() *string {
	return r.Country
}

// GetPincode returns the pincode
func (r *CreateAddressRequest) GetPincode() *string {
	return r.Pincode
}

// GetFullAddress returns the full address
func (r *CreateAddressRequest) GetFullAddress() *string {
	return r.FullAddress
}
