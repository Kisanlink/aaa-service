package user

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/Kisanlink/aaa-service/controller/handler"
	"github.com/gin-gonic/gin"
)

// OTPData holds the OTP and its verification status.
type OTPData struct {
	OTP      string
	Verified bool
}

// stores OTPs in-memory
var (
	otpStorage = make(map[string]OTPData)
	otpMutex   = sync.Mutex{}
)

type PasswordResetFlowRequest struct {
	Username    string `json:"username" binding:"required"`
	OTP         string `json:"otp,omitempty"`
	NewPassword string `json:"new_password,omitempty"`
}

// generateOTP creates a random 6-digit OTP.
func generateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// PasswordResetHandler handles the forgot password, OTP verification, and password reset flows.
func (s *Server) PasswordResetHandler(c *gin.Context) {
	var req PasswordResetFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Find the user
	user, err := s.UserRepo.FindUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
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
		handler.SendOTP(*user.CountryCode, mobileNumberStr, otp)
		c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})
		return
	}

	// Case 2: OTP Verification - verify OTP when provided and new password is not given.
	if req.OTP != "" && req.NewPassword == "" {
		otpMutex.Lock()
		otpData, exists := otpStorage[key]
		otpMutex.Unlock()

		if !exists || otpData.OTP != req.OTP {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
			return
		}

		// Mark OTP as verified without deleting it.
		otpMutex.Lock()
		otpData.Verified = true
		otpStorage[key] = otpData
		otpMutex.Unlock()

		c.JSON(http.StatusOK, gin.H{"message": "OTP verified. Proceed to reset password."})
		return
	}

	// Case 3: Reset Password - require both OTP and new password.
	if req.NewPassword != "" {
		if req.OTP == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "OTP is required for resetting password"})
			return
		}

		otpMutex.Lock()
		otpData, exists := otpStorage[key]
		otpMutex.Unlock()

		// Check that OTP exists, matches, and is verified.
		if !exists || otpData.OTP != req.OTP || !otpData.Verified {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or unverified OTP"})
			return
		}

		// Proceed to reset the password.
		hashedPassword, err := HashPassword(req.NewPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		err = s.UserRepo.UpdatePassword(c.Request.Context(), user.ID, hashedPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}

		// Delete OTP data after successful password reset.
		otpMutex.Lock()
		delete(otpStorage, key)
		otpMutex.Unlock()

		c.JSON(http.StatusOK, gin.H{
			"status_code": http.StatusOK,
			"success":     true,
			"message":     "Password reset successfully",
		})
		return
	}

	// Fallback response if none of the cases match.
	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
}
