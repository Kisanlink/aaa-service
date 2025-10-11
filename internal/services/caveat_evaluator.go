package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/gorm"
)

// CaveatEvaluator evaluates caveats for authorization decisions
type CaveatEvaluator struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewCaveatEvaluator creates a new CaveatEvaluator instance
func NewCaveatEvaluator(db *gorm.DB, logger *zap.Logger) *CaveatEvaluator {
	return &CaveatEvaluator{
		db:     db,
		logger: logger,
	}
}

// EvaluateTimeCaveat evaluates time-based caveats
func (ce *CaveatEvaluator) EvaluateTimeCaveat(ctx context.Context, startsAt, endsAt *time.Time, checkTime time.Time) (bool, string) {
	// Check if before start time
	if startsAt != nil && checkTime.Before(*startsAt) {
		return false, fmt.Sprintf("Access not yet active (starts at %s)", startsAt.Format(time.RFC3339))
	}

	// Check if after end time
	if endsAt != nil && checkTime.After(*endsAt) {
		return false, fmt.Sprintf("Access expired (ended at %s)", endsAt.Format(time.RFC3339))
	}

	return true, "Time caveat satisfied"
}

// EvaluateAttributeCaveat evaluates attribute-based caveats
func (ce *CaveatEvaluator) EvaluateAttributeCaveat(ctx context.Context, requiredAttrs, providedAttrs map[string]interface{}) (bool, string) {
	if len(requiredAttrs) == 0 {
		return true, "No attributes required"
	}

	if len(providedAttrs) == 0 {
		return false, "Required attributes not provided"
	}

	// Check each required attribute
	for key, requiredValue := range requiredAttrs {
		providedValue, exists := providedAttrs[key]
		if !exists {
			return false, fmt.Sprintf("Missing required attribute: %s", key)
		}

		// Compare values (simple equality for now, could be enhanced)
		if !ce.compareAttributeValues(requiredValue, providedValue) {
			return false, fmt.Sprintf("Attribute mismatch: %s (required: %v, provided: %v)",
				key, requiredValue, providedValue)
		}
	}

	return true, "All required attributes satisfied"
}

// EvaluateColumnCaveat evaluates column-based caveats
func (ce *CaveatEvaluator) EvaluateColumnCaveat(ctx context.Context,
	allowedGroups []string, requestedColumns []string, tableName string) (bool, []string, []string, string) {

	if len(requestedColumns) == 0 {
		return true, []string{}, []string{}, "No columns requested"
	}

	if len(allowedGroups) == 0 {
		return false, []string{}, requestedColumns, "No column groups allowed"
	}

	// Load column groups from database
	allowedColumns := make(map[string]bool)
	for _, groupName := range allowedGroups {
		var columnGroup models.ColumnGroup
		if err := ce.db.WithContext(ctx).
			Preload("ColumnMembers").
			Where("name = ? AND table = ?", groupName, tableName).
			First(&columnGroup).Error; err != nil {
			ce.logger.Warn("Failed to load column group",
				zap.String("group", groupName),
				zap.String("table", tableName),
				zap.Error(err))
			continue
		}

		// Add columns from this group to allowed set
		for _, member := range columnGroup.ColumnMembers {
			if member.IsActive {
				allowedColumns[member.ColumnName] = true
			}
		}
	}

	// Check requested columns against allowed
	allowed := []string{}
	denied := []string{}

	for _, col := range requestedColumns {
		if allowedColumns[col] {
			allowed = append(allowed, col)
		} else {
			denied = append(denied, col)
		}
	}

	if len(denied) > 0 {
		return false, allowed, denied, fmt.Sprintf("Access denied to columns: %v", denied)
	}

	return true, allowed, []string{}, "All requested columns allowed"
}

// compareAttributeValues compares two attribute values
func (ce *CaveatEvaluator) compareAttributeValues(required, provided interface{}) bool {
	// Simple equality check for now
	// Could be enhanced to support:
	// - Range comparisons for numbers
	// - Pattern matching for strings
	// - Set membership checks for lists
	return fmt.Sprintf("%v", required) == fmt.Sprintf("%v", provided)
}

// ExtractCaveatContext extracts caveat evaluation context from a struct
func (ce *CaveatEvaluator) ExtractCaveatContext(caveatStruct *structpb.Struct) (map[string]interface{}, error) {
	if caveatStruct == nil {
		return nil, nil
	}

	context := make(map[string]interface{})

	for key, value := range caveatStruct.Fields {
		switch v := value.Kind.(type) {
		case *structpb.Value_StringValue:
			context[key] = v.StringValue
		case *structpb.Value_NumberValue:
			context[key] = v.NumberValue
		case *structpb.Value_BoolValue:
			context[key] = v.BoolValue
		case *structpb.Value_StructValue:
			// Recursively extract nested struct
			nested, err := ce.ExtractCaveatContext(v.StructValue)
			if err != nil {
				return nil, err
			}
			context[key] = nested
		case *structpb.Value_ListValue:
			// Extract list values
			list := make([]interface{}, 0, len(v.ListValue.Values))
			for _, item := range v.ListValue.Values {
				switch iv := item.Kind.(type) {
				case *structpb.Value_StringValue:
					list = append(list, iv.StringValue)
				case *structpb.Value_NumberValue:
					list = append(list, iv.NumberValue)
				case *structpb.Value_BoolValue:
					list = append(list, iv.BoolValue)
				default:
					list = append(list, item.String())
				}
			}
			context[key] = list
		}
	}

	return context, nil
}

// LoadPrincipalAttributes loads attributes for a principal
func (ce *CaveatEvaluator) LoadPrincipalAttributes(ctx context.Context, principalID string) (map[string]interface{}, error) {
	attributes := make(map[string]interface{})

	// Load attributes from database
	var attrs []models.Attribute
	if err := ce.db.WithContext(ctx).
		Where("subject_id = ? AND subject_type = ? AND is_active = ?",
			principalID, models.AttributeSubjectPrincipal, true).
		Find(&attrs).Error; err != nil {
		return nil, fmt.Errorf("failed to load principal attributes: %w", err)
	}

	// Process attributes
	now := time.Now()
	for _, attr := range attrs {
		// Skip expired attributes
		if attr.ExpiresAt != nil && now.After(*attr.ExpiresAt) {
			continue
		}

		// Extract value based on type
		if attr.Value != nil {
			if val, ok := attr.Value["value"]; ok {
				attributes[attr.Key] = val
			} else {
				// Use entire value object if no "value" field
				attributes[attr.Key] = attr.Value
			}
		}
	}

	return attributes, nil
}

// LoadResourceAttributes loads attributes for a resource
func (ce *CaveatEvaluator) LoadResourceAttributes(ctx context.Context, resourceID string) (map[string]interface{}, error) {
	attributes := make(map[string]interface{})

	// Load attributes from database
	var attrs []models.Attribute
	if err := ce.db.WithContext(ctx).
		Where("subject_id = ? AND subject_type = ? AND is_active = ?",
			resourceID, models.AttributeSubjectResource, true).
		Find(&attrs).Error; err != nil {
		return nil, fmt.Errorf("failed to load resource attributes: %w", err)
	}

	// Process attributes
	now := time.Now()
	for _, attr := range attrs {
		// Skip expired attributes
		if attr.ExpiresAt != nil && now.After(*attr.ExpiresAt) {
			continue
		}

		// Extract value
		if attr.Value != nil {
			if val, ok := attr.Value["value"]; ok {
				attributes[attr.Key] = val
			} else {
				attributes[attr.Key] = attr.Value
			}
		}
	}

	return attributes, nil
}

// CaveatResult represents the result of caveat evaluation
type CaveatResult struct {
	Name      string
	Satisfied bool
	Reason    string
	Details   map[string]interface{}
}

// EvaluateAllCaveats evaluates all caveats for a binding
func (ce *CaveatEvaluator) EvaluateAllCaveats(ctx context.Context, binding *models.Binding,
	checkTime time.Time, principalAttrs map[string]interface{}, requestedColumns []string) (*CaveatResult, error) {

	if binding.Caveat == nil {
		return &CaveatResult{
			Name:      "none",
			Satisfied: true,
			Reason:    "No caveats to evaluate",
		}, nil
	}

	// Check time-based caveats
	if startsAt, hasStart := (*binding.Caveat)["starts_at"]; hasStart {
		if endsAt, hasEnd := (*binding.Caveat)["ends_at"]; hasEnd {
			// Parse times
			startTime, err := parseTime(startsAt)
			if err != nil {
				return nil, fmt.Errorf("invalid starts_at: %w", err)
			}

			endTime, err := parseTime(endsAt)
			if err != nil {
				return nil, fmt.Errorf("invalid ends_at: %w", err)
			}

			satisfied, reason := ce.EvaluateTimeCaveat(ctx, startTime, endTime, checkTime)
			if !satisfied {
				return &CaveatResult{
					Name:      "within_time",
					Satisfied: false,
					Reason:    reason,
					Details: map[string]interface{}{
						"starts_at":    startTime,
						"ends_at":      endTime,
						"current_time": checkTime,
					},
				}, nil
			}
		}
	}

	// Check attribute-based caveats
	if requiredAttrs, hasAttrs := (*binding.Caveat)["required_attributes"]; hasAttrs {
		reqAttrs, ok := requiredAttrs.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid required_attributes format")
		}

		satisfied, reason := ce.EvaluateAttributeCaveat(ctx, reqAttrs, principalAttrs)
		if !satisfied {
			return &CaveatResult{
				Name:      "has_attributes",
				Satisfied: false,
				Reason:    reason,
				Details: map[string]interface{}{
					"required": reqAttrs,
					"provided": principalAttrs,
				},
			}, nil
		}
	}

	// Check column-based caveats
	if columnGroups, hasColumns := (*binding.Caveat)["column_groups"]; hasColumns {
		groups, ok := columnGroups.([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid column_groups format")
		}

		// Convert to string slice
		groupNames := make([]string, 0, len(groups))
		for _, g := range groups {
			if groupName, ok := g.(string); ok {
				groupNames = append(groupNames, groupName)
			}
		}

		// Evaluate column access
		satisfied, allowed, denied, reason := ce.EvaluateColumnCaveat(ctx,
			groupNames, requestedColumns, binding.ResourceType)

		if !satisfied {
			return &CaveatResult{
				Name:      "has_columns",
				Satisfied: false,
				Reason:    reason,
				Details: map[string]interface{}{
					"allowed_groups":    groupNames,
					"requested_columns": requestedColumns,
					"allowed_columns":   allowed,
					"denied_columns":    denied,
				},
			}, nil
		}
	}

	// All caveats satisfied
	return &CaveatResult{
		Name:      "all",
		Satisfied: true,
		Reason:    "All caveats satisfied",
	}, nil
}

// parseTime parses a time value from various formats
func parseTime(value interface{}) (*time.Time, error) {
	switch v := value.(type) {
	case string:
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, err
		}
		return &t, nil
	case float64:
		t := time.Unix(int64(v), 0)
		return &t, nil
	case int64:
		t := time.Unix(v, 0)
		return &t, nil
	default:
		return nil, fmt.Errorf("unsupported time format: %T", value)
	}
}
