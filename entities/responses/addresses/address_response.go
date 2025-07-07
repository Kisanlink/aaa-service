package addresses

import (
	"time"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/aaa-service/entities/responses"
)

// AddressResponse represents an address response
type AddressResponse struct {
	responses.Response
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	House       *string `json:"house,omitempty"`
	Street      *string `json:"street,omitempty"`
	Landmark    *string `json:"landmark,omitempty"`
	Locality    *string `json:"locality,omitempty"`
	Village     *string `json:"village,omitempty"`
	SubDistrict *string `json:"sub_district,omitempty"`
	District    *string `json:"district,omitempty"`
	State       *string `json:"state,omitempty"`
	PostOffice  *string `json:"post_office,omitempty"`
	Pincode     *string `json:"pincode,omitempty"`
	Country     *string `json:"country,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// NewAddressResponse creates a new AddressResponse from an Address model
func NewAddressResponse(address *models.Address) *AddressResponse {
	return &AddressResponse{
		ID:          address.ID,
		UserID:      address.UserID,
		House:       address.House,
		Street:      address.Street,
		Landmark:    address.Landmark,
		Locality:    address.Locality,
		Village:     address.Village,
		SubDistrict: address.SubDistrict,
		District:    address.District,
		State:       address.State,
		PostOffice:  address.PostOffice,
		Pincode:     address.Pincode,
		Country:     address.Country,
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

// GetUserID returns the user ID
func (r *AddressResponse) GetUserID() string {
	return r.UserID
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

// GetLocality returns the locality
func (r *AddressResponse) GetLocality() *string {
	return r.Locality
}

// GetVillage returns the village
func (r *AddressResponse) GetVillage() *string {
	return r.Village
}

// GetSubDistrict returns the sub district
func (r *AddressResponse) GetSubDistrict() *string {
	return r.SubDistrict
}

// GetDistrict returns the district
func (r *AddressResponse) GetDistrict() *string {
	return r.District
}

// GetState returns the state
func (r *AddressResponse) GetState() *string {
	return r.State
}

// GetPostOffice returns the post office
func (r *AddressResponse) GetPostOffice() *string {
	return r.PostOffice
}

// GetPincode returns the pincode
func (r *AddressResponse) GetPincode() *string {
	return r.Pincode
}

// GetCountry returns the country
func (r *AddressResponse) GetCountry() *string {
	return r.Country
}

// GetCreatedAt returns the created at timestamp
func (r *AddressResponse) GetCreatedAt() string {
	return r.CreatedAt
}

// GetUpdatedAt returns the updated at timestamp
func (r *AddressResponse) GetUpdatedAt() string {
	return r.UpdatedAt
}
