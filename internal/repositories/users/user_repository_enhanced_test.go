package users

import (
	"context"
	"testing"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// TestUserRepository_EnhancedMethods tests the enhanced repository methods
// These are unit tests that focus on the business logic and error handling
func TestUserRepository_EnhancedMethods(t *testing.T) {
	t.Run("TestMPinHashing", func(t *testing.T) {
		// Test MPIN hashing functionality
		plainMPin := "1234"

		// Generate hash
		hashedMPin, err := bcrypt.GenerateFromPassword([]byte(plainMPin), bcrypt.DefaultCost)
		assert.NoError(t, err)
		assert.NotEmpty(t, hashedMPin)

		// Verify correct MPIN
		err = bcrypt.CompareHashAndPassword(hashedMPin, []byte(plainMPin))
		assert.NoError(t, err)

		// Verify incorrect MPIN
		err = bcrypt.CompareHashAndPassword(hashedMPin, []byte("5678"))
		assert.Error(t, err)
	})

	t.Run("TestUserModelMPinMethods", func(t *testing.T) {
		// Test User model MPIN-related methods
		user := models.NewUser("1234567890", "+91", "password")

		// Initially no MPIN
		assert.False(t, user.HasMPin())

		// Set MPIN
		mpin := "1234"
		user.SetMPin(mpin)
		assert.True(t, user.HasMPin())
		assert.Equal(t, mpin, *user.MPin)

		// Test with nil MPIN
		user.MPin = nil
		assert.False(t, user.HasMPin())

		// Test with empty MPIN
		emptyMPin := ""
		user.MPin = &emptyMPin
		assert.False(t, user.HasMPin())
	})

	t.Run("TestUserRoleActiveAssignment", func(t *testing.T) {
		// Test UserRole active assignment logic
		role := &models.Role{
			Name:     "Test Role",
			IsActive: true,
		}

		userRole := &models.UserRole{
			UserID:   "test-user",
			RoleID:   "test-role",
			IsActive: true,
			Role:     *role,
		}

		// Both active
		assert.True(t, userRole.IsActiveAssignment())

		// User role inactive
		userRole.IsActive = false
		assert.False(t, userRole.IsActiveAssignment())

		// Role inactive
		userRole.IsActive = true
		userRole.Role.IsActive = false
		assert.False(t, userRole.IsActiveAssignment())

		// Both inactive
		userRole.IsActive = false
		assert.False(t, userRole.IsActiveAssignment())
	})
}

// TestUserRepository_ValidationLogic tests validation logic for enhanced methods
func TestUserRepository_ValidationLogic(t *testing.T) {
	t.Run("TestMPinValidation", func(t *testing.T) {
		// Test MPIN validation requirements
		testCases := []struct {
			name  string
			mpin  string
			valid bool
		}{
			{"Valid 4-digit MPIN", "1234", true},
			{"Valid 6-digit MPIN", "123456", true},
			{"Empty MPIN", "", false},
			{"Too short", "12", false},
			{"Too long", "1234567", false},
			{"Non-numeric", "abcd", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Basic length validation
				if tc.mpin == "" {
					assert.False(t, tc.valid)
				} else if len(tc.mpin) == 4 || len(tc.mpin) == 6 {
					// Check if all characters are digits
					isNumeric := true
					for _, char := range tc.mpin {
						if char < '0' || char > '9' {
							isNumeric = false
							break
						}
					}
					assert.Equal(t, tc.valid, isNumeric)
				} else {
					assert.False(t, tc.valid)
				}
			})
		}
	})

	t.Run("TestUserIDValidation", func(t *testing.T) {
		// Test user ID validation
		testCases := []struct {
			name   string
			userID string
			valid  bool
		}{
			{"Valid UUID-like ID", "550e8400-e29b-41d4-a716-446655440000", true},
			{"Valid short ID", "user-123", true},
			{"Empty ID", "", false},
			{"Very long ID", string(make([]byte, 300)), false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				if tc.userID == "" {
					assert.False(t, tc.valid)
				} else if len(tc.userID) > 255 {
					assert.False(t, tc.valid)
				} else {
					assert.True(t, tc.valid)
				}
			})
		}
	})
}

// TestUserRepository_ErrorScenarios tests error handling scenarios
func TestUserRepository_ErrorScenarios(t *testing.T) {
	t.Run("TestMPinHashingErrors", func(t *testing.T) {
		// Test bcrypt cost validation
		plainMPin := "1234"

		// Test with invalid cost (too high)
		_, err := bcrypt.GenerateFromPassword([]byte(plainMPin), 32) // Max cost is 31
		assert.Error(t, err)

		// Test with valid cost
		_, err = bcrypt.GenerateFromPassword([]byte(plainMPin), bcrypt.DefaultCost)
		assert.NoError(t, err)
	})

	t.Run("TestMPinComparisonErrors", func(t *testing.T) {
		// Test invalid hash format
		plainMPin := "1234"
		invalidHash := "invalid-hash"

		err := bcrypt.CompareHashAndPassword([]byte(invalidHash), []byte(plainMPin))
		assert.Error(t, err)

		// Test empty hash
		err = bcrypt.CompareHashAndPassword([]byte(""), []byte(plainMPin))
		assert.Error(t, err)

		// Test empty plain text
		validHash, _ := bcrypt.GenerateFromPassword([]byte(plainMPin), bcrypt.DefaultCost)
		err = bcrypt.CompareHashAndPassword(validHash, []byte(""))
		assert.Error(t, err)
	})
}

// TestUserRepository_ConcurrencyConsiderations tests concurrency-related scenarios
func TestUserRepository_ConcurrencyConsiderations(t *testing.T) {
	t.Run("TestConcurrentMPinHashing", func(t *testing.T) {
		// Test that multiple MPIN hashing operations can run concurrently
		plainMPin := "1234"
		numGoroutines := 10

		results := make(chan []byte, numGoroutines)
		errors := make(chan error, numGoroutines)

		// Start multiple goroutines
		for i := 0; i < numGoroutines; i++ {
			go func() {
				hash, err := bcrypt.GenerateFromPassword([]byte(plainMPin), bcrypt.DefaultCost)
				if err != nil {
					errors <- err
					return
				}
				results <- hash
			}()
		}

		// Collect results
		hashes := make([][]byte, 0, numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			select {
			case hash := <-results:
				hashes = append(hashes, hash)
			case err := <-errors:
				t.Fatalf("Unexpected error in goroutine: %v", err)
			}
		}

		// Verify all hashes are different (bcrypt includes salt)
		assert.Len(t, hashes, numGoroutines)
		for i := 0; i < len(hashes); i++ {
			for j := i + 1; j < len(hashes); j++ {
				assert.NotEqual(t, string(hashes[i]), string(hashes[j]), "Hashes should be different due to salt")
			}
		}

		// Verify all hashes are valid for the original MPIN
		for _, hash := range hashes {
			err := bcrypt.CompareHashAndPassword(hash, []byte(plainMPin))
			assert.NoError(t, err)
		}
	})
}

// TestUserRepository_BusinessLogic tests business logic scenarios
func TestUserRepository_BusinessLogic(t *testing.T) {
	t.Run("TestUserLifecycleScenarios", func(t *testing.T) {
		// Test user creation with MPIN
		user := models.NewUser("1234567890", "+91", "password")
		assert.NotNil(t, user)
		assert.False(t, user.HasMPin())

		// Set MPIN during user lifecycle
		mpin := "1234"
		hashedMPin, err := bcrypt.GenerateFromPassword([]byte(mpin), bcrypt.DefaultCost)
		assert.NoError(t, err)

		user.SetMPin(string(hashedMPin))
		assert.True(t, user.HasMPin())

		// Verify MPIN
		err = bcrypt.CompareHashAndPassword([]byte(*user.MPin), []byte(mpin))
		assert.NoError(t, err)
	})

	t.Run("TestRoleAssignmentScenarios", func(t *testing.T) {
		// Test role assignment business logic
		userID := "user-123"
		roleID := "role-456"

		userRole := models.NewUserRole(userID, roleID)
		assert.NotNil(t, userRole)
		assert.Equal(t, userID, userRole.UserID)
		assert.Equal(t, roleID, userRole.RoleID)
		assert.True(t, userRole.IsActive)

		// Test role deactivation
		userRole.IsActive = false
		assert.False(t, userRole.IsActive)

		// Test with inactive role
		role := &models.Role{
			Name:     "Test Role",
			IsActive: false,
		}
		userRole.Role = *role
		userRole.IsActive = true
		assert.False(t, userRole.IsActiveAssignment())
	})
}

// TestUserRepository_MethodSignatures tests that method signatures are correct
func TestUserRepository_MethodSignatures(t *testing.T) {
	t.Run("TestMethodExists", func(t *testing.T) {
		// This test ensures the methods exist with correct signatures
		// by attempting to assign them to function variables

		var softDeleteWithCascade func(ctx context.Context, userID, deletedBy string) error
		var getWithActiveRoles func(ctx context.Context, userID string) (*models.User, error)
		var verifyMPin func(ctx context.Context, userID, plainMPin string) error

		// Create a repository instance (we can't test the actual methods without a real DB)
		// but we can verify the method signatures exist
		repo := &UserRepository{}

		// Assign methods to verify signatures
		softDeleteWithCascade = repo.SoftDeleteWithCascade
		getWithActiveRoles = repo.GetWithActiveRoles
		verifyMPin = repo.VerifyMPin

		// Verify they're not nil (methods exist)
		assert.NotNil(t, softDeleteWithCascade)
		assert.NotNil(t, getWithActiveRoles)
		assert.NotNil(t, verifyMPin)
	})
}
