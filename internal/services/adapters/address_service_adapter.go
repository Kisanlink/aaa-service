package adapters

import (
	"context"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
)

// AddressServiceAdapter adapts the existing AddressService to the KYC AddressService interface
type AddressServiceAdapter struct {
	service interfaces.AddressService
}

// NewAddressServiceAdapter creates a new address service adapter
func NewAddressServiceAdapter(service interfaces.AddressService) *AddressServiceAdapter {
	return &AddressServiceAdapter{service: service}
}

// CreateAddress creates a new address record
func (a *AddressServiceAdapter) CreateAddress(ctx context.Context, address *models.Address) error {
	return a.service.CreateAddress(ctx, address)
}

// FindOrCreateAddress finds an existing address by full_address or creates a new one
// Returns (addressID string, wasCreated bool, error)
func (a *AddressServiceAdapter) FindOrCreateAddress(ctx context.Context, address *models.Address) (string, bool, error) {
	return a.service.FindOrCreateAddress(ctx, address)
}
