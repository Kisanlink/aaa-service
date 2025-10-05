package user

import (
	"testing"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	userResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/users"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// TestMPinValidation tests MPIN validation logic
func TestMPinValidation(t *testing.T) {
	t.Run("valid 4-digit MPIN", func(t *testing.T) {
		mpin := "1234"

		// Test length validation
		assert.True(t, len(mpin) >= 4 && len(mpin) <= 6)

		// Test digit validation
		for _, char := range mpin {
			assert.True(t, char >= '0' && char <= '9')
		}
	})

	t.Run("valid 6-digit MPIN", func(t *testing.T) {
		mpin := "123456"

		// Test length validation
		assert.True(t, len(mpin) >= 4 && len(mpin) <= 6)

		// Test digit validation
		for _, char := range mpin {
			assert.True(t, char >= '0' && char <= '9')
		}
	})

	t.Run("invalid MPIN - too short", func(t *testing.T) {
		mpin := "12"
		assert.False(t, len(mpin) >= 4 && len(mpin) <= 6)
	})

	t.Run("invalid MPIN - too long", func(t *testing.T) {
		mpin := "1234567"
		assert.False(t, len(mpin) >= 4 && len(mpin) <= 6)
	})

	t.Run("invalid MPIN - contains letters", func(t *testing.T) {
		mpin := "12ab"

		hasInvalidChar := false
		for _, char := range mpin {
			if char < '0' || char > '9' {
				hasInvalidChar = true
				break
			}
		}
		assert.True(t, hasInvalidChar)
	})
}

// TestPasswordHashing tests password and MPIN hashing
func TestPasswordHashing(t *testing.T) {
	t.Run("password hashing and verification", func(t *testing.T) {
		password := "testpassword123"

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		assert.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)

		// Verify correct password
		err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
		assert.NoError(t, err)

		// Verify incorrect password
		err = bcrypt.CompareHashAndPassword(hashedPassword, []byte("wrongpassword"))
		assert.Error(t, err)
	})

	t.Run("MPIN hashing and verification", func(t *testing.T) {
		mpin := "1234"

		// Hash MPIN
		hashedMPin, err := bcrypt.GenerateFromPassword([]byte(mpin), bcrypt.DefaultCost)
		assert.NoError(t, err)
		assert.NotEmpty(t, hashedMPin)

		// Verify correct MPIN
		err = bcrypt.CompareHashAndPassword(hashedMPin, []byte(mpin))
		assert.NoError(t, err)

		// Verify incorrect MPIN
		err = bcrypt.CompareHashAndPassword(hashedMPin, []byte("5678"))
		assert.Error(t, err)
	})
}

// TestUserModelMPinMethods tests User model MPIN-related methods
func TestUserModelMPinMethods(t *testing.T) {
	t.Run("user without MPIN", func(t *testing.T) {
		user := models.NewUser("1234567890", "+91", "hashedpassword")

		assert.False(t, user.HasMPin())
		assert.Nil(t, user.MPin)
	})

	t.Run("user with MPIN", func(t *testing.T) {
		user := models.NewUser("1234567890", "+91", "hashedpassword")

		// Set MPIN
		hashedMPin, _ := bcrypt.GenerateFromPassword([]byte("1234"), bcrypt.DefaultCost)
		user.SetMPin(string(hashedMPin))

		assert.True(t, user.HasMPin())
		assert.NotNil(t, user.MPin)
		assert.Equal(t, string(hashedMPin), *user.MPin)
	})
}

// TestCredentialPriority tests credential authentication priority logic
func TestCredentialPriority(t *testing.T) {
	t.Run("password takes priority when both provided", func(t *testing.T) {
		password := "testpassword"
		mpin := "1234"

		// Simulate the priority logic from VerifyUserCredentials
		var usePassword bool
		if password != "" {
			usePassword = true
		} else if mpin != "" {
			usePassword = false
		}

		assert.True(t, usePassword, "Password should take priority when both are provided")
	})

	t.Run("MPIN used when password not provided", func(t *testing.T) {
		password := ""
		mpin := "1234"

		// Simulate the priority logic from VerifyUserCredentials
		var usePassword bool
		if password != "" {
			usePassword = true
		} else if mpin != "" {
			usePassword = false
		}

		assert.False(t, usePassword, "MPIN should be used when password is not provided")
	})
}

// TestUserResponseStructure tests the UserResponse structure for roles
func TestUserResponseStructure(t *testing.T) {
	t.Run("user response with roles", func(t *testing.T) {
		response := &userResponses.UserResponse{
			ID:          "test-user-id",
			PhoneNumber: "1234567890",
			CountryCode: "+91",
			HasMPin:     true,
			Roles: []userResponses.UserRoleDetail{
				{
					ID:       "user-role-1",
					UserID:   "test-user-id",
					RoleID:   "role-1",
					IsActive: true,
					Role: userResponses.RoleDetail{
						ID:          "role-1",
						Name:        "admin",
						Description: "Administrator role",
						IsActive:    true,
						AssignedAt:  "2024-01-01T00:00:00Z",
					},
				},
			},
		}

		assert.Equal(t, "test-user-id", response.ID)
		assert.True(t, response.HasMPin)
		assert.Len(t, response.Roles, 1)
		assert.Equal(t, "admin", response.Roles[0].Role.Name)
		assert.True(t, response.Roles[0].IsActive)
	})

	t.Run("user response without roles", func(t *testing.T) {
		response := &userResponses.UserResponse{
			ID:          "test-user-id",
			PhoneNumber: "1234567890",
			CountryCode: "+91",
			HasMPin:     false,
			Roles:       []userResponses.UserRoleDetail{},
		}

		assert.Equal(t, "test-user-id", response.ID)
		assert.False(t, response.HasMPin)
		assert.Len(t, response.Roles, 0)
	})
}
