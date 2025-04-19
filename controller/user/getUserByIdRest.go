package user

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type GetUserByIdResponse struct {
	StatusCode    int    `json:"status_code"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	Data          *User  `json:"data"`
	DataTimeStamp string `json:"data_time_stamp"`
}

func (s *Server) GetUserByIdRestApi(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": http.StatusBadRequest,
			"success":     false,
			"message":     "ID is required",
			"data":        nil,
		})
		return
	}

	user, err := s.UserRepo.FindExistingUserByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": http.StatusInternalServerError,
			"success":     false,
			"message":     "failed to fetch user",
			"data":        nil,
		})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status_code": http.StatusNotFound,
			"success":     false,
			"message":     "user not found",
			"data":        nil,
		})
		return
	}

	rolePermissions, err := s.UserRepo.FindUsageRights(ctx, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": http.StatusInternalServerError,
			"success":     false,
			"message":     "failed to fetch role permissions",
			"data":        nil,
		})
		return
	}

	// Convert role permissions to the new structure
	var rolesResponse []RoleResponse
	for roleName, permissions := range rolePermissions {
		uniquePerms := make(map[string]PermissionResponse)
		for _, perm := range permissions {
			key := perm.Name + ":" + perm.Action + ":" + perm.Resource
			if _, exists := uniquePerms[key]; !exists {
				uniquePerms[key] = PermissionResponse{
					Name:        perm.Name,
					Description: perm.Description,
					Action:      perm.Action,
					Source:      perm.Source,
					Resource:    perm.Resource,
				}
			}
		}

		var permsSlice []PermissionResponse
		for _, perm := range uniquePerms {
			permsSlice = append(permsSlice, perm)
		}

		rolesResponse = append(rolesResponse, RoleResponse{
			RoleName:    roleName,
			Permissions: permsSlice,
		})
	}

	var address *Address
	if user.AddressID != nil {
		addr, err := s.UserRepo.GetAddressByID(ctx, *user.AddressID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status_code": http.StatusInternalServerError,
				"success":     false,
				"message":     "failed to fetch address",
				"data":        nil,
			})
			return
		}

		if addr != nil {
			address = &Address{
				ID:          addr.ID,
				House:       safeString(addr.House),
				Street:      safeString(addr.Street),
				Landmark:    safeString(addr.Landmark),
				PostOffice:  safeString(addr.PostOffice),
				Subdistrict: safeString(addr.Subdistrict),
				District:    safeString(addr.District),
				VTC:         safeString(addr.VTC),
				State:       safeString(addr.State),
				Country:     safeString(addr.Country),
				Pincode:     safeString(addr.Pincode),
				FullAddress: safeString(addr.FullAddress),
				CreatedAt:   addr.CreatedAt.Format(time.RFC3339),
				UpdatedAt:   addr.UpdatedAt.Format(time.RFC3339),
			}
		}
	}

	responseUser := &User{
		ID:             user.ID,
		Username:       user.Username,
		Password:       "",
		IsValidated:    user.IsValidated,
		CreatedAt:      user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      user.UpdatedAt.Format(time.RFC3339),
		RolePermission: rolesResponse,
		AadhaarNumber:  safeString(user.AadhaarNumber),
		Status:         safeString(user.Status),
		Name:           safeString(user.Name),
		CareOf:         safeString(user.CareOf),
		DateOfBirth:    safeString(user.DateOfBirth),
		Photo:          safeString(user.Photo),
		EmailHash:      safeString(user.EmailHash),
		ShareCode:      safeString(user.ShareCode),
		YearOfBirth:    safeString(user.YearOfBirth),
		Message:        safeString(user.Message),
		MobileNumber:   user.MobileNumber,
		CountryCode:    safeString(user.CountryCode),
		Address:        address,
	}

	response := GetUserByIdResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "User fetched successfully",
		Data:          responseUser,
		DataTimeStamp: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}
