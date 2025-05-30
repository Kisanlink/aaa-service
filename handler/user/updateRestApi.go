package user

import (
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// UpdateUserRestApi updates user information
// @Summary Update user
// @Description Updates user information by ID. Only provided fields will be updated (partial update supported).
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Param request body model.UpdateUserRequest true "User update data (partial updates allowed)"
// @Success 200 {object} helper.Response{data=model.UserRes} "User updated successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid ID or request body"
// @Failure 404 {object} helper.ErrorResponse "User not found"
// @Failure 500 {object} helper.ErrorResponse "Failed to update user or fetch related data"
// @Router /users/{id} [put]
func (s *UserHandler) UpdateUserRestApi(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"ID is required"})
		return
	}

	var req model.User
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request body"})
		return
	}

	existingUser, err := s.userService.FindExistingUserByID(id)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}
	if existingUser == nil {
		helper.SendErrorResponse(c.Writer, http.StatusNotFound, []string{"User not found"})
		return
	}

	// Update fields
	if req.Username != "" {
		existingUser.Username = req.Username
	}
	if req.AadhaarNumber != nil && *req.AadhaarNumber != "" {
		existingUser.AadhaarNumber = req.AadhaarNumber
	}
	if req.Status != nil && *req.Status != "" {
		existingUser.Status = req.Status
	}
	if req.Name != nil && *req.Name != "" {
		existingUser.Name = req.Name
	}
	if req.CareOf != nil && *req.CareOf != "" {
		existingUser.CareOf = req.CareOf
	}
	if req.DateOfBirth != nil && *req.DateOfBirth != "" {
		existingUser.DateOfBirth = req.DateOfBirth
	}
	if req.Photo != nil && *req.Photo != "" {
		existingUser.Photo = req.Photo
	}
	if req.EmailHash != nil && *req.EmailHash != "" {
		existingUser.EmailHash = req.EmailHash
	}
	if req.ShareCode != nil && *req.ShareCode != "" {
		existingUser.ShareCode = req.ShareCode
	}
	if req.YearOfBirth != nil && *req.YearOfBirth != "" {
		existingUser.YearOfBirth = req.YearOfBirth
	}
	if req.Message != nil && *req.Message != "" {
		existingUser.Message = req.Message
	}
	if req.MobileNumber != 0 {
		existingUser.MobileNumber = req.MobileNumber
	}

	if err := s.userService.UpdateUser(*existingUser); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}
	rolesResponse, err := s.userService.GetUserRolesWithPermissions(req.ID)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}
	// Fetch address if exists
	var address *model.AddressRes
	if existingUser.AddressID != nil {
		addr, err := s.userService.GetAddressByID(*existingUser.AddressID)
		if err != nil {
			helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
			return
		}
		if addr != nil {
			address = &model.AddressRes{
				ID:          addr.ID,
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
			}
		}
	}

	responseUser := &model.UserRes{
		ID:            existingUser.ID,
		Username:      existingUser.Username,
		Password:      "",
		IsValidated:   existingUser.IsValidated,
		CreatedAt:     existingUser.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:     existingUser.UpdatedAt.Format(time.RFC3339Nano),
		AadhaarNumber: helper.SafeString(existingUser.AadhaarNumber),
		Status:        helper.SafeString(existingUser.Status),
		Name:          helper.SafeString(existingUser.Name),
		CareOf:        helper.SafeString(existingUser.CareOf),
		DateOfBirth:   helper.SafeString(existingUser.DateOfBirth),
		Photo:         helper.SafeString(existingUser.Photo),
		EmailHash:     helper.SafeString(existingUser.EmailHash),
		ShareCode:     helper.SafeString(existingUser.ShareCode),
		YearOfBirth:   helper.SafeString(existingUser.YearOfBirth),
		Message:       helper.SafeString(existingUser.Message),
		MobileNumber:  existingUser.MobileNumber,
		CountryCode:   helper.SafeString(existingUser.CountryCode),
		Address:       address,
		Roles:         rolesResponse.Roles,
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "User updated successfully", responseUser)
}
