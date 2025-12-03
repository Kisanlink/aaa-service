package contacts

import (
	"net/http"
	"strconv"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/requests/contacts"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	contactService "github.com/Kisanlink/aaa-service/v2/internal/services/contacts"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ContactHandler handles HTTP requests for contact operations
type ContactHandler struct {
	contactService *contactService.ContactService
	validator      interfaces.Validator
	responder      interfaces.Responder
	logger         *zap.Logger
}

// NewContactHandler creates a new ContactHandler instance
func NewContactHandler(
	contactService *contactService.ContactService,
	validator interfaces.Validator,
	responder interfaces.Responder,
	logger *zap.Logger,
) *ContactHandler {
	return &ContactHandler{
		contactService: contactService,
		validator:      validator,
		responder:      responder,
		logger:         logger,
	}
}

// CreateContact handles POST /contacts
//
//	@Summary		Create a new contact
//	@Description	Create a new contact with the provided information
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			contact	body		contacts.CreateContactRequest	true	"Contact creation data"
//	@Success		201		{object}	contacts.ContactResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		409		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/contacts [post]
func (h *ContactHandler) CreateContact(c *gin.Context) {
	h.logger.Info("Creating contact")

	var req contacts.CreateContactRequest
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

	// Additional validation using validator service
	if err := h.validator.ValidateStruct(&req); err != nil {
		h.logger.Error("Struct validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Create contact through service
	contactResponse, err := h.contactService.CreateContact(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create contact", zap.Error(err))
		if validationErr, ok := err.(*errors.ValidationError); ok {
			h.responder.SendValidationError(c, []string{validationErr.Error()})
			return
		}
		if conflictErr, ok := err.(*errors.ConflictError); ok {
			h.responder.SendError(c, http.StatusConflict, conflictErr.Error(), conflictErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Contact created successfully", zap.String("contactID", contactResponse.ID))
	h.responder.SendSuccess(c, http.StatusCreated, contactResponse)
}

// GetContact handles GET /contacts/:id
//
//	@Summary		Get a contact by ID
//	@Description	Retrieve a contact by its ID
//	@Tags			contacts
//	@Produce		json
//	@Param			id	path		string	true	"Contact ID"
//	@Success		200	{object}	contacts.ContactResponse
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		404	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Router			/api/v1/contacts/{id} [get]
func (h *ContactHandler) GetContact(c *gin.Context) {
	contactID := c.Param("id")
	if contactID == "" {
		h.responder.SendValidationError(c, []string{"contact ID is required"})
		return
	}

	h.logger.Info("Getting contact", zap.String("contactID", contactID))

	// Get contact through service
	contactResponse, err := h.contactService.GetContact(c.Request.Context(), contactID)
	if err != nil {
		h.logger.Error("Failed to get contact", zap.Error(err))
		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusNotFound, notFoundErr.Error(), notFoundErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, contactResponse)
}

// UpdateContact handles PUT /contacts/:id
//
//	@Summary		Update a contact
//	@Description	Update an existing contact with the provided information
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Contact ID"
//	@Param			contact	body		contacts.UpdateContactRequest	true	"Contact update data"
//	@Success		200		{object}	contacts.ContactResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/contacts/{id} [put]
func (h *ContactHandler) UpdateContact(c *gin.Context) {
	contactID := c.Param("id")
	if contactID == "" {
		h.responder.SendValidationError(c, []string{"contact ID is required"})
		return
	}

	h.logger.Info("Updating contact", zap.String("contactID", contactID))

	var req contacts.UpdateContactRequest
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

	// Additional validation using validator service
	if err := h.validator.ValidateStruct(&req); err != nil {
		h.logger.Error("Struct validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Update contact through service
	contactResponse, err := h.contactService.UpdateContact(c.Request.Context(), contactID, &req)
	if err != nil {
		h.logger.Error("Failed to update contact", zap.Error(err))
		if validationErr, ok := err.(*errors.ValidationError); ok {
			h.responder.SendValidationError(c, []string{validationErr.Error()})
			return
		}
		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusNotFound, notFoundErr.Error(), notFoundErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Contact updated successfully", zap.String("contactID", contactID))
	h.responder.SendSuccess(c, http.StatusOK, contactResponse)
}

// DeleteContact handles DELETE /contacts/:id
//
//	@Summary		Delete a contact
//	@Description	Soft delete a contact by ID
//	@Tags			contacts
//	@Produce		json
//	@Param			id	path	string	true	"Contact ID"
//	@Success		204	"Contact deleted successfully"
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		404	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Router			/api/v1/contacts/{id} [delete]
func (h *ContactHandler) DeleteContact(c *gin.Context) {
	contactID := c.Param("id")
	if contactID == "" {
		h.responder.SendValidationError(c, []string{"contact ID is required"})
		return
	}

	// Get user ID from context (assuming it's set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.responder.SendError(c, http.StatusUnauthorized, "user ID not found in context", nil)
		return
	}

	h.logger.Info("Deleting contact", zap.String("contactID", contactID), zap.String("deletedBy", userID.(string)))

	// Delete contact through service
	err := h.contactService.DeleteContact(c.Request.Context(), contactID, userID.(string))
	if err != nil {
		h.logger.Error("Failed to delete contact", zap.Error(err))
		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusNotFound, notFoundErr.Error(), notFoundErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Contact deleted successfully", zap.String("contactID", contactID))
	c.Status(http.StatusNoContent)
}

// ListContacts handles GET /contacts
//
//	@Summary		List contacts
//	@Description	Retrieve a paginated list of contacts
//	@Tags			contacts
//	@Produce		json
//	@Param			limit	query		int	false	"Number of contacts to return (default: 10, max: 100)"
//	@Param			offset	query		int	false	"Number of contacts to skip (default: 0)"
//	@Success		200		{object}	contacts.ContactListResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/contacts [get]
func (h *ContactHandler) ListContacts(c *gin.Context) {
	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	h.logger.Info("Listing contacts", zap.Int("limit", limit), zap.Int("offset", offset))

	// List contacts through service
	contactsResponse, err := h.contactService.ListContacts(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list contacts", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	h.responder.SendPaginatedResponse(c, contactsResponse.Contacts, int(contactsResponse.Total), limit, offset)
}

// GetContactsByUser handles GET /contacts/user/:userID
//
//	@Summary		Get contacts by user
//	@Description	Retrieve contacts for a specific user
//	@Tags			contacts
//	@Produce		json
//	@Param			userID	path		string	true	"User ID"
//	@Param			limit	query		int		false	"Number of contacts to return (default: 10, max: 100)"
//	@Param			offset	query		int		false	"Number of contacts to skip (default: 0)"
//	@Success		200		{object}	contacts.ContactListResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v1/contacts/user/{userID} [get]
func (h *ContactHandler) GetContactsByUser(c *gin.Context) {
	userID := c.Param("userID")
	if userID == "" {
		h.responder.SendValidationError(c, []string{"user ID is required"})
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	h.logger.Info("Getting contacts by user", zap.String("userID", userID), zap.Int("limit", limit), zap.Int("offset", offset))

	// Get contacts by user through service
	contactsResponse, err := h.contactService.GetContactsByUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get contacts by user", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	h.responder.SendPaginatedResponse(c, contactsResponse.Contacts, int(contactsResponse.Total), limit, offset)
}
