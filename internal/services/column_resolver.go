package services

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/gorm"
)

// ColumnResolver handles column-level authorization
type ColumnResolver struct {
	db     *gorm.DB
	logger *zap.Logger
	cache  map[string]*ColumnCache // Simple in-memory cache
}

// ColumnCache caches column group information
type ColumnCache struct {
	Groups  map[string]*models.ColumnGroup
	Columns map[string][]string // column name -> group names
}

// ColumnGroupAccess represents column group access information
type ColumnGroupAccess struct {
	Name        string
	Description string
	Columns     []string
}

// NewColumnResolver creates a new ColumnResolver instance
func NewColumnResolver(db *gorm.DB, logger *zap.Logger) *ColumnResolver {
	return &ColumnResolver{
		db:     db,
		logger: logger,
		cache:  make(map[string]*ColumnCache),
	}
}

// CheckColumns checks if a principal can access the requested columns
func (cr *ColumnResolver) CheckColumns(ctx context.Context, principalID, tableName, resourceID string,
	requestedColumns []string, action string, attributes *structpb.Struct) (bool, []string, []string, []string, error) {

	// Load column groups for the table
	tableCache, err := cr.loadTableCache(ctx, tableName)
	if err != nil {
		return false, nil, nil, nil, fmt.Errorf("failed to load table cache: %w", err)
	}

	// Get principal's column groups for this table
	allowedGroups, err := cr.getPrincipalColumnGroups(ctx, principalID, tableName, resourceID, action)
	if err != nil {
		return false, nil, nil, nil, fmt.Errorf("failed to get principal column groups: %w", err)
	}

	// Build allowed column set
	allowedColumns := make(map[string]bool)
	for _, groupName := range allowedGroups {
		if group, exists := tableCache.Groups[groupName]; exists {
			for _, member := range group.ColumnMembers {
				if member.IsActive {
					allowedColumns[member.ColumnName] = true
				}
			}
		}
	}

	// Check requested columns
	allowed := []string{}
	denied := []string{}

	for _, col := range requestedColumns {
		if allowedColumns[col] {
			allowed = append(allowed, col)
		} else {
			denied = append(denied, col)
		}
	}

	// Overall decision
	allAllowed := len(denied) == 0

	cr.logger.Info("Column check completed",
		zap.String("principal", principalID),
		zap.String("table", tableName),
		zap.Int("requested", len(requestedColumns)),
		zap.Int("allowed", len(allowed)),
		zap.Int("denied", len(denied)),
		zap.Bool("result", allAllowed))

	return allAllowed, allowed, denied, allowedGroups, nil
}

// ListAllowedColumns lists all columns a principal can access
func (cr *ColumnResolver) ListAllowedColumns(ctx context.Context, principalID, tableName, resourceID string,
	action string, attributes *structpb.Struct) ([]string, []*ColumnGroupAccess, error) {

	// Load column groups for the table
	tableCache, err := cr.loadTableCache(ctx, tableName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load table cache: %w", err)
	}

	// Get principal's column groups for this table
	allowedGroups, err := cr.getPrincipalColumnGroups(ctx, principalID, tableName, resourceID, action)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get principal column groups: %w", err)
	}

	// Build allowed column set and group access info
	allowedColumns := make(map[string]bool)
	groupAccess := make([]*ColumnGroupAccess, 0, len(allowedGroups))

	for _, groupName := range allowedGroups {
		if group, exists := tableCache.Groups[groupName]; exists {
			columns := []string{}
			for _, member := range group.ColumnMembers {
				if member.IsActive {
					allowedColumns[member.ColumnName] = true
					columns = append(columns, member.ColumnName)
				}
			}

			groupAccess = append(groupAccess, &ColumnGroupAccess{
				Name:        group.Name,
				Description: group.Description,
				Columns:     columns,
			})
		}
	}

	// Convert map to slice
	columnList := make([]string, 0, len(allowedColumns))
	for col := range allowedColumns {
		columnList = append(columnList, col)
	}

	cr.logger.Info("Listed allowed columns",
		zap.String("principal", principalID),
		zap.String("table", tableName),
		zap.Int("column_count", len(columnList)),
		zap.Int("group_count", len(groupAccess)))

	return columnList, groupAccess, nil
}

// loadTableCache loads or retrieves cached column group information for a table
func (cr *ColumnResolver) loadTableCache(ctx context.Context, tableName string) (*ColumnCache, error) {
	// Check cache first
	if cache, exists := cr.cache[tableName]; exists {
		return cache, nil
	}

	// Load from database
	var columnGroups []models.ColumnGroup
	if err := cr.db.WithContext(ctx).
		Preload("ColumnMembers").
		Where("table = ? AND is_active = ?", tableName, true).
		Find(&columnGroups).Error; err != nil {
		return nil, fmt.Errorf("failed to load column groups: %w", err)
	}

	// Build cache
	cache := &ColumnCache{
		Groups:  make(map[string]*models.ColumnGroup),
		Columns: make(map[string][]string),
	}

	for i := range columnGroups {
		group := &columnGroups[i]
		cache.Groups[group.Name] = group

		// Build reverse index: column -> groups
		for _, member := range group.ColumnMembers {
			if member.IsActive {
				if _, exists := cache.Columns[member.ColumnName]; !exists {
					cache.Columns[member.ColumnName] = []string{}
				}
				cache.Columns[member.ColumnName] = append(cache.Columns[member.ColumnName], group.Name)
			}
		}
	}

	// Store in cache
	cr.cache[tableName] = cache

	cr.logger.Debug("Loaded table cache",
		zap.String("table", tableName),
		zap.Int("group_count", len(cache.Groups)),
		zap.Int("column_count", len(cache.Columns)))

	return cache, nil
}

// getPrincipalColumnGroups gets the column groups a principal has access to
func (cr *ColumnResolver) getPrincipalColumnGroups(ctx context.Context,
	principalID, tableName, resourceID string, action string) ([]string, error) {

	// This would typically query SpiceDB or the binding table to determine
	// which column groups the principal has access to for this table

	// For now, we'll query bindings directly
	var bindings []models.Binding
	query := cr.db.WithContext(ctx).
		Where("subject_id = ? AND resource_type = ? AND is_active = ?",
			principalID, "aaa/table", true)

	if resourceID != "" {
		query = query.Where("(resource_id = ? OR resource_id IS NULL)", resourceID)
	}

	if err := query.Find(&bindings).Error; err != nil {
		return nil, fmt.Errorf("failed to load bindings: %w", err)
	}

	// Extract column groups from caveats
	groupSet := make(map[string]bool)

	for _, binding := range bindings {
		if binding.Caveat == nil {
			continue
		}

		// Check for column groups in caveat
		if columnGroups, hasColumns := (*binding.Caveat)["column_groups"]; hasColumns {
			switch groups := columnGroups.(type) {
			case []interface{}:
				for _, g := range groups {
					if groupName, ok := g.(string); ok {
						groupSet[groupName] = true
					}
				}
			case []string:
				for _, groupName := range groups {
					groupSet[groupName] = true
				}
			}
		}
	}

	// Also check group memberships for inherited access
	var groupMemberships []models.GroupMembership
	if err := cr.db.WithContext(ctx).
		Preload("Group").
		Where("principal_id = ? AND principal_type = ? AND is_active = ?",
			principalID, "user", true).
		Find(&groupMemberships).Error; err != nil {
		cr.logger.Warn("Failed to load group memberships", zap.Error(err))
	}

	// Check bindings for groups
	for _, membership := range groupMemberships {
		var groupBindings []models.Binding
		if err := cr.db.WithContext(ctx).
			Where("subject_id = ? AND subject_type = ? AND resource_type = ? AND is_active = ?",
				membership.GroupID, "group", "aaa/table", true).
			Find(&groupBindings).Error; err != nil {
			continue
		}

		for _, binding := range groupBindings {
			if binding.Caveat == nil {
				continue
			}

			// Extract column groups
			if columnGroups, hasColumns := (*binding.Caveat)["column_groups"]; hasColumns {
				switch groups := columnGroups.(type) {
				case []interface{}:
					for _, g := range groups {
						if groupName, ok := g.(string); ok {
							groupSet[groupName] = true
						}
					}
				case []string:
					for _, groupName := range groups {
						groupSet[groupName] = true
					}
				}
			}
		}
	}

	// Convert set to slice
	groups := make([]string, 0, len(groupSet))
	for group := range groupSet {
		groups = append(groups, group)
	}

	return groups, nil
}

// InvalidateCache invalidates the cache for a table
func (cr *ColumnResolver) InvalidateCache(tableName string) {
	delete(cr.cache, tableName)
	cr.logger.Debug("Invalidated table cache", zap.String("table", tableName))
}

// InvalidateAllCache invalidates all cached data
func (cr *ColumnResolver) InvalidateAllCache() {
	cr.cache = make(map[string]*ColumnCache)
	cr.logger.Debug("Invalidated all table caches")
}

// GetColumnGroupsForTable gets all column groups for a table
func (cr *ColumnResolver) GetColumnGroupsForTable(ctx context.Context, tableName string) ([]*models.ColumnGroup, error) {
	cache, err := cr.loadTableCache(ctx, tableName)
	if err != nil {
		return nil, err
	}

	groups := make([]*models.ColumnGroup, 0, len(cache.Groups))
	for _, group := range cache.Groups {
		groups = append(groups, group)
	}

	return groups, nil
}

// GetColumnsInGroup gets all columns in a specific group
func (cr *ColumnResolver) GetColumnsInGroup(ctx context.Context, tableName, groupName string) ([]string, error) {
	cache, err := cr.loadTableCache(ctx, tableName)
	if err != nil {
		return nil, err
	}

	group, exists := cache.Groups[groupName]
	if !exists {
		return nil, fmt.Errorf("column group %s not found for table %s", groupName, tableName)
	}

	columns := make([]string, 0, len(group.ColumnMembers))
	for _, member := range group.ColumnMembers {
		if member.IsActive {
			columns = append(columns, member.ColumnName)
		}
	}

	return columns, nil
}

// GetGroupsForColumn gets all groups that contain a specific column
func (cr *ColumnResolver) GetGroupsForColumn(ctx context.Context, tableName, columnName string) ([]string, error) {
	cache, err := cr.loadTableCache(ctx, tableName)
	if err != nil {
		return nil, err
	}

	groups, exists := cache.Columns[columnName]
	if !exists {
		return []string{}, nil
	}

	return groups, nil
}
