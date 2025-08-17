package events

import (
	"context"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// EventRepository handles database operations for Event entities
type EventRepository struct {
	*base.BaseFilterableRepository[*models.Event]
	dbManager db.DBManager
}

// NewEventRepository creates a new EventRepository instance
func NewEventRepository(dbManager db.DBManager) *EventRepository {
	baseRepo := base.NewBaseFilterableRepository[*models.Event]()
	baseRepo.SetDBManager(dbManager)
	return &EventRepository{
		BaseFilterableRepository: baseRepo,
		dbManager:                dbManager,
	}
}

// Create creates a new event using the base repository
func (r *EventRepository) Create(ctx context.Context, event *models.Event) error {
	return r.BaseFilterableRepository.Create(ctx, event)
}

// GetByID retrieves an event by ID using the base repository
func (r *EventRepository) GetByID(ctx context.Context, id string) (*models.Event, error) {
	event := &models.Event{}
	return r.BaseFilterableRepository.GetByID(ctx, id, event)
}

// Update updates an existing event using the base repository
func (r *EventRepository) Update(ctx context.Context, event *models.Event) error {
	return r.BaseFilterableRepository.Update(ctx, event)
}

// Delete deletes an event by ID using the base repository
func (r *EventRepository) Delete(ctx context.Context, id string) error {
	event := &models.Event{}
	return r.BaseFilterableRepository.Delete(ctx, id, event)
}

// List retrieves events with pagination using database-level filtering
func (r *EventRepository) List(ctx context.Context, limit, offset int) ([]*models.Event, error) {
	filter := base.NewFilterBuilder().
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Count returns the total number of events using database-level counting
func (r *EventRepository) Count(ctx context.Context) (int64, error) {
	filter := base.NewFilter()
	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// Exists checks if an event exists by ID using the base repository
func (r *EventRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.Exists(ctx, id)
}

// SoftDelete soft deletes an event by ID using the base repository
func (r *EventRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	return r.BaseFilterableRepository.SoftDelete(ctx, id, deletedBy)
}

// Restore restores a soft-deleted event using the base repository
func (r *EventRepository) Restore(ctx context.Context, id string) error {
	return r.BaseFilterableRepository.Restore(ctx, id)
}

// ListWithDeleted retrieves events including soft-deleted ones using the base repository
func (r *EventRepository) ListWithDeleted(ctx context.Context, limit, offset int) ([]*models.Event, error) {
	return r.BaseFilterableRepository.ListWithDeleted(ctx, limit, offset)
}

// CountWithDeleted returns count including soft-deleted events using the base repository
func (r *EventRepository) CountWithDeleted(ctx context.Context) (int64, error) {
	return r.BaseFilterableRepository.CountWithDeleted(ctx)
}

// ExistsWithDeleted checks if event exists including soft-deleted ones using the base repository
func (r *EventRepository) ExistsWithDeleted(ctx context.Context, id string) (bool, error) {
	return r.BaseFilterableRepository.ExistsWithDeleted(ctx, id)
}

// GetByCreatedBy gets events by creator using the base repository
func (r *EventRepository) GetByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*models.Event, error) {
	return r.BaseFilterableRepository.GetByCreatedBy(ctx, createdBy, limit, offset)
}

// GetByUpdatedBy gets events by updater using the base repository
func (r *EventRepository) GetByUpdatedBy(ctx context.Context, updatedBy string, limit, offset int) ([]*models.Event, error) {
	return r.BaseFilterableRepository.GetByUpdatedBy(ctx, updatedBy, limit, offset)
}

// GetByType retrieves events by type
func (r *EventRepository) GetByType(ctx context.Context, eventType string, limit, offset int) ([]*models.Event, error) {
	filter := base.NewFilterBuilder().
		Where("type", "=", eventType).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByService retrieves events by service name
func (r *EventRepository) GetByService(ctx context.Context, serviceName string, limit, offset int) ([]*models.Event, error) {
	filter := base.NewFilterBuilder().
		Where("service_name", "=", serviceName).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByUserID retrieves events by user ID
func (r *EventRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Event, error) {
	filter := base.NewFilterBuilder().
		Where("user_id", "=", userID).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByDateRange retrieves events within a date range
func (r *EventRepository) GetByDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*models.Event, error) {
	filter := base.NewFilterBuilder().
		Where("created_at", ">=", startDate).
		Where("created_at", "<=", endDate).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}
