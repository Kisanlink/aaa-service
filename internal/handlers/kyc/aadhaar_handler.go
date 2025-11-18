package kyc

import (
	"net/http"
	"strings"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/requests/kyc"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	kycService "github.com/Kisanlink/aaa-service/v2/internal/services/kyc"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles KYC/Aadhaar verification HTTP requests
type Handler struct {
	kycService *kycService.Service
	validator  interfaces.Validator
	responder  interfaces.Responder
	logger     *zap.Logger
}

// NewHandler creates a new KYC handler
func NewHandler(
	kycService *kycService.Service,
	validator interfaces.Validator,
	responder interfaces.Responder,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		kycService: kycService,
		validator:  validator,
		responder:  responder,
		logger:     logger,
	}
}

// GenerateOTP handles POST /api/v1/kyc/aadhaar/otp
//
//	@Summary		Generate OTP for Aadhaar verification
//	@Description	Generate OTP for Aadhaar verification. Requires user consent and valid Aadhaar number.
//	@Tags			kyc
//	@Accept			json
//	@Produce		json
//	@Param			request	body		kyc.GenerateOTPRequest	true	"Generate OTP request with Aadhaar number and consent"
//	@Success		200		{object}	kyc.GenerateOTPResponse
//	@Failure		400		{object}	map[string]interface{}	"Invalid request or validation error"
//	@Failure		401		{object}	map[string]interface{}	"Unauthorized - authentication required"
//	@Failure		429		{object}	map[string]interface{}	"Rate limit exceeded"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/kyc/aadhaar/otp [post]
//	@Security		Bearer
func (h *Handler) GenerateOTP(c *gin.Context) {
	h.logger.Info("Processing generate OTP request")

	// Parse request body
	var req kyc.GenerateOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body for OTP generation",
			zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Warn("OTP generation request validation failed",
			zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Additional validation using validator service
	if err := h.validator.ValidateStruct(&req); err != nil {
		h.logger.Warn("Struct validation failed for OTP generation",
			zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "Authentication required",
			errors.NewSecureUnauthorizedError())
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		h.responder.SendError(c, http.StatusUnauthorized, "Authentication required",
			errors.NewSecureUnauthorizedError())
		return
	}

	// Get auth token from header
	authToken := c.GetHeader("Authorization")

	// Call service
	resp, err := h.kycService.GenerateOTP(c.Request.Context(), &req, userIDStr, authToken)
	if err != nil {
		h.logger.Error("Failed to generate OTP",
			zap.String("user_id", userIDStr),
			zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.logger.Info("OTP generated successfully",
		zap.String("user_id", userIDStr),
		zap.String("reference_id", resp.ReferenceID))
	h.responder.SendSuccess(c, http.StatusOK, resp)
}

// VerifyOTP handles POST /api/v1/kyc/aadhaar/otp/verify
//
//	@Summary		Verify Aadhaar OTP
//	@Description	Verify OTP and complete Aadhaar verification. Updates user profile with verified Aadhaar data.
//	@Tags			kyc
//	@Accept			json
//	@Produce		json
//	@Param			request	body		kyc.VerifyOTPRequest	true	"Verify OTP request with reference ID and OTP"
//	@Success		200		{object}	kyc.VerifyOTPResponse
//	@Failure		400		{object}	map[string]interface{}	"Invalid OTP or validation error"
//	@Failure		401		{object}	map[string]interface{}	"Unauthorized - authentication required"
//	@Failure		404		{object}	map[string]interface{}	"Reference ID not found"
//	@Failure		429		{object}	map[string]interface{}	"Max OTP attempts exceeded"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/kyc/aadhaar/otp/verify [post]
//	@Security		Bearer
func (h *Handler) VerifyOTP(c *gin.Context) {
	h.logger.Info("Processing verify OTP request")

	// Parse request body
	var req kyc.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body for OTP verification",
			zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Warn("OTP verification request validation failed",
			zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Additional validation using validator service
	if err := h.validator.ValidateStruct(&req); err != nil {
		h.logger.Warn("Struct validation failed for OTP verification",
			zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "Authentication required",
			errors.NewSecureUnauthorizedError())
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		h.responder.SendError(c, http.StatusUnauthorized, "Authentication required",
			errors.NewSecureUnauthorizedError())
		return
	}

	// Get auth token from header
	authToken := c.GetHeader("Authorization")

	// Call service
	resp, err := h.kycService.VerifyOTP(c.Request.Context(), &req, userIDStr, authToken)
	if err != nil {
		h.logger.Error("Failed to verify OTP",
			zap.String("user_id", userIDStr),
			zap.String("reference_id", req.ReferenceID),
			zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.logger.Info("OTP verified successfully",
		zap.String("user_id", userIDStr),
		zap.String("profile_id", resp.ProfileID),
		zap.String("address_id", resp.AddressID))
	h.responder.SendSuccess(c, http.StatusOK, resp)
}

// GetKYCStatus handles GET /api/v1/kyc/status/:user_id
//
//	@Summary		Get KYC verification status
//	@Description	Retrieve KYC verification status for a user. Users can only view their own status unless they have admin privileges.
//	@Tags			kyc
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		string	true	"User ID"
//	@Success		200		{object}	kyc.KYCStatusResponse
//	@Failure		400		{object}	map[string]interface{}	"Invalid request"
//	@Failure		401		{object}	map[string]interface{}	"Unauthorized - authentication required"
//	@Failure		403		{object}	map[string]interface{}	"Forbidden - can only view own status"
//	@Failure		404		{object}	map[string]interface{}	"User not found"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/kyc/status/{user_id} [get]
//	@Security		Bearer
func (h *Handler) GetKYCStatus(c *gin.Context) {
	// Get user ID from path parameter
	targetUserID := c.Param("user_id")
	if targetUserID == "" {
		h.logger.Warn("Missing user ID in path parameter")
		h.responder.SendValidationError(c, []string{"user_id is required"})
		return
	}

	h.logger.Info("Processing get KYC status request",
		zap.String("target_user_id", targetUserID))

	// Get authenticated user ID from context
	authUserID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		h.responder.SendError(c, http.StatusUnauthorized, "Authentication required",
			errors.NewSecureUnauthorizedError())
		return
	}

	authUserIDStr, ok := authUserID.(string)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		h.responder.SendError(c, http.StatusUnauthorized, "Authentication required",
			errors.NewSecureUnauthorizedError())
		return
	}

	// Check authorization (user can only view own status, unless admin)
	// For now, allow users to only view their own status
	if authUserIDStr != targetUserID {
		// TODO: Check if user has admin role to allow viewing other users' status
		h.logger.Warn("User attempting to access another user's KYC status",
			zap.String("auth_user_id", authUserIDStr),
			zap.String("target_user_id", targetUserID))
		h.responder.SendError(c, http.StatusForbidden,
			"You can only view your own KYC status",
			errors.NewSecureForbiddenError())
		return
	}

	// Call service
	resp, err := h.kycService.GetKYCStatus(c.Request.Context(), targetUserID)
	if err != nil {
		h.logger.Error("Failed to get KYC status",
			zap.String("user_id", targetUserID),
			zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.logger.Info("KYC status retrieved successfully",
		zap.String("user_id", targetUserID),
		zap.String("kyc_status", resp.KYCStatus))
	h.responder.SendSuccess(c, http.StatusOK, resp)
}

// handleServiceError maps service errors to HTTP responses
func (h *Handler) handleServiceError(c *gin.Context, err error) {
	// Use custom error types from pkg/errors
	switch e := err.(type) {
	case *errors.ValidationError:
		h.responder.SendValidationError(c, []string{e.Error()})
		return
	case *errors.BadRequestError:
		// Check for specific rate limit or OTP errors
		errMsg := strings.ToLower(e.Error())
		if strings.Contains(errMsg, "rate limit") {
			h.responder.SendError(c, http.StatusTooManyRequests, e.Error(), e)
			return
		}
		h.responder.SendError(c, http.StatusBadRequest, e.Error(), e)
		return
	case *errors.UnauthorizedError:
		h.responder.SendError(c, http.StatusUnauthorized, e.Error(), e)
		return
	case *errors.ForbiddenError:
		h.responder.SendError(c, http.StatusForbidden, e.Error(), e)
		return
	case *errors.NotFoundError:
		h.responder.SendError(c, http.StatusNotFound, e.Error(), e)
		return
	case *errors.ConflictError:
		h.responder.SendError(c, http.StatusConflict, e.Error(), e)
		return
	case *errors.InternalError:
		h.responder.SendInternalError(c, e)
		return
	default:
		// For unknown errors, log and return internal error
		h.logger.Error("Unhandled error type in KYC handler",
			zap.Error(err),
			zap.String("error_type", "unknown"))
		h.responder.SendInternalError(c, err)
		return
	}
}
