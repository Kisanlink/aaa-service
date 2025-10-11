package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"go.uber.org/zap"
)

// MaintenanceMode represents the maintenance mode configuration
type MaintenanceMode struct {
	Enabled    bool       `json:"enabled"`
	Message    string     `json:"message"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	EnabledBy  string     `json:"enabled_by"`
	Reason     string     `json:"reason"`
	AllowAdmin bool       `json:"allow_admin"`
	AllowRead  bool       `json:"allow_read"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// MaintenanceService manages application maintenance mode
type MaintenanceService struct {
	cacheService interfaces.CacheService
	logger       interfaces.Logger
	cacheKey     string
}

// NewMaintenanceService creates a new maintenance service instance
func NewMaintenanceService(
	cacheService interfaces.CacheService,
	logger interfaces.Logger,
) *MaintenanceService {
	return &MaintenanceService{
		cacheService: cacheService,
		logger:       logger,
		cacheKey:     "system:maintenance_mode",
	}
}

// IsMaintenanceMode checks if the system is currently in maintenance mode
func (s *MaintenanceService) IsMaintenanceMode(ctx context.Context) (bool, interface{}, error) {
	s.logger.Debug("Checking maintenance mode status")

	// Try to get maintenance mode from cache
	if cachedMode, found := s.cacheService.Get(s.cacheKey); found {
		if mode, ok := cachedMode.(*MaintenanceMode); ok {
			// Check if maintenance window has expired
			if mode.EndTime != nil && time.Now().After(*mode.EndTime) {
				s.logger.Info("Maintenance window expired, disabling maintenance mode")
				if err := s.DisableMaintenanceMode(ctx, "system"); err != nil {
					s.logger.Error("Failed to auto-disable expired maintenance mode", zap.Error(err))
				}
				return false, nil, nil
			}
			return mode.Enabled, mode, nil
		}
	}

	// Default to not in maintenance mode
	return false, nil, nil
}

// EnableMaintenanceMode enables maintenance mode with the specified configuration
func (s *MaintenanceService) EnableMaintenanceMode(ctx context.Context, config interface{}) error {
	s.logger.Info("Enabling maintenance mode")

	if config == nil {
		return errors.NewValidationError("maintenance mode configuration is required")
	}

	// Type assertion for configuration
	var maintenanceConfig *MaintenanceMode

	// Handle different input types
	switch v := config.(type) {
	case *MaintenanceMode:
		maintenanceConfig = v
	case map[string]interface{}:
		// Convert map to MaintenanceMode struct
		maintenanceConfig = &MaintenanceMode{}
		if enabled, ok := v["enabled"].(bool); ok {
			maintenanceConfig.Enabled = enabled
		}
		if message, ok := v["message"].(string); ok {
			maintenanceConfig.Message = message
		}
		if reason, ok := v["reason"].(string); ok {
			maintenanceConfig.Reason = reason
		}
		if enabledBy, ok := v["enabled_by"].(string); ok {
			maintenanceConfig.EnabledBy = enabledBy
		}
		if allowAdmin, ok := v["allow_admin"].(bool); ok {
			maintenanceConfig.AllowAdmin = allowAdmin
		}
		if allowRead, ok := v["allow_read"].(bool); ok {
			maintenanceConfig.AllowRead = allowRead
		}
		if endTime, ok := v["end_time"].(time.Time); ok {
			maintenanceConfig.EndTime = &endTime
		}
	default:
		return errors.NewValidationError("invalid maintenance mode configuration type")
	}

	s.logger.Info("Enabling maintenance mode",
		zap.String("enabled_by", maintenanceConfig.EnabledBy),
		zap.String("reason", maintenanceConfig.Reason))

	// Set default values
	if maintenanceConfig.Message == "" {
		maintenanceConfig.Message = "System is currently under maintenance. Please try again later."
	}

	maintenanceConfig.Enabled = true
	maintenanceConfig.StartTime = time.Now()
	maintenanceConfig.UpdatedAt = time.Now()

	// Store in cache with appropriate TTL
	ttl := 86400 // 24 hours default
	if maintenanceConfig.EndTime != nil {
		duration := time.Until(*maintenanceConfig.EndTime)
		if duration > 0 {
			ttl = int(duration.Seconds()) + 300 // Add 5 minutes buffer
		}
	}

	if err := s.cacheService.Set(s.cacheKey, maintenanceConfig, ttl); err != nil {
		s.logger.Error("Failed to enable maintenance mode", zap.Error(err))
		return errors.NewInternalError(fmt.Errorf("failed to enable maintenance mode: %w", err))
	}

	s.logger.Info("Maintenance mode enabled successfully")
	return nil
}

// DisableMaintenanceMode disables maintenance mode
func (s *MaintenanceService) DisableMaintenanceMode(ctx context.Context, disabledBy string) error {
	s.logger.Info("Disabling maintenance mode", zap.String("disabled_by", disabledBy))

	// Check if maintenance mode is currently enabled
	enabled, modeInterface, err := s.IsMaintenanceMode(ctx)
	if err != nil {
		return err
	}

	if !enabled {
		return errors.NewConflictError("maintenance mode is not currently enabled")
	}

	// Update the configuration to disabled
	if modeInterface != nil {
		if mode, ok := modeInterface.(*MaintenanceMode); ok {
			mode.Enabled = false
			mode.UpdatedAt = time.Now()
			// Keep the record for a short time for audit purposes
			if err := s.cacheService.Set(s.cacheKey+"_last", mode, 3600); err != nil {
				s.logger.Warn("Failed to store last maintenance mode record", zap.Error(err))
			}
		}
	}

	// Remove from cache
	if err := s.cacheService.Delete(s.cacheKey); err != nil {
		s.logger.Error("Failed to disable maintenance mode", zap.Error(err))
		return errors.NewInternalError(fmt.Errorf("failed to disable maintenance mode: %w", err))
	}

	s.logger.Info("Maintenance mode disabled successfully")
	return nil
}

// GetMaintenanceStatus returns the current maintenance mode status and configuration
func (s *MaintenanceService) GetMaintenanceStatus(ctx context.Context) (interface{}, error) {
	s.logger.Debug("Getting maintenance mode status")

	enabled, modeInterface, err := s.IsMaintenanceMode(ctx)
	if err != nil {
		return nil, err
	}

	if !enabled || modeInterface == nil {
		return &MaintenanceMode{
			Enabled:   false,
			Message:   "System is operating normally",
			UpdatedAt: time.Now(),
		}, nil
	}

	return modeInterface, nil
}

// IsUserAllowedDuringMaintenance checks if a user should be allowed access during maintenance
func (s *MaintenanceService) IsUserAllowedDuringMaintenance(ctx context.Context, userID string, isAdmin bool, isReadOperation bool) (bool, error) {
	enabled, modeInterface, err := s.IsMaintenanceMode(ctx)
	if err != nil {
		return false, err
	}

	if !enabled {
		return true, nil // Not in maintenance mode, allow all
	}

	if modeInterface == nil {
		return false, errors.NewInternalError(fmt.Errorf("maintenance mode enabled but configuration not found"))
	}

	// Type assert to get the actual maintenance mode configuration
	mode, ok := modeInterface.(*MaintenanceMode)
	if !ok {
		return false, errors.NewInternalError(fmt.Errorf("invalid maintenance mode configuration type"))
	}

	// Check admin bypass
	if isAdmin && mode.AllowAdmin {
		s.logger.Debug("Allowing admin user during maintenance", zap.String("userID", userID))
		return true, nil
	}

	// Check read operation bypass
	if isReadOperation && mode.AllowRead {
		s.logger.Debug("Allowing read operation during maintenance", zap.String("userID", userID))
		return true, nil
	}

	return false, nil
}

// UpdateMaintenanceMessage updates the maintenance message without changing other settings
func (s *MaintenanceService) UpdateMaintenanceMessage(ctx context.Context, message string, updatedBy string) error {
	enabled, modeInterface, err := s.IsMaintenanceMode(ctx)
	if err != nil {
		return err
	}

	if !enabled || modeInterface == nil {
		return errors.NewConflictError("maintenance mode is not currently enabled")
	}

	// Type assert to get the actual maintenance mode configuration
	mode, ok := modeInterface.(*MaintenanceMode)
	if !ok {
		return errors.NewInternalError(fmt.Errorf("invalid maintenance mode configuration type"))
	}

	mode.Message = message
	mode.UpdatedAt = time.Now()

	// Calculate remaining TTL
	ttl := 86400 // Default 24 hours
	if mode.EndTime != nil {
		duration := time.Until(*mode.EndTime)
		if duration > 0 {
			ttl = int(duration.Seconds()) + 300
		}
	}

	if err := s.cacheService.Set(s.cacheKey, mode, ttl); err != nil {
		return errors.NewInternalError(fmt.Errorf("failed to update maintenance message: %w", err))
	}

	s.logger.Info("Maintenance message updated",
		zap.String("updated_by", updatedBy),
		zap.String("new_message", message))

	return nil
}
