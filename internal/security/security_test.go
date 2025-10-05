package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityUtils_HashPassword(t *testing.T) {
	su := NewSecurityUtils(12, 12)

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "ValidPassword123!",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := su.HashPassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.NotEqual(t, tt.password, hash)
			}
		})
	}
}

func TestSecurityUtils_VerifyPassword(t *testing.T) {
	su := NewSecurityUtils(12, 12)
	password := "TestPassword123!"

	hash, err := su.HashPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			wantErr:  false,
		},
		{
			name:     "incorrect password",
			password: "WrongPassword",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "empty hash",
			password: password,
			hash:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := su.VerifyPassword(tt.password, tt.hash)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityUtils_HashMPin(t *testing.T) {
	su := NewSecurityUtils(12, 12)

	tests := []struct {
		name    string
		mpin    string
		wantErr bool
	}{
		{
			name:    "valid 4-digit MPIN",
			mpin:    "1234",
			wantErr: false,
		},
		{
			name:    "valid 6-digit MPIN",
			mpin:    "123456",
			wantErr: false,
		},
		{
			name:    "empty MPIN",
			mpin:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := su.HashMPin(tt.mpin)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.NotEqual(t, tt.mpin, hash)
			}
		})
	}
}

func TestSecurityUtils_GenerateSecureToken(t *testing.T) {
	su := NewSecurityUtils(12, 12)

	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{
			name:    "valid length",
			length:  32,
			wantErr: false,
		},
		{
			name:    "zero length",
			length:  0,
			wantErr: true,
		},
		{
			name:    "negative length",
			length:  -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := su.GenerateSecureToken(tt.length)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestSecurityUtils_GenerateSecureNumericCode(t *testing.T) {
	su := NewSecurityUtils(12, 12)

	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{
			name:    "valid length 4",
			length:  4,
			wantErr: false,
		},
		{
			name:    "valid length 6",
			length:  6,
			wantErr: false,
		},
		{
			name:    "zero length",
			length:  0,
			wantErr: true,
		},
		{
			name:    "too long",
			length:  11,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := su.GenerateSecureNumericCode(tt.length)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, code)
			} else {
				assert.NoError(t, err)
				assert.Len(t, code, tt.length)
				// Verify all characters are digits
				for _, char := range code {
					assert.True(t, char >= '0' && char <= '9')
				}
			}
		})
	}
}

func TestSecurityUtils_AssessPasswordStrength(t *testing.T) {
	su := NewSecurityUtils(12, 12)

	tests := []struct {
		name     string
		password string
		expected PasswordStrength
	}{
		{
			name:     "very strong password",
			password: "MyVeryStr0ng&SecureP@ssw0rd!",
			expected: VeryStrongPassword,
		},
		{
			name:     "strong password",
			password: "StrongP@ssw0rd123",
			expected: StrongPassword,
		},
		{
			name:     "medium password",
			password: "Password123",
			expected: WeakPassword, // This password contains "password" which is penalized
		},
		{
			name:     "weak password",
			password: "password",
			expected: WeakPassword,
		},
		{
			name:     "very weak password",
			password: "123",
			expected: WeakPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strength := su.AssessPasswordStrength(tt.password)
			assert.Equal(t, tt.expected, strength)
		})
	}
}

func TestSecurityUtils_ValidateMPinStrength(t *testing.T) {
	su := NewSecurityUtils(12, 12)

	tests := []struct {
		name    string
		mpin    string
		wantErr bool
	}{
		{
			name:    "valid 4-digit MPIN",
			mpin:    "1357",
			wantErr: false,
		},
		{
			name:    "valid 6-digit MPIN",
			mpin:    "135792",
			wantErr: false,
		},
		{
			name:    "weak MPIN - all same digits",
			mpin:    "1111",
			wantErr: true,
		},
		{
			name:    "weak MPIN - sequential",
			mpin:    "1234",
			wantErr: true,
		},
		{
			name:    "weak MPIN - reverse sequential",
			mpin:    "4321",
			wantErr: true,
		},
		{
			name:    "invalid length",
			mpin:    "12345",
			wantErr: true,
		},
		{
			name:    "non-numeric",
			mpin:    "12ab",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := su.ValidateMPinStrength(tt.mpin)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityUtils_SanitizeInput(t *testing.T) {
	su := NewSecurityUtils(12, 12)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal input",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "input with null bytes",
			input:    "Hello\x00World",
			expected: "HelloWorld",
		},
		{
			name:     "input with control characters",
			input:    "Hello\x01\x02World",
			expected: "HelloWorld",
		},
		{
			name:     "input with allowed whitespace",
			input:    "Hello\t\n\rWorld",
			expected: "Hello\t\n\rWorld",
		},
		{
			name:     "input with extra whitespace",
			input:    "  Hello   World  ",
			expected: "Hello   World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := su.SanitizeInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityUtils_IsValidEmail(t *testing.T) {
	su := NewSecurityUtils(12, 12)

	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{
			name:     "valid email",
			email:    "user@example.com",
			expected: true,
		},
		{
			name:     "valid email with subdomain",
			email:    "user@mail.example.com",
			expected: true,
		},
		{
			name:     "invalid email - no @",
			email:    "userexample.com",
			expected: false,
		},
		{
			name:     "invalid email - no domain",
			email:    "user@",
			expected: false,
		},
		{
			name:     "invalid email - no TLD",
			email:    "user@example",
			expected: false,
		},
		{
			name:     "empty email",
			email:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := su.IsValidEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityUtils_IsValidPhoneNumber(t *testing.T) {
	su := NewSecurityUtils(12, 12)

	tests := []struct {
		name     string
		phone    string
		expected bool
	}{
		{
			name:     "valid Indian mobile",
			phone:    "9876543210",
			expected: true,
		},
		{
			name:     "valid Indian mobile with formatting",
			phone:    "+91-9876543210",
			expected: true,
		},
		{
			name:     "invalid Indian mobile - starts with 5",
			phone:    "5876543210",
			expected: false,
		},
		{
			name:     "valid international number",
			phone:    "1234567890123",
			expected: true,
		},
		{
			name:     "invalid - too short",
			phone:    "123456789",
			expected: false,
		},
		{
			name:     "invalid - too long",
			phone:    "1234567890123456",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := su.IsValidPhoneNumber(tt.phone)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityUtils_SecureCompare(t *testing.T) {
	su := NewSecurityUtils(12, 12)

	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{
			name:     "identical strings",
			a:        "hello",
			b:        "hello",
			expected: true,
		},
		{
			name:     "different strings",
			a:        "hello",
			b:        "world",
			expected: false,
		},
		{
			name:     "empty strings",
			a:        "",
			b:        "",
			expected: true,
		},
		{
			name:     "one empty string",
			a:        "hello",
			b:        "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := su.SecureCompare(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPasswordStrengthString(t *testing.T) {
	tests := []struct {
		name     string
		strength PasswordStrength
		expected string
	}{
		{
			name:     "weak password",
			strength: WeakPassword,
			expected: "Weak",
		},
		{
			name:     "medium password",
			strength: MediumPassword,
			expected: "Medium",
		},
		{
			name:     "strong password",
			strength: StrongPassword,
			expected: "Strong",
		},
		{
			name:     "very strong password",
			strength: VeryStrongPassword,
			expected: "Very Strong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPasswordStrengthString(tt.strength)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests for performance-critical functions
func BenchmarkHashPassword(b *testing.B) {
	su := NewSecurityUtils(12, 12)
	password := "TestPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = su.HashPassword(password)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	su := NewSecurityUtils(12, 12)
	password := "TestPassword123!"
	hash, _ := su.HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = su.VerifyPassword(password, hash)
	}
}

func BenchmarkGenerateSecureToken(b *testing.B) {
	su := NewSecurityUtils(12, 12)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = su.GenerateSecureToken(32)
	}
}

func BenchmarkSanitizeInput(b *testing.B) {
	su := NewSecurityUtils(12, 12)
	input := "Hello\x00\x01World with some\ttabs and\nnewlines"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = su.SanitizeInput(input)
	}
}
