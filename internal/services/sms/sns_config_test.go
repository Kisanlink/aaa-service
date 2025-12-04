package sms

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSNSConfig(t *testing.T) {
	config := DefaultSNSConfig()

	assert.Equal(t, "ap-south-1", config.Region)
	assert.False(t, config.Enabled, "SMS should be disabled by default for safety")
	assert.Equal(t, "KISANLINK", config.SenderID)
	assert.Equal(t, SNSMessageTypeTransactional, config.MessageType)
	assert.Equal(t, 10*time.Minute, config.OTPExpiry)
	assert.Equal(t, 5, config.MaxSMSPerHour)
	assert.Equal(t, 20, config.MaxSMSPerDay)
}

func TestSNSConfig_Validate(t *testing.T) {
	t.Run("sets defaults for empty values", func(t *testing.T) {
		config := &SNSConfig{}
		err := config.Validate()

		assert.NoError(t, err)
		assert.Equal(t, "ap-south-1", config.Region)
		assert.Equal(t, "KISANLINK", config.SenderID)
		assert.Equal(t, SNSMessageTypeTransactional, config.MessageType)
		assert.Equal(t, 10*time.Minute, config.OTPExpiry)
		assert.Equal(t, 5, config.MaxSMSPerHour)
		assert.Equal(t, 20, config.MaxSMSPerDay)
	})

	t.Run("truncates sender ID if too long", func(t *testing.T) {
		config := &SNSConfig{
			SenderID: "VERYLONGSENDERID",
		}
		err := config.Validate()

		assert.NoError(t, err)
		assert.Len(t, config.SenderID, 11)
		assert.Equal(t, "VERYLONGSEN", config.SenderID)
	})

	t.Run("preserves valid values", func(t *testing.T) {
		config := &SNSConfig{
			Region:        "us-east-1",
			Enabled:       true,
			SenderID:      "MYAPP",
			MessageType:   SNSMessageTypePromotional,
			OTPExpiry:     5 * time.Minute,
			MaxSMSPerHour: 10,
			MaxSMSPerDay:  50,
		}
		err := config.Validate()

		assert.NoError(t, err)
		assert.Equal(t, "us-east-1", config.Region)
		assert.True(t, config.Enabled)
		assert.Equal(t, "MYAPP", config.SenderID)
		assert.Equal(t, SNSMessageTypePromotional, config.MessageType)
		assert.Equal(t, 5*time.Minute, config.OTPExpiry)
		assert.Equal(t, 10, config.MaxSMSPerHour)
		assert.Equal(t, 50, config.MaxSMSPerDay)
	})
}

func TestSNSConfig_IsTransactional(t *testing.T) {
	t.Run("returns true for transactional", func(t *testing.T) {
		config := &SNSConfig{MessageType: SNSMessageTypeTransactional}
		assert.True(t, config.IsTransactional())
	})

	t.Run("returns false for promotional", func(t *testing.T) {
		config := &SNSConfig{MessageType: SNSMessageTypePromotional}
		assert.False(t, config.IsTransactional())
	})
}
