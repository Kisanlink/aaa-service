package actions

import (
	"context"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// ActionRepository handles database operations for Action entities
type ActionRepository struct {
	*base.BaseFilterableRepository[*models.Action]
	dbManager db.DBManager
}

// NewActionRepository creates a new ActionRepository instance
func NewActionRepository(dbManager db.DBManager) *ActionRepository {
	baseRepo := base.NewBaseFilterableRepository[*models.Action]()
	baseRepo.SetDBManager(dbManager) // Connect the base repository to the actual database
	return &ActionRepository{
		BaseFilterableRepository: baseRepo,
		dbManager:                dbManager,
	}
}

// ActionRepositoryInterface defines the contract for action repository operations
type ActionRepositoryInterface interface {
	// Create operations
	Create(ctx context.Context, action *models.Action) error
	CreateBatch(ctx context.Context, actions []*models.Action) error
	CreateStaticAction(ctx context.Context, name, description, category string) (*models.Action, error)
	CreateDynamicAction(ctx context.Context, name, description, category, serviceID string) (*models.Action, error)

	// Read operations
	GetByID(ctx context.Context, id string) (*models.Action, error)
	GetByName(ctx context.Context, name string) (*models.Action, error)
	List(ctx context.Context, filter *base.Filter) ([]*models.Action, error)
	Count(ctx context.Context, filter *base.Filter) (int64, error)
	Exists(ctx context.Context, id string) (bool, error)
	GetByCategory(ctx context.Context, category string, limit, offset int) ([]*models.Action, error)
	GetStaticActions(ctx context.Context, limit, offset int) ([]*models.Action, error)
	GetDynamicActions(ctx context.Context, limit, offset int) ([]*models.Action, error)
	GetByServiceID(ctx context.Context, serviceID string, limit, offset int) ([]*models.Action, error)
	GetActiveActions(ctx context.Context, limit, offset int) ([]*models.Action, error)

	// Update operations
	Update(ctx context.Context, action *models.Action) error
	UpdateName(ctx context.Context, id, name string) error
	UpdateDescription(ctx context.Context, id, description string) error
	UpdateCategory(ctx context.Context, id, category string) error
	Activate(ctx context.Context, id string) error
	Deactivate(ctx context.Context, id string) error

	// Delete operations
	Delete(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string, deletedBy string) error
	Restore(ctx context.Context, id string) error
	DeleteBatch(ctx context.Context, ids []string) error
	SoftDeleteBatch(ctx context.Context, ids []string, deletedBy string) error
}
