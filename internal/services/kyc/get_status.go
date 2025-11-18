package kyc

import (
	"context"
	"fmt"

	kycResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/kyc"
	"go.uber.org/zap"
)

// getKYCStatus retrieves KYC verification status for a user
func (s *Service) getKYCStatus(ctx context.Context, userID string) (*kycResponses.KYCStatusResponse, error) {
	s.logger.Info("Getting KYC status",
		zap.String("user_id", userID))

	// Validate user ID
	if userID == "" {
		s.logger.Error("User ID is required for KYC status check")
		return nil, fmt.Errorf("user ID is required")
	}

	// Get most recent verification record for user
	verification, err := s.aadhaarRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get verification status",
			zap.String("user_id", userID),
			zap.Error(err))

		// Log audit event for status check failure
		s.auditService.LogUserActionWithError(ctx, userID, "aadhaar_status_check_failed", "aadhaar_verification", "", err, map[string]interface{}{
			"user_id": userID,
		})

		return nil, fmt.Errorf("verification record not found")
	}

	// Log audit event for status check
	s.auditService.LogUserAction(ctx, userID, "aadhaar_status_checked", "aadhaar_verification", verification.ID, map[string]interface{}{
		"kyc_status":          verification.KYCStatus,
		"verification_status": verification.VerificationStatus,
		"attempts":            verification.Attempts,
	})

	s.logger.Info("KYC status retrieved successfully",
		zap.String("user_id", userID),
		zap.String("kyc_status", verification.KYCStatus),
		zap.String("verification_status", verification.VerificationStatus),
		zap.Int("attempts", verification.Attempts))

	// Build response
	return &kycResponses.KYCStatusResponse{
		StatusCode:              200,
		UserID:                  userID,
		KYCStatus:               verification.KYCStatus,
		AadhaarVerified:         verification.VerificationStatus == "VERIFIED",
		AadhaarVerifiedAt:       verification.OTPVerifiedAt,
		VerificationAttempts:    verification.Attempts,
		LastVerificationAttempt: verification.LastAttemptAt,
	}, nil
}
