package user

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

type OTPData struct {
	OTP      string
	Verified bool
}

var (
	otpStorage = make(map[string]OTPData)
	otpMutex   = sync.Mutex{}
)

// generateOTP creates a random 6-digit OTP.
func generateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// PasswordResetHandler handles password reset flow
// @Summary Password reset flow
// @Description Handles the complete password reset flow in three steps: 1) Request OTP, 2) Verify OTP, 3) Reset password. Each step requires different request parameters.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body model.PasswordResetFlowRequest true "Password reset request"
// @Success 200 {object} object "Success responses vary by step: 1) 'OTP sent successfully', 2) 'OTP verified. Proceed to reset password.', 3) 'Password reset successfully'"
// @Failure 400 {object} helper.ErrorResponse "Invalid request body or parameters"
// @Failure 401 {object} helper.ErrorResponse "Invalid or expired OTP"
// @Failure 404 {object} helper.ErrorResponse "User not found"
// @Failure 500 {object} helper.ErrorResponse "Internal server error"
// @Router /forgot-password [post]
func (s *UserHandler) PasswordResetHandler(c *gin.Context) {
	var req model.PasswordResetFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request body"})
		return
	}

	// Find the user
	user, err := s.userService.FindUserByUsername(req.Username)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusNotFound, []string{err.Error()})
		return
	}
	key := fmt.Sprintf("otp:%s", user.ID)

	// Case 1: Forgot Password - send OTP if no OTP and no new password provided.
	if req.OTP == "" && req.NewPassword == "" {
		otp := generateOTP()

		otpMutex.Lock()
		otpStorage[key] = OTPData{OTP: otp, Verified: false}
		otpMutex.Unlock()

		// Simulate expiry after 10 minutes.
		go func() {
			time.Sleep(10 * time.Minute)
			otpMutex.Lock()
			delete(otpStorage, key)
			otpMutex.Unlock()
		}()

		mobileNumberStr := fmt.Sprintf("%v", user.MobileNumber)
		if user.CountryCode == nil {
			helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Country code is missing"})
			return
		}
		client.SendOTP(*user.CountryCode, mobileNumberStr, otp)

		helper.SendSuccessResponse(c.Writer, http.StatusOK, "OTP sent successfully", nil)
		return
	}

	if req.OTP != "" && req.NewPassword == "" {
		otpMutex.Lock()
		otpData, exists := otpStorage[key]
		otpMutex.Unlock()

		if !exists || otpData.OTP != req.OTP {
			helper.SendErrorResponse(c.Writer, http.StatusUnauthorized, []string{"Invalid or expired OTP"})
			return
		}

		otpMutex.Lock()
		otpData.Verified = true
		otpStorage[key] = otpData
		otpMutex.Unlock()

		helper.SendSuccessResponse(c.Writer, http.StatusOK, "OTP verified. Proceed to reset password.", nil)
		return
	}

	// Case 3: Reset Password - require both OTP and new password.
	if req.NewPassword != "" {
		if req.OTP == "" {
			helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"OTP is required for resetting password"})
			return
		}

		otpMutex.Lock()
		otpData, exists := otpStorage[key]
		otpMutex.Unlock()

		// Check that OTP exists, matches, and is verified.
		if !exists || otpData.OTP != req.OTP || !otpData.Verified {
			helper.SendErrorResponse(c.Writer, http.StatusUnauthorized, []string{"Invalid or unverified OTP"})
			return
		}

		// Proceed to reset the password.
		hashedPassword, err := helper.HashPassword(req.NewPassword)
		if err != nil {
			helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to hash password"})
			return
		}

		if err := s.userService.UpdatePassword(user.ID, hashedPassword); err != nil {
			helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
			return
		}

		// Delete OTP data after successful password reset.
		otpMutex.Lock()
		delete(otpStorage, key)
		otpMutex.Unlock()

		helper.SendSuccessResponse(c.Writer, http.StatusOK, "Password reset successfully", nil)
		return
	}

	// Fallback response if none of the cases match.
	helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request parameters"})
}
