package user

import (
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Source      string `json:"source"`
	Resource    string `json:"resource"`
}

type Role struct {
	RoleName    string       `json:"role_name"`
	Permissions []Permission `json:"permissions"`
}

type UserResponse struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	IsValidated    bool   `json:"is_validated"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	RolePermission []Role `json:"role_permissions"`

	// Detailed fields (only included when user_details=true)
	AadhaarNumber string   `json:"aadhaar_number,omitempty"`
	Status        string   `json:"status,omitempty"`
	Name          string   `json:"name,omitempty"`
	CareOf        string   `json:"care_of,omitempty"`
	DateOfBirth   string   `json:"date_of_birth,omitempty"`
	Photo         string   `json:"photo,omitempty"`
	EmailHash     string   `json:"email_hash,omitempty"`
	ShareCode     string   `json:"share_code,omitempty"`
	YearOfBirth   string   `json:"year_of_birth,omitempty"`
	Message       string   `json:"message,omitempty"`
	MobileNumber  *uint64  `json:"mobile_number,omitempty"`
	CountryCode   string   `json:"country_code,omitempty"`
	Address       *Address `json:"address,omitempty"`
}
type LoginResponse struct {
	StatusCode    int          `json:"status_code"`
	Success       bool         `json:"success"`
	Message       string       `json:"message"`
	Data          UserResponse `json:"data"`
	DataTimeStamp string       `json:"data_time_stamp"`
}

func ConvertAndDeduplicateRolePermissions(input map[string][]model.Permission) []Role {
	var roles []Role

	for roleName, permissions := range input {
		uniquePerms := make(map[string]Permission)

		// Deduplicate permissions
		for _, perm := range permissions {
			key := perm.Name + "|" + perm.Action + "|" + perm.Resource
			if _, exists := uniquePerms[key]; !exists {
				uniquePerms[key] = Permission{
					Name:        perm.Name,
					Description: perm.Description,
					Action:      perm.Action,
					Source:      perm.Source,
					Resource:    perm.Resource,
				}
			}
		}

		// Convert map to slice
		var permSlice []Permission
		for _, perm := range uniquePerms {
			permSlice = append(permSlice, perm)
		}

		roles = append(roles, Role{
			RoleName:    roleName,
			Permissions: permSlice,
		})
	}

	return roles
}

func (s *Server) LoginRestApi(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate inputs
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	if !helper.IsValidUsername(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username '" + req.Username + "' contains invalid characters. Only a-z, A-Z, 0-9, /, _, |, -, =, + are allowed, and spaces are prohibited.",
		})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	// Find user
	existingUser, err := s.UserRepo.FindUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Get role permissions
	rolePermissions, err := s.UserRepo.FindUsageRights(c.Request.Context(), existingUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user permissions"})
		return
	}

	// Check if detailed user info is requested
	includeDetails := c.Query("user_details") == "true"

	// Prepare base user response
	userResponse := UserResponse{
		ID:             existingUser.ID,
		Username:       existingUser.Username,
		IsValidated:    existingUser.IsValidated,
		CreatedAt:      existingUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      existingUser.UpdatedAt.Format(time.RFC3339),
		RolePermission: ConvertAndDeduplicateRolePermissions(rolePermissions),
	}

	// Only fetch and include additional details if requested
	if includeDetails {
		// Get address if exists
		var address *Address
		if existingUser.AddressID != nil {
			addr, err := s.UserRepo.GetAddressByID(c.Request.Context(), *existingUser.AddressID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch address"})
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

		// Add detailed fields to the response
		userResponse.AadhaarNumber = safeString(existingUser.AadhaarNumber)
		userResponse.Status = safeString(existingUser.Status)
		userResponse.Name = safeString(existingUser.Name)
		userResponse.CareOf = safeString(existingUser.CareOf)
		userResponse.DateOfBirth = safeString(existingUser.DateOfBirth)
		userResponse.Photo = safeString(existingUser.Photo)
		userResponse.EmailHash = safeString(existingUser.EmailHash)
		userResponse.ShareCode = safeString(existingUser.ShareCode)
		userResponse.YearOfBirth = safeString(existingUser.YearOfBirth)
		userResponse.Message = safeString(existingUser.Message)
		userResponse.MobileNumber = &existingUser.MobileNumber
		userResponse.CountryCode = safeString(existingUser.CountryCode)
		userResponse.Address = address
	}

	// Set auth headers
	if err := helper.SetAuthHeadersWithTokensRest(
		c,
		existingUser.ID,
		existingUser.Username,
		existingUser.IsValidated,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set auth headers"})
		return
	}

	// Prepare response
	response := LoginResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "Login successful",
		DataTimeStamp: time.Now().Format(time.RFC3339),
		Data:          userResponse,
	}

	c.JSON(http.StatusOK, response)
}
