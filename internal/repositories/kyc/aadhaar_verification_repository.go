package kyc

import (
	"bytes"
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AadhaarVerificationRepository handles database operations for Aadhaar verification entities
type AadhaarVerificationRepository interface {
	// Create new verification record
	Create(ctx context.Context, verification *models.AadhaarVerification) error

	// Get verification by ID
	GetByID(ctx context.Context, id string) (*models.AadhaarVerification, error)

	// Get verification by user ID
	GetByUserID(ctx context.Context, userID string) (*models.AadhaarVerification, error)

	// Get verification by reference ID (from Sandbox API)
	GetByReferenceID(ctx context.Context, referenceID string) (*models.AadhaarVerification, error)

	// Update verification record with all fields
	Update(ctx context.Context, verification *models.AadhaarVerification) error

	// Update verification status (PENDING, VERIFIED, FAILED)
	UpdateStatus(ctx context.Context, id string, status string) error

	// Increment verification attempts counter
	IncrementAttempts(ctx context.Context, id string) error

	// Create OTP attempt record
	CreateOTPAttempt(ctx context.Context, attempt *models.OTPAttempt) error

	// Get all OTP attempts for a verification
	GetOTPAttempts(ctx context.Context, verificationID string) ([]models.OTPAttempt, error)

	// Photo storage using kisanlink-db S3 manager
	UploadPhoto(ctx context.Context, userID string, photoData []byte, fileName string) (string, error)
	DeletePhoto(ctx context.Context, photoURL string) error
}

type aadhaarVerificationRepository struct {
	dbManager db.DBManager
	s3Manager *db.S3Manager
	logger    *zap.Logger
}

// NewAadhaarVerificationRepository creates a new AadhaarVerificationRepository instance
func NewAadhaarVerificationRepository(dbManager db.DBManager, s3Manager *db.S3Manager, logger *zap.Logger) AadhaarVerificationRepository {
	return &aadhaarVerificationRepository{
		dbManager: dbManager,
		s3Manager: s3Manager,
		logger:    logger,
	}
}

// getDB retrieves the database connection from the DBManager
func (r *aadhaarVerificationRepository) getDB(ctx context.Context, readOnly bool) (*gorm.DB, error) {
	// Try to get the database from the database manager
	if postgresMgr, ok := r.dbManager.(interface {
		GetDB(context.Context, bool) (*gorm.DB, error)
	}); ok {
		return postgresMgr.GetDB(ctx, readOnly)
	}

	return nil, fmt.Errorf("database manager does not support GetDB method")
}

// Create creates a new aadhaar verification record
func (r *aadhaarVerificationRepository) Create(ctx context.Context, verification *models.AadhaarVerification) error {
	db, err := r.getDB(ctx, false)
	if err != nil {
		r.logger.Error("Failed to get database connection", zap.Error(err))
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	if err := db.WithContext(ctx).Create(verification).Error; err != nil {
		r.logger.Error("Failed to create aadhaar verification", zap.Error(err))
		return fmt.Errorf("failed to create aadhaar verification: %w", err)
	}

	return nil
}

// GetByID retrieves an aadhaar verification by ID
func (r *aadhaarVerificationRepository) GetByID(ctx context.Context, id string) (*models.AadhaarVerification, error) {
	db, err := r.getDB(ctx, true)
	if err != nil {
		r.logger.Error("Failed to get database connection",
			zap.String("id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	verification := &models.AadhaarVerification{}
	if err := db.WithContext(ctx).Where("id = ?", id).First(verification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("aadhaar verification not found with id: %s", id)
		}
		r.logger.Error("Failed to get aadhaar verification by ID",
			zap.String("id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get aadhaar verification with id %s: %w", id, err)
	}
	return verification, nil
}

// GetByUserID retrieves an aadhaar verification by user ID
func (r *aadhaarVerificationRepository) GetByUserID(ctx context.Context, userID string) (*models.AadhaarVerification, error) {
	db, err := r.getDB(ctx, true)
	if err != nil {
		r.logger.Error("Failed to get database connection",
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	verification := &models.AadhaarVerification{}
	if err := db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(verification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("aadhaar verification not found for user: %s", userID)
		}
		r.logger.Error("Failed to get aadhaar verification by user ID",
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get aadhaar verification for user %s: %w", userID, err)
	}

	return verification, nil
}

// GetByReferenceID retrieves an aadhaar verification by reference ID
func (r *aadhaarVerificationRepository) GetByReferenceID(ctx context.Context, referenceID string) (*models.AadhaarVerification, error) {
	db, err := r.getDB(ctx, true)
	if err != nil {
		r.logger.Error("Failed to get database connection",
			zap.String("reference_id", referenceID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	verification := &models.AadhaarVerification{}
	if err := db.WithContext(ctx).
		Where("reference_id = ?", referenceID).
		First(verification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("aadhaar verification not found with reference_id: %s", referenceID)
		}
		r.logger.Error("Failed to get aadhaar verification by reference ID",
			zap.String("reference_id", referenceID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get aadhaar verification with reference_id %s: %w", referenceID, err)
	}

	return verification, nil
}

// Update updates the entire verification record
func (r *aadhaarVerificationRepository) Update(ctx context.Context, verification *models.AadhaarVerification) error {
	if verification == nil {
		return fmt.Errorf("verification cannot be nil")
	}

	if verification.ID == "" {
		return fmt.Errorf("verification ID is required")
	}

	db, err := r.getDB(ctx, false)
	if err != nil {
		r.logger.Error("Failed to get database connection",
			zap.String("id", verification.ID),
			zap.Error(err))
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Update all fields
	result := db.WithContext(ctx).
		Model(&models.AadhaarVerification{}).
		Where("id = ?", verification.ID).
		Updates(map[string]interface{}{
			"verification_status": verification.VerificationStatus,
			"kyc_status":          verification.KYCStatus,
			"otp_verified_at":     verification.OTPVerifiedAt,
			"photo_url":           verification.PhotoURL,
			"name":                verification.Name,
			"gender":              verification.Gender,
			"date_of_birth":       verification.DateOfBirth,
			"full_address":        verification.FullAddress,
			"address_json":        verification.AddressJSON,
			"updated_by":          verification.UpdatedBy,
			"updated_at":          verification.UpdatedAt,
		})

	if result.Error != nil {
		r.logger.Error("Failed to update verification",
			zap.String("id", verification.ID),
			zap.Error(result.Error))
		return fmt.Errorf("failed to update verification: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("aadhaar verification not found with id: %s", verification.ID)
	}

	r.logger.Info("Updated verification record",
		zap.String("id", verification.ID),
		zap.String("verification_status", verification.VerificationStatus),
		zap.String("kyc_status", verification.KYCStatus),
		zap.String("name", verification.Name))

	return nil
}

// UpdateStatus updates the verification status
func (r *aadhaarVerificationRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	db, err := r.getDB(ctx, false)
	if err != nil {
		r.logger.Error("Failed to get database connection",
			zap.String("id", id),
			zap.Error(err))
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	result := db.WithContext(ctx).
		Model(&models.AadhaarVerification{}).
		Where("id = ?", id).
		Update("verification_status", status)

	if result.Error != nil {
		r.logger.Error("Failed to update verification status",
			zap.String("id", id),
			zap.String("status", status),
			zap.Error(result.Error))
		return fmt.Errorf("failed to update verification status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("aadhaar verification not found with id: %s", id)
	}

	r.logger.Info("Updated verification status",
		zap.String("id", id),
		zap.String("status", status))

	return nil
}

// IncrementAttempts increments the verification attempts counter
func (r *aadhaarVerificationRepository) IncrementAttempts(ctx context.Context, id string) error {
	db, err := r.getDB(ctx, false)
	if err != nil {
		r.logger.Error("Failed to get database connection",
			zap.String("id", id),
			zap.Error(err))
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	result := db.WithContext(ctx).
		Model(&models.AadhaarVerification{}).
		Where("id = ?", id).
		UpdateColumn("attempts", gorm.Expr("attempts + ?", 1))

	if result.Error != nil {
		r.logger.Error("Failed to increment verification attempts",
			zap.String("id", id),
			zap.Error(result.Error))
		return fmt.Errorf("failed to increment attempts: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("aadhaar verification not found with id: %s", id)
	}

	r.logger.Debug("Incremented verification attempts",
		zap.String("id", id))

	return nil
}

// CreateOTPAttempt creates a new OTP attempt record
func (r *aadhaarVerificationRepository) CreateOTPAttempt(ctx context.Context, attempt *models.OTPAttempt) error {
	db, err := r.getDB(ctx, false)
	if err != nil {
		r.logger.Error("Failed to get database connection",
			zap.String("verification_id", attempt.AadhaarVerificationID),
			zap.Error(err))
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	if err := db.WithContext(ctx).Create(attempt).Error; err != nil {
		r.logger.Error("Failed to create OTP attempt",
			zap.String("verification_id", attempt.AadhaarVerificationID),
			zap.Int("attempt_number", attempt.AttemptNumber),
			zap.Error(err))
		return fmt.Errorf("failed to create OTP attempt: %w", err)
	}

	r.logger.Info("Created OTP attempt",
		zap.String("verification_id", attempt.AadhaarVerificationID),
		zap.Int("attempt_number", attempt.AttemptNumber),
		zap.String("status", attempt.Status))

	return nil
}

// GetOTPAttempts retrieves all OTP attempts for a verification
func (r *aadhaarVerificationRepository) GetOTPAttempts(ctx context.Context, verificationID string) ([]models.OTPAttempt, error) {
	db, err := r.getDB(ctx, true)
	if err != nil {
		r.logger.Error("Failed to get database connection",
			zap.String("verification_id", verificationID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	var attempts []models.OTPAttempt
	if err := db.WithContext(ctx).
		Where("aadhaar_verification_id = ?", verificationID).
		Order("created_at DESC").
		Find(&attempts).Error; err != nil {
		r.logger.Error("Failed to get OTP attempts",
			zap.String("verification_id", verificationID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get OTP attempts: %w", err)
	}

	r.logger.Debug("Retrieved OTP attempts",
		zap.String("verification_id", verificationID),
		zap.Int("count", len(attempts)))

	return attempts, nil
}

// UploadPhoto uploads a photo to S3 using kisanlink-db S3 manager
func (r *aadhaarVerificationRepository) UploadPhoto(ctx context.Context, userID string, photoData []byte, fileName string) (string, error) {
	if r.s3Manager == nil {
		r.logger.Error("S3 manager not initialized")
		return "", fmt.Errorf("S3 manager not initialized")
	}

	// Folder structure: aadhaar/photos/{userID}/{fileName}
	key := fmt.Sprintf("aadhaar/photos/%s/%s", userID, fileName)

	// Create a reader from photo data
	reader := bytes.NewReader(photoData)

	// Upload file to S3
	if err := r.s3Manager.UploadFile(ctx, key, reader, "image/jpeg", nil); err != nil {
		r.logger.Error("Failed to upload photo to S3",
			zap.String("user_id", userID),
			zap.String("file_name", fileName),
			zap.String("key", key),
			zap.Error(err))
		return "", fmt.Errorf("failed to upload photo: %w", err)
	}

	// Generate presigned URL (valid for 7 days)
	photoURL, err := r.s3Manager.GetPresignedURL(ctx, key, 7*24*60*60*1000000000) // 7 days in nanoseconds
	if err != nil {
		r.logger.Error("Failed to generate presigned URL",
			zap.String("key", key),
			zap.Error(err))
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	r.logger.Info("Photo uploaded successfully",
		zap.String("user_id", userID),
		zap.String("file_name", fileName),
		zap.String("key", key))

	return photoURL, nil
}

// DeletePhoto deletes a photo from S3
func (r *aadhaarVerificationRepository) DeletePhoto(ctx context.Context, photoURL string) error {
	if r.s3Manager == nil {
		r.logger.Error("S3 manager not initialized")
		return fmt.Errorf("S3 manager not initialized")
	}

	// Extract key from URL (assuming URL format includes the key)
	// For simplicity, we'll assume photoURL is either a full URL or just the key
	// In production, you'd parse the URL properly
	key := photoURL

	if err := r.s3Manager.Delete(ctx, key); err != nil {
		r.logger.Error("Failed to delete photo from S3",
			zap.String("photo_url", photoURL),
			zap.Error(err))
		return fmt.Errorf("failed to delete photo: %w", err)
	}

	r.logger.Info("Photo deleted successfully",
		zap.String("photo_url", photoURL))

	return nil
}
