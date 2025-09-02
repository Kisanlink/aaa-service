package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"go.uber.org/zap"
)

// ErrorMonitoringService provides error monitoring and alerting capabilities
type ErrorMonitoringService struct {
	logger       *zap.Logger
	auditService *AuditService
	cacheService interfaces.CacheService

	// Error tracking
	errorCounts    map[string]*ErrorCounter
	errorCountsMux sync.RWMutex

	// Alert thresholds
	config *ErrorMonitoringConfig
}

// ErrorMonitoringConfig holds configuration for error monitoring
type ErrorMonitoringConfig struct {
	// Alert thresholds
	ErrorRateThreshold        float64 // Errors per minute to trigger alert
	CriticalErrorThreshold    int     // Critical errors to trigger immediate alert
	SecurityErrorThreshold    int     // Security errors to trigger alert
	ConsecutiveErrorThreshold int     // Consecutive errors to trigger alert

	// Time windows
	MonitoringWindow time.Duration // Window for error rate calculation
	AlertCooldown    time.Duration // Cooldown between alerts
	CleanupInterval  time.Duration // Interval to clean up old error data

	// Alert configuration
	EnableEmailAlerts bool
	EnableSlackAlerts bool
	EnableSMSAlerts   bool
	AlertRecipients   []string

	// Error categorization
	CriticalErrorTypes []string
	SecurityErrorTypes []string
}

// ErrorCounter tracks error counts and rates
type ErrorCounter struct {
	Count            int
	LastOccurrence   time.Time
	FirstOccurrence  time.Time
	ErrorType        string
	LastError        error
	ConsecutiveCount int
	LastAlertTime    time.Time
}

// ErrorAlert represents an error alert
type ErrorAlert struct {
	ID         string
	Type       string
	Severity   string
	Message    string
	ErrorCount int
	TimeWindow time.Duration
	Timestamp  time.Time
	Context    map[string]interface{}
}

// NewErrorMonitoringService creates a new error monitoring service
func NewErrorMonitoringService(
	logger *zap.Logger,
	auditService *AuditService,
	cacheService interfaces.CacheService,
) *ErrorMonitoringService {
	config := &ErrorMonitoringConfig{
		ErrorRateThreshold:        10.0, // 10 errors per minute
		CriticalErrorThreshold:    5,    // 5 critical errors
		SecurityErrorThreshold:    3,    // 3 security errors
		ConsecutiveErrorThreshold: 10,   // 10 consecutive errors

		MonitoringWindow: 5 * time.Minute,
		AlertCooldown:    15 * time.Minute,
		CleanupInterval:  1 * time.Hour,

		EnableEmailAlerts: true,
		EnableSlackAlerts: false,
		EnableSMSAlerts:   false,

		CriticalErrorTypes: []string{
			"INTERNAL_ERROR",
			"DATABASE_ERROR",
			"AUTHENTICATION_SYSTEM_ERROR",
		},
		SecurityErrorTypes: []string{
			"UNAUTHORIZED",
			"FORBIDDEN",
			"SUSPICIOUS_ACTIVITY",
			"RATE_LIMIT_VIOLATION",
		},
	}

	service := &ErrorMonitoringService{
		logger:       logger,
		auditService: auditService,
		cacheService: cacheService,
		errorCounts:  make(map[string]*ErrorCounter),
		config:       config,
	}

	// Start background cleanup routine
	go service.startCleanupRoutine()

	return service
}

// RecordError records an error for monitoring
func (s *ErrorMonitoringService) RecordError(ctx context.Context, err error, requestID, userID, path string) {
	if err == nil {
		return
	}

	errorType := s.getErrorType(err)
	errorKey := fmt.Sprintf("%s:%s", errorType, path)

	s.errorCountsMux.Lock()
	defer s.errorCountsMux.Unlock()

	counter, exists := s.errorCounts[errorKey]
	if !exists {
		counter = &ErrorCounter{
			ErrorType:       errorType,
			FirstOccurrence: time.Now(),
		}
		s.errorCounts[errorKey] = counter
	}

	// Update counter
	counter.Count++
	counter.LastOccurrence = time.Now()
	counter.LastError = err
	counter.ConsecutiveCount++

	// Log error details
	s.logger.Error("Error recorded for monitoring",
		zap.String("request_id", requestID),
		zap.String("user_id", userID),
		zap.String("path", path),
		zap.String("error_type", errorType),
		zap.Error(err),
		zap.Int("total_count", counter.Count),
		zap.Int("consecutive_count", counter.ConsecutiveCount),
	)

	// Check if we need to trigger alerts
	s.checkAndTriggerAlerts(ctx, errorKey, counter, requestID, userID, path)

	// Log to audit service for security errors
	if s.isSecurityError(errorType) {
		s.auditService.LogSecurityEvent(ctx, userID, "error_occurred", path, false, map[string]interface{}{
			"error_type":    errorType,
			"error_message": err.Error(),
			"request_id":    requestID,
		})
	}
}

// RecordSuccessfulRequest records a successful request (resets consecutive error count)
func (s *ErrorMonitoringService) RecordSuccessfulRequest(path string) {
	s.errorCountsMux.Lock()
	defer s.errorCountsMux.Unlock()

	// Reset consecutive error counts for this path
	for key, counter := range s.errorCounts {
		if fmt.Sprintf("%s:%s", counter.ErrorType, path) == key {
			counter.ConsecutiveCount = 0
		}
	}
}

// GetErrorStatistics returns error statistics for monitoring dashboards
func (s *ErrorMonitoringService) GetErrorStatistics(ctx context.Context, timeWindow time.Duration) (*ErrorStatistics, error) {
	s.errorCountsMux.RLock()
	defer s.errorCountsMux.RUnlock()

	cutoff := time.Now().Add(-timeWindow)
	stats := &ErrorStatistics{
		TimeWindow:   timeWindow,
		GeneratedAt:  time.Now(),
		ErrorsByType: make(map[string]int),
		ErrorsByPath: make(map[string]int),
		TotalErrors:  0,
	}

	for key, counter := range s.errorCounts {
		if counter.LastOccurrence.After(cutoff) {
			stats.TotalErrors += counter.Count
			stats.ErrorsByType[counter.ErrorType] += counter.Count

			// Extract path from key (format: "errorType:path")
			if colonIndex := len(counter.ErrorType) + 1; colonIndex < len(key) {
				path := key[colonIndex:]
				stats.ErrorsByPath[path] += counter.Count
			}
		}
	}

	// Calculate error rate (errors per minute)
	if timeWindow.Minutes() > 0 {
		stats.ErrorRate = float64(stats.TotalErrors) / timeWindow.Minutes()
	}

	return stats, nil
}

// checkAndTriggerAlerts checks if alerts should be triggered based on error patterns
func (s *ErrorMonitoringService) checkAndTriggerAlerts(ctx context.Context, errorKey string, counter *ErrorCounter, requestID, userID, path string) {
	now := time.Now()

	// Check cooldown period
	if now.Sub(counter.LastAlertTime) < s.config.AlertCooldown {
		return
	}

	// Check for critical errors
	if s.isCriticalError(counter.ErrorType) && counter.Count >= s.config.CriticalErrorThreshold {
		alert := &ErrorAlert{
			ID:         fmt.Sprintf("critical_%s_%d", errorKey, now.Unix()),
			Type:       "CRITICAL_ERROR",
			Severity:   "CRITICAL",
			Message:    fmt.Sprintf("Critical error threshold exceeded: %d occurrences of %s", counter.Count, counter.ErrorType),
			ErrorCount: counter.Count,
			TimeWindow: s.config.MonitoringWindow,
			Timestamp:  now,
			Context: map[string]interface{}{
				"error_type": counter.ErrorType,
				"path":       path,
				"user_id":    userID,
				"request_id": requestID,
				"last_error": counter.LastError.Error(),
			},
		}
		s.triggerAlert(ctx, alert)
		counter.LastAlertTime = now
		return
	}

	// Check for security errors
	if s.isSecurityError(counter.ErrorType) && counter.Count >= s.config.SecurityErrorThreshold {
		alert := &ErrorAlert{
			ID:         fmt.Sprintf("security_%s_%d", errorKey, now.Unix()),
			Type:       "SECURITY_ERROR",
			Severity:   "HIGH",
			Message:    fmt.Sprintf("Security error threshold exceeded: %d occurrences of %s", counter.Count, counter.ErrorType),
			ErrorCount: counter.Count,
			TimeWindow: s.config.MonitoringWindow,
			Timestamp:  now,
			Context: map[string]interface{}{
				"error_type": counter.ErrorType,
				"path":       path,
				"user_id":    userID,
				"request_id": requestID,
			},
		}
		s.triggerAlert(ctx, alert)
		counter.LastAlertTime = now
		return
	}

	// Check for consecutive errors
	if counter.ConsecutiveCount >= s.config.ConsecutiveErrorThreshold {
		alert := &ErrorAlert{
			ID:         fmt.Sprintf("consecutive_%s_%d", errorKey, now.Unix()),
			Type:       "CONSECUTIVE_ERRORS",
			Severity:   "MEDIUM",
			Message:    fmt.Sprintf("Consecutive error threshold exceeded: %d consecutive %s errors", counter.ConsecutiveCount, counter.ErrorType),
			ErrorCount: counter.ConsecutiveCount,
			TimeWindow: s.config.MonitoringWindow,
			Timestamp:  now,
			Context: map[string]interface{}{
				"error_type": counter.ErrorType,
				"path":       path,
				"user_id":    userID,
				"request_id": requestID,
			},
		}
		s.triggerAlert(ctx, alert)
		counter.LastAlertTime = now
		return
	}

	// Check error rate
	errorRate := s.calculateErrorRate(errorKey)
	if errorRate > s.config.ErrorRateThreshold {
		alert := &ErrorAlert{
			ID:         fmt.Sprintf("rate_%s_%d", errorKey, now.Unix()),
			Type:       "HIGH_ERROR_RATE",
			Severity:   "MEDIUM",
			Message:    fmt.Sprintf("High error rate detected: %.2f errors/minute for %s", errorRate, counter.ErrorType),
			ErrorCount: counter.Count,
			TimeWindow: s.config.MonitoringWindow,
			Timestamp:  now,
			Context: map[string]interface{}{
				"error_type": counter.ErrorType,
				"error_rate": errorRate,
				"path":       path,
				"user_id":    userID,
				"request_id": requestID,
			},
		}
		s.triggerAlert(ctx, alert)
		counter.LastAlertTime = now
	}
}

// triggerAlert triggers an alert through configured channels
func (s *ErrorMonitoringService) triggerAlert(ctx context.Context, alert *ErrorAlert) {
	// Log the alert
	s.logger.Error("Error alert triggered",
		zap.String("alert_id", alert.ID),
		zap.String("alert_type", alert.Type),
		zap.String("severity", alert.Severity),
		zap.String("message", alert.Message),
		zap.Int("error_count", alert.ErrorCount),
		zap.Any("context", alert.Context),
	)

	// Log to audit service
	s.auditService.LogSystemEvent(ctx, "error_alert_triggered", "monitoring", true, map[string]interface{}{
		"alert_id":    alert.ID,
		"alert_type":  alert.Type,
		"severity":    alert.Severity,
		"message":     alert.Message,
		"error_count": alert.ErrorCount,
		"context":     alert.Context,
	})

	// Send alerts through configured channels
	if s.config.EnableEmailAlerts {
		s.sendEmailAlert(alert)
	}

	if s.config.EnableSlackAlerts {
		s.sendSlackAlert(alert)
	}

	if s.config.EnableSMSAlerts && alert.Severity == "CRITICAL" {
		s.sendSMSAlert(alert)
	}
}

// Helper methods

func (s *ErrorMonitoringService) getErrorType(err error) string {
	switch err.(type) {
	case *errors.ValidationError:
		return "VALIDATION_ERROR"
	case *errors.BadRequestError:
		return "BAD_REQUEST"
	case *errors.UnauthorizedError:
		return "UNAUTHORIZED"
	case *errors.ForbiddenError:
		return "FORBIDDEN"
	case *errors.NotFoundError:
		return "NOT_FOUND"
	case *errors.ConflictError:
		return "CONFLICT"
	case *errors.InternalError:
		return "INTERNAL_ERROR"
	default:
		return "GENERIC_ERROR"
	}
}

func (s *ErrorMonitoringService) isCriticalError(errorType string) bool {
	for _, criticalType := range s.config.CriticalErrorTypes {
		if errorType == criticalType {
			return true
		}
	}
	return false
}

func (s *ErrorMonitoringService) isSecurityError(errorType string) bool {
	for _, securityType := range s.config.SecurityErrorTypes {
		if errorType == securityType {
			return true
		}
	}
	return false
}

func (s *ErrorMonitoringService) calculateErrorRate(errorKey string) float64 {
	counter, exists := s.errorCounts[errorKey]
	if !exists {
		return 0
	}

	duration := time.Since(counter.FirstOccurrence)
	if duration.Minutes() == 0 {
		return 0
	}

	return float64(counter.Count) / duration.Minutes()
}

func (s *ErrorMonitoringService) startCleanupRoutine() {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanupOldErrors()
	}
}

func (s *ErrorMonitoringService) cleanupOldErrors() {
	s.errorCountsMux.Lock()
	defer s.errorCountsMux.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour) // Keep errors for 24 hours

	for key, counter := range s.errorCounts {
		if counter.LastOccurrence.Before(cutoff) {
			delete(s.errorCounts, key)
		}
	}

	s.logger.Debug("Cleaned up old error monitoring data")
}

// Alert sending methods (placeholder implementations)

func (s *ErrorMonitoringService) sendEmailAlert(alert *ErrorAlert) {
	// Placeholder for email alert implementation
	s.logger.Info("Email alert would be sent",
		zap.String("alert_id", alert.ID),
		zap.String("message", alert.Message),
	)
}

func (s *ErrorMonitoringService) sendSlackAlert(alert *ErrorAlert) {
	// Placeholder for Slack alert implementation
	s.logger.Info("Slack alert would be sent",
		zap.String("alert_id", alert.ID),
		zap.String("message", alert.Message),
	)
}

func (s *ErrorMonitoringService) sendSMSAlert(alert *ErrorAlert) {
	// Placeholder for SMS alert implementation
	s.logger.Info("SMS alert would be sent",
		zap.String("alert_id", alert.ID),
		zap.String("message", alert.Message),
	)
}

// ErrorStatistics holds error statistics
type ErrorStatistics struct {
	TimeWindow   time.Duration  `json:"time_window"`
	GeneratedAt  time.Time      `json:"generated_at"`
	TotalErrors  int            `json:"total_errors"`
	ErrorRate    float64        `json:"error_rate"`
	ErrorsByType map[string]int `json:"errors_by_type"`
	ErrorsByPath map[string]int `json:"errors_by_path"`
}
