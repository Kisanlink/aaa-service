package contacts

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	contactRequests "github.com/Kisanlink/aaa-service/internal/entities/requests/contacts"
	contactResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/contacts"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	contactRepo "github.com/Kisanlink/aaa-service/internal/repositories/contacts"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"go.uber.org/zap"
)

// ContactService handles business logic for Contact entities
type ContactService struct {
	contactRepo *contactRepo.ContactRepository
	cache       interfaces.CacheService
	logger      interfaces.Logger
	validator   interfaces.Validator
}

// NewContactService creates a new ContactService instance
func NewContactService(
	repo *contactRepo.ContactRepository,
	cache interfaces.CacheService,
	logger interfaces.Logger,
	validator interfaces.Validator,
) *ContactService {
	return &ContactService{
		contactRepo: repo,
		cache:       cache,
		logger:      logger,
		validator:   validator,
	}
}

// CreateContact creates a new contact with proper validation and business logic
func (s *ContactService) CreateContact(ctx context.Context, req *contactRequests.CreateContactRequest) (*contactResponses.ContactResponse, error) {
	s.logger.Info("Creating new contact", zap.String("type", req.Type), zap.String("userID", req.UserID))

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.Error("Contact creation validation failed", zap.Error(err))
		return nil, errors.NewValidationError("invalid contact data", err.Error())
	}

	// Create contact model using the new constructor
	contact := models.NewContact(req.UserID, req.Type, req.Value)
	contact.Description = req.Description
	contact.IsPrimary = req.IsPrimary
	contact.IsActive = req.IsActive
	contact.CountryCode = req.CountryCode

	// Save contact to repository
	err := s.contactRepo.Create(ctx, contact)
	if err != nil {
		s.logger.Error("Failed to create contact in repository", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	s.logger.Info("Contact created successfully", zap.String("contact_id", contact.ID))

	// Convert to response format
	response := &contactResponses.ContactResponse{
		ID:          contact.ID,
		UserID:      contact.UserID,
		Type:        contact.Type,
		Value:       contact.Value,
		Description: contact.Description,
		IsPrimary:   contact.IsPrimary,
		IsActive:    contact.IsActive,
		IsVerified:  contact.IsVerified,
		VerifiedAt:  contact.VerifiedAt,
		VerifiedBy:  contact.VerifiedBy,
		CountryCode: contact.CountryCode,
		CreatedAt:   contact.CreatedAt,
		UpdatedAt:   contact.UpdatedAt,
	}

	return response, nil
}

// GetContact retrieves a contact by ID
func (s *ContactService) GetContact(ctx context.Context, id string) (*contactResponses.ContactResponse, error) {
	s.logger.Info("Getting contact", zap.String("id", id))

	// Try to get from cache first
	cacheKey := fmt.Sprintf("contact:%s", id)
	if cached, exists := s.cache.Get(cacheKey); exists && cached != nil {
		if contact, ok := cached.(*contactResponses.ContactResponse); ok {
			s.logger.Debug("Contact retrieved from cache", zap.String("id", id))
			return contact, nil
		}
	}

	// Get from repository
	contact, err := s.contactRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get contact from repository", zap.Error(err))
		return nil, errors.NewNotFoundError("contact not found")
	}

	if contact == nil {
		return nil, errors.NewNotFoundError("contact not found")
	}

	// Convert to response format
	response := &contactResponses.ContactResponse{
		ID:          contact.ID,
		UserID:      contact.UserID,
		Type:        contact.Type,
		Value:       contact.Value,
		Description: contact.Description,
		IsPrimary:   contact.IsPrimary,
		IsActive:    contact.IsActive,
		IsVerified:  contact.IsVerified,
		VerifiedAt:  contact.VerifiedAt,
		VerifiedBy:  contact.VerifiedBy,
		CountryCode: contact.CountryCode,
		CreatedAt:   contact.CreatedAt,
		UpdatedAt:   contact.UpdatedAt,
	}

	// Cache the response
	if err := s.cache.Set(cacheKey, response, 300); err != nil {
		s.logger.Warn("Failed to cache contact response", zap.Error(err))
	}

	return response, nil
}

// UpdateContact updates an existing contact
func (s *ContactService) UpdateContact(ctx context.Context, id string, req *contactRequests.UpdateContactRequest) (*contactResponses.ContactResponse, error) {
	s.logger.Info("Updating contact", zap.String("id", id))

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.Error("Contact update validation failed", zap.Error(err))
		return nil, errors.NewValidationError("invalid contact data", err.Error())
	}

	// Get existing contact
	existingContact, err := s.contactRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get contact for update", zap.Error(err))
		return nil, errors.NewNotFoundError("contact not found")
	}

	if existingContact == nil {
		return nil, errors.NewNotFoundError("contact not found")
	}

	// Update fields if provided
	if req.Type != nil {
		existingContact.Type = *req.Type
	}
	if req.Value != nil {
		existingContact.Value = *req.Value
	}
	if req.Description != nil {
		existingContact.Description = req.Description
	}
	if req.IsPrimary != nil {
		existingContact.IsPrimary = *req.IsPrimary
	}
	if req.IsActive != nil {
		existingContact.IsActive = *req.IsActive
	}
	if req.CountryCode != nil {
		existingContact.CountryCode = req.CountryCode
	}
	if req.IsVerified != nil {
		existingContact.IsVerified = *req.IsVerified
	}
	if req.VerifiedBy != nil {
		existingContact.VerifiedBy = req.VerifiedBy
	}

	// Save updated contact
	err = s.contactRepo.Update(ctx, existingContact)
	if err != nil {
		s.logger.Error("Failed to update contact in repository", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	s.logger.Info("Contact updated successfully", zap.String("contact_id", id))

	// Clear cache
	cacheKey := fmt.Sprintf("contact:%s", id)
	if err := s.cache.Delete(cacheKey); err != nil {
		s.logger.Warn("Failed to delete contact cache", zap.Error(err))
	}

	// Convert to response format
	response := &contactResponses.ContactResponse{
		ID:          existingContact.ID,
		UserID:      existingContact.UserID,
		Type:        existingContact.Type,
		Value:       existingContact.Value,
		Description: existingContact.Description,
		IsPrimary:   existingContact.IsPrimary,
		IsActive:    existingContact.IsActive,
		IsVerified:  existingContact.IsVerified,
		VerifiedAt:  existingContact.VerifiedAt,
		VerifiedBy:  existingContact.VerifiedBy,
		CountryCode: existingContact.CountryCode,
		CreatedAt:   existingContact.CreatedAt,
		UpdatedAt:   existingContact.UpdatedAt,
	}

	return response, nil
}

// DeleteContact deletes a contact by ID
func (s *ContactService) DeleteContact(ctx context.Context, id string, deletedBy string) error {
	s.logger.Info("Deleting contact", zap.String("id", id), zap.String("deleted_by", deletedBy))

	// Check if contact exists
	exists, err := s.contactRepo.Exists(ctx, id)
	if err != nil {
		s.logger.Error("Failed to check contact existence", zap.Error(err))
		return errors.NewInternalError(err)
	}

	if !exists {
		return errors.NewNotFoundError("contact not found")
	}

	// Soft delete the contact
	err = s.contactRepo.SoftDelete(ctx, id, deletedBy)
	if err != nil {
		s.logger.Error("Failed to delete contact", zap.Error(err))
		return errors.NewInternalError(err)
	}

	s.logger.Info("Contact deleted successfully", zap.String("contact_id", id))

	// Clear cache
	cacheKey := fmt.Sprintf("contact:%s", id)
	if err := s.cache.Delete(cacheKey); err != nil {
		s.logger.Warn("Failed to delete contact cache", zap.Error(err))
	}

	return nil
}

// ListContacts retrieves contacts with pagination
func (s *ContactService) ListContacts(ctx context.Context, limit, offset int) (*contactResponses.ContactListResponse, error) {
	s.logger.Info("Listing contacts", zap.Int("limit", limit), zap.Int("offset", offset))

	// Get contacts from repository
	contacts, err := s.contactRepo.List(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list contacts", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Get total count
	total, err := s.contactRepo.Count(ctx)
	if err != nil {
		s.logger.Error("Failed to count contacts", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Convert to response format
	contactResponseList := make([]*contactResponses.ContactResponse, len(contacts))
	for i, contact := range contacts {
		contactResponseList[i] = &contactResponses.ContactResponse{
			ID:          contact.ID,
			UserID:      contact.UserID,
			Type:        contact.Type,
			Value:       contact.Value,
			Description: contact.Description,
			IsPrimary:   contact.IsPrimary,
			IsActive:    contact.IsActive,
			IsVerified:  contact.IsVerified,
			VerifiedAt:  contact.VerifiedAt,
			VerifiedBy:  contact.VerifiedBy,
			CountryCode: contact.CountryCode,
			CreatedAt:   contact.CreatedAt,
			UpdatedAt:   contact.UpdatedAt,
		}
	}

	response := &contactResponses.ContactListResponse{
		Contacts: contactResponseList,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}

	return response, nil
}

// GetContactsByUser retrieves contacts by user ID
func (s *ContactService) GetContactsByUser(ctx context.Context, userID string, limit, offset int) (*contactResponses.ContactListResponse, error) {
	s.logger.Info("Getting contacts by user", zap.String("userID", userID))

	// Get contacts from repository
	contacts, err := s.contactRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get contacts by user", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Convert to response format
	contactResponseList := make([]*contactResponses.ContactResponse, len(contacts))
	for i, contact := range contacts {
		contactResponseList[i] = &contactResponses.ContactResponse{
			ID:          contact.ID,
			UserID:      contact.UserID,
			Type:        contact.Type,
			Value:       contact.Value,
			Description: contact.Description,
			IsPrimary:   contact.IsPrimary,
			IsActive:    contact.IsActive,
			CreatedAt:   contact.CreatedAt,
			UpdatedAt:   contact.UpdatedAt,
		}
	}

	response := &contactResponses.ContactListResponse{
		Contacts: contactResponseList,
		Total:    int64(len(contacts)),
		Limit:    limit,
		Offset:   offset,
	}

	return response, nil
}
