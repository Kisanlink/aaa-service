package sms

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockSMSDeliveryRepository mocks the SMS delivery repository
type MockSMSDeliveryRepository struct {
	mock.Mock
}

func (m *MockSMSDeliveryRepository) Create(ctx context.Context, log *models.SMSDeliveryLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockSMSDeliveryRepository) GetByID(ctx context.Context, id string) (*models.SMSDeliveryLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SMSDeliveryLog), args.Error(1)
}

func (m *MockSMSDeliveryRepository) UpdateStatus(ctx context.Context, id, status string, snsMessageID, failureReason *string) error {
	args := m.Called(ctx, id, status, snsMessageID, failureReason)
	return args.Error(0)
}

func (m *MockSMSDeliveryRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.SMSDeliveryLog, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SMSDeliveryLog), args.Error(1)
}

func (m *MockSMSDeliveryRepository) GetByPhoneMasked(ctx context.Context, maskedPhone string, since time.Time) ([]*models.SMSDeliveryLog, error) {
	args := m.Called(ctx, maskedPhone, since)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SMSDeliveryLog), args.Error(1)
}

func (m *MockSMSDeliveryRepository) CountRecentByPhone(ctx context.Context, maskedPhone string, since time.Time) (int64, error) {
	args := m.Called(ctx, maskedPhone, since)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSMSDeliveryRepository) CountRecentByPhoneAndType(ctx context.Context, maskedPhone, messageType string, since time.Time) (int64, error) {
	args := m.Called(ctx, maskedPhone, messageType, since)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSMSDeliveryRepository) GetRecentFailures(ctx context.Context, maskedPhone string, since time.Time) ([]*models.SMSDeliveryLog, error) {
	args := m.Called(ctx, maskedPhone, since)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SMSDeliveryLog), args.Error(1)
}

func (m *MockSMSDeliveryRepository) CleanupOldLogs(ctx context.Context, retentionDays int) (int64, error) {
	args := m.Called(ctx, retentionDays)
	return args.Get(0).(int64), args.Error(1)
}

// MockAuditService mocks the audit service
type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) LogUserAction(ctx context.Context, userID, action, resource, resourceID string, details map[string]interface{}) {
	m.Called(ctx, userID, action, resource, resourceID, details)
}

func (m *MockAuditService) LogUserActionWithError(ctx context.Context, userID, action, resource, resourceID string, err error, details map[string]interface{}) {
	m.Called(ctx, userID, action, resource, resourceID, err, details)
}

func (m *MockAuditService) LogRoleChange(ctx context.Context, adminUserID, action, targetUserID, roleID string, details map[string]interface{}) {
	m.Called(ctx, adminUserID, action, targetUserID, roleID, details)
}

func (m *MockAuditService) LogPermissionChange(ctx context.Context, adminUserID, action, roleID, permissionID string, details map[string]interface{}) {
	m.Called(ctx, adminUserID, action, roleID, permissionID, details)
}

func (m *MockAuditService) LogSystemEvent(ctx context.Context, action, resource, resourceID, message string, details map[string]interface{}) {
	m.Called(ctx, action, resource, resourceID, message, details)
}

func (m *MockAuditService) GetUserAuditLogs(ctx context.Context, userID string, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0), args.Error(1)
}

func (m *MockAuditService) GetResourceAuditLogs(ctx context.Context, resourceType, resourceID string, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, resourceType, resourceID, limit, offset)
	return args.Get(0), args.Error(1)
}

func (m *MockAuditService) GetAuditLogsByAction(ctx context.Context, action string, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, action, limit, offset)
	return args.Get(0), args.Error(1)
}

func (m *MockAuditService) GetAuditLogsByTimeRange(ctx context.Context, startTime, endTime time.Time, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, startTime, endTime, limit, offset)
	return args.Get(0), args.Error(1)
}

func TestSNSService_ValidatePhoneNumber(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultSNSConfig()
	config.Enabled = true

	// Create a minimal service for testing validation
	svc := &SNSService{
		config: config,
		logger: logger,
	}

	tests := []struct {
		name        string
		phoneNumber string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid Indian number",
			phoneNumber: "+919876543210",
			wantErr:     false,
		},
		{
			name:        "valid US number",
			phoneNumber: "+14155552671",
			wantErr:     false,
		},
		{
			name:        "valid UK number",
			phoneNumber: "+447700900123",
			wantErr:     false,
		},
		{
			name:        "missing plus sign",
			phoneNumber: "919876543210",
			wantErr:     true,
			errContains: "E.164 format",
		},
		{
			name:        "too short",
			phoneNumber: "+9198765",
			wantErr:     true,
			errContains: "E.164 format",
		},
		{
			name:        "too long",
			phoneNumber: "+9198765432101234567",
			wantErr:     true,
			errContains: "E.164 format",
		},
		{
			name:        "contains non-digits",
			phoneNumber: "+91-9876-543210",
			wantErr:     true,
			errContains: "E.164 format",
		},
		{
			name:        "empty string",
			phoneNumber: "",
			wantErr:     true,
			errContains: "E.164 format",
		},
		{
			name:        "starts with zero",
			phoneNumber: "+0987654321",
			wantErr:     true,
			errContains: "E.164 format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidatePhoneNumber(tt.phoneNumber)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSNSService_SendOTP_Disabled(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockSMSDeliveryRepository)

	config := DefaultSNSConfig()
	config.Enabled = false // Disabled

	svc := &SNSService{
		config:       config,
		deliveryRepo: nil, // Won't be used since SMS is disabled
		logger:       logger,
	}

	err := svc.SendOTP(context.Background(), "+919876543210", "123456")
	assert.NoError(t, err, "Should succeed silently when SMS is disabled")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestSNSService_SendOTP_InvalidPhoneNumber(t *testing.T) {
	logger := zap.NewNop()

	config := DefaultSNSConfig()
	config.Enabled = true

	svc := &SNSService{
		config: config,
		logger: logger,
	}

	err := svc.SendOTP(context.Background(), "invalid-phone", "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "E.164 format")
}

func TestSNSService_SendOTP_RateLimitExceeded_Hourly(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockSMSDeliveryRepository)

	config := DefaultSNSConfig()
	config.Enabled = true
	config.MaxSMSPerHour = 5

	svc := &SNSService{
		config:       config,
		deliveryRepo: mockRepo,
		logger:       logger,
	}

	// Mock hourly count at limit
	mockRepo.On("CountRecentByPhone", mock.Anything, mock.Anything, mock.Anything).Return(int64(5), nil).Once()

	err := svc.SendOTP(context.Background(), "+919876543210", "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded")
	assert.Contains(t, err.Error(), "per hour")
}

func TestSNSService_SendOTP_RateLimitExceeded_Daily(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockSMSDeliveryRepository)

	config := DefaultSNSConfig()
	config.Enabled = true
	config.MaxSMSPerHour = 10
	config.MaxSMSPerDay = 20

	svc := &SNSService{
		config:       config,
		deliveryRepo: mockRepo,
		logger:       logger,
	}

	// Mock hourly count within limit
	mockRepo.On("CountRecentByPhone", mock.Anything, mock.Anything, mock.Anything).Return(int64(3), nil).Once()
	// Mock daily count at limit
	mockRepo.On("CountRecentByPhone", mock.Anything, mock.Anything, mock.Anything).Return(int64(20), nil).Once()

	err := svc.SendOTP(context.Background(), "+919876543210", "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded")
	assert.Contains(t, err.Error(), "per day")
}

func TestSNSService_SendSecurityAlert_Disabled(t *testing.T) {
	logger := zap.NewNop()

	config := DefaultSNSConfig()
	config.Enabled = false

	svc := &SNSService{
		config: config,
		logger: logger,
	}

	err := svc.SendSecurityAlert(context.Background(), "+919876543210", "Security alert message")
	assert.NoError(t, err, "Should succeed silently when SMS is disabled")
}

func TestSNSService_SendBulkSMS_Disabled(t *testing.T) {
	logger := zap.NewNop()

	config := DefaultSNSConfig()
	config.Enabled = false

	svc := &SNSService{
		config: config,
		logger: logger,
	}

	phoneNumbers := []string{"+919876543210", "+919876543211"}
	err := svc.SendBulkSMS(context.Background(), phoneNumbers, "Bulk message")
	assert.NoError(t, err, "Should succeed silently when SMS is disabled")
}

func TestSNSService_IsEnabled(t *testing.T) {
	t.Run("returns true when enabled", func(t *testing.T) {
		config := DefaultSNSConfig()
		config.Enabled = true
		svc := &SNSService{config: config}
		assert.True(t, svc.IsEnabled())
	})

	t.Run("returns false when disabled", func(t *testing.T) {
		config := DefaultSNSConfig()
		config.Enabled = false
		svc := &SNSService{config: config}
		assert.False(t, svc.IsEnabled())
	})
}

func TestSNSService_GetConfig(t *testing.T) {
	config := &SNSConfig{
		Region:   "us-west-2",
		Enabled:  true,
		SenderID: "TESTAPP",
	}
	svc := &SNSService{config: config}

	returnedConfig := svc.GetConfig()
	assert.Equal(t, config, returnedConfig)
	assert.Equal(t, "us-west-2", returnedConfig.Region)
	assert.Equal(t, "TESTAPP", returnedConfig.SenderID)
}

func TestSNSService_checkRateLimit(t *testing.T) {
	logger := zap.NewNop()

	t.Run("passes when within limits", func(t *testing.T) {
		mockRepo := new(MockSMSDeliveryRepository)
		config := DefaultSNSConfig()
		config.MaxSMSPerHour = 5
		config.MaxSMSPerDay = 20

		svc := &SNSService{
			config:       config,
			deliveryRepo: mockRepo,
			logger:       logger,
		}

		// Both counts within limits
		mockRepo.On("CountRecentByPhone", mock.Anything, mock.Anything, mock.Anything).Return(int64(2), nil).Once()
		mockRepo.On("CountRecentByPhone", mock.Anything, mock.Anything, mock.Anything).Return(int64(10), nil).Once()

		err := svc.checkRateLimit(context.Background(), "+91-XXXX-XXX-3210")
		assert.NoError(t, err)
	})

	t.Run("continues on rate limit check error", func(t *testing.T) {
		mockRepo := new(MockSMSDeliveryRepository)
		config := DefaultSNSConfig()

		svc := &SNSService{
			config:       config,
			deliveryRepo: mockRepo,
			logger:       logger,
		}

		// Simulate database error
		mockRepo.On("CountRecentByPhone", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), errors.New("db error")).Once()
		mockRepo.On("CountRecentByPhone", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), errors.New("db error")).Once()

		err := svc.checkRateLimit(context.Background(), "+91-XXXX-XXX-3210")
		assert.NoError(t, err, "Should not block on rate limit check failure")
	})
}
