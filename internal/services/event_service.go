package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// EventService manages the immutable audit event chain
type EventService struct {
	db            *gorm.DB
	logger        *zap.Logger
	sequenceMutex sync.Mutex
	lastHash      string
	lastSequence  int64
}

// NewEventService creates a new EventService instance
func NewEventService(db *gorm.DB, logger *zap.Logger) *EventService {
	es := &EventService{
		db:     db,
		logger: logger,
	}

	// Load last event to initialize sequence and hash
	if err := es.loadLastEvent(context.Background()); err != nil {
		logger.Warn("Failed to load last event", zap.Error(err))
	}

	return es
}

// CreateEvent creates a new event in the chain
func (es *EventService) CreateEvent(ctx context.Context, actorID, actorType string,
	kind models.EventKind, resourceType, resourceID string,
	organizationID *string, payload models.EventPayload) (*models.Event, error) {

	es.sequenceMutex.Lock()
	defer es.sequenceMutex.Unlock()

	// Create the event
	event := models.NewEvent(actorID, actorType, kind, resourceType, resourceID, payload)
	event.OrganizationID = organizationID
	event.SequenceNum = es.lastSequence + 1

	// Set previous hash
	if es.lastHash != "" {
		event.PrevHash = &es.lastHash
	}

	// Compute hash for this event
	if err := event.SetHash(); err != nil {
		return nil, fmt.Errorf("failed to compute event hash: %w", err)
	}

	// Extract request metadata from context
	if requestID := getRequestID(ctx); requestID != "" {
		event.RequestID = &requestID
	}
	if sourceIP := getSourceIP(ctx); sourceIP != "" {
		event.SourceIP = &sourceIP
	}
	if userAgent := getUserAgent(ctx); userAgent != "" {
		event.UserAgent = &userAgent
	}

	// Save to database
	if err := es.db.WithContext(ctx).Create(event).Error; err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	// Update last event tracking
	es.lastHash = event.Hash
	es.lastSequence = event.SequenceNum

	es.logger.Info("Event created",
		zap.String("event_id", event.ID),
		zap.String("kind", string(event.Kind)),
		zap.String("resource", fmt.Sprintf("%s/%s", resourceType, resourceID)),
		zap.Int64("sequence", event.SequenceNum))

	return event, nil
}

// CreateBindingEvent creates an event for binding operations
func (es *EventService) CreateBindingEvent(ctx context.Context, binding *models.Binding,
	action string, actorID string) (*models.Event, error) {

	var kind models.EventKind
	switch action {
	case "CREATE":
		kind = models.EventKindBindingCreated
	case "UPDATE":
		kind = models.EventKindBindingUpdated
	case "DELETE":
		kind = models.EventKindBindingDeleted
	case "ROLLBACK":
		kind = models.EventKindBindingRolledBack
	default:
		return nil, fmt.Errorf("unknown binding action: %s", action)
	}

	payload := models.EventPayload{
		"binding_id":    binding.ID,
		"subject_id":    binding.SubjectID,
		"subject_type":  binding.SubjectType,
		"binding_type":  binding.BindingType,
		"resource_type": binding.ResourceType,
	}

	if binding.ResourceID != nil {
		payload["resource_id"] = *binding.ResourceID
	}
	if binding.RoleID != nil {
		payload["role_id"] = *binding.RoleID
	}
	if binding.PermissionID != nil {
		payload["permission_id"] = *binding.PermissionID
	}
	if binding.Caveat != nil {
		payload["caveat"] = *binding.Caveat
	}

	// For delete events, include the full binding data
	if action == "DELETE" {
		payload["binding"] = binding
	}

	return es.CreateEvent(ctx, actorID, "user", kind,
		binding.ResourceType, binding.ID, &binding.OrganizationID, payload)
}

// CreateGroupEvent creates an event for group operations
func (es *EventService) CreateGroupEvent(ctx context.Context, groupID string,
	kind models.EventKind, actorID string, payload models.EventPayload) (*models.Event, error) {

	// Add group ID to payload
	if payload == nil {
		payload = make(models.EventPayload)
	}
	payload["group_id"] = groupID

	return es.CreateEvent(ctx, actorID, "user", kind, "aaa/group", groupID, nil, payload)
}

// CreateResourceEvent creates an event for resource operations
func (es *EventService) CreateResourceEvent(ctx context.Context, resource *models.Resource,
	kind models.EventKind, actorID string, additionalPayload map[string]interface{}) (*models.Event, error) {

	payload := models.EventPayload{
		"resource_id":   resource.ID,
		"resource_name": resource.Name,
		"resource_type": resource.Type,
	}

	if resource.ParentID != nil {
		payload["parent_id"] = *resource.ParentID
	}
	if resource.OwnerID != nil {
		payload["owner_id"] = *resource.OwnerID
	}

	// Merge additional payload
	for k, v := range additionalPayload {
		payload[k] = v
	}

	var orgID *string
	if resource.Type == "aaa/organization" {
		orgID = &resource.ID
	}

	return es.CreateEvent(ctx, actorID, "user", kind, resource.Type, resource.ID, orgID, payload)
}

// VerifyEventChain verifies the integrity of the event chain
func (es *EventService) VerifyEventChain(ctx context.Context, startSequence, endSequence int64) (bool, []string, error) {
	errors := []string{}

	// Load events in sequence order
	var events []models.Event
	if err := es.db.WithContext(ctx).
		Where("sequence_num >= ? AND sequence_num <= ?", startSequence, endSequence).
		Order("sequence_num ASC").
		Find(&events).Error; err != nil {
		return false, nil, fmt.Errorf("failed to load events: %w", err)
	}

	if len(events) == 0 {
		return true, []string{}, nil // No events to verify
	}

	// Verify each event
	var prevHash *string
	for i, event := range events {
		// Verify sequence number
		if i > 0 && event.SequenceNum != events[i-1].SequenceNum+1 {
			errors = append(errors, fmt.Sprintf("Sequence gap at %d (expected %d)",
				event.SequenceNum, events[i-1].SequenceNum+1))
		}

		// Verify previous hash linkage
		if event.PrevHash != nil && prevHash != nil {
			if *event.PrevHash != *prevHash {
				errors = append(errors, fmt.Sprintf("Hash chain broken at sequence %d", event.SequenceNum))
			}
		}

		// Verify event hash
		if err := event.VerifyHash(); err != nil {
			errors = append(errors, fmt.Sprintf("Invalid hash at sequence %d: %v", event.SequenceNum, err))
		}

		prevHash = &event.Hash
	}

	isValid := len(errors) == 0

	es.logger.Info("Event chain verification completed",
		zap.Int64("start_sequence", startSequence),
		zap.Int64("end_sequence", endSequence),
		zap.Bool("valid", isValid),
		zap.Int("error_count", len(errors)))

	return isValid, errors, nil
}

// CreateCheckpoint creates a checkpoint of the event chain
func (es *EventService) CreateCheckpoint(ctx context.Context, createdByID string) (*models.EventCheckpoint, error) {
	es.sequenceMutex.Lock()
	defer es.sequenceMutex.Unlock()

	// Get the last event
	var lastEvent models.Event
	if err := es.db.WithContext(ctx).
		Order("sequence_num DESC").
		First(&lastEvent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no events to checkpoint")
		}
		return nil, fmt.Errorf("failed to get last event: %w", err)
	}

	// Count total events
	var eventCount int64
	if err := es.db.WithContext(ctx).
		Model(&models.Event{}).
		Count(&eventCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count events: %w", err)
	}

	// Calculate Merkle root (simplified - in production, build full Merkle tree)
	merkleRoot := es.calculateMerkleRoot(ctx, lastEvent.SequenceNum)

	// Create checkpoint
	checkpoint := models.NewEventCheckpoint(
		lastEvent.ID,
		lastEvent.SequenceNum,
		lastEvent.Hash,
		merkleRoot,
		eventCount,
		createdByID,
	)

	if err := es.db.WithContext(ctx).Create(checkpoint).Error; err != nil {
		return nil, fmt.Errorf("failed to create checkpoint: %w", err)
	}

	es.logger.Info("Event checkpoint created",
		zap.String("checkpoint_id", checkpoint.ID),
		zap.Int64("last_sequence", lastEvent.SequenceNum),
		zap.Int64("event_count", eventCount))

	return checkpoint, nil
}

// GetEvents retrieves events with filters
func (es *EventService) GetEvents(ctx context.Context, filters EventFilters) ([]*models.Event, error) {
	query := es.db.WithContext(ctx)

	if filters.ActorID != "" {
		query = query.Where("actor_id = ?", filters.ActorID)
	}
	if filters.ResourceType != "" {
		query = query.Where("resource_type = ?", filters.ResourceType)
	}
	if filters.ResourceID != "" {
		query = query.Where("resource_id = ?", filters.ResourceID)
	}
	if filters.OrganizationID != "" {
		query = query.Where("organization_id = ?", filters.OrganizationID)
	}
	if len(filters.Kinds) > 0 {
		query = query.Where("kind IN ?", filters.Kinds)
	}
	if !filters.StartTime.IsZero() {
		query = query.Where("occurred_at >= ?", filters.StartTime)
	}
	if !filters.EndTime.IsZero() {
		query = query.Where("occurred_at <= ?", filters.EndTime)
	}

	// Apply ordering
	if filters.OrderBy == "" {
		filters.OrderBy = "sequence_num DESC"
	}
	query = query.Order(filters.OrderBy)

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var events []*models.Event
	if err := query.Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	return events, nil
}

// ReplayEvents replays events to rebuild state
func (es *EventService) ReplayEvents(ctx context.Context, upToTime time.Time,
	resourceTypes []string, handler EventHandler) error {

	query := es.db.WithContext(ctx).
		Where("occurred_at <= ?", upToTime).
		Order("sequence_num ASC")

	if len(resourceTypes) > 0 {
		query = query.Where("resource_type IN ?", resourceTypes)
	}

	// Process in batches
	const batchSize = 100
	offset := 0
	eventsProcessed := 0

	for {
		var events []models.Event
		if err := query.Offset(offset).Limit(batchSize).Find(&events).Error; err != nil {
			return fmt.Errorf("failed to load events batch: %w", err)
		}

		if len(events) == 0 {
			break // No more events
		}

		// Process each event
		for _, event := range events {
			if err := handler(ctx, &event); err != nil {
				es.logger.Error("Failed to replay event",
					zap.String("event_id", event.ID),
					zap.Int64("sequence", event.SequenceNum),
					zap.Error(err))
				// Continue with other events
			}
			eventsProcessed++
		}

		offset += batchSize
	}

	es.logger.Info("Event replay completed",
		zap.Time("up_to", upToTime),
		zap.Int("events_processed", eventsProcessed))

	return nil
}

// Helper functions

func (es *EventService) loadLastEvent(ctx context.Context) error {
	var lastEvent models.Event
	if err := es.db.WithContext(ctx).
		Order("sequence_num DESC").
		First(&lastEvent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// No events yet
			es.lastSequence = 0
			es.lastHash = ""
			return nil
		}
		return err
	}

	es.lastSequence = lastEvent.SequenceNum
	es.lastHash = lastEvent.Hash
	return nil
}

func (es *EventService) calculateMerkleRoot(ctx context.Context, upToSequence int64) string {
	// Simplified Merkle root calculation
	// In production, build a proper Merkle tree

	var hashes []string
	es.db.WithContext(ctx).
		Model(&models.Event{}).
		Where("sequence_num <= ?", upToSequence).
		Order("sequence_num ASC").
		Pluck("hash", &hashes)

	if len(hashes) == 0 {
		return ""
	}

	// Concatenate all hashes and compute final hash
	combined := ""
	for _, h := range hashes {
		combined += h
	}

	hash := sha256.Sum256([]byte(combined))
	return fmt.Sprintf("%x", hash)
}

// EventFilters defines filters for querying events
type EventFilters struct {
	ActorID        string
	ResourceType   string
	ResourceID     string
	OrganizationID string
	Kinds          []models.EventKind
	StartTime      time.Time
	EndTime        time.Time
	OrderBy        string
	Limit          int
	Offset         int
}

// EventHandler is a function that processes an event during replay
type EventHandler func(ctx context.Context, event *models.Event) error

// Context helpers for request metadata
func getRequestID(ctx context.Context) string {
	if val := ctx.Value("request_id"); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getSourceIP(ctx context.Context) string {
	if val := ctx.Value("source_ip"); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getUserAgent(ctx context.Context) string {
	if val := ctx.Value("user_agent"); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
