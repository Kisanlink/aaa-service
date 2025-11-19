package kyc

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	kycRequests "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/kyc"
	kycResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/kyc"
	"go.uber.org/zap"
)

// generateOTP generates OTP for Aadhaar verification
func (s *Service) generateOTP(ctx context.Context, req *kycRequests.GenerateOTPRequest, userID, authToken string) (*kycResponses.GenerateOTPResponse, error) {
	s.logger.Info("Generating OTP for Aadhaar verification",
		zap.String("user_id", userID),
		zap.String("aadhaar_masked", maskAadhaar(req.AadhaarNumber)))

	// 1. Validate request
	if err := req.Validate(); err != nil {
		s.logger.Error("OTP generation request validation failed",
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// 2. Check rate limits (check recent attempts for this Aadhaar number)
	if err := s.checkRateLimit(ctx, req.AadhaarNumber); err != nil {
		s.logger.Warn("Rate limit exceeded",
			zap.String("user_id", userID),
			zap.String("aadhaar_masked", maskAadhaar(req.AadhaarNumber)),
			zap.Error(err))
		return nil, err
	}

	// 3. Call Sandbox API to generate OTP
	sandboxResp, err := s.sandboxClient.GenerateOTP(ctx, req.AadhaarNumber, req.Consent.String(), authToken)
	if err != nil {
		s.logger.Error("Failed to generate OTP via Sandbox API",
			zap.String("user_id", userID),
			zap.String("aadhaar_masked", maskAadhaar(req.AadhaarNumber)),
			zap.Error(err))

		// Log audit event for failure
		s.auditService.LogUserActionWithError(ctx, userID, "aadhaar_otp_generation_failed", "aadhaar_verification", "", err, map[string]interface{}{
			"aadhaar_masked": maskAadhaar(req.AadhaarNumber),
		})

		return nil, fmt.Errorf("failed to generate OTP: %w", err)
	}

	// 4. Create verification record in database
	verification := &models.AadhaarVerification{
		ID:                 generateVerificationID(),
		UserID:             userID,
		AadhaarNumber:      req.AadhaarNumber,
		ReferenceID:        strconv.Itoa(sandboxResp.Data.ReferenceID),
		TransactionID:      sandboxResp.TransactionID,
		VerificationStatus: "PENDING",
		KYCStatus:          "PENDING",
		OTPRequestedAt:     timePtr(time.Now()),
		Attempts:           0,
		CreatedBy:          userID,
	}

	if err := s.aadhaarRepo.Create(ctx, verification); err != nil {
		s.logger.Error("Failed to create verification record",
			zap.String("user_id", userID),
			zap.String("verification_id", verification.ID),
			zap.Error(err))

		// Log audit event for failure
		s.auditService.LogUserActionWithError(ctx, userID, "aadhaar_verification_record_creation_failed", "aadhaar_verification", verification.ID, err, map[string]interface{}{
			"reference_id":   verification.ReferenceID,
			"transaction_id": verification.TransactionID,
		})

		return nil, fmt.Errorf("failed to save verification record: %w", err)
	}

	// 5. Log audit event for success
	s.auditService.LogUserAction(ctx, userID, "aadhaar_otp_generated", "aadhaar_verification", verification.ID, map[string]interface{}{
		"reference_id":   verification.ReferenceID,
		"transaction_id": verification.TransactionID,
		"aadhaar_masked": maskAadhaar(req.AadhaarNumber),
	})

	// 6. Calculate OTP expiration time
	expiresAt := time.Now().Add(time.Duration(s.config.OTPExpirationSeconds) * time.Second).Unix()

	s.logger.Info("OTP generated successfully",
		zap.String("user_id", userID),
		zap.String("verification_id", verification.ID),
		zap.String("reference_id", verification.ReferenceID),
		zap.Int64("expires_at", expiresAt))

	// 7. Return response
	return &kycResponses.GenerateOTPResponse{
		StatusCode:    200,
		Message:       sandboxResp.Data.Message,
		ReferenceID:   verification.ReferenceID,
		TransactionID: sandboxResp.TransactionID,
		Timestamp:     sandboxResp.Timestamp,
		ExpiresAt:     expiresAt,
	}, nil
}

// checkRateLimit checks if user has exceeded OTP generation rate limits
func (s *Service) checkRateLimit(ctx context.Context, aadhaarNumber string) error {
	// TODO: Implement comprehensive rate limiting logic
	// For now, we'll just return nil to allow all requests
	// In production, this should:
	// 1. Check if there are more than OTPMaxAttempts in the last OTPCooldownSeconds
	// 2. Track by both user_id and aadhaar_number
	// 3. Use Redis or in-memory cache for rate limit tracking
	// 4. Return appropriate error if limit exceeded

	s.logger.Debug("Rate limit check passed",
		zap.String("aadhaar_masked", maskAadhaar(aadhaarNumber)))

	return nil
}

// generateVerificationID generates a unique verification ID
func generateVerificationID() string {
	return fmt.Sprintf("VERIFY%d", time.Now().UnixNano())
}
