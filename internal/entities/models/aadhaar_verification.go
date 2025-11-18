package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// AadhaarVerification represents an Aadhaar verification record for KYC purposes.
// It stores the complete lifecycle of an Aadhaar OTP verification flow including
// OTP generation, verification status, and extracted KYC data.
type AadhaarVerification struct {
	ID                 string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	UserID             string         `gorm:"type:varchar(255);not null;index" json:"user_id"`
	AadhaarNumber      string         `gorm:"type:varchar(12)" json:"aadhaar_number,omitempty"`
	TransactionID      string         `gorm:"type:varchar(255);unique" json:"transaction_id"`
	ReferenceID        string         `gorm:"type:varchar(255);unique" json:"reference_id"`
	OTPRequestedAt     *time.Time     `json:"otp_requested_at,omitempty"`
	OTPVerifiedAt      *time.Time     `json:"otp_verified_at,omitempty"`
	VerificationStatus string         `gorm:"type:varchar(50);default:'PENDING'" json:"verification_status"`
	KYCStatus          string         `gorm:"type:varchar(50);default:'PENDING'" json:"kyc_status"`
	PhotoURL           string         `gorm:"type:text" json:"photo_url,omitempty"`
	Name               string         `gorm:"type:varchar(255)" json:"name,omitempty"`
	DateOfBirth        *time.Time     `json:"date_of_birth,omitempty"`
	Gender             string         `gorm:"type:varchar(20)" json:"gender,omitempty"`
	FullAddress        string         `gorm:"type:text" json:"full_address,omitempty"`
	AddressJSON        AadhaarAddress `gorm:"type:jsonb" json:"address,omitempty"`
	Attempts           int            `gorm:"default:0" json:"attempts"`
	LastAttemptAt      *time.Time     `json:"last_attempt_at,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          *time.Time     `gorm:"index" json:"deleted_at,omitempty"`
	CreatedBy          string         `gorm:"type:varchar(255)" json:"created_by,omitempty"`
	UpdatedBy          string         `gorm:"type:varchar(255)" json:"updated_by,omitempty"`

	// Relationships
	User        *User        `gorm:"foreignKey:UserID" json:"-"`
	OTPAttempts []OTPAttempt `gorm:"foreignKey:AadhaarVerificationID" json:"otp_attempts,omitempty"`
}

// AadhaarAddress represents the structured address data from Aadhaar verification.
// This struct is stored as JSONB in the database and requires custom marshaling.
type AadhaarAddress struct {
	House    string `json:"house"`
	Street   string `json:"street"`
	Landmark string `json:"landmark"`
	District string `json:"district"`
	State    string `json:"state"`
	Pincode  int    `json:"pincode"`
	Country  string `json:"country"`
}

// Value implements the driver.Valuer interface for GORM JSONB support.
// It marshals the AadhaarAddress struct to JSON bytes for database storage.
func (a AadhaarAddress) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface for GORM JSONB support.
// It unmarshals JSON bytes from the database into the AadhaarAddress struct.
func (a *AadhaarAddress) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, a)
}

// TableName specifies the table name for the AadhaarVerification model.
func (AadhaarVerification) TableName() string {
	return "aadhaar_verifications"
}
