package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/entities/models"
	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"go.uber.org/zap"
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
		s.logger.Error("Failed to create address", zap.Error(err))
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("address:%s", createReq.ID))

	s.logger.Info("Address created successfully", zap.String("addressID", createReq.ID))
	return createReq, nil
}

// GetAddressByID retrieves an address by ID with caching
func (s *AddressService) GetAddressByID(ctx context.Context, addressID string) (interface{}, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("address:%s", addressID)
	if cached, exists := s.cacheService.Get(cacheKey); exists {
		if address, ok := cached.(*models.Address); ok {
			s.logger.Debug("Address retrieved from cache", zap.String("addressID", addressID))
			return address, nil
		}
	}

	// Get from database
	address, err := s.addressRepo.GetByID(ctx, addressID)
	if err != nil {
		s.logger.Error("Failed to get address by ID", zap.String("addressID", addressID), zap.Error(err))
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	// Cache the result
	s.cacheService.Set(cacheKey, address, 300) // Cache for 5 minutes

	return address, nil
}

// GetAddressByUserID retrieves addresses by user ID
func (s *AddressService) GetAddressByUserID(ctx context.Context, userID string) (interface{}, error) {
	s.logger.Info("Getting addresses by user ID", zap.String("userID", userID))

	addresses, err := s.addressRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get addresses by user ID", zap.String("userID", userID), zap.Error(err))
		return nil, fmt.Errorf("failed to get addresses: %w", err)
	}

	return addresses, nil
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

	// Check if address exists by trying to get it
	_, err := s.addressRepo.GetByID(ctx, updateReq.ID)
	if err != nil {
		s.logger.Error("Failed to check address existence", zap.String("addressID", updateReq.ID), zap.Error(err))
		return nil, errors.NewNotFoundError("address not found")
	}

	// Build full address
	updateReq.BuildFullAddress()

	// Update address in database
	if err := s.addressRepo.Update(ctx, updateReq); err != nil {
		s.logger.Error("Failed to update address", zap.String("addressID", updateReq.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to update address: %w", err)
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("address:%s", updateReq.ID))

	s.logger.Info("Address updated successfully", zap.String("addressID", updateReq.ID))
	return updateReq, nil
}

// DeleteAddress soft deletes an address
func (s *AddressService) DeleteAddress(ctx context.Context, addressID string) error {
	s.logger.Info("Deleting address")

	// Delete address
	if err := s.addressRepo.Delete(ctx, addressID); err != nil {
		s.logger.Error("Failed to delete address")
		return fmt.Errorf("failed to delete address: %w", err)
	}

	// Clear cache
	s.cacheService.Delete(fmt.Sprintf("address:%s", addressID))

	s.logger.Info("Address deleted successfully")
	return nil
}

// ListAddresses lists addresses with filters
func (s *AddressService) ListAddresses(ctx context.Context, filters interface{}) (interface{}, error) {
	s.logger.Info("Listing addresses")

	// Default pagination
	limit, offset := 10, 0

	// Extract limit and offset from filters if available
	if filterMap, ok := filters.(map[string]interface{}); ok {
		if l, exists := filterMap["limit"]; exists {
			if limitInt, ok := l.(int); ok {
				limit = limitInt
			}
		}
		if o, exists := filterMap["offset"]; exists {
			if offsetInt, ok := o.(int); ok {
				offset = offsetInt
			}
		}
	}

	addresses, err := s.addressRepo.List(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list addresses", zap.Error(err))
		return nil, fmt.Errorf("failed to list addresses: %w", err)
	}

	s.logger.Info("Address listing completed", zap.Int("count", len(addresses)))
	return addresses, nil
}

// SearchAddresses searches addresses by keyword
func (s *AddressService) SearchAddresses(ctx context.Context, keyword string, limit, offset int) (interface{}, error) {
	s.logger.Info("Searching addresses", zap.String("keyword", keyword), zap.Int("limit", limit), zap.Int("offset", offset))

	if strings.TrimSpace(keyword) == "" {
		return nil, fmt.Errorf("search keyword cannot be empty")
	}

	// Search addresses using the correct repository method
	addresses, err := s.addressRepo.Search(ctx, keyword, limit, offset)
	if err != nil {
		s.logger.Error("Failed to search addresses", zap.Error(err))
		return nil, fmt.Errorf("failed to search addresses: %w", err)
	}

	s.logger.Info("Address search completed", zap.Int("count", len(addresses)))
	return addresses, nil
}

// ValidateAddress validates an address
func (s *AddressService) ValidateAddress(ctx context.Context, address interface{}) error {
	s.logger.Info("Validating address")

	addr, ok := address.(*models.Address)
	if !ok {
		return fmt.Errorf("invalid address type")
	}

	return s.validateAddress(addr)
}

// GeocodingAddress performs geocoding on an address
func (s *AddressService) GeocodingAddress(ctx context.Context, address interface{}) (interface{}, error) {
	s.logger.Info("Geocoding address")

	addr, ok := address.(*models.Address)
	if !ok {
		return nil, fmt.Errorf("invalid address type")
	}

	// Placeholder implementation - integrate with actual geocoding service
	result := map[string]interface{}{
		"address":   addr,
		"latitude":  0.0,
		"longitude": 0.0,
		"accuracy":  "placeholder",
	}

	s.logger.Info("Address geocoding completed", zap.String("addressID", addr.ID))
	return result, nil
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
