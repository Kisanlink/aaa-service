package addresses

import (
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/internal/entities/responses"
)

// AddressResponse represents an address response
type AddressResponse struct {
	responses.Response
	ID          string  `json:"id"`
	House       *string `json:"house,omitempty"`
	Street      *string `json:"street,omitempty"`
	Landmark    *string `json:"landmark,omitempty"`
	PostOffice  *string `json:"post_office,omitempty"`
	Subdistrict *string `json:"subdistrict,omitempty"`
	District    *string `json:"district,omitempty"`
	VTC         *string `json:"vtc,omitempty"` // Village/Town/City
	State       *string `json:"state,omitempty"`
	Country     *string `json:"country,omitempty"`
	Pincode     *string `json:"pincode,omitempty"`
	FullAddress *string `json:"full_address,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// NewAddressResponse creates a new AddressResponse from an Address model
func NewAddressResponse(address *models.Address) *AddressResponse {
	return &AddressResponse{
		ID:          address.ID,
		House:       address.House,
		Street:      address.Street,
		Landmark:    address.Landmark,
		PostOffice:  address.PostOffice,
		Subdistrict: address.Subdistrict,
		District:    address.District,
		VTC:         address.VTC,
		State:       address.State,
		Country:     address.Country,
		Pincode:     address.Pincode,
		FullAddress: address.FullAddress,
		CreatedAt:   address.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   address.UpdatedAt.Format(time.RFC3339),
	}
}

// NewAddressResponseFromModel creates a new AddressResponse from an Address model
func NewAddressResponseFromModel(address *models.Address) *AddressResponse {
	return NewAddressResponse(address)
}

// GetID returns the address ID
func (r *AddressResponse) GetID() string {
	return r.ID
}

// GetHouse returns the house
func (r *AddressResponse) GetHouse() *string {
	return r.House
}

// GetStreet returns the street
func (r *AddressResponse) GetStreet() *string {
	return r.Street
}

// GetLandmark returns the landmark
func (r *AddressResponse) GetLandmark() *string {
	return r.Landmark
}

// GetPostOffice returns the post office
func (r *AddressResponse) GetPostOffice() *string {
	return r.PostOffice
}

// GetSubdistrict returns the subdistrict
func (r *AddressResponse) GetSubdistrict() *string {
	return r.Subdistrict
}

// GetDistrict returns the district
func (r *AddressResponse) GetDistrict() *string {
	return r.District
}

// GetVTC returns the village/town/city
func (r *AddressResponse) GetVTC() *string {
	return r.VTC
}

// GetState returns the state
func (r *AddressResponse) GetState() *string {
	return r.State
}

// GetCountry returns the country
func (r *AddressResponse) GetCountry() *string {
	return r.Country
}

// GetPincode returns the pincode
func (r *AddressResponse) GetPincode() *string {
	return r.Pincode
}

// GetFullAddress returns the full address
func (r *AddressResponse) GetFullAddress() *string {
	return r.FullAddress
}

// GetCreatedAt returns the created at timestamp
func (r *AddressResponse) GetCreatedAt() string {
	return r.CreatedAt
}

// GetUpdatedAt returns the updated at timestamp
func (r *AddressResponse) GetUpdatedAt() string {
	return r.UpdatedAt
}
