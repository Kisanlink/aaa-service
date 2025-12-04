package models

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// SMSDeliveryLog stores audit records for all SMS delivery attempts
type SMSDeliveryLog struct {
	*base.BaseModel
	UserID            *string                `json:"user_id" gorm:"type:varchar(255);index"`
	PhoneNumberMasked string                 `json:"phone_number_masked" gorm:"type:varchar(20);not null"`
	MessageType       string                 `json:"message_type" gorm:"type:varchar(50);not null"`
	SNSMessageID      *string                `json:"sns_message_id" gorm:"type:varchar(100)"`
	Status            string                 `json:"status" gorm:"type:varchar(20);not null;default:pending"`
	FailureReason     *string                `json:"failure_reason" gorm:"type:text"`
	SentAt            time.Time              `json:"sent_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	DeliveredAt       *time.Time             `json:"delivered_at"`
	RequestDetails    map[string]interface{} `json:"request_details" gorm:"type:jsonb"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

// SMS delivery status constants
const (
	SMSStatusPending   = "pending"
	SMSStatusSent      = "sent"
	SMSStatusDelivered = "delivered"
	SMSStatusFailed    = "failed"
)

// SMS message type constants
const (
	SMSTypePasswordResetOTP = "password_reset_otp"
	SMSTypeSecurityAlert    = "security_alert"
	SMSTypeBulk             = "bulk"
	SMSTypeVerification     = "verification"
)

// NewSMSDeliveryLog creates a new SMSDeliveryLog instance
func NewSMSDeliveryLog(messageType, maskedPhone string) *SMSDeliveryLog {
	return &SMSDeliveryLog{
		BaseModel:         base.NewBaseModel("SMSLOG", hash.Medium),
		PhoneNumberMasked: maskedPhone,
		MessageType:       messageType,
		Status:            SMSStatusPending,
		SentAt:            time.Now(),
		RequestDetails:    make(map[string]interface{}),
	}
}

// NewSMSDeliveryLogWithUser creates a new SMSDeliveryLog with user association
func NewSMSDeliveryLogWithUser(userID, messageType, maskedPhone string) *SMSDeliveryLog {
	log := NewSMSDeliveryLog(messageType, maskedPhone)
	log.UserID = &userID
	return log
}

// TableName specifies the table name for SMSDeliveryLog
func (s *SMSDeliveryLog) TableName() string {
	return "sms_delivery_logs"
}

// GetTableIdentifier returns the table identifier for SMSDeliveryLog
func (s *SMSDeliveryLog) GetTableIdentifier() string {
	return "SMSLOG"
}

// GetTableSize returns the table size for SMSDeliveryLog
func (s *SMSDeliveryLog) GetTableSize() hash.TableSize {
	return hash.Medium
}

// BeforeCreate is called before creating a new SMS delivery log
func (s *SMSDeliveryLog) BeforeCreate() error {
	if s.SentAt.IsZero() {
		s.SentAt = time.Now()
	}
	if s.RequestDetails == nil {
		s.RequestDetails = make(map[string]interface{})
	}
	return s.BaseModel.BeforeCreate()
}

// BeforeUpdate is called before updating an SMS delivery log
func (s *SMSDeliveryLog) BeforeUpdate() error {
	return s.BaseModel.BeforeUpdate()
}

// BeforeDelete is called before deleting an SMS delivery log
func (s *SMSDeliveryLog) BeforeDelete() error {
	return s.BaseModel.BeforeDelete()
}

// BeforeSoftDelete is called before soft deleting an SMS delivery log
func (s *SMSDeliveryLog) BeforeSoftDelete() error {
	return s.BaseModel.BeforeSoftDelete()
}

// GORM Hooks
func (s *SMSDeliveryLog) BeforeCreateGORM(tx *gorm.DB) error {
	return s.BeforeCreate()
}

func (s *SMSDeliveryLog) BeforeUpdateGORM(tx *gorm.DB) error {
	return s.BeforeUpdate()
}

func (s *SMSDeliveryLog) BeforeDeleteGORM(tx *gorm.DB) error {
	return s.BeforeDelete()
}

// MarkAsSent marks the SMS as successfully sent
func (s *SMSDeliveryLog) MarkAsSent(snsMessageID string) {
	s.Status = SMSStatusSent
	s.SNSMessageID = &snsMessageID
}

// MarkAsDelivered marks the SMS as delivered
func (s *SMSDeliveryLog) MarkAsDelivered() {
	s.Status = SMSStatusDelivered
	now := time.Now()
	s.DeliveredAt = &now
}

// MarkAsFailed marks the SMS as failed with a reason
func (s *SMSDeliveryLog) MarkAsFailed(reason string) {
	s.Status = SMSStatusFailed
	s.FailureReason = &reason
}

// IsSent checks if the SMS was sent
func (s *SMSDeliveryLog) IsSent() bool {
	return s.Status == SMSStatusSent || s.Status == SMSStatusDelivered
}

// IsFailed checks if the SMS failed
func (s *SMSDeliveryLog) IsFailed() bool {
	return s.Status == SMSStatusFailed
}

// IsPending checks if the SMS is pending
func (s *SMSDeliveryLog) IsPending() bool {
	return s.Status == SMSStatusPending
}

// AddDetail adds a key-value pair to the request details
func (s *SMSDeliveryLog) AddDetail(key string, value interface{}) {
	if s.RequestDetails == nil {
		s.RequestDetails = make(map[string]interface{})
	}
	s.RequestDetails[key] = value
}

// GetDetail retrieves a value from the request details
func (s *SMSDeliveryLog) GetDetail(key string) (interface{}, bool) {
	if s.RequestDetails == nil {
		return nil, false
	}
	value, exists := s.RequestDetails[key]
	return value, exists
}

// MaskPhoneNumber masks a phone number for storage
// e.g., "9876543210" -> "XXXX-XXX-3210"
func MaskPhoneNumber(phone string) string {
	if len(phone) < 4 {
		return "****"
	}
	return "XXXX-XXX-" + phone[len(phone)-4:]
}

// MaskPhoneNumberE164 masks an E.164 formatted phone number
// e.g., "+919876543210" -> "+91-XXXX-XXX-3210"
func MaskPhoneNumberE164(phone string) string {
	if len(phone) < 5 {
		return "****"
	}
	// For simplicity, assume country code is 2-3 digits after +
	if len(phone) > 4 {
		countryCode := phone[:3] // e.g., "+91"
		lastFour := phone[len(phone)-4:]
		return countryCode + "-XXXX-XXX-" + lastFour
	}
	return "****"
}
