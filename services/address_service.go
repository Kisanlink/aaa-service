package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/Kisanlink/aaa-service/pkg/errors"
)

// AddressService implements the AddressService interface
type AddressService struct {
	addressRepo  interfaces.AddressRepository
	cacheService interfaces.CacheService
	logger       interfaces.Logger
	validator    interfaces.Validator
}

// NewAddressService creates a new AddressService instance
func NewAddressService(
	addressRepo interfaces.AddressRepository,
	cacheService interfaces.CacheService,
	logger interfaces.Logger,
	validator interfaces.Validator,
) interfaces.AddressService {
	return &AddressService{
		addressRepo:  addressRepo,
		cacheService: cacheService,
		logger:       logger,
		validator:    validator,
	}
}

// CreateAddress creates a new address
func (s *AddressService) CreateAddress(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("Creating new address")

	// Type assertion for request
	createReq, ok := req.(*models.Address)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	// Validate address
	if err := s.validateAddress(createReq); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Build full address
	createReq.BuildFullAddress()

	// Create address in database
	if err := s.addressRepo.Create(ctx, createReq); err != nil {
		s.logger.Error("Failed to create address", "error", err)
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("address:%s", createReq.ID))

	s.logger.Info("Address created successfully", "addressID", createReq.ID)
	return createReq, nil
}

// GetAddressByID retrieves an address by ID with caching
func (s *AddressService) GetAddressByID(ctx context.Context, addressID string) (interface{}, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("address:%s", addressID)
	if cached, exists := s.cacheService.Get(cacheKey); exists {
		if address, ok := cached.(*models.Address); ok {
			s.logger.Debug("Address retrieved from cache", "addressID", addressID)
			return address, nil
		}
	}

	// Get from database
	address, err := s.addressRepo.GetByID(ctx, addressID)
	if err != nil {
		s.logger.Error("Failed to get address by ID", "addressID", addressID, "error", err)
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	// Cache the result
	s.cacheService.Set(cacheKey, address, 300) // Cache for 5 minutes

	return address, nil
}

// UpdateAddress updates an existing address
func (s *AddressService) UpdateAddress(ctx context.Context, req interface{}) (interface{}, error) {
	s.logger.Info("Updating address")

	// Type assertion for request
	updateReq, ok := req.(*models.Address)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	// Validate address
	if err := s.validateAddress(updateReq); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if address exists
	exists, err := s.addressRepo.Exists(ctx, updateReq.ID)
	if err != nil {
		s.logger.Error("Failed to check address existence", "addressID", updateReq.ID, "error", err)
		return nil, fmt.Errorf("failed to check address existence: %w", err)
	}

	if !exists {
		return nil, errors.NewNotFoundError("address not found")
	}

	// Build full address
	updateReq.BuildFullAddress()

	// Update address in database
	if err := s.addressRepo.Update(ctx, updateReq); err != nil {
		s.logger.Error("Failed to update address", "addressID", updateReq.ID, "error", err)
		return nil, fmt.Errorf("failed to update address: %w", err)
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("address:%s", updateReq.ID))

	s.logger.Info("Address updated successfully", "addressID", updateReq.ID)
	return updateReq, nil
}

// DeleteAddress soft deletes an address
func (s *AddressService) DeleteAddress(ctx context.Context, addressID string) (interface{}, error) {
	s.logger.Info("Deleting address", "addressID", addressID)

	// Check if address exists
	exists, err := s.addressRepo.Exists(ctx, addressID)
	if err != nil {
		s.logger.Error("Failed to check address existence", "addressID", addressID, "error", err)
		return nil, fmt.Errorf("failed to check address existence: %w", err)
	}

	if !exists {
		return nil, errors.NewNotFoundError("address not found")
	}

	// Delete address
	if err := s.addressRepo.Delete(ctx, addressID); err != nil {
		s.logger.Error("Failed to delete address", "addressID", addressID, "error", err)
		return nil, fmt.Errorf("failed to delete address: %w", err)
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("address:%s", addressID))

	s.logger.Info("Address deleted successfully", "addressID", addressID)
	return map[string]string{"message": "Address deleted successfully"}, nil
}

// SearchAddresses searches addresses by keyword
func (s *AddressService) SearchAddresses(ctx context.Context, keyword string, limit, offset int) (interface{}, error) {
	s.logger.Info("Searching addresses", "keyword", keyword, "limit", limit, "offset", offset)

	if strings.TrimSpace(keyword) == "" {
		return nil, fmt.Errorf("search keyword cannot be empty")
	}

	// Search addresses
	addresses, err := s.addressRepo.SearchByKeyword(ctx, keyword, limit, offset)
	if err != nil {
		s.logger.Error("Failed to search addresses", "error", err)
		return nil, fmt.Errorf("failed to search addresses: %w", err)
	}

	s.logger.Info("Address search completed", "count", len(addresses))
	return addresses, nil
}

// Helper methods

func (s *AddressService) validateAddress(address *models.Address) error {
	if address == nil {
		return fmt.Errorf("address cannot be nil")
	}

	// Basic validation - at least one field should be present
	if (address.House == nil || *address.House == "") &&
		(address.Street == nil || *address.Street == "") &&
		(address.District == nil || *address.District == "") &&
		(address.State == nil || *address.State == "") &&
		(address.Country == nil || *address.Country == "") {
		return fmt.Errorf("at least one address field must be provided")
	}

	return nil
}
