package sms

import "time"

// SNSConfig holds AWS SNS configuration for SMS delivery
type SNSConfig struct {
	// Region is the AWS region for SNS (e.g., "ap-south-1")
	Region string

	// Enabled is the feature flag to enable/disable SMS
	Enabled bool

	// SenderID is the SMS sender ID (max 11 characters, e.g., "KISANLINK")
	// Note: SenderID support varies by country
	SenderID string

	// MessageType is the SNS SMS type: "Transactional" or "Promotional"
	// Transactional: Higher delivery priority, used for OTPs
	// Promotional: Lower cost, subject to DND restrictions
	MessageType string

	// OTPExpiry is the validity duration for OTPs
	OTPExpiry time.Duration

	// MaxSMSPerHour is the rate limit per phone number per hour
	MaxSMSPerHour int

	// MaxSMSPerDay is the rate limit per phone number per day
	MaxSMSPerDay int
}

// SNS message type constants
const (
	SNSMessageTypeTransactional = "Transactional"
	SNSMessageTypePromotional   = "Promotional"
)

// DefaultSNSConfig returns the default configuration for SNS SMS
func DefaultSNSConfig() *SNSConfig {
	return &SNSConfig{
		Region:        "ap-south-1",
		Enabled:       false, // Disabled by default for safety
		SenderID:      "KISANLINK",
		MessageType:   SNSMessageTypeTransactional,
		OTPExpiry:     10 * time.Minute,
		MaxSMSPerHour: 5,
		MaxSMSPerDay:  20,
	}
}

// Validate validates the SNS configuration
func (c *SNSConfig) Validate() error {
	if c.Region == "" {
		c.Region = "ap-south-1"
	}
	if c.SenderID == "" {
		c.SenderID = "KISANLINK"
	}
	if len(c.SenderID) > 11 {
		c.SenderID = c.SenderID[:11]
	}
	if c.MessageType == "" {
		c.MessageType = SNSMessageTypeTransactional
	}
	if c.OTPExpiry == 0 {
		c.OTPExpiry = 10 * time.Minute
	}
	if c.MaxSMSPerHour == 0 {
		c.MaxSMSPerHour = 5
	}
	if c.MaxSMSPerDay == 0 {
		c.MaxSMSPerDay = 20
	}
	return nil
}

// IsTransactional returns true if the message type is transactional
func (c *SNSConfig) IsTransactional() bool {
	return c.MessageType == SNSMessageTypeTransactional
}
