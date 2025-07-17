package addresses

import (
	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/aaa-service/entities/requests"
)

// UpdateAddressRequest represents a request to update an address
type UpdateAddressRequest struct {
	*requests.BaseRequest
	ID          string  `json:"id" validate:"required,min=1"`
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
}

// NewUpdateAddressRequest creates a new UpdateAddressRequest instance
func NewUpdateAddressRequest(
	id string,
	house, street, landmark, postOffice, subdistrict, district, vtc, state, country, pincode, fullAddress *string,
	protocol, operation, version, requestID string,
	headers map[string][]string,
	body interface{},
	context map[string]interface{},
) *UpdateAddressRequest {
	return &UpdateAddressRequest{
		BaseRequest: requests.NewBaseRequest(
			protocol,
			operation,
			version,
			requestID,
			"UpdateAddress",
			headers,
			body,
			context,
		),
		ID:          id,
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

// Validate validates the UpdateAddressRequest
func (r *UpdateAddressRequest) Validate() error {
	if r.ID == "" {
		return requests.NewValidationError("id", "Address ID is required")
	}

	// Validate pincode format if provided
	if r.Pincode != nil && len(*r.Pincode) != 6 {
		return requests.NewValidationError("pincode", "Pincode must be 6 digits")
	}

	return nil
}

// ToModel converts the request to an Address model
func (r *UpdateAddressRequest) ToModel() *models.Address {
	address := models.NewAddress()
	address.ID = r.ID
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

// GetID returns the address ID
func (r *UpdateAddressRequest) GetID() string {
	return r.ID
}
