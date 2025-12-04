package sms

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// OTP length constants
const (
	DefaultOTPLength = 6
	MinOTPLength     = 4
	MaxOTPLength     = 8
)

// GenerateNumericOTP generates a cryptographically secure numeric OTP of the specified length
func GenerateNumericOTP(length int) (string, error) {
	if length < MinOTPLength {
		length = MinOTPLength
	}
	if length > MaxOTPLength {
		length = MaxOTPLength
	}

	// Calculate the range: e.g., for 6 digits: min=100000, max=999999
	min := int64(1)
	for i := 1; i < length; i++ {
		min *= 10
	}
	max := min*10 - 1

	// Generate random number in range [min, max]
	rangeSize := big.NewInt(max - min + 1)
	n, err := rand.Int(rand.Reader, rangeSize)
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	otp := n.Int64() + min
	format := fmt.Sprintf("%%0%dd", length)
	return fmt.Sprintf(format, otp), nil
}

// GenerateOTP generates a 6-digit OTP (convenience function)
func GenerateOTP() (string, error) {
	return GenerateNumericOTP(DefaultOTPLength)
}

// GenerateAlphanumericOTP generates a cryptographically secure alphanumeric OTP
// Useful for scenarios where alphanumeric codes are preferred
func GenerateAlphanumericOTP(length int) (string, error) {
	if length < MinOTPLength {
		length = MinOTPLength
	}
	if length > MaxOTPLength {
		length = MaxOTPLength
	}

	// Character set: digits + uppercase letters (excluding ambiguous characters I, O)
	const charset = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	charsetLen := big.NewInt(int64(len(charset)))

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate alphanumeric OTP: %w", err)
		}
		result[i] = charset[n.Int64()]
	}

	return string(result), nil
}
