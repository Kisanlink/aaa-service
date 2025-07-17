package addresses

import (
	"net/http"
	"strconv"

	"github.com/Kisanlink/aaa-service/entities/requests/addresses"
	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AddressHandler handles address-related HTTP requests
type AddressHandler struct {
	addressService interfaces.AddressService
	validator      interfaces.Validator
	responder      interfaces.Responder
	logger         *zap.Logger
}

// NewAddressHandler creates a new AddressHandler instance
func NewAddressHandler(
	addressService interfaces.AddressService,
	validator interfaces.Validator,
	responder interfaces.Responder,
	logger *zap.Logger,
) *AddressHandler {
	return &AddressHandler{
		addressService: addressService,
		validator:      validator,
		responder:      responder,
		logger:         logger,
	}
}

// CreateAddress handles POST /addresses
func (h *AddressHandler) CreateAddress(c *gin.Context) {
	h.logger.Info("Creating address")

	var req addresses.CreateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Request validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Convert request to model
	addressModel := req.ToModel()

	// Create address through service
	err := h.addressService.CreateAddress(c.Request.Context(), addressModel)
	if err != nil {
		h.logger.Error("Failed to create address", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Address created successfully", zap.String("userID", req.UserID))
	h.responder.SendSuccess(c, http.StatusCreated, map[string]string{"message": "Address created successfully"})
}

// GetAddress handles GET /addresses/:id
func (h *AddressHandler) GetAddress(c *gin.Context) {
	addressID := c.Param("id")
	h.logger.Info("Getting address by ID", zap.String("addressID", addressID))

	if addressID == "" {
		h.responder.SendValidationError(c, []string{"address ID is required"})
		return
	}

	// Get address through service
	result, err := h.addressService.GetAddressByID(c.Request.Context(), addressID)
	if err != nil {
		h.logger.Error("Failed to get address", zap.String("addressID", addressID), zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Address retrieved successfully", zap.String("addressID", addressID))
	h.responder.SendSuccess(c, http.StatusOK, result)
}

// UpdateAddress handles PUT /addresses/:id
func (h *AddressHandler) UpdateAddress(c *gin.Context) {
	addressID := c.Param("id")
	h.logger.Info("Updating address", zap.String("addressID", addressID))

	if addressID == "" {
		h.responder.SendValidationError(c, []string{"address ID is required"})
		return
	}

	var req addresses.UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Set the ID from the URL parameter
	req.ID = addressID

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Request validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Convert request to model
	addressModel := req.ToModel()

	// Update address through service
	err := h.addressService.UpdateAddress(c.Request.Context(), addressModel)
	if err != nil {
		h.logger.Error("Failed to update address", zap.String("addressID", addressID), zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Address updated successfully", zap.String("addressID", addressID))
	h.responder.SendSuccess(c, http.StatusOK, map[string]string{"message": "Address updated successfully"})
}

// DeleteAddress handles DELETE /addresses/:id
func (h *AddressHandler) DeleteAddress(c *gin.Context) {
	addressID := c.Param("id")
	h.logger.Info("Deleting address", zap.String("addressID", addressID))

	if addressID == "" {
		h.responder.SendValidationError(c, []string{"address ID is required"})
		return
	}

	// Delete address through service
	if err := h.addressService.DeleteAddress(c.Request.Context(), addressID); err != nil {
		h.logger.Error("Failed to delete address", zap.String("addressID", addressID), zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	result := map[string]interface{}{
		"message": "Address deleted successfully",
	}

	h.logger.Info("Address deleted successfully", zap.String("addressID", addressID))
	h.responder.SendSuccess(c, http.StatusOK, result)
}

// SearchAddresses handles GET /addresses/search
func (h *AddressHandler) SearchAddresses(c *gin.Context) {
	query := c.Query("q")
	h.logger.Info("Searching addresses", zap.String("query", query))

	if query == "" {
		h.responder.SendValidationError(c, []string{"search query is required"})
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		h.responder.SendValidationError(c, []string{"invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		h.responder.SendValidationError(c, []string{"invalid offset parameter"})
		return
	}

	// Search addresses through service
	result, err := h.addressService.SearchAddresses(c.Request.Context(), query, limit, offset)
	if err != nil {
		h.logger.Error("Failed to search addresses", zap.String("query", query), zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Address search completed", zap.String("query", query))
	h.responder.SendSuccess(c, http.StatusOK, result)
}
