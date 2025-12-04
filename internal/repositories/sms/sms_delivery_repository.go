package sms

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SMSDeliveryRepository handles database operations for SMS delivery logs
type SMSDeliveryRepository struct {
	dbManager db.DBManager
	logger    *zap.Logger
}

// NewSMSDeliveryRepository creates a new SMSDeliveryRepository
func NewSMSDeliveryRepository(dbManager db.DBManager, logger *zap.Logger) *SMSDeliveryRepository {
	return &SMSDeliveryRepository{
		dbManager: dbManager,
		logger:    logger,
	}
}

// getDB is a helper method to get the database connection
func (r *SMSDeliveryRepository) getDB(ctx context.Context, readOnly bool) (*gorm.DB, error) {
	if postgresMgr, ok := r.dbManager.(interface {
		GetDB(context.Context, bool) (*gorm.DB, error)
	}); ok {
		return postgresMgr.GetDB(ctx, readOnly)
	}
	return nil, fmt.Errorf("database manager does not support GetDB method")
}

// Create creates a new SMS delivery log
func (r *SMSDeliveryRepository) Create(ctx context.Context, log *models.SMSDeliveryLog) error {
	if err := r.dbManager.Create(ctx, log); err != nil {
		r.logger.Error("Failed to create SMS delivery log",
			zap.Error(err),
			zap.String("message_type", log.MessageType))
		return fmt.Errorf("failed to create SMS delivery log: %w", err)
	}
	return nil
}

// GetByID retrieves an SMS delivery log by ID
func (r *SMSDeliveryRepository) GetByID(ctx context.Context, id string) (*models.SMSDeliveryLog, error) {
	var log models.SMSDeliveryLog
	if err := r.dbManager.GetByID(ctx, id, &log); err != nil {
		return nil, fmt.Errorf("failed to get SMS delivery log: %w", err)
	}
	return &log, nil
}

// UpdateStatus updates the status of an SMS delivery log
func (r *SMSDeliveryRepository) UpdateStatus(ctx context.Context, id, status string, snsMessageID, failureReason *string) error {
	db, err := r.getDB(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if snsMessageID != nil {
		updates["sns_message_id"] = *snsMessageID
	}
	if failureReason != nil {
		updates["failure_reason"] = *failureReason
	}
	if status == models.SMSStatusDelivered {
		now := time.Now()
		updates["delivered_at"] = now
	}

	result := db.WithContext(ctx).
		Model(&models.SMSDeliveryLog{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("Failed to update SMS delivery log status",
			zap.Error(result.Error),
			zap.String("id", id),
			zap.String("status", status))
		return fmt.Errorf("failed to update SMS delivery log: %w", result.Error)
	}
	return nil
}

// GetByUserID retrieves SMS delivery logs by user ID with pagination
func (r *SMSDeliveryRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.SMSDeliveryLog, error) {
	db, err := r.getDB(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	var logs []*models.SMSDeliveryLog
	result := db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("sent_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get SMS delivery logs by user ID: %w", result.Error)
	}
	return logs, nil
}

// GetByPhoneMasked retrieves SMS delivery logs by masked phone number since a given time
func (r *SMSDeliveryRepository) GetByPhoneMasked(ctx context.Context, maskedPhone string, since time.Time) ([]*models.SMSDeliveryLog, error) {
	db, err := r.getDB(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	var logs []*models.SMSDeliveryLog
	result := db.WithContext(ctx).
		Where("phone_number_masked = ? AND sent_at >= ?", maskedPhone, since).
		Order("sent_at DESC").
		Find(&logs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get SMS delivery logs by phone: %w", result.Error)
	}
	return logs, nil
}

// CountRecentByPhone counts recent SMS attempts for rate limiting
func (r *SMSDeliveryRepository) CountRecentByPhone(ctx context.Context, maskedPhone string, since time.Time) (int64, error) {
	db, err := r.getDB(ctx, true)
	if err != nil {
		return 0, fmt.Errorf("failed to get database connection: %w", err)
	}

	var count int64
	result := db.WithContext(ctx).
		Model(&models.SMSDeliveryLog{}).
		Where("phone_number_masked = ? AND sent_at >= ?", maskedPhone, since).
		Count(&count)

	if result.Error != nil {
		r.logger.Warn("Failed to count recent SMS by phone",
			zap.Error(result.Error),
			zap.String("phone_masked", maskedPhone))
		return 0, fmt.Errorf("failed to count recent SMS: %w", result.Error)
	}
	return count, nil
}

// CountRecentByPhoneAndType counts recent SMS attempts by phone and message type
func (r *SMSDeliveryRepository) CountRecentByPhoneAndType(ctx context.Context, maskedPhone, messageType string, since time.Time) (int64, error) {
	db, err := r.getDB(ctx, true)
	if err != nil {
		return 0, fmt.Errorf("failed to get database connection: %w", err)
	}

	var count int64
	result := db.WithContext(ctx).
		Model(&models.SMSDeliveryLog{}).
		Where("phone_number_masked = ? AND message_type = ? AND sent_at >= ?", maskedPhone, messageType, since).
		Count(&count)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to count recent SMS by phone and type: %w", result.Error)
	}
	return count, nil
}

// GetRecentFailures retrieves recent failed SMS attempts for a phone
func (r *SMSDeliveryRepository) GetRecentFailures(ctx context.Context, maskedPhone string, since time.Time) ([]*models.SMSDeliveryLog, error) {
	db, err := r.getDB(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	var logs []*models.SMSDeliveryLog
	result := db.WithContext(ctx).
		Where("phone_number_masked = ? AND status = ? AND sent_at >= ?", maskedPhone, models.SMSStatusFailed, since).
		Order("sent_at DESC").
		Find(&logs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get recent failures: %w", result.Error)
	}
	return logs, nil
}

// CleanupOldLogs removes logs older than the specified retention period
func (r *SMSDeliveryRepository) CleanupOldLogs(ctx context.Context, retentionDays int) (int64, error) {
	db, err := r.getDB(ctx, false)
	if err != nil {
		return 0, fmt.Errorf("failed to get database connection: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result := db.WithContext(ctx).
		Where("sent_at < ?", cutoff).
		Delete(&models.SMSDeliveryLog{})

	if result.Error != nil {
		r.logger.Error("Failed to cleanup old SMS logs",
			zap.Error(result.Error),
			zap.Int("retention_days", retentionDays))
		return 0, fmt.Errorf("failed to cleanup old SMS logs: %w", result.Error)
	}

	r.logger.Info("Cleaned up old SMS delivery logs",
		zap.Int64("deleted_count", result.RowsAffected),
		zap.Int("retention_days", retentionDays))

	return result.RowsAffected, nil
}
