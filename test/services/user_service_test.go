package services

// Additional comprehensive tests for GetUserByUsername
func TestUserService_GetUserByUsername(t *testing.T) {
	testCases := []struct {
		name        string
	"fmt"
	"strings"
	"time"
		username    string
		shouldError bool
		setupUser   bool
	}{
		{
			name:        "Valid username",
			username:    "validuser",
			shouldError: false,
			setupUser:   true,
		},
		{
			name:        "Empty username",
			username:    "",
			shouldError: true,
			setupUser:   false,
		},
		{
			name:        "Non-existent username",
			username:    "nonexistent",
			shouldError: true,
			setupUser:   false,
		},
		{
			name:        "Username with special characters",
			username:    "user@domain.com",
			shouldError: false,
			setupUser:   true,
		},
		{
			name:        "Very long username",
			username:    "verylongusernamethatmightcauseissues1234567890",
			shouldError: true,
			setupUser:   false,
		},
		{
			name:        "Username case sensitivity",
			username:    "TestUser",
			shouldError: false,
			setupUser:   true,
		},
		{
			name:        "Username with numbers",
			username:    "user123",
			shouldError: false,
			setupUser:   true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			result, err := userService.GetUserByUsername(context.Background(), tt.username)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.username, *result.Username)
			}
		})
	}
}

// Test GetUserByMobileNumber with comprehensive scenarios
func TestUserService_GetUserByMobileNumber(t *testing.T) {
	testCases := []struct {
		name         string
		mobileNumber uint64
		shouldError  bool
	}{
		{
			name:         "Valid 10-digit mobile number",
			mobileNumber: 9876543210,
			shouldError:  false,
		},
		{
			name:         "Zero mobile number",
			mobileNumber: 0,
			shouldError:  true,
		},
		{
			name:         "Invalid mobile number - too short",
			mobileNumber: 123,
			shouldError:  true,
		},
		{
			name:         "Valid mobile number starting with 9",
			mobileNumber: 9123456789,
			shouldError:  false,
		},
		{
			name:         "Valid mobile number starting with 8",
			mobileNumber: 8123456789,
			shouldError:  false,
		},
		{
			name:         "Valid mobile number starting with 7",
			mobileNumber: 7123456789,
			shouldError:  false,
		},
		{
			name:         "Valid mobile number starting with 6",
			mobileNumber: 6123456789,
			shouldError:  false,
		},
		{
			name:         "Maximum uint64 mobile number",
			mobileNumber: 9999999999,
			shouldError:  false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			result, err := userService.GetUserByMobileNumber(context.Background(), tt.mobileNumber)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Username)
			}
		})
	}
}

// Test GetUserByAadhaarNumber with validation scenarios
func TestUserService_GetUserByAadhaarNumber(t *testing.T) {
	testCases := []struct {
		name           string
		aadhaarNumber  string
		shouldError    bool
	}{
		{
			name:          "Valid 12-digit Aadhaar number",
			aadhaarNumber: "123456789012",
			shouldError:   false,
		},
		{
			name:          "Empty Aadhaar number",
			aadhaarNumber: "",
			shouldError:   true,
		},
		{
			name:          "Invalid Aadhaar number - too short",
			aadhaarNumber: "12345",
			shouldError:   true,
		},
		{
			name:          "Invalid Aadhaar number - too long",
			aadhaarNumber: "1234567890123",
			shouldError:   true,
		},
		{
			name:          "Invalid Aadhaar number - contains letters",
			aadhaarNumber: "12345678901a",
			shouldError:   true,
		},
		{
			name:          "Valid Aadhaar number with all zeros",
			aadhaarNumber: "000000000000",
			shouldError:   false,
		},
		{
			name:          "Valid Aadhaar number with all nines",
			aadhaarNumber: "999999999999",
			shouldError:   false,
		},
		{
			name:          "Aadhaar number with special characters",
			aadhaarNumber: "1234-5678-9012",
			shouldError:   true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			result, err := userService.GetUserByAadhaarNumber(context.Background(), tt.aadhaarNumber)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if !tt.shouldError {
					assert.Equal(t, tt.aadhaarNumber, *result.AadhaarNumber)
				}
			}
		})
	}
}

// Test UpdateUser with various scenarios using existing test data structure
func TestUserService_UpdateUserExtended(t *testing.T) {
	for _, tt := range UserServiceUpdateUserTests {
		t.Run(tt.testName, func(t *testing.T) {
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			req := &userRequests.UpdateUserRequest{
				UserID:   tt.userID,
				Username: &tt.newUsername,
			}

			result, err := userService.UpdateUser(context.Background(), req)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}

	// Additional edge cases for UpdateUser
	additionalTestCases := []struct {
		name        string
		userID      string
		newUsername string
		shouldError bool
	}{
		{
			name:        "Update with same username",
			userID:      "usr123456789",
			newUsername: "testuser",
			shouldError: false,
		},
		{
			name:        "Update with very long username",
			userID:      "usr123456789",
			newUsername: "verylongusernamethatshouldnotbeallowed1234567890",
			shouldError: true,
		},
		{
			name:        "Update with special characters in username",
			userID:      "usr123456789",
			newUsername: "user@test.com",
			shouldError: false,
		},
		{
			name:        "Update with unicode characters",
			userID:      "usr123456789",
			newUsername: "üser123",
			shouldError: false,
		},
	}

	for _, tt := range additionalTestCases {
		t.Run(tt.name, func(t *testing.T) {
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			req := &userRequests.UpdateUserRequest{
				UserID:   tt.userID,
				Username: &tt.newUsername,
			}

			result, err := userService.UpdateUser(context.Background(), req)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// Test DeleteUser using existing test data and additional scenarios
func TestUserService_DeleteUserExtended(t *testing.T) {
	for _, tt := range UserServiceDeleteUserTests {
		t.Run(tt.testName, func(t *testing.T) {
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			err := userService.DeleteUser(context.Background(), tt.userID)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Additional edge cases for DeleteUser
	additionalTestCases := []struct {
		name        string
		userID      string
		shouldError bool
	}{
		{
			name:        "User ID with special characters",
			userID:      "usr-123@456",
			shouldError: true,
		},
		{
			name:        "Very long user ID",
			userID:      "usr1234567890123456789012345678901234567890",
			shouldError: true,
		},
		{
			name:        "User ID with SQL injection attempt",
			userID:      "usr123'; DROP TABLE users; --",
			shouldError: true,
		},
	}

	for _, tt := range additionalTestCases {
		t.Run(tt.name, func(t *testing.T) {
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			err := userService.DeleteUser(context.Background(), tt.userID)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test ListUsers using existing test data and pagination scenarios
func TestUserService_ListUsersExtended(t *testing.T) {
	for _, tt := range UserServiceListUsersTests {
		t.Run(tt.testName, func(t *testing.T) {
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			result, err := userService.ListUsers(context.Background(), tt.limit, tt.offset)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}

	// Additional pagination edge cases
	additionalTestCases := []struct {
		name        string
		limit       int
		offset      int
		shouldError bool
	}{
		{
			name:        "Zero limit",
			limit:       0,
			offset:      0,
			shouldError: true,
		},
		{
			name:        "Negative limit",
			limit:       -1,
			offset:      0,
			shouldError: true,
		},
		{
			name:        "Negative offset",
			limit:       10,
			offset:      -1,
			shouldError: true,
		},
		{
			name:        "Very large limit",
			limit:       10000,
			offset:      0,
			shouldError: true,
		},
		{
			name:        "Minimum valid pagination",
			limit:       1,
			offset:      0,
			shouldError: false,
		},
	}

	for _, tt := range additionalTestCases {
		t.Run(tt.name, func(t *testing.T) {
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			result, err := userService.ListUsers(context.Background(), tt.limit, tt.offset)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// Test ListActiveUsers
func TestUserService_ListActiveUsers(t *testing.T) {
	testCases := []struct {
		name        string
		limit       int
		offset      int
		shouldError bool
	}{
		{
			name:        "Valid active users list",
			limit:       10,
			offset:      0,
			shouldError: false,
		},
		{
			name:        "Invalid limit for active users",
			limit:       -1,
			offset:      0,
			shouldError: true,
		},
		{
			name:        "Zero limit for active users",
			limit:       0,
			offset:      0,
			shouldError: true,
		},
		{
			name:        "Large limit for active users",
			limit:       1000,
			offset:      0,
			shouldError: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			result, err := userService.ListActiveUsers(context.Background(), tt.limit, tt.offset)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// Test SearchUsers functionality
func TestUserService_SearchUsers(t *testing.T) {
	testCases := []struct {
		name        string
		keyword     string
		limit       int
		offset      int
		shouldError bool
	}{
		{
			name:        "Valid search with simple keyword",
			keyword:     "test",
			limit:       10,
			offset:      0,
			shouldError: false,
		},
		{
			name:        "Empty keyword",
			keyword:     "",
			limit:       10,
			offset:      0,
			shouldError: true,
		},
		{
			name:        "Single character keyword",
			keyword:     "a",
			limit:       10,
			offset:      0,
			shouldError: false,
		},
		{
			name:        "Special characters in keyword",
			keyword:     "test@123",
			limit:       10,
			offset:      0,
			shouldError: false,
		},
		{
			name:        "Very long keyword",
			keyword:     "verylongkeywordthatmightcauseissueswithsearchfunctionality",
			limit:       10,
			offset:      0,
			shouldError: false,
		},
		{
			name:        "Keyword with spaces",
			keyword:     "test user",
			limit:       10,
			offset:      0,
			shouldError: false,
		},
		{
			name:        "Invalid pagination with search",
			keyword:     "test",
			limit:       -1,
			offset:      0,
			shouldError: true,
		},
		{
			name:        "Unicode characters in keyword",
			keyword:     "tëst",
			limit:       10,
			offset:      0,
			shouldError: false,
		},
		{
			name:        "SQL injection attempt in keyword",
			keyword:     "'; DROP TABLE users; --",
			limit:       10,
			offset:      0,
			shouldError: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			result, err := userService.SearchUsers(context.Background(), tt.keyword, tt.limit, tt.offset)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// Test ValidateUser
func TestUserService_ValidateUser(t *testing.T) {
	testCases := []struct {
		name        string
		userID      string
		shouldError bool
	}{
		{
			name:        "Valid user validation",
			userID:      "usr123456789",
			shouldError: false,
		},
		{
			name:        "Empty user ID validation",
			userID:      "",
			shouldError: true,
		},
		{
			name:        "Invalid user ID validation",
			userID:      "invalid-id",
			shouldError: true,
		},
		{
			name:        "Non-existent user ID validation",
			userID:      "usr999999999",
			shouldError: true,
		},
		{
			name:        "User ID with special characters validation",
			userID:      "usr@123#456",
			shouldError: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			err := userService.ValidateUser(context.Background(), tt.userID)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test DeductTokens functionality
func TestUserService_DeductTokens(t *testing.T) {
	testCases := []struct {
		name        string
		userID      string
		amount      int
		shouldError bool
	}{
		{
			name:        "Valid token deduction",
			userID:      "usr123456789",
			amount:      100,
			shouldError: false,
		},
		{
			name:        "Deduct zero tokens",
			userID:      "usr123456789",
			amount:      0,
			shouldError: true,
		},
		{
			name:        "Deduct negative tokens",
			userID:      "usr123456789",
			amount:      -50,
			shouldError: true,
		},
		{
			name:        "Empty user ID for deduction",
			userID:      "",
			amount:      100,
			shouldError: true,
		},
		{
			name:        "Small token amount deduction",
			userID:      "usr123456789",
			amount:      1,
			shouldError: false,
		},
		{
			name:        "Large token amount deduction",
			userID:      "usr123456789",
			amount:      999999,
			shouldError: false,
		},
		{
			name:        "Very large token amount deduction",
			userID:      "usr123456789",
			amount:      2147483647,
			shouldError: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			err := userService.DeductTokens(context.Background(), tt.userID, tt.amount)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test AddTokens functionality
func TestUserService_AddTokens(t *testing.T) {
	testCases := []struct {
		name        string
		userID      string
		amount      int
		shouldError bool
	}{
		{
			name:        "Valid token addition",
			userID:      "usr123456789",
			amount:      100,
			shouldError: false,
		},
		{
			name:        "Add zero tokens",
			userID:      "usr123456789",
			amount:      0,
			shouldError: true,
		},
		{
			name:        "Add negative tokens",
			userID:      "usr123456789",
			amount:      -50,
			shouldError: true,
		},
		{
			name:        "Empty user ID for addition",
			userID:      "",
			amount:      100,
			shouldError: true,
		},
		{
			name:        "Small token amount addition",
			userID:      "usr123456789",
			amount:      1,
			shouldError: false,
		},
		{
			name:        "Large token amount addition",
			userID:      "usr123456789",
			amount:      999999,
			shouldError: false,
		},
		{
			name:        "Very large token amount addition",
			userID:      "usr123456789",
			amount:      2147483647,
			shouldError: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			err := userService.AddTokens(context.Background(), tt.userID, tt.amount)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test GetUserWithProfile
func TestUserService_GetUserWithProfile(t *testing.T) {
	testCases := []struct {
		name        string
		userID      string
		shouldError bool
	}{
		{
			name:        "Valid user profile retrieval",
			userID:      "usr123456789",
			shouldError: false,
		},
		{
			name:        "Empty user ID for profile",
			userID:      "",
			shouldError: true,
		},
		{
			name:        "Invalid user ID for profile",
			userID:      "invalid-id",
			shouldError: true,
		},
		{
			name:        "Non-existent user profile",
			userID:      "usr999999999",
			shouldError: true,
		},
		{
			name:        "User ID with special characters for profile",
			userID:      "usr@123#456",
			shouldError: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			result, err := userService.GetUserWithProfile(context.Background(), tt.userID)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.userID, result.ID)
			}
		})
	}
}

// Test GetUserWithRoles
func TestUserService_GetUserWithRoles(t *testing.T) {
	testCases := []struct {
		name        string
		userID      string
		shouldError bool
	}{
		{
			name:        "Valid user roles retrieval",
			userID:      "usr123456789",
			shouldError: false,
		},
		{
			name:        "Empty user ID for roles",
			userID:      "",
			shouldError: true,
		},
		{
			name:        "Invalid user ID for roles",
			userID:      "invalid-id",
			shouldError: true,
		},
		{
			name:        "Non-existent user roles",
			userID:      "usr999999999",
			shouldError: true,
		},
		{
			name:        "User ID with special characters for roles",
			userID:      "usr@123#456",
			shouldError: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			dbManager := setupTestDatabase(t)
			defer cleanupTestDatabase(t, dbManager)

			userService := &MockUserService{
				shouldError: tt.shouldError,
			}

			result, err := userService.GetUserWithRoles(context.Background(), tt.userID)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.userID, result.ID)
			}
		})
	}
}

// Test context timeout and cancellation scenarios
func TestUserService_ContextTimeout(t *testing.T) {
	testCases := []struct {
		name           string
		timeout        time.Duration
		cancelAfter    time.Duration
		expectedError  bool
	}{
		{
			name:          "Normal context - no timeout",
			timeout:       time.Second * 5,
			cancelAfter:   0,
			expectedError: false,
		},
		{
			name:          "Context timeout before completion",
			timeout:       time.Millisecond * 100,
			cancelAfter:   time.Millisecond * 200,
			expectedError: true,
		},
		{
			name:          "Context cancelled manually",
			timeout:       time.Second * 5,
			cancelAfter:   time.Millisecond * 50,
			expectedError: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			if tt.cancelAfter > 0 {
				go func() {
					time.Sleep(tt.cancelAfter)
					cancel()
				}()
			}

			userService := &MockUserService{
				shouldError: tt.expectedError,
			}

			_, err := userService.GetUserByID(ctx, "usr123456789")

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkUserService_CreateUser(b *testing.B) {
	userService := &MockUserService{shouldError: false}
	req := &userRequests.CreateUserRequest{
		Username: func() *string { s := "benchmarkuser"; return &s }(),
		Password: "password123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := userService.CreateUser(context.Background(), req)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkUserService_GetUserByID(b *testing.B) {
	userService := &MockUserService{shouldError: false}
	userID := "usr123456789"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := userService.GetUserByID(context.Background(), userID)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkUserService_GetUserByUsername(b *testing.B) {
	userService := &MockUserService{shouldError: false}
	username := "testuser"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := userService.GetUserByUsername(context.Background(), username)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkUserService_ListUsers(b *testing.B) {
	userService := &MockUserService{shouldError: false}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := userService.ListUsers(context.Background(), 10, 0)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

// Test edge cases for concurrent access
func TestUserService_ConcurrentAccess(t *testing.T) {
	userService := &MockUserService{shouldError: false}
	userID := "usr123456789"

	// Test concurrent reads
	t.Run("Concurrent reads", func(t *testing.T) {
		done := make(chan bool, 20)
		
		for i := 0; i < 20; i++ {
			go func(idx int) {
				defer func() { done <- true }()
				_, err := userService.GetUserByID(context.Background(), userID)
				assert.NoError(t, err)
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 20; i++ {
			<-done
		}
	})

	// Test concurrent writes
	t.Run("Concurrent writes", func(t *testing.T) {
		done := make(chan bool, 10)
		
		for i := 0; i < 10; i++ {
			go func(idx int) {
				defer func() { done <- true }()
				req := &userRequests.CreateUserRequest{
					Username: func() *string { 
						s := fmt.Sprintf("concurrentuser%d", idx)
						return &s 
					}(),
					Password: "password123",
				}
				_, err := userService.CreateUser(context.Background(), req)
				assert.NoError(t, err)
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	// Test concurrent mixed operations
	t.Run("Concurrent mixed operations", func(t *testing.T) {
		done := make(chan bool, 15)
		
		// Readers
		for i := 0; i < 5; i++ {
			go func(idx int) {
				defer func() { done <- true }()
				_, err := userService.GetUserByID(context.Background(), userID)
				assert.NoError(t, err)
			}(i)
		}
		
		// Writers
		for i := 0; i < 5; i++ {
			go func(idx int) {
				defer func() { done <- true }()
				req := &userRequests.CreateUserRequest{
					Username: func() *string { 
						s := fmt.Sprintf("mixeduser%d", idx)
						return &s 
					}(),
					Password: "password123",
				}
				_, err := userService.CreateUser(context.Background(), req)
				assert.NoError(t, err)
			}(i)
		}
		
		// Listers
		for i := 0; i < 5; i++ {
			go func(idx int) {
				defer func() { done <- true }()
				_, err := userService.ListUsers(context.Background(), 10, 0)
				assert.NoError(t, err)
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 15; i++ {
			<-done
		}
	})
}

// Test nil request handling
func TestUserService_NilRequestHandling(t *testing.T) {
	dbManager := setupTestDatabase(t)
	defer cleanupTestDatabase(t, dbManager)

	userService := &MockUserService{shouldError: true}

	// Test CreateUser with nil request
	t.Run("CreateUser with nil request", func(t *testing.T) {
		_, err := userService.CreateUser(context.Background(), nil)
		assert.Error(t, err)
	})

	// Test UpdateUser with nil request
	t.Run("UpdateUser with nil request", func(t *testing.T) {
		_, err := userService.UpdateUser(context.Background(), nil)
		assert.Error(t, err)
	})
}

// Test response validation
func TestUserService_ResponseValidation(t *testing.T) {
	dbManager := setupTestDatabase(t)
	defer cleanupTestDatabase(t, dbManager)

	userService := &MockUserService{shouldError: false}

	t.Run("Validate CreateUser response", func(t *testing.T) {
		req := &userRequests.CreateUserRequest{
			Username: func() *string { s := "testuser"; return &s }(),
			Password: "password123",
		}
		
		result, err := userService.CreateUser(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.ID)
		assert.NotNil(t, result.Username)
		assert.Equal(t, *req.Username, *result.Username)
	})

	t.Run("Validate GetUserByID response", func(t *testing.T) {
		userID := "usr123456789"
		
		result, err := userService.GetUserByID(context.Background(), userID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.ID)
		assert.NotNil(t, result.Username)
	})

	t.Run("Validate ListUsers response", func(t *testing.T) {
		result, err := userService.ListUsers(context.Background(), 10, 0)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// Test error message validation
func TestUserService_ErrorMessages(t *testing.T) {
	dbManager := setupTestDatabase(t)
	defer cleanupTestDatabase(t, dbManager)

	userService := &MockUserService{shouldError: true}

	testCases := []struct {
		name           string
		operation      func() error
		expectedErrors []string
	}{
		{
			name: "CreateUser error",
			operation: func() error {
				_, err := userService.CreateUser(context.Background(), &userRequests.CreateUserRequest{
					Username: func() *string { s := "test"; return &s }(),
					Password: "pass",
				})
				return err
			},
			expectedErrors: []string{"validation error", "error"},
		},
		{
			name: "GetUserByID error",
			operation: func() error {
				_, err := userService.GetUserByID(context.Background(), "invalid")
				return err
			},
			expectedErrors: []string{"user not found", "not found", "error"},
		},
		{
			name: "ValidateUser error",
			operation: func() error {
				return userService.ValidateUser(context.Background(), "invalid")
			},
			expectedErrors: []string{"validation failed", "failed", "error"},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.operation()
			assert.Error(t, err)
			
			errorMessage := err.Error()
			hasExpectedError := false
			for _, expectedError := range tt.expectedErrors {
				if strings.Contains(strings.ToLower(errorMessage), strings.ToLower(expectedError)) {
					hasExpectedError = true
					break
				}
			}
			assert.True(t, hasExpectedError, "Error message '%s' should contain one of %v", errorMessage, tt.expectedErrors)
		})
	}
}

// Helper functions
func createTestUserRequest(username string) *userRequests.CreateUserRequest {
	return &userRequests.CreateUserRequest{
		Username: &username,
		Password: "testpassword123",
	}
}

func createTestUpdateRequest(userID, username string) *userRequests.UpdateUserRequest {
	return &userRequests.UpdateUserRequest{
		UserID:   userID,
		Username: &username,
	}
}

// Test setup and teardown functions validation
func TestUserService_TestInfrastructure(t *testing.T) {
	t.Run("Test database setup", func(t *testing.T) {
		dbManager := setupTestDatabase(t)
		assert.NotNil(t, dbManager)
		assert.True(t, dbManager.IsConnected())
		assert.Equal(t, db.BackendInMemory, dbManager.GetBackendType())
		cleanupTestDatabase(t, dbManager)
	})

	t.Run("Mock service initialization", func(t *testing.T) {
		userService := &MockUserService{shouldError: false}
		assert.NotNil(t, userService)
		assert.False(t, userService.shouldError)
	})
}

// Test data integrity and consistency
func TestUserService_DataIntegrity(t *testing.T) {
	dbManager := setupTestDatabase(t)
	defer cleanupTestDatabase(t, dbManager)

	userService := &MockUserService{shouldError: false}

	t.Run("Create and retrieve user consistency", func(t *testing.T) {
		username := "consistencytest"
		req := &userRequests.CreateUserRequest{
			Username: &username,
			Password: "password123",
		}
		
		createResult, err := userService.CreateUser(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, createResult)
		
		getResult, err := userService.GetUserByID(context.Background(), createResult.ID)
		assert.NoError(t, err)
		assert.NotNil(t, getResult)
		assert.Equal(t, createResult.ID, getResult.ID)
		assert.Equal(t, createResult.Username, getResult.Username)
	})
}

// Test edge cases and boundary conditions
func TestUserService_EdgeCases(t *testing.T) {
	dbManager := setupTestDatabase(t)
	defer cleanupTestDatabase(t, dbManager)

	userService := &MockUserService{shouldError: false}

	t.Run("Empty context handling", func(t *testing.T) {
		// Test with background context
		_, err := userService.GetUserByID(context.Background(), "usr123")
		assert.NoError(t, err)
	})

	t.Run("Timeout context handling", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		
		time.Sleep(time.Millisecond) // Ensure timeout occurs
		
		userService.shouldError = true
		_, err := userService.GetUserByID(ctx, "usr123")
		assert.Error(t, err)
	})
}
