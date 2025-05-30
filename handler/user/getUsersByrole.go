package user

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// GetUsersByRoleRestApi retrieves users by role name or role ID (both optional)
// @Summary Get users by role
// @Description Retrieves a list of users filtered by role name or role ID (both optional)
// @Tags Users
// @Accept json
// @Produce json
// @Param roleName query string false "Role name to filter users"
// @Param roleId query string false "Role ID to filter users"
// @Param page query int false "Page number (starts from 1)"
// @Param limit query int false "Number of items per page"
// @Success 200 {object} helper.Response{data=[]model.UserRes} "Users fetched successfully"
// @Failure 400 {object} helper.ErrorResponse "Bad request if both roleName and roleId are provided"
// @Failure 404 {object} helper.ErrorResponse "Role not found when searching by name"
// @Failure 500 {object} helper.ErrorResponse "Internal server error"
// @Router /users/by-role [get]
func (s *UserHandler) GetUsersByRoleRestApi(c *gin.Context) {
	// Get query parameters
	roleName := c.Query("roleName")
	roleIdStr := c.Query("roleId")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0"))

	// Validate that not both roleName and roleId are provided
	if roleName != "" && roleIdStr != "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Please provide either roleName or roleId, not both"})
		return
	}

	var roleId string
	var err error

	// If roleName is provided, find the role by name first
	if roleName != "" {
		role, err := s.RoleService.GetRoleByName(roleName)
		if err != nil {
			helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
			return
		}
		roleId = (role.ID)
	}
	users, err := s.userService.GetUsersByRole(roleId, page, limit)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to fetch users"})
		return
	}

	var responseUsers []model.UserRes
	for _, user := range users {
		rolesResponse, err := s.userService.GetUserRolesWithPermissions(user.ID)
		if err != nil {
			helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
			return
		}

		var address *model.AddressRes
		if user.AddressID != nil {
			addr, err := s.userService.GetAddressByID(*user.AddressID)
			if err != nil {
				helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
				return
			}

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

		responseUser := model.UserRes{
			ID:            user.ID,
			Username:      user.Username,
			IsValidated:   user.IsValidated,
			CreatedAt:     user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     user.UpdatedAt.Format(time.RFC3339),
			AadhaarNumber: helper.SafeString(user.AadhaarNumber),
			Status:        helper.SafeString(user.Status),
			Name:          helper.SafeString(user.Name),
			CareOf:        helper.SafeString(user.CareOf),
			DateOfBirth:   helper.SafeString(user.DateOfBirth),
			Photo:         helper.SafeString(user.Photo),
			EmailHash:     helper.SafeString(user.EmailHash),
			ShareCode:     helper.SafeString(user.ShareCode),
			YearOfBirth:   helper.SafeString(user.YearOfBirth),
			Message:       helper.SafeString(user.Message),
			MobileNumber:  user.MobileNumber,
			CountryCode:   helper.SafeString(user.CountryCode),
			Address:       address,
			Roles:         rolesResponse.Roles,
		}

		responseUsers = append(responseUsers, responseUser)
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Users fetched successfully", responseUsers)
}
