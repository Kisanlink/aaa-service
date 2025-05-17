package user

import (
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// GetUserByIdRestApi retrieves a user by ID
// @Summary Get user by ID
// @Description Retrieves a single user's details including roles, permissions, and address information by their unique ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID" example(123e4567-e89b-12d3-a456-426614174000)
// @Success 200 {object} helper.Response{data=model.UserRes} "User fetched successfully"
// @Failure 400 {object} helper.Response "ID is required"
// @Failure 404 {object} helper.Response "User not found"
// @Failure 500 {object} helper.Response "Internal server error when fetching user or related data"
// @Route /users/{id} [get]
func (s *UserHandler) GetUserByIdRestApi(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"ID is required"})
		return
	}

	user, err := s.userService.FindExistingUserByID(id)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to fetch user"})
		return
	}

	if user == nil {
		helper.SendErrorResponse(c.Writer, http.StatusNotFound, []string{"User not found"})
		return
	}

	rolePermissions, err := s.userService.FindUsageRights(user.ID)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to fetch role permissions"})
		return
	}

	// Convert role permissions to the new structure
	var rolesResponse []model.RoleResp
	for roleName, permissions := range rolePermissions {
		uniquePerms := make(map[string]model.Permission)
		for _, perm := range permissions {
			key := perm.Name + ":" + perm.Action + ":" + perm.Resource
			if _, exists := uniquePerms[key]; !exists {
				uniquePerms[key] = model.Permission{
					Name:        perm.Name,
					Description: perm.Description,
					Action:      perm.Action,
					Source:      perm.Source,
					Resource:    perm.Resource,
				}
			}
		}

		var permsSlice []model.Permission
		for _, perm := range uniquePerms {
			permsSlice = append(permsSlice, perm)
		}

		rolesResponse = append(rolesResponse, model.RoleResp{
			RoleName:    roleName,
			Permissions: permsSlice,
		})
	}

	var address *model.AddressRes
	if user.AddressID != nil {
		addr, err := s.userService.GetAddressByID(*user.AddressID)
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

	responseUser := &model.UserRes{
		ID:             user.ID,
		Username:       user.Username,
		Password:       "",
		IsValidated:    user.IsValidated,
		CreatedAt:      user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      user.UpdatedAt.Format(time.RFC3339),
		RolePermission: rolesResponse,
		AadhaarNumber:  helper.SafeString(user.AadhaarNumber),
		Status:         helper.SafeString(user.Status),
		Name:           helper.SafeString(user.Name),
		CareOf:         helper.SafeString(user.CareOf),
		DateOfBirth:    helper.SafeString(user.DateOfBirth),
		Photo:          helper.SafeString(user.Photo),
		EmailHash:      helper.SafeString(user.EmailHash),
		ShareCode:      helper.SafeString(user.ShareCode),
		YearOfBirth:    helper.SafeString(user.YearOfBirth),
		Message:        helper.SafeString(user.Message),
		MobileNumber:   user.MobileNumber,
		CountryCode:    helper.SafeString(user.CountryCode),
		Address:        address,
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "User fetched successfully", responseUser)
}
