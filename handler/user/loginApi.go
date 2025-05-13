package user

import (
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// LoginRestApi authenticates a user and returns JWT tokens
// @Summary User login
// @Description Authenticates a user with username and password, returns JWT tokens in response headers and optional user details in body
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Login credentials"
// @Param user_details query boolean false "Include full user details" default(false)
// @Success 200 {object} helper.Response{data=model.UserResponse} "Login successful"
// @Header 200 {string} Authorization "Bearer access token"
// @Header 200 {string} Refresh-Token "Refresh token"
// @Failure 400 {object} helper.Response "Invalid request body or missing credentials"
// @Failure 401 {object} helper.Response "Invalid credentials"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /login [post]
func (s *UserHandler) LoginRestApi(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request body"})
		return
	}

	// Validate inputs
	if req.Username == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Username is required"})
		return
	}

	if !helper.IsValidUsername(req.Username) {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest,
			[]string{"Username '" + req.Username + "' contains invalid characters. Only a-z, A-Z, 0-9, /, _, |, -, =, + are allowed, and spaces are prohibited."})
		return
	}

	if req.Password == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Password is required"})
		return
	}

	// Find user
	existingUser, err := s.userService.FindUserByUsername(req.Username)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusUnauthorized, []string{"Invalid credentials"})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password))
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusUnauthorized, []string{"Invalid credentials"})
		return
	}

	// Get role permissions
	rolePermissions, err := s.userService.FindUsageRights(existingUser.ID)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to fetch user permissions"})
		return
	}

	// Check if detailed user info is requested
	includeDetails := c.Query("user_details") == "true"

	// Prepare base user response
	userResponse := model.UserResponse{
		ID:             existingUser.ID,
		Username:       existingUser.Username,
		IsValidated:    existingUser.IsValidated,
		CreatedAt:      existingUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      existingUser.UpdatedAt.Format(time.RFC3339),
		RolePermission: helper.ConvertAndDeduplicateRolePermissions(rolePermissions),
	}

	// Only fetch and include additional details if requested
	if includeDetails {
		// Get address if exists
		var address *model.AddressRes
		if existingUser.AddressID != nil {
			addr, err := s.userService.GetAddressByID(*existingUser.AddressID)
			if err != nil {
				helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to fetch address"})
				return
			}

			if addr != nil {
				address = &model.AddressRes{
					ID:          addr.ID,
					House:       helper.SafeString(addr.House),
					Street:      helper.SafeString(addr.Street),
					Landmark:    helper.SafeString(addr.Landmark),
					PostOffice:  helper.SafeString(addr.PostOffice),
					Subdistrict: helper.SafeString(addr.Subdistrict),
					District:    helper.SafeString(addr.District),
					VTC:         helper.SafeString(addr.VTC),
					State:       helper.SafeString(addr.State),
					Country:     helper.SafeString(addr.Country),
					Pincode:     helper.SafeString(addr.Pincode),
					FullAddress: helper.SafeString(addr.FullAddress),
					CreatedAt:   addr.CreatedAt.Format(time.RFC3339),
					UpdatedAt:   addr.UpdatedAt.Format(time.RFC3339),
				}
			}
		}

		// Add detailed fields to the response
		userResponse.AadhaarNumber = helper.SafeString(existingUser.AadhaarNumber)
		userResponse.Status = helper.SafeString(existingUser.Status)
		userResponse.Name = helper.SafeString(existingUser.Name)
		userResponse.CareOf = helper.SafeString(existingUser.CareOf)
		userResponse.DateOfBirth = helper.SafeString(existingUser.DateOfBirth)
		userResponse.Photo = helper.SafeString(existingUser.Photo)
		userResponse.EmailHash = helper.SafeString(existingUser.EmailHash)
		userResponse.ShareCode = helper.SafeString(existingUser.ShareCode)
		userResponse.YearOfBirth = helper.SafeString(existingUser.YearOfBirth)
		userResponse.Message = helper.SafeString(existingUser.Message)
		userResponse.MobileNumber = &existingUser.MobileNumber
		userResponse.CountryCode = helper.SafeString(existingUser.CountryCode)
		userResponse.Address = address
	}

	// Set auth headers
	if err := helper.SetAuthHeadersWithTokensRest(
		c,
		existingUser.ID,
		existingUser.Username,
		existingUser.IsValidated,
	); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to set auth headers"})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Login successful", userResponse)
}
