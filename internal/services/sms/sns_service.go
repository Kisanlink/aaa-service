package sms

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"go.uber.org/zap"
)

// SMSDeliveryRepository defines the interface for SMS delivery logging
type SMSDeliveryRepository interface {
	Create(ctx context.Context, log *models.SMSDeliveryLog) error
	GetByID(ctx context.Context, id string) (*models.SMSDeliveryLog, error)
	UpdateStatus(ctx context.Context, id, status string, snsMessageID, failureReason *string) error
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.SMSDeliveryLog, error)
	GetByPhoneMasked(ctx context.Context, maskedPhone string, since time.Time) ([]*models.SMSDeliveryLog, error)
	CountRecentByPhone(ctx context.Context, maskedPhone string, since time.Time) (int64, error)
	CountRecentByPhoneAndType(ctx context.Context, maskedPhone, messageType string, since time.Time) (int64, error)
	GetRecentFailures(ctx context.Context, maskedPhone string, since time.Time) ([]*models.SMSDeliveryLog, error)
	CleanupOldLogs(ctx context.Context, retentionDays int) (int64, error)
}

// SNSService implements interfaces.SMSService using AWS SNS
type SNSService struct {
	client       *sns.Client
	config       *SNSConfig
	deliveryRepo SMSDeliveryRepository
	auditService interfaces.AuditService
	logger       *zap.Logger
}

// NewSNSService creates a new SNS SMS service
func NewSNSService(
	ctx context.Context,
	snsConfig *SNSConfig,
	deliveryRepo SMSDeliveryRepository,
	auditService interfaces.AuditService,
	logger *zap.Logger,
) (*SNSService, error) {
	if snsConfig == nil {
		snsConfig = DefaultSNSConfig()
	}
	if err := snsConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid SNS config: %w", err)
	}

	// Load AWS config using default credential chain
	// Priority: IAM role > environment variables > shared credentials file
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(snsConfig.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sns.NewFromConfig(awsCfg)

	logger.Info("SNS SMS service initialized",
		zap.String("region", snsConfig.Region),
		zap.String("sender_id", snsConfig.SenderID),
		zap.Bool("enabled", snsConfig.Enabled))

	return &SNSService{
		client:       client,
		config:       snsConfig,
		deliveryRepo: deliveryRepo,
		auditService: auditService,
		logger:       logger,
	}, nil
}

// SendOTP sends a password reset OTP via SMS
func (s *SNSService) SendOTP(ctx context.Context, phoneNumber, otp string) error {
	if !s.config.Enabled {
		s.logger.Info("SMS service disabled, skipping OTP send",
			zap.String("phone_masked", models.MaskPhoneNumber(phoneNumber)))
		return nil
	}

	// Validate phone number
	if err := s.ValidatePhoneNumber(phoneNumber); err != nil {
		return err
	}

	maskedPhone := models.MaskPhoneNumberE164(phoneNumber)

	// Check rate limits
	if err := s.checkRateLimit(ctx, maskedPhone); err != nil {
		return err
	}

	// Create delivery log
	deliveryLog := models.NewSMSDeliveryLog(models.SMSTypePasswordResetOTP, maskedPhone)
	deliveryLog.AddDetail("otp_length", len(otp))

	if err := s.deliveryRepo.Create(ctx, deliveryLog); err != nil {
		s.logger.Warn("Failed to create SMS delivery log", zap.Error(err))
		// Continue anyway - don't block SMS sending
	}

	// Format message
	message := fmt.Sprintf("Your password reset code is %s. Valid for %d minutes. Do not share this code.",
		otp, int(s.config.OTPExpiry.Minutes()))

	// Send SMS
	messageID, err := s.sendSMS(ctx, phoneNumber, message)
	if err != nil {
		failureReason := err.Error()
		_ = s.deliveryRepo.UpdateStatus(ctx, deliveryLog.ID, models.SMSStatusFailed, nil, &failureReason)

		s.logger.Error("Failed to send OTP SMS",
			zap.String("phone_masked", maskedPhone),
			zap.Error(err))

		// Log audit event
		if s.auditService != nil {
			s.auditService.LogUserActionWithError(ctx, "", "sms_otp_send_failed", "sms", deliveryLog.ID, err, map[string]interface{}{
				"phone_masked": maskedPhone,
				"message_type": models.SMSTypePasswordResetOTP,
			})
		}

		return fmt.Errorf("failed to send OTP SMS: %w", err)
	}

	// Update delivery log with success
	_ = s.deliveryRepo.UpdateStatus(ctx, deliveryLog.ID, models.SMSStatusSent, &messageID, nil)

	s.logger.Info("OTP SMS sent successfully",
		zap.String("phone_masked", maskedPhone),
		zap.String("message_id", messageID))

	// Log audit event
	if s.auditService != nil {
		s.auditService.LogUserAction(ctx, "", "sms_otp_sent", "sms", deliveryLog.ID, map[string]interface{}{
			"phone_masked": maskedPhone,
			"message_id":   messageID,
			"message_type": models.SMSTypePasswordResetOTP,
		})
	}

	return nil
}

// SendSecurityAlert sends a security alert SMS
func (s *SNSService) SendSecurityAlert(ctx context.Context, phoneNumber, alertMessage string) error {
	if !s.config.Enabled {
		s.logger.Info("SMS service disabled, skipping security alert")
		return nil
	}

	if err := s.ValidatePhoneNumber(phoneNumber); err != nil {
		return err
	}

	maskedPhone := models.MaskPhoneNumberE164(phoneNumber)
	deliveryLog := models.NewSMSDeliveryLog(models.SMSTypeSecurityAlert, maskedPhone)

	if err := s.deliveryRepo.Create(ctx, deliveryLog); err != nil {
		s.logger.Warn("Failed to create SMS delivery log", zap.Error(err))
	}

	messageID, err := s.sendSMS(ctx, phoneNumber, alertMessage)
	if err != nil {
		failureReason := err.Error()
		_ = s.deliveryRepo.UpdateStatus(ctx, deliveryLog.ID, models.SMSStatusFailed, nil, &failureReason)
		return fmt.Errorf("failed to send security alert: %w", err)
	}

	_ = s.deliveryRepo.UpdateStatus(ctx, deliveryLog.ID, models.SMSStatusSent, &messageID, nil)

	s.logger.Info("Security alert SMS sent",
		zap.String("phone_masked", maskedPhone),
		zap.String("message_id", messageID))

	return nil
}

// SendBulkSMS sends SMS to multiple recipients
func (s *SNSService) SendBulkSMS(ctx context.Context, phoneNumbers []string, message string) error {
	if !s.config.Enabled {
		s.logger.Info("SMS service disabled, skipping bulk SMS")
		return nil
	}

	var errors []error
	successCount := 0

	for _, phone := range phoneNumbers {
		if err := s.ValidatePhoneNumber(phone); err != nil {
			errors = append(errors, fmt.Errorf("invalid phone %s: %w", models.MaskPhoneNumberE164(phone), err))
			continue
		}

		maskedPhone := models.MaskPhoneNumberE164(phone)
		deliveryLog := models.NewSMSDeliveryLog(models.SMSTypeBulk, maskedPhone)
		_ = s.deliveryRepo.Create(ctx, deliveryLog)

		messageID, err := s.sendSMS(ctx, phone, message)
		if err != nil {
			errors = append(errors, err)
			failureReason := err.Error()
			_ = s.deliveryRepo.UpdateStatus(ctx, deliveryLog.ID, models.SMSStatusFailed, nil, &failureReason)
			s.logger.Warn("Failed to send bulk SMS to recipient",
				zap.String("phone_masked", maskedPhone),
				zap.Error(err))
		} else {
			successCount++
			_ = s.deliveryRepo.UpdateStatus(ctx, deliveryLog.ID, models.SMSStatusSent, &messageID, nil)
		}
	}

	s.logger.Info("Bulk SMS completed",
		zap.Int("total", len(phoneNumbers)),
		zap.Int("success", successCount),
		zap.Int("failed", len(errors)))

	if len(errors) > 0 {
		return fmt.Errorf("failed to send %d of %d messages", len(errors), len(phoneNumbers))
	}
	return nil
}

// ValidatePhoneNumber validates phone number format (E.164)
func (s *SNSService) ValidatePhoneNumber(phoneNumber string) error {
	// E.164 format: +[country code][number], e.g., +919876543210
	// Length: 8-15 digits after the +
	e164Regex := regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
	if !e164Regex.MatchString(phoneNumber) {
		return fmt.Errorf("invalid phone number format: must be E.164 format (e.g., +919876543210)")
	}
	return nil
}

// sendSMS is the internal method that actually sends SMS via AWS SNS
func (s *SNSService) sendSMS(ctx context.Context, phoneNumber, message string) (string, error) {
	input := &sns.PublishInput{
		PhoneNumber: aws.String(phoneNumber),
		Message:     aws.String(message),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"AWS.SNS.SMS.SenderID": {
				DataType:    aws.String("String"),
				StringValue: aws.String(s.config.SenderID),
			},
			"AWS.SNS.SMS.SMSType": {
				DataType:    aws.String("String"),
				StringValue: aws.String(s.config.MessageType),
			},
		},
	}

	output, err := s.client.Publish(ctx, input)
	if err != nil {
		return "", fmt.Errorf("SNS publish failed: %w", err)
	}

	if output.MessageId == nil {
		return "", fmt.Errorf("SNS returned nil message ID")
	}

	return *output.MessageId, nil
}

// checkRateLimit checks if the phone number has exceeded rate limits
func (s *SNSService) checkRateLimit(ctx context.Context, maskedPhone string) error {
	// Check hourly limit
	hourAgo := time.Now().Add(-1 * time.Hour)
	hourlyCount, err := s.deliveryRepo.CountRecentByPhone(ctx, maskedPhone, hourAgo)
	if err != nil {
		s.logger.Warn("Failed to check hourly rate limit", zap.Error(err))
		// Don't block on rate limit check failure
	} else if hourlyCount >= int64(s.config.MaxSMSPerHour) {
		return fmt.Errorf("SMS rate limit exceeded: maximum %d SMS per hour", s.config.MaxSMSPerHour)
	}

	// Check daily limit
	dayAgo := time.Now().Add(-24 * time.Hour)
	dailyCount, err := s.deliveryRepo.CountRecentByPhone(ctx, maskedPhone, dayAgo)
	if err != nil {
		s.logger.Warn("Failed to check daily rate limit", zap.Error(err))
	} else if dailyCount >= int64(s.config.MaxSMSPerDay) {
		return fmt.Errorf("SMS rate limit exceeded: maximum %d SMS per day", s.config.MaxSMSPerDay)
	}

	return nil
}

// IsEnabled returns whether SMS service is enabled
func (s *SNSService) IsEnabled() bool {
	return s.config.Enabled
}

// GetConfig returns the current SNS configuration
func (s *SNSService) GetConfig() *SNSConfig {
	return s.config
}
