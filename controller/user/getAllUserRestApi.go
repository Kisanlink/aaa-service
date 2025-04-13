package user

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Address struct {
	ID          string `json:"id"`
	House       string `json:"house"`
	Street      string `json:"street"`
	Landmark    string `json:"landmark"`
	PostOffice  string `json:"post_office"`
	Subdistrict string `json:"subdistrict"`
	District    string `json:"district"`
	VTC         string `json:"vtc"`
	State       string `json:"state"`
	Country     string `json:"country"`
	Pincode     string `json:"pincode"`
	FullAddress string `json:"full_address"`
}

type User struct {
	ID              string                          `json:"id"`
	Username        string                          `json:"username"`
	Password        string                          `json:"password"`
	IsValidated     bool                            `json:"is_validated"`
	CreatedAt       string                          `json:"created_at"`
	UpdatedAt       string                          `json:"updated_at"`
	RolePermissions map[string][]PermissionResponse `json:"role_permissions"`
	AadhaarNumber   string                          `json:"aadhaar_number"`
	Status          string                          `json:"status"`
	Name            string                          `json:"name"`
	CareOf          string                          `json:"care_of"`
	DateOfBirth     string                          `json:"date_of_birth"`
	Photo           string                          `json:"photo"`
	EmailHash       string                          `json:"email_hash"`
	ShareCode       string                          `json:"share_code"`
	YearOfBirth     string                          `json:"year_of_birth"`
	Message         string                          `json:"message"`
	MobileNumber    uint64                          `json:"mobile_number"`
	CountryCode     string                          `json:"country_code"`
	Address         *Address                        `json:"address"`
}

type GetUserResponse struct {
	StatusCode    int    `json:"status_code"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	Data          []User `json:"data"`
	DataTimeStamp string `json:"data_time_stamp"`
}

func (s *Server) GetUserRestApi(c *gin.Context) {
	ctx := c.Request.Context()

	users, err := s.UserRepo.GetUsers(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": http.StatusInternalServerError,
			"success":     false,
			"message":     "Failed to fetch users",
			"data":        nil,
		})
		return
	}

	var responseUsers []User
	for _, user := range users {
		// Get role permissions
		rolePermissions, err := s.UserRepo.FindUsageRights(ctx, user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status_code": http.StatusInternalServerError,
				"success":     false,
				"message":     "Failed to fetch user permissions",
				"data":        nil,
			})
			return
		}

		// Process and deduplicate permissions
		processedRolePermissions := make(map[string][]PermissionResponse)
		for role, permissions := range rolePermissions {
			uniquePerms := make(map[string]PermissionResponse)
			for _, perm := range permissions {
				key := perm.Name + ":" + perm.Action + ":" + perm.Resource
				uniquePerms[key] = PermissionResponse{
					Name:        perm.Name,
					Description: perm.Description,
					Action:      perm.Action,
					Source:      perm.Source,
					Resource:    perm.Resource,
				}
			}
			// Convert unique permissions map to slice
			var permsSlice []PermissionResponse
			for _, perm := range uniquePerms {
				permsSlice = append(permsSlice, perm)
			}
			processedRolePermissions[role] = permsSlice
		}

		var address *Address
		if user.AddressID != nil {
			addr, err := s.UserRepo.GetAddressByID(ctx, *user.AddressID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status_code": http.StatusInternalServerError,
					"success":     false,
					"message":     "Failed to fetch address",
					"data":        nil,
				})
				return
			}

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

		responseUser := User{
			ID:              user.ID,
			Username:        user.Username,
			Password:        "", // Don't return password in response
			IsValidated:     user.IsValidated,
			CreatedAt:       user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       user.UpdatedAt.Format(time.RFC3339),
			RolePermissions: processedRolePermissions,
			AadhaarNumber:   safeString(user.AadhaarNumber),
			Status:          safeString(user.Status),
			Name:            safeString(user.Name),
			CareOf:          safeString(user.CareOf),
			DateOfBirth:     safeString(user.DateOfBirth),
			Photo:           safeString(user.Photo),
			EmailHash:       safeString(user.EmailHash),
			ShareCode:       safeString(user.ShareCode),
			YearOfBirth:     safeString(user.YearOfBirth),
			Message:         safeString(user.Message),
			MobileNumber:    user.MobileNumber,
			CountryCode:     safeString(user.CountryCode),
			Address:         address,
		}

		responseUsers = append(responseUsers, responseUser)
	}

	response := GetUserResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "Users fetched successfully",
		Data:          responseUsers,
		DataTimeStamp: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}
