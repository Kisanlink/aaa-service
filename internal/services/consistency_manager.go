package services

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ConsistencyManager manages consistency requirements for PostgreSQL operations
type ConsistencyManager struct {
	logger *zap.Logger
}

// NewConsistencyManager creates a new ConsistencyManager instance
func NewConsistencyManager(logger *zap.Logger) *ConsistencyManager {
	return &ConsistencyManager{
		logger: logger,
	}
}

// GetConsistency determines the appropriate consistency requirement
func (cm *ConsistencyManager) GetConsistency(token string, useStrict bool) (string, error) {
	// For PostgreSQL, we'll use transaction isolation levels
	if useStrict {
		return "SERIALIZABLE", nil
	}

	// If a token is provided, use it for consistency tracking
	if token != "" {
		return token, nil
	}

	// Default to read committed for better performance
	return "READ COMMITTED", nil
}

// GetCurrentToken gets the current consistency token from PostgreSQL
func (cm *ConsistencyManager) GetCurrentToken(ctx context.Context) (string, error) {
	// For PostgreSQL, we'll use a timestamp-based token
	token := fmt.Sprintf("pg_%d", time.Now().UnixNano())
	cm.logger.Debug("Generated PostgreSQL consistency token", zap.String("token", token))
	return token, nil
}

// WaitForConsistency waits until a write is fully consistent
func (cm *ConsistencyManager) WaitForConsistency(ctx context.Context, token string, timeout time.Duration) error {
	if token == "" {
		return fmt.Errorf("token is required for waiting on consistency")
	}

	// For PostgreSQL, we'll simulate consistency by waiting a short time
	// In a real implementation, you might check transaction logs or use WAL
	select {
	case <-time.After(100 * time.Millisecond):
		cm.logger.Debug("Consistency wait completed", zap.String("token", token))
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// WriteWithConsistency performs a write operation and optionally waits for consistency
func (cm *ConsistencyManager) WriteWithConsistency(ctx context.Context,
	writeFunc func() (string, error), waitForConsistency bool) (string, error) {

	// Perform the write operation
	token, err := writeFunc()
	if err != nil {
		return "", err
	}

	// If strict consistency is requested, wait for it
	if waitForConsistency && token != "" {
		if err := cm.WaitForConsistency(ctx, token, 5*time.Second); err != nil {
			// Log the error but don't fail the operation
			cm.logger.Warn("Failed to wait for consistency",
				zap.String("token", token),
				zap.Error(err))
		}
	}

	return token, nil
}

// ConsistencyMode represents different consistency modes
type ConsistencyMode string

const (
	// ConsistencyModeEventual provides best performance with eventual consistency
	ConsistencyModeEventual ConsistencyMode = "eventual"

	// ConsistencyModeStrict ensures full consistency at the cost of latency
	ConsistencyModeStrict ConsistencyMode = "strict"

	// ConsistencyModeBounded provides consistency within a time bound
	ConsistencyModeBounded ConsistencyMode = "bounded"
)

// GetConsistencyForMode returns the appropriate consistency based on mode
func (cm *ConsistencyManager) GetConsistencyForMode(mode ConsistencyMode, token string) string {
	switch mode {
	case ConsistencyModeStrict:
		return "SERIALIZABLE"

	case ConsistencyModeBounded:
		if token != "" {
			return token
		}
		// Fall through to eventual if no token
		fallthrough

	case ConsistencyModeEventual:
		fallthrough

	default:
		return "READ COMMITTED"
	}
}

// DetermineConsistencyMode determines the appropriate consistency mode based on the resource
func (cm *ConsistencyManager) DetermineConsistencyMode(resourceType string, isCritical bool) ConsistencyMode {
	// Critical operations always use strict consistency
	if isCritical {
		return ConsistencyModeStrict
	}

	// Certain resource types require strict consistency
	switch resourceType {
	case "aaa/organization", "aaa/role", "aaa/binding":
		return ConsistencyModeStrict

	case "aaa/user", "aaa/group", "aaa/permission":
		return ConsistencyModeBounded

	default:
		return ConsistencyModeEventual
	}
}

// ConsistencyConfig holds configuration for consistency management
type ConsistencyConfig struct {
	DefaultMode           ConsistencyMode
	StrictTimeout         time.Duration
	BoundedTimeout        time.Duration
	CriticalResourceTypes []string
}

// NewConsistencyConfig creates a default consistency configuration
func NewConsistencyConfig() *ConsistencyConfig {
	return &ConsistencyConfig{
		DefaultMode:    ConsistencyModeEventual,
		StrictTimeout:  5 * time.Second,
		BoundedTimeout: 2 * time.Second,
		CriticalResourceTypes: []string{
			"aaa/organization",
			"aaa/role",
			"aaa/binding",
		},
	}
}

// IsCriticalResource checks if a resource type is critical
func (cc *ConsistencyConfig) IsCriticalResource(resourceType string) bool {
	for _, critical := range cc.CriticalResourceTypes {
		if critical == resourceType {
			return true
		}
	}
	return false
}
