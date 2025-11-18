package models

import "time"

// OTPAttempt represents a single OTP verification attempt for Aadhaar KYC.
// This table is used for audit logging, security monitoring, and tracking
// failed verification attempts to prevent brute force attacks.
type OTPAttempt struct {
	ID                    string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	AadhaarVerificationID string    `gorm:"type:varchar(255);not null;index" json:"aadhaar_verification_id"`
	AttemptNumber         int       `gorm:"not null" json:"attempt_number"`
	OTPValue              string    `gorm:"type:varchar(255)" json:"-"` // Hashed, never expose in JSON
	IPAddress             string    `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent             string    `gorm:"type:text" json:"user_agent,omitempty"`
	Status                string    `gorm:"type:varchar(50)" json:"status"` // SUCCESS, FAILED, EXPIRED
	FailedReason          string    `gorm:"type:varchar(255)" json:"failed_reason,omitempty"`
	CreatedAt             time.Time `json:"created_at"`

	// Relationships
	AadhaarVerification *AadhaarVerification `gorm:"foreignKey:AadhaarVerificationID" json:"-"`
}

// TableName specifies the table name for the OTPAttempt model.
func (OTPAttempt) TableName() string {
	return "otp_attempts"
}
