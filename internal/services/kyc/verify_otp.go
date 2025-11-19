package kyc

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	kycRequests "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/kyc"
	kycResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/kyc"
	"go.uber.org/zap"
)

// verifyOTP verifies OTP and updates user profile with Aadhaar data
func (s *Service) verifyOTP(ctx context.Context, req *kycRequests.VerifyOTPRequest, userID, authToken string) (*kycResponses.VerifyOTPResponse, error) {
	s.logger.Info("Verifying OTP for Aadhaar",
		zap.String("user_id", userID),
		zap.String("reference_id", req.ReferenceID))

	// 1. Validate request
	if err := req.Validate(); err != nil {
		s.logger.Error("OTP verification request validation failed",
			zap.String("user_id", userID),
			zap.String("reference_id", req.ReferenceID),
			zap.Error(err))
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// 2. Get verification record
	verification, err := s.aadhaarRepo.GetByReferenceID(ctx, req.ReferenceID)
	if err != nil {
		s.logger.Error("Verification record not found",
			zap.String("user_id", userID),
			zap.String("reference_id", req.ReferenceID),
			zap.Error(err))

		// Log audit event for failure
		s.auditService.LogUserActionWithError(ctx, userID, "aadhaar_otp_verification_failed", "aadhaar_verification", "", err, map[string]interface{}{
			"reference_id": req.ReferenceID,
			"reason":       "verification_record_not_found",
		})

		return nil, fmt.Errorf("verification record not found")
	}

	// 3. Check if verification belongs to this user
	if verification.UserID != userID {
		s.logger.Warn("User trying to verify another user's Aadhaar",
			zap.String("user_id", userID),
			zap.String("verification_user_id", verification.UserID),
			zap.String("reference_id", req.ReferenceID))

		// Log audit event for unauthorized access attempt
		s.auditService.LogUserActionWithError(ctx, userID, "aadhaar_otp_verification_unauthorized", "aadhaar_verification", verification.ID,
			fmt.Errorf("unauthorized access"), map[string]interface{}{
				"reference_id":         req.ReferenceID,
				"verification_user_id": verification.UserID,
			})

		return nil, fmt.Errorf("unauthorized: verification does not belong to this user")
	}

	// 4. Check if OTP has expired
	if verification.OTPRequestedAt != nil {
		elapsed := time.Since(*verification.OTPRequestedAt)
		if elapsed > time.Duration(s.config.OTPExpirationSeconds)*time.Second {
			s.logger.Warn("OTP has expired",
				zap.String("user_id", userID),
				zap.String("reference_id", req.ReferenceID),
				zap.Duration("elapsed", elapsed),
				zap.Int("expiration_seconds", s.config.OTPExpirationSeconds))

			// Log audit event for expired OTP
			s.auditService.LogUserActionWithError(ctx, userID, "aadhaar_otp_expired", "aadhaar_verification", verification.ID,
				fmt.Errorf("OTP expired"), map[string]interface{}{
					"reference_id":       req.ReferenceID,
					"elapsed_seconds":    int(elapsed.Seconds()),
					"expiration_seconds": s.config.OTPExpirationSeconds,
				})

			return nil, fmt.Errorf("OTP has expired")
		}
	}

	// 5. Check attempt limits
	if verification.Attempts >= s.config.OTPMaxAttempts {
		s.logger.Warn("Maximum OTP attempts exceeded",
			zap.String("user_id", userID),
			zap.String("reference_id", req.ReferenceID),
			zap.Int("attempts", verification.Attempts),
			zap.Int("max_attempts", s.config.OTPMaxAttempts))

		// Log audit event for max attempts exceeded
		s.auditService.LogUserActionWithError(ctx, userID, "aadhaar_otp_max_attempts_exceeded", "aadhaar_verification", verification.ID,
			fmt.Errorf("max attempts exceeded"), map[string]interface{}{
				"reference_id": req.ReferenceID,
				"attempts":     verification.Attempts,
				"max_attempts": s.config.OTPMaxAttempts,
			})

		return nil, fmt.Errorf("maximum OTP attempts exceeded")
	}

	// 6. Call Sandbox API to verify OTP
	sandboxResp, err := s.sandboxClient.VerifyOTP(ctx, req.ReferenceID, req.OTP, authToken)
	if err != nil {
		// Record failed attempt
		if incrementErr := s.aadhaarRepo.IncrementAttempts(ctx, verification.ID); incrementErr != nil {
			s.logger.Error("Failed to increment attempts counter",
				zap.String("verification_id", verification.ID),
				zap.Error(incrementErr))
		}

		s.logger.Error("Failed to verify OTP via Sandbox API",
			zap.String("user_id", userID),
			zap.String("reference_id", req.ReferenceID),
			zap.Error(err))

		// Log audit event for verification failure
		s.auditService.LogUserActionWithError(ctx, userID, "aadhaar_otp_verification_failed", "aadhaar_verification", verification.ID, err, map[string]interface{}{
			"reference_id": req.ReferenceID,
			"attempts":     verification.Attempts + 1,
		})

		return nil, fmt.Errorf("OTP verification failed: %w", err)
	}

	// 7. Upload photo to S3 (decode from base64)
	photoURL := ""
	if sandboxResp.Data.Photo != "" {
		photoData, err := base64.StdEncoding.DecodeString(sandboxResp.Data.Photo)
		if err != nil {
			s.logger.Warn("Failed to decode photo from base64, continuing without photo",
				zap.String("user_id", userID),
				zap.Error(err))
		} else {
			fileName := fmt.Sprintf("aadhaar_%d.jpg", time.Now().Unix())
			photoURL, err = s.aadhaarRepo.UploadPhoto(ctx, userID, photoData, fileName)
			if err != nil {
				s.logger.Warn("Failed to upload photo to S3, continuing without photo",
					zap.String("user_id", userID),
					zap.Error(err))
				photoURL = "" // Continue without photo (non-critical failure)
			} else {
				s.logger.Info("Photo uploaded successfully",
					zap.String("user_id", userID),
					zap.String("photo_url", photoURL))
			}
		}
	}

	// 8. Create address from Aadhaar data first (before updating profile)
	addressID := ""
	addressID, err = s.createAddress(ctx, userID, &sandboxResp.Data)
	if err != nil {
		s.logger.Warn("Failed to create address, continuing",
			zap.String("user_id", userID),
			zap.Error(err))
		addressID = "" // Not critical, continue without address
	} else {
		s.logger.Info("Address created successfully",
			zap.String("user_id", userID),
			zap.String("address_id", addressID))
	}

	// 9. Update user profile with Aadhaar data and link address
	if err := s.updateUserProfile(ctx, userID, &sandboxResp.Data, photoURL, addressID); err != nil {
		s.logger.Error("Failed to update user profile",
			zap.String("user_id", userID),
			zap.Error(err))

		// Log audit event for profile update failure
		s.auditService.LogUserActionWithError(ctx, userID, "aadhaar_profile_update_failed", "user_profile", userID, err, map[string]interface{}{
			"reference_id": req.ReferenceID,
		})

		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	// 10. Update verification status in database
	now := time.Now()
	verification.OTPVerifiedAt = &now
	verification.VerificationStatus = "VERIFIED"
	verification.KYCStatus = "VERIFIED"
	verification.PhotoURL = photoURL
	verification.Name = sandboxResp.Data.Name
	verification.Gender = sandboxResp.Data.Gender
	verification.FullAddress = sandboxResp.Data.FullAddress

	// Parse date of birth - Sandbox API returns DD-MM-YYYY format
	if sandboxResp.Data.DateOfBirth != "" {
		dob, err := time.Parse("02-01-2006", sandboxResp.Data.DateOfBirth)
		if err != nil {
			s.logger.Warn("Failed to parse date of birth",
				zap.String("date_of_birth", sandboxResp.Data.DateOfBirth),
				zap.String("expected_format", "DD-MM-YYYY"),
				zap.Error(err))
		} else {
			verification.DateOfBirth = &dob
			s.logger.Info("Date of birth parsed successfully",
				zap.String("user_id", userID),
				zap.Time("dob", dob))
		}
	}

	// Map address JSON
	verification.AddressJSON = models.AadhaarAddress{
		House:    sandboxResp.Data.Address.House,
		Street:   sandboxResp.Data.Address.Street,
		Landmark: sandboxResp.Data.Address.Landmark,
		District: sandboxResp.Data.Address.District,
		State:    sandboxResp.Data.Address.State,
		Pincode:  sandboxResp.Data.Address.Pincode,
		Country:  sandboxResp.Data.Address.Country,
	}

	verification.UpdatedBy = userID

	if err := s.aadhaarRepo.UpdateStatus(ctx, verification.ID, "VERIFIED"); err != nil {
		s.logger.Error("Failed to update verification status",
			zap.String("verification_id", verification.ID),
			zap.Error(err))
		// Continue anyway, as user profile is already updated
	}

	// 11. Log audit event for successful verification
	s.auditService.LogUserAction(ctx, userID, "aadhaar_verified", "aadhaar_verification", verification.ID, map[string]interface{}{
		"reference_id": verification.ReferenceID,
		"profile_id":   userID,
		"address_id":   addressID,
		"photo_url":    photoURL,
		"name":         verification.Name,
	})

	s.logger.Info("OTP verified successfully",
		zap.String("user_id", userID),
		zap.String("verification_id", verification.ID),
		zap.String("reference_id", req.ReferenceID),
		zap.String("name", verification.Name))

	// 12. Fetch profile, address, and contacts to include in response
	profile, address, contacts := s.fetchUserData(ctx, userID, addressID)

	// 13. Return response with complete user data
	return &kycResponses.VerifyOTPResponse{
		StatusCode: 200,
		Message:    "OTP verification successful",
		ProfileID:  userID,
		AddressID:  addressID,
		Profile:    profile,
		Address:    address,
		Contacts:   contacts,
		AadhaarData: &kycResponses.AadhaarData{
			Name:        sandboxResp.Data.Name,
			Gender:      sandboxResp.Data.Gender,
			DateOfBirth: sandboxResp.Data.DateOfBirth,
			YearOfBirth: sandboxResp.Data.YOB,
			CareOf:      sandboxResp.Data.CareOf,
			FullAddress: sandboxResp.Data.FullAddress,
			PhotoURL:    photoURL,
			ShareCode:   sandboxResp.Data.ShareCode,
			Status:      sandboxResp.Data.Status,
			Address:     mapSandboxAddress(&sandboxResp.Data.Address),
		},
	}, nil
}

// updateUserProfile updates user profile with verified Aadhaar data
func (s *Service) updateUserProfile(ctx context.Context, userID string, kycData *KYCData, photoURL string, addressID string) error {
	s.logger.Info("Updating user profile with Aadhaar data",
		zap.String("user_id", userID),
		zap.String("name", kycData.Name),
		zap.String("address_id", addressID))

	updates := map[string]interface{}{
		"is_validated":        true,
		"full_name":           kycData.Name,
		"aadhaar_verified":    true,
		"aadhaar_verified_at": time.Now(),
		"kyc_status":          "VERIFIED",
	}

	if photoURL != "" {
		updates["photo_url"] = photoURL
	}

	// Link the address to the user profile
	if addressID != "" {
		updates["address_id"] = addressID
	}

	if err := s.userService.Update(ctx, userID, updates); err != nil {
		s.logger.Error("Failed to update user profile",
			zap.String("user_id", userID),
			zap.Error(err))
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	s.logger.Info("User profile updated successfully",
		zap.String("user_id", userID))

	return nil
}

// createAddress finds or creates address from Aadhaar data
// If an address with the same full_address already exists, reuses it
func (s *Service) createAddress(ctx context.Context, userID string, kycData *KYCData) (string, error) {
	s.logger.Info("Finding or creating address from Aadhaar data",
		zap.String("user_id", userID))

	// Create address model using NewAddress()
	address := models.NewAddress()
	address.UserID = userID

	// Set address fields (all are pointers)
	if kycData.Address.House != "" {
		address.House = &kycData.Address.House
	}
	if kycData.Address.Street != "" {
		address.Street = &kycData.Address.Street
	}
	if kycData.Address.Landmark != "" {
		address.Landmark = &kycData.Address.Landmark
	}
	if kycData.Address.District != "" {
		address.District = &kycData.Address.District
	}
	if kycData.Address.State != "" {
		address.State = &kycData.Address.State
	}
	if kycData.Address.Country != "" {
		address.Country = &kycData.Address.Country
	}
	if kycData.Address.Pincode > 0 {
		pincode := strconv.Itoa(kycData.Address.Pincode)
		address.Pincode = &pincode
	}

	// Build full address for logging
	fullAddr := address.BuildFullAddress()

	// Set created by (CreatedBy is string, not pointer)
	address.CreatedBy = userID

	// Find existing address or create new one
	addressID, wasCreated, err := s.addressService.FindOrCreateAddress(ctx, address)
	if err != nil {
		s.logger.Error("Failed to find or create address",
			zap.String("user_id", userID),
			zap.String("full_address", fullAddr),
			zap.Error(err))
		return "", fmt.Errorf("failed to find or create address: %w", err)
	}

	if wasCreated {
		s.logger.Info("New address created",
			zap.String("user_id", userID),
			zap.String("address_id", addressID),
			zap.String("full_address", fullAddr))
	} else {
		s.logger.Info("Reusing existing address",
			zap.String("user_id", userID),
			zap.String("address_id", addressID),
			zap.String("full_address", fullAddr))
	}

	return addressID, nil
}

// fetchUserData fetches profile, address, and contacts for the user
func (s *Service) fetchUserData(ctx context.Context, userID, addressID string) (*models.UserProfile, *models.Address, []*models.Contact) {
	var profile *models.UserProfile
	var address *models.Address
	var contacts []*models.Contact

	// Fetch user profile
	if p, err := s.userService.GetProfile(ctx, userID); err == nil {
		profile = p
		s.logger.Info("Fetched user profile for response",
			zap.String("user_id", userID),
			zap.String("profile_id", p.ID))
	} else {
		s.logger.Warn("Failed to fetch user profile for response",
			zap.String("user_id", userID),
			zap.Error(err))
	}

	// Fetch address
	if addressID != "" {
		if a, err := s.addressService.GetAddressByID(ctx, addressID); err == nil {
			address = a
			s.logger.Info("Fetched address for response",
				zap.String("user_id", userID),
				zap.String("address_id", addressID))
		} else {
			s.logger.Warn("Failed to fetch address for response",
				zap.String("user_id", userID),
				zap.String("address_id", addressID),
				zap.Error(err))
		}
	}

	// Note: Contacts are not available from Aadhaar, return empty slice
	// Contacts must be added separately through the contacts API
	contacts = []*models.Contact{}

	return profile, address, contacts
}

// mapSandboxAddress maps Sandbox address to response address
func mapSandboxAddress(addr *SandboxAddress) *kycResponses.AadhaarAddr {
	if addr == nil {
		return nil
	}

	return &kycResponses.AadhaarAddr{
		House:    addr.House,
		Street:   addr.Street,
		Landmark: addr.Landmark,
		District: addr.District,
		State:    addr.State,
		Pincode:  addr.Pincode,
		Country:  addr.Country,
	}
}
