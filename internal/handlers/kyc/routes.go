package kyc

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers KYC routes with the router
// The authMiddleware parameter should be the JWT authentication middleware
// that validates tokens and sets "user_id" in the Gin context
func RegisterRoutes(router *gin.Engine, handler *Handler, authMiddleware gin.HandlerFunc) {
	// KYC routes group
	kyc := router.Group("/api/v1/kyc")

	// Apply authentication middleware to all KYC routes
	// All KYC operations require authenticated users
	kyc.Use(authMiddleware)

	// Aadhaar verification routes
	{
		// Generate OTP for Aadhaar verification
		// POST /api/v1/kyc/aadhaar/otp
		kyc.POST("/aadhaar/otp", handler.GenerateOTP)

		// Verify OTP and complete Aadhaar verification
		// POST /api/v1/kyc/aadhaar/otp/verify
		kyc.POST("/aadhaar/otp/verify", handler.VerifyOTP)
	}

	// KYC status routes
	{
		// Get KYC verification status for a user
		// GET /api/v1/kyc/status/:user_id
		kyc.GET("/status/:user_id", handler.GetKYCStatus)
	}
}
