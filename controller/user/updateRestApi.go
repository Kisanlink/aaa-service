package user

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type UpdateUserRequest struct {
	Username      string `json:"username"`
	AadhaarNumber string `json:"aadhaar_number"`
	Status        string `json:"status"`
	Name          string `json:"name"`
	CareOf        string `json:"care_of"`
	DateOfBirth   string `json:"date_of_birth"`
	Photo         string `json:"photo"`
	EmailHash     string `json:"email_hash"`
	ShareCode     string `json:"share_code"`
	YearOfBirth   string `json:"year_of_birth"`
	Message       string `json:"message"`
	MobileNumber  uint64 `json:"mobile_number"`
}

type UpdateUserResponse struct {
	StatusCode    int    `json:"status_code"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	Data          *User  `json:"data"`
	DataTimeStamp string `json:"data_time_stamp"`
}

func (s *Server) UpdateUserRestApi(c *gin.Context) {
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

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": http.StatusBadRequest,
			"success":     false,
			"message":     "Invalid request body",
			"data":        nil,
		})
		return
	}

	existingUser, err := s.UserRepo.FindExistingUserByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": http.StatusInternalServerError,
			"success":     false,
			"message":     "failed to fetch user",
			"data":        nil,
		})
		return
	}

	if existingUser == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status_code": http.StatusNotFound,
			"success":     false,
			"message":     "user not found",
			"data":        nil,
		})
		return
	}

	if req.Username != "" {
		existingUser.Username = req.Username
	}
	if req.AadhaarNumber != "" {
		existingUser.AadhaarNumber = &req.AadhaarNumber
	}
	if req.Status != "" {
		existingUser.Status = &req.Status
	}
	if req.Name != "" {
		existingUser.Name = &req.Name
	}
	if req.CareOf != "" {
		existingUser.CareOf = &req.CareOf
	}
	if req.DateOfBirth != "" {
		existingUser.DateOfBirth = &req.DateOfBirth
	}
	if req.Photo != "" {
		existingUser.Photo = &req.Photo
	}
	if req.EmailHash != "" {
		existingUser.EmailHash = &req.EmailHash
	}
	if req.ShareCode != "" {
		existingUser.ShareCode = &req.ShareCode
	}
	if req.YearOfBirth != "" {
		existingUser.YearOfBirth = &req.YearOfBirth
	}
	if req.Message != "" {
		existingUser.Message = &req.Message
	}
	if req.MobileNumber != 0 {
		existingUser.MobileNumber = req.MobileNumber
	}

	if err := s.UserRepo.UpdateUser(ctx, *existingUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": http.StatusInternalServerError,
			"success":     false,
			"message":     "failed to update user",
			"data":        nil,
		})
		return
	}

	rolePermissions, err := s.UserRepo.FindUsageRights(ctx, existingUser.ID)
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
	if existingUser.AddressID != nil {
		addr, err := s.UserRepo.GetAddressByID(ctx, *existingUser.AddressID)
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
			}
		}
	}

	responseUser := &User{
		ID:            existingUser.ID,
		Username:      existingUser.Username,
		Password:      "",
		IsValidated:   existingUser.IsValidated,
		CreatedAt:     existingUser.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:     existingUser.UpdatedAt.Format(time.RFC3339Nano),
		Roles:         rolesResponse,
		AadhaarNumber: safeString(existingUser.AadhaarNumber),
		Status:        safeString(existingUser.Status),
		Name:          safeString(existingUser.Name),
		CareOf:        safeString(existingUser.CareOf),
		DateOfBirth:   safeString(existingUser.DateOfBirth),
		Photo:         safeString(existingUser.Photo),
		EmailHash:     safeString(existingUser.EmailHash),
		ShareCode:     safeString(existingUser.ShareCode),
		YearOfBirth:   safeString(existingUser.YearOfBirth),
		Message:       safeString(existingUser.Message),
		MobileNumber:  existingUser.MobileNumber,
		CountryCode:   safeString(existingUser.CountryCode),
		Address:       address,
	}

	response := UpdateUserResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "User updated successfully",
		Data:          responseUser,
		DataTimeStamp: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}
