package kyc

import (
	"context"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	kycRequests "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/kyc"
	kycResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/kyc"
	"github.com/Kisanlink/aaa-service/v2/internal/repositories/kyc"
	"go.uber.org/zap"
)

// Config holds KYC service configuration
type Config struct {
	OTPExpirationSeconds int
	OTPMaxAttempts       int
	OTPCooldownSeconds   int
	PhotoMaxSizeMB       int
}

// Service implements KYC operations for Aadhaar verification
type Service struct {
	aadhaarRepo    kyc.AadhaarVerificationRepository
	userService    UserService
	addressService AddressService
	sandboxClient  *SandboxClient
	auditService   AuditService
	logger         *zap.Logger
	config         *Config
}

// NewService creates a new KYC service instance with all required dependencies
func NewService(
	aadhaarRepo kyc.AadhaarVerificationRepository,
	userService UserService,
	addressService AddressService,
	sandboxClient *SandboxClient,
	auditService AuditService,
	logger *zap.Logger,
	config *Config,
) *Service {
	return &Service{
		aadhaarRepo:    aadhaarRepo,
		userService:    userService,
		addressService: addressService,
		sandboxClient:  sandboxClient,
		auditService:   auditService,
		logger:         logger,
		config:         config,
	}
}

// UserService defines the interface for user profile operations
type UserService interface {
	// Update updates user profile fields
	Update(ctx context.Context, userID string, updates map[string]interface{}) error
	// GetProfile retrieves the user profile by user ID
	GetProfile(ctx context.Context, userID string) (*models.UserProfile, error)
}

// AddressService defines the interface for address operations
type AddressService interface {
	// CreateAddress creates a new address record
	CreateAddress(ctx context.Context, address *models.Address) error
	// FindOrCreateAddress finds an existing address by full_address or creates a new one
	// Returns (addressID string, wasCreated bool, error)
	FindOrCreateAddress(ctx context.Context, address *models.Address) (string, bool, error)
	// GetAddressByID retrieves an address by ID
	GetAddressByID(ctx context.Context, addressID string) (*models.Address, error)
}

// AuditService defines the interface for audit logging
type AuditService interface {
	// LogUserAction logs a successful user action
	LogUserAction(ctx context.Context, userID, action, resource, resourceID string, details map[string]interface{})
	// LogUserActionWithError logs a failed user action
	LogUserActionWithError(ctx context.Context, userID, action, resource, resourceID string, err error, details map[string]interface{})
}

// GenerateOTP generates OTP for Aadhaar verification
func (s *Service) GenerateOTP(ctx context.Context, req *kycRequests.GenerateOTPRequest, userID, authToken string) (*kycResponses.GenerateOTPResponse, error) {
	// Implementation in generate_otp.go
	return s.generateOTP(ctx, req, userID, authToken)
}

// VerifyOTP verifies OTP and updates user profile with Aadhaar data
func (s *Service) VerifyOTP(ctx context.Context, req *kycRequests.VerifyOTPRequest, userID, authToken string) (*kycResponses.VerifyOTPResponse, error) {
	// Implementation in verify_otp.go
	return s.verifyOTP(ctx, req, userID, authToken)
}

// GetKYCStatus retrieves KYC verification status for a user
func (s *Service) GetKYCStatus(ctx context.Context, userID string) (*kycResponses.KYCStatusResponse, error) {
	// Implementation in get_status.go
	return s.getKYCStatus(ctx, userID)
}

// Helper function to create a time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}
