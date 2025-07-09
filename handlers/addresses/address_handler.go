package addresses

import (
	"fmt"
	"net/http"
	"strconv"

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

// CreateAddressRequest represents a request to create an address
type CreateAddressRequest struct {
	House       *string `json:"house,omitempty"`
	Street      *string `json:"street,omitempty"`
	Landmark    *string `json:"landmark,omitempty"`
	PostOffice  *string `json:"post_office,omitempty"`
	Subdistrict *string `json:"subdistrict,omitempty"`
	District    *string `json:"district,omitempty"`
	VTC         *string `json:"vtc,omitempty"`
	State       *string `json:"state,omitempty"`
	Country     *string `json:"country,omitempty"`
	Pincode     *string `json:"pincode,omitempty"`
	FullAddress *string `json:"full_address,omitempty"`
	UserID      string  `json:"user_id" validate:"required"`
}

// Validate validates the CreateAddressRequest
func (r *CreateAddressRequest) Validate() error {
	if r.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	// At least one address field should be provided
	if r.House == nil && r.Street == nil && r.FullAddress == nil {
		return fmt.Errorf("at least one address field is required")
	}
	return nil
}

// CreateAddress handles POST /addresses
func (h *AddressHandler) CreateAddress(c *gin.Context) {
	h.logger.Info("Creating address")

	var req CreateAddressRequest
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

	// TODO: Create address through service when fully implemented
	// For now, return mock response
	result := map[string]interface{}{
		"id":      "addr_" + req.UserID,
		"user_id": req.UserID,
		"message": "Address created successfully",
	}

	h.logger.Info("Address created successfully", zap.String("userID", req.UserID))
	h.responder.SendSuccess(c, http.StatusCreated, result)
}

// GetAddress handles GET /addresses/:id
func (h *AddressHandler) GetAddress(c *gin.Context) {
	addressID := c.Param("id")
	h.logger.Info("Getting address by ID", zap.String("addressID", addressID))

	if addressID == "" {
		h.responder.SendValidationError(c, []string{"address ID is required"})
		return
	}

	// TODO: Get address through service when fully implemented
	result := map[string]interface{}{
		"id":       addressID,
		"house":    "123",
		"street":   "Main Street",
		"district": "Example District",
		"state":    "Example State",
		"pincode":  "123456",
		"message":  "Address retrieved successfully",
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

	var req CreateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// TODO: Update address through service when fully implemented
	result := map[string]interface{}{
		"id":      addressID,
		"message": "Address updated successfully",
	}

	h.logger.Info("Address updated successfully", zap.String("addressID", addressID))
	h.responder.SendSuccess(c, http.StatusOK, result)
}

// DeleteAddress handles DELETE /addresses/:id
func (h *AddressHandler) DeleteAddress(c *gin.Context) {
	addressID := c.Param("id")
	h.logger.Info("Deleting address", zap.String("addressID", addressID))

	if addressID == "" {
		h.responder.SendValidationError(c, []string{"address ID is required"})
		return
	}

	// TODO: Delete address through service when fully implemented
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

	// TODO: Search addresses through service when fully implemented
	result := map[string]interface{}{
		"addresses": []interface{}{},
		"total":     0,
		"limit":     limit,
		"offset":    offset,
		"query":     query,
		"message":   "Address search completed",
	}

	h.logger.Info("Address search completed", zap.String("query", query))
	h.responder.SendSuccess(c, http.StatusOK, result)
}
