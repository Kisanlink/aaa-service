package services

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TupleWriter processes events and writes tuples to PostgreSQL
type TupleWriter struct {
	dbManager    *db.DatabaseManager
	compiler     *TupleCompiler
	logger       *zap.Logger
	eventCursor  int64
	cursorMutex  sync.RWMutex
	stopChan     chan struct{}
	isRunning    bool
	runningMutex sync.RWMutex
}

// NewTupleWriter creates a new TupleWriter instance
func NewTupleWriter(dbManager *db.DatabaseManager, logger *zap.Logger) *TupleWriter {
	// Get the PostgreSQL manager to pass to the compiler
	postgresManager := dbManager.GetPostgresManager()
	if postgresManager == nil {
		logger.Error("PostgreSQL manager not available for TupleWriter")
		return nil
	}

	// Get GORM DB for the compiler
	gormDB, err := postgresManager.GetDB(context.Background(), false)
	if err != nil {
		logger.Error("Failed to get database connection for TupleWriter", zap.Error(err))
		return nil
	}

	compiler := NewTupleCompiler(gormDB, logger)

	return &TupleWriter{
		dbManager: dbManager,
		compiler:  compiler,
		logger:    logger,
		stopChan:  make(chan struct{}),
	}
}

// Start begins processing events
func (tw *TupleWriter) Start(ctx context.Context) error {
	tw.runningMutex.Lock()
	if tw.isRunning {
		tw.runningMutex.Unlock()
		return fmt.Errorf("tuple writer is already running")
	}
	tw.isRunning = true
	tw.runningMutex.Unlock()

	// Load cursor position
	if err := tw.loadCursor(ctx); err != nil {
		tw.logger.Warn("Failed to load cursor, starting from beginning", zap.Error(err))
		tw.eventCursor = 0
	}

	// Start processing loop
	go tw.processLoop(ctx)

	tw.logger.Info("Tuple writer started", zap.Int64("cursor", tw.eventCursor))
	return nil
}

// Stop stops processing events
func (tw *TupleWriter) Stop() {
	tw.runningMutex.Lock()
	defer tw.runningMutex.Unlock()

	if !tw.isRunning {
		return
	}

	close(tw.stopChan)
	tw.isRunning = false
	tw.logger.Info("Tuple writer stopped")
}

// processLoop continuously processes events
func (tw *TupleWriter) processLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second) // Poll every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-tw.stopChan:
			return

		case <-ticker.C:
			if err := tw.processEvents(ctx); err != nil {
				tw.logger.Error("Failed to process events", zap.Error(err))
			}

		case <-ctx.Done():
			return
		}
	}
}

// processEvents processes new events since the last cursor
func (tw *TupleWriter) processEvents(ctx context.Context) error {
	// Get the PostgreSQL manager from the database manager
	postgresManager := tw.dbManager.GetPostgresManager()
	if postgresManager == nil {
		return fmt.Errorf("PostgreSQL manager not available")
	}

	// Get the GORM DB connection
	gormDB, err := postgresManager.GetDB(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Get events since the last cursor
	var events []models.Event
	err = gormDB.WithContext(ctx).
		Where("id > ?", tw.eventCursor).
		Order("id ASC").
		Limit(100).
		Find(&events).Error

	if err != nil {
		return fmt.Errorf("failed to fetch events: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	tw.logger.Debug("Processing events", zap.Int("count", len(events)))

	// Process each event
	for _, event := range events {
		if err := tw.processEvent(ctx, &event); err != nil {
			tw.logger.Error("Failed to process event",
				zap.String("event_id", event.ID),
				zap.Error(err))
			continue
		}

		// Update cursor - convert string ID to int64 for cursor tracking
		tw.cursorMutex.Lock()
		if eventID, err := strconv.ParseInt(event.ID, 10, 64); err == nil && eventID > tw.eventCursor {
			tw.eventCursor = eventID
		}
		tw.cursorMutex.Unlock()
	}

	// Save cursor position
	if err := tw.saveCursor(ctx); err != nil {
		tw.logger.Warn("Failed to save cursor", zap.Error(err))
	}

	return nil
}

// processEvent processes a single event
func (tw *TupleWriter) processEvent(ctx context.Context, event *models.Event) error {
	switch event.Kind {
	case models.EventKindBindingCreated:
		return tw.handleBindingCreated(ctx, event)

	case models.EventKindBindingUpdated:
		return tw.handleBindingUpdated(ctx, event)

	case models.EventKindBindingDeleted:
		return tw.handleBindingDeleted(ctx, event)

	case models.EventKindRoleCreated, models.EventKindRoleUpdated, models.EventKindRoleDeleted:
		return tw.handleRolePermissionChanged(ctx, event)

	default:
		tw.logger.Debug("Skipping unknown event kind", zap.String("kind", string(event.Kind)))
		return nil
	}
}

// handleBindingCreated processes binding creation events
func (tw *TupleWriter) handleBindingCreated(ctx context.Context, event *models.Event) error {
	// This function is no longer needed as we're using PostgreSQL RBAC directly
	// The logic for handling binding creation in PostgreSQL would go here.
	// For now, we'll just log the action.
	tw.logger.Info("Skipping handleBindingCreated as we're using PostgreSQL RBAC directly",
		zap.String("event_id", event.ID))
	return nil
}

// handleBindingUpdated processes binding update events
func (tw *TupleWriter) handleBindingUpdated(ctx context.Context, event *models.Event) error {
	// This function is no longer needed as we're using PostgreSQL RBAC directly
	// The logic for handling binding updates in PostgreSQL would go here.
	// For now, we'll just log the action.
	tw.logger.Info("Skipping handleBindingUpdated as we're using PostgreSQL RBAC directly",
		zap.String("event_id", event.ID))
	return nil
}

// handleBindingDeleted processes binding deletion events
func (tw *TupleWriter) handleBindingDeleted(ctx context.Context, event *models.Event) error {
	// This function is no longer needed as we're using PostgreSQL RBAC directly
	// The logic for handling binding deletion in PostgreSQL would go here.
	// For now, we'll just log the action.
	tw.logger.Info("Skipping handleBindingDeleted as we're using PostgreSQL RBAC directly",
		zap.String("event_id", event.ID))
	return nil
}

// handleRolePermissionChanged processes role permission change events
func (tw *TupleWriter) handleRolePermissionChanged(ctx context.Context, event *models.Event) error {
	// This function is no longer needed as we're using PostgreSQL RBAC directly
	// The logic for handling role permission changes in PostgreSQL would go here.
	// For now, we'll just log the action.
	tw.logger.Info("Skipping handleRolePermissionChanged as we're using PostgreSQL RBAC directly",
		zap.String("event_id", event.ID))
	return nil
}

// loadCursor loads the last processed event cursor
func (tw *TupleWriter) loadCursor(ctx context.Context) error {
	// Get the PostgreSQL manager from the database manager
	postgresManager := tw.dbManager.GetPostgresManager()
	if postgresManager == nil {
		return fmt.Errorf("PostgreSQL manager not available")
	}

	// Get the GORM DB connection
	gormDB, err := postgresManager.GetDB(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// In production, this would load from a persistent store
	// For now, we'll start from the last event
	var lastEvent models.Event
	if err := gormDB.WithContext(ctx).
		Order("id DESC").
		First(&lastEvent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			tw.eventCursor = 0
			return nil
		}
		return err
	}

	// Convert string ID to int64 for cursor tracking
	if eventID, err := strconv.ParseInt(lastEvent.ID, 10, 64); err == nil {
		tw.eventCursor = eventID
	} else {
		tw.eventCursor = 0
	}
	return nil
}

// saveCursor saves the current event cursor
func (tw *TupleWriter) saveCursor(ctx context.Context) error {
	// In production, this would save to a persistent store
	// For now, we'll just log it
	tw.logger.Debug("Saving cursor", zap.Int64("cursor", tw.eventCursor))
	return nil
}
