package sms

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateOTP(t *testing.T) {
	t.Run("generates 6-digit OTP by default", func(t *testing.T) {
		otp, err := GenerateOTP()
		require.NoError(t, err)
		assert.Len(t, otp, 6)

		// Verify all characters are digits
		for _, c := range otp {
			assert.True(t, c >= '0' && c <= '9', "OTP should only contain digits")
		}
	})

	t.Run("generates unique OTPs", func(t *testing.T) {
		otps := make(map[string]bool)
		for i := 0; i < 100; i++ {
			otp, err := GenerateOTP()
			require.NoError(t, err)
			otps[otp] = true
		}
		// With 100 random 6-digit numbers, we should have at least 95 unique values
		assert.Greater(t, len(otps), 90, "OTPs should be unique")
	})
}

func TestGenerateNumericOTP(t *testing.T) {
	tests := []struct {
		name           string
		length         int
		expectedLength int
	}{
		{"4-digit OTP", 4, 4},
		{"5-digit OTP", 5, 5},
		{"6-digit OTP", 6, 6},
		{"7-digit OTP", 7, 7},
		{"8-digit OTP", 8, 8},
		{"below minimum defaults to 4", 2, 4},
		{"above maximum defaults to 8", 10, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otp, err := GenerateNumericOTP(tt.length)
			require.NoError(t, err)
			assert.Len(t, otp, tt.expectedLength)

			// Verify all characters are digits
			for _, c := range otp {
				assert.True(t, c >= '0' && c <= '9', "OTP should only contain digits")
			}
		})
	}
}

func TestGenerateNumericOTP_NoLeadingZeros(t *testing.T) {
	// Generate many OTPs and verify they don't start with leading zeros
	// (since we generate from min to max range)
	for i := 0; i < 100; i++ {
		otp, err := GenerateNumericOTP(6)
		require.NoError(t, err)
		assert.NotEqual(t, '0', otp[0], "OTP should not have leading zeros")
	}
}

func TestGenerateAlphanumericOTP(t *testing.T) {
	t.Run("generates alphanumeric OTP of correct length", func(t *testing.T) {
		otp, err := GenerateAlphanumericOTP(6)
		require.NoError(t, err)
		assert.Len(t, otp, 6)
	})

	t.Run("excludes ambiguous characters", func(t *testing.T) {
		// Generate many OTPs and check none contain I or O
		for i := 0; i < 100; i++ {
			otp, err := GenerateAlphanumericOTP(8)
			require.NoError(t, err)
			assert.NotContains(t, otp, "I", "OTP should not contain ambiguous character I")
			assert.NotContains(t, otp, "O", "OTP should not contain ambiguous character O")
		}
	})

	t.Run("contains only valid characters", func(t *testing.T) {
		validChars := "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"
		otp, err := GenerateAlphanumericOTP(8)
		require.NoError(t, err)

		for _, c := range otp {
			found := false
			for _, valid := range validChars {
				if c == valid {
					found = true
					break
				}
			}
			assert.True(t, found, "Character %c should be in valid character set", c)
		}
	})

	t.Run("respects length limits", func(t *testing.T) {
		// Below minimum
		otp, err := GenerateAlphanumericOTP(2)
		require.NoError(t, err)
		assert.Len(t, otp, MinOTPLength)

		// Above maximum
		otp, err = GenerateAlphanumericOTP(20)
		require.NoError(t, err)
		assert.Len(t, otp, MaxOTPLength)
	})
}

func BenchmarkGenerateOTP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateOTP()
	}
}

func BenchmarkGenerateNumericOTP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateNumericOTP(6)
	}
}

func BenchmarkGenerateAlphanumericOTP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateAlphanumericOTP(8)
	}
}
