package principals

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	principalRequests "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/principals"
	principalResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/principals"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/organizations"
	principalRepo "github.com/Kisanlink/aaa-service/v2/internal/repositories/principals"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"go.uber.org/zap"
)

// Service handles business logic for principal and service operations
type Service struct {
	orgRepo       *organizations.OrganizationRepository
	principalRepo *principalRepo.PrincipalRepository
	serviceRepo   *principalRepo.ServiceRepository
	validator     interfaces.Validator
	logger        *zap.Logger
}

// NewPrincipalService creates a new principal service instance
func NewPrincipalService(
	orgRepo *organizations.OrganizationRepository,
	principalRepo *principalRepo.PrincipalRepository,
	serviceRepo *principalRepo.ServiceRepository,
	validator interfaces.Validator,
	logger *zap.Logger,
) *Service {
	return &Service{
		orgRepo:       orgRepo,
		principalRepo: principalRepo,
		serviceRepo:   serviceRepo,
		validator:     validator,
		logger:        logger,
	}
}

// CreatePrincipal creates a new principal with proper validation and business logic
func (s *Service) CreatePrincipal(ctx context.Context, req *principalRequests.CreatePrincipalRequest) (*principalResponses.PrincipalResponse, error) {
	s.logger.Info("Creating new principal", zap.String("name", req.Name), zap.String("type", req.Type))

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.Error("Principal creation validation failed", zap.Error(err))
		return nil, errors.NewValidationError("invalid principal data", err.Error())
	}

	// Validate principal type and ID consistency
	if err := s.validatePrincipalTypeAndID(req); err != nil {
		s.logger.Error("Principal type and ID validation failed", zap.Error(err))
		return nil, err
	}

	// Verify organization exists and is active if specified
	if req.OrganizationID != nil && *req.OrganizationID != "" {
		org, err := s.orgRepo.GetByID(ctx, *req.OrganizationID)
		if err != nil || org == nil {
			s.logger.Warn("Organization not found", zap.String("org_id", *req.OrganizationID))
			return nil, errors.NewNotFoundError("organization not found")
		}
		if !org.IsActive {
			s.logger.Warn("Organization is inactive", zap.String("org_id", *req.OrganizationID))
			return nil, errors.NewValidationError("cannot create principal in inactive organization")
		}
	}

	// Create principal model
	principal := models.NewPrincipal(models.PrincipalType(req.Type), req.Name)
	principal.OrganizationID = req.OrganizationID
	principal.Metadata = req.Metadata

	// Set type-specific ID
	if req.Type == "user" && req.UserID != nil {
		principal.UserID = req.UserID
	} else if req.Type == "service" && req.ServiceID != nil {
		principal.ServiceID = req.ServiceID
	}

	// Save principal to repository
	if err := s.principalRepo.Create(ctx, principal); err != nil {
		s.logger.Error("Failed to create principal in repository", zap.Error(err))
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.NewConflictError("principal with this information already exists")
		}
		return nil, errors.NewInternalError(err)
	}

	s.logger.Info("Principal created successfully",
		zap.String("principal_id", principal.ID),
		zap.String("name", principal.Name),
		zap.String("type", string(principal.Type)))

	// Convert to response format
	response := &principalResponses.PrincipalResponse{
		ID:             principal.ID,
		Type:           string(principal.Type),
		UserID:         principal.UserID,
		ServiceID:      principal.ServiceID,
		Name:           principal.Name,
		OrganizationID: principal.OrganizationID,
		IsActive:       principal.IsActive,
		Metadata:       principal.Metadata,
		CreatedAt:      &principal.CreatedAt,
		UpdatedAt:      &principal.UpdatedAt,
	}

	return response, nil
}

// CreateService creates a new service with proper validation and business logic
func (s *Service) CreateService(ctx context.Context, req *principalRequests.CreateServiceRequest) (*principalResponses.ServiceResponse, error) {
	s.logger.Info("Creating new service", zap.String("name", req.Name))

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.Error("Service creation validation failed", zap.Error(err))
		return nil, errors.NewValidationError("invalid service data", err.Error())
	}

	// Verify organization exists and is active
	org, err := s.orgRepo.GetByID(ctx, req.OrganizationID)
	if err != nil || org == nil {
		s.logger.Warn("Organization not found", zap.String("org_id", req.OrganizationID))
		return nil, errors.NewNotFoundError("organization not found")
	}
	if !org.IsActive {
		s.logger.Warn("Organization is inactive", zap.String("org_id", req.OrganizationID))
		return nil, errors.NewValidationError("cannot create service in inactive organization")
	}

	// Hash the API key for security
	hashedAPIKey := s.hashAPIKey(req.APIKey)

	// Create service model
	service := models.NewService(req.Name, req.Description, req.OrganizationID, hashedAPIKey)
	service.Metadata = req.Metadata

	// Save service to repository
	if err := s.serviceRepo.Create(ctx, service); err != nil {
		s.logger.Error("Failed to create service in repository", zap.Error(err))
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.NewConflictError("service with this information already exists")
		}
		return nil, errors.NewInternalError(err)
	}

	s.logger.Info("Service created successfully",
		zap.String("service_id", service.ID),
		zap.String("name", service.Name),
		zap.String("org_id", service.OrganizationID))

	// Convert to response format
	response := &principalResponses.ServiceResponse{
		ID:             service.ID,
		Name:           service.Name,
		Description:    service.Description,
		OrganizationID: service.OrganizationID,
		IsActive:       service.IsActive,
		Metadata:       service.Metadata,
		CreatedAt:      &service.CreatedAt,
		UpdatedAt:      &service.UpdatedAt,
	}

	return response, nil
}

// GetPrincipal retrieves a principal by ID
func (s *Service) GetPrincipal(ctx context.Context, principalID string) (*principalResponses.PrincipalResponse, error) {
	s.logger.Info("Retrieving principal", zap.String("principal_id", principalID))

	// Get principal from repository
	principal, err := s.principalRepo.GetByID(ctx, principalID)
	if err != nil {
		s.logger.Error("Failed to retrieve principal", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	if principal == nil {
		return nil, errors.NewNotFoundError("principal not found")
	}

	// Convert to response format
	response := &principalResponses.PrincipalResponse{
		ID:             principal.ID,
		Type:           string(principal.Type),
		UserID:         principal.UserID,
		ServiceID:      principal.ServiceID,
		Name:           principal.Name,
		OrganizationID: principal.OrganizationID,
		IsActive:       principal.IsActive,
		Metadata:       principal.Metadata,
		CreatedAt:      &principal.CreatedAt,
		UpdatedAt:      &principal.UpdatedAt,
	}

	return response, nil
}

// GetService retrieves a service by ID
func (s *Service) GetService(ctx context.Context, serviceID string) (*principalResponses.ServiceResponse, error) {
	s.logger.Info("Retrieving service", zap.String("service_id", serviceID))

	// Get service from repository
	service, err := s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		s.logger.Error("Failed to retrieve service", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	if service == nil {
		return nil, errors.NewNotFoundError("service not found")
	}

	// Convert to response format
	response := &principalResponses.ServiceResponse{
		ID:             service.ID,
		Name:           service.Name,
		Description:    service.Description,
		OrganizationID: service.OrganizationID,
		IsActive:       service.IsActive,
		Metadata:       service.Metadata,
		CreatedAt:      &service.CreatedAt,
		UpdatedAt:      &service.UpdatedAt,
	}

	return response, nil
}

// UpdatePrincipal updates an existing principal
func (s *Service) UpdatePrincipal(ctx context.Context, principalID string, req *principalRequests.UpdatePrincipalRequest) (*principalResponses.PrincipalResponse, error) {
	s.logger.Info("Updating principal", zap.String("principal_id", principalID))

	// Validate request
	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.Error("Principal update validation failed", zap.Error(err))
		return nil, errors.NewValidationError("invalid principal data", err.Error())
	}

	// Get existing principal
	principal, err := s.principalRepo.GetByID(ctx, principalID)
	if err != nil {
		s.logger.Error("Failed to retrieve principal for update", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	if principal == nil {
		return nil, errors.NewNotFoundError("principal not found")
	}

	// Update fields
	if req.Name != nil {
		principal.Name = *req.Name
	}
	if req.OrganizationID != nil {
		principal.OrganizationID = req.OrganizationID
	}
	if req.IsActive != nil {
		principal.IsActive = *req.IsActive
	}
	if req.Metadata != nil {
		principal.Metadata = req.Metadata
	}

	// Save changes
	if err := s.principalRepo.Update(ctx, principal); err != nil {
		s.logger.Error("Failed to update principal", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	s.logger.Info("Principal updated successfully", zap.String("principal_id", principalID))

	// Convert to response format
	response := &principalResponses.PrincipalResponse{
		ID:             principal.ID,
		Type:           string(principal.Type),
		UserID:         principal.UserID,
		ServiceID:      principal.ServiceID,
		Name:           principal.Name,
		OrganizationID: principal.OrganizationID,
		IsActive:       principal.IsActive,
		Metadata:       principal.Metadata,
		CreatedAt:      &principal.CreatedAt,
		UpdatedAt:      &principal.UpdatedAt,
	}

	return response, nil
}

// DeletePrincipal deletes a principal
func (s *Service) DeletePrincipal(ctx context.Context, principalID string, deletedBy string) error {
	s.logger.Info("Deleting principal", zap.String("principal_id", principalID))

	// Get principal to check if it exists
	principal, err := s.principalRepo.GetByID(ctx, principalID)
	if err != nil {
		s.logger.Error("Failed to retrieve principal for deletion", zap.Error(err))
		return errors.NewInternalError(err)
	}

	if principal == nil {
		return errors.NewNotFoundError("principal not found")
	}

	// Soft delete the principal
	if err := s.principalRepo.SoftDelete(ctx, principalID, deletedBy); err != nil {
		s.logger.Error("Failed to delete principal", zap.Error(err))
		return errors.NewInternalError(err)
	}

	s.logger.Info("Principal deleted successfully", zap.String("principal_id", principalID))
	return nil
}

// DeleteService deletes a service
func (s *Service) DeleteService(ctx context.Context, serviceID string, deletedBy string) error {
	s.logger.Info("Deleting service", zap.String("service_id", serviceID))

	// Get service to check if it exists
	service, err := s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		s.logger.Error("Failed to retrieve service for deletion", zap.Error(err))
		return errors.NewInternalError(err)
	}

	if service == nil {
		return errors.NewNotFoundError("service not found")
	}

	// Soft delete the service
	if err := s.serviceRepo.SoftDelete(ctx, serviceID, deletedBy); err != nil {
		s.logger.Error("Failed to delete service", zap.Error(err))
		return errors.NewInternalError(err)
	}

	s.logger.Info("Service deleted successfully", zap.String("service_id", serviceID))
	return nil
}

// ListPrincipals retrieves principals with pagination and filtering
func (s *Service) ListPrincipals(ctx context.Context, limit, offset int, principalType, organizationID string) ([]*principalResponses.PrincipalResponse, error) {
	s.logger.Info("Listing principals", zap.Int("limit", limit), zap.Int("offset", offset))

	var principals []*models.Principal
	var err error

	// Apply filters based on parameters
	if principalType != "" && organizationID != "" {
		principals, err = s.principalRepo.GetPrincipalsByTypeAndOrganization(ctx, principalType, organizationID, limit, offset)
	} else if principalType != "" {
		principals, err = s.principalRepo.GetByType(ctx, principalType, limit, offset)
	} else if organizationID != "" {
		principals, err = s.principalRepo.GetByOrganization(ctx, organizationID, limit, offset)
	} else {
		principals, err = s.principalRepo.ListActive(ctx, limit, offset)
	}

	if err != nil {
		s.logger.Error("Failed to list principals", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Convert to response format
	responses := make([]*principalResponses.PrincipalResponse, len(principals))
	for i, principal := range principals {
		responses[i] = &principalResponses.PrincipalResponse{
			ID:             principal.ID,
			Type:           string(principal.Type),
			UserID:         principal.UserID,
			ServiceID:      principal.ServiceID,
			Name:           principal.Name,
			OrganizationID: principal.OrganizationID,
			IsActive:       principal.IsActive,
			Metadata:       principal.Metadata,
			CreatedAt:      &principal.CreatedAt,
			UpdatedAt:      &principal.UpdatedAt,
		}
	}

	return responses, nil
}

// CountPrincipals returns the count of principals matching the filters
func (s *Service) CountPrincipals(ctx context.Context, principalType, organizationID string) (int64, error) {
	if principalType != "" && organizationID != "" {
		return s.principalRepo.CountByTypeAndOrganization(ctx, principalType, organizationID)
	} else if principalType != "" {
		return s.principalRepo.CountByType(ctx, principalType)
	} else if organizationID != "" {
		return s.principalRepo.CountByOrganization(ctx, organizationID)
	}
	return s.principalRepo.CountActive(ctx)
}

// ListServices retrieves services with pagination and filtering
func (s *Service) ListServices(ctx context.Context, limit, offset int, organizationID string) ([]*principalResponses.ServiceResponse, error) {
	s.logger.Info("Listing services", zap.Int("limit", limit), zap.Int("offset", offset))

	var services []*models.Service
	var err error

	// Apply filters based on parameters
	if organizationID != "" {
		services, err = s.serviceRepo.GetByOrganization(ctx, organizationID, limit, offset)
	} else {
		services, err = s.serviceRepo.ListActive(ctx, limit, offset)
	}

	if err != nil {
		s.logger.Error("Failed to list services", zap.Error(err))
		return nil, errors.NewInternalError(err)
	}

	// Convert to response format
	responses := make([]*principalResponses.ServiceResponse, len(services))
	for i, service := range services {
		responses[i] = &principalResponses.ServiceResponse{
			ID:             service.ID,
			Name:           service.Name,
			Description:    service.Description,
			OrganizationID: service.OrganizationID,
			IsActive:       service.IsActive,
			Metadata:       service.Metadata,
			CreatedAt:      &service.CreatedAt,
			UpdatedAt:      &service.UpdatedAt,
		}
	}

	return responses, nil
}

// CountServices returns the count of services matching the filters
func (s *Service) CountServices(ctx context.Context, organizationID string) (int64, error) {
	if organizationID != "" {
		return s.serviceRepo.CountByOrganization(ctx, organizationID)
	}
	return s.serviceRepo.CountActive(ctx)
}

// GenerateAPIKey generates a secure API key for services
func (s *Service) GenerateAPIKey() (string, error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to hex string
	apiKey := hex.EncodeToString(bytes)
	return apiKey, nil
}

// ValidateAPIKey validates an API key against a stored hash
func (s *Service) ValidateAPIKey(apiKey, storedHash string) bool {
	// Hash the provided API key
	hashedAPIKey := s.hashAPIKey(apiKey)

	// Compare with stored hash
	return hashedAPIKey == storedHash
}

// hashAPIKey hashes an API key using SHA-256
func (s *Service) hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// validatePrincipalTypeAndID validates that the principal type and ID are consistent
func (s *Service) validatePrincipalTypeAndID(req *principalRequests.CreatePrincipalRequest) error {
	if req.Type == "user" {
		if req.UserID == nil || *req.UserID == "" {
			return errors.NewValidationError("user ID is required for user-type principals")
		}
		if req.ServiceID != nil && *req.ServiceID != "" {
			return errors.NewValidationError("service ID should not be set for user-type principals")
		}
	} else if req.Type == "service" {
		if req.ServiceID == nil || *req.ServiceID == "" {
			return errors.NewValidationError("service ID is required for service-type principals")
		}
		if req.UserID != nil && *req.UserID != "" {
			return errors.NewValidationError("user ID should not be set for service-type principals")
		}
	} else {
		return errors.NewValidationError("invalid principal type")
	}

	return nil
}
