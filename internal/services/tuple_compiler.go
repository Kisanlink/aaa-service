package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TupleCompiler compiles bindings and roles into PostgreSQL tuples
type TupleCompiler struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewTupleCompiler creates a new TupleCompiler instance
func NewTupleCompiler(db *gorm.DB, logger *zap.Logger) *TupleCompiler {
	return &TupleCompiler{
		db:     db,
		logger: logger,
	}
}

// CompiledTuple represents a PostgreSQL tuple with caveats
type CompiledTuple struct {
	ResourceType string
	ResourceID   string
	Relation     string
	SubjectType  string
	SubjectID    string
	Caveat       *CompiledCaveat
}

// CompiledCaveat represents a PostgreSQL caveat
type CompiledCaveat struct {
	Name    string
	Context map[string]interface{}
}

// CompileBinding compiles a binding into PostgreSQL tuples
func (tc *TupleCompiler) CompileBinding(ctx context.Context, binding *models.Binding) ([]*CompiledTuple, error) {
	tuples := []*CompiledTuple{}

	// Determine the subject
	subjectType, subjectID, err := tc.resolveSubject(ctx, binding)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve subject: %w", err)
	}

	// Compile based on binding type
	switch binding.BindingType {
	case models.BindingTypeRole:
		if binding.RoleID == nil {
			return nil, fmt.Errorf("role_id is required for role bindings")
		}
		roleTuples, err := tc.compileRoleBinding(ctx, binding, subjectType, subjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to compile role binding: %w", err)
		}
		tuples = append(tuples, roleTuples...)

	case models.BindingTypePermission:
		if binding.PermissionID == nil {
			return nil, fmt.Errorf("permission_id is required for permission bindings")
		}
		permTuple, err := tc.compilePermissionBinding(ctx, binding, subjectType, subjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to compile permission binding: %w", err)
		}
		tuples = append(tuples, permTuple)

	default:
		return nil, fmt.Errorf("unknown binding type: %s", binding.BindingType)
	}

	return tuples, nil
}

// resolveSubject determines the PostgreSQL subject from the binding
func (tc *TupleCompiler) resolveSubject(ctx context.Context, binding *models.Binding) (string, string, error) {
	switch binding.SubjectType {
	case models.BindingSubjectUser:
		return "aaa/user", binding.SubjectID, nil

	case models.BindingSubjectGroup:
		return "aaa/group", binding.SubjectID, nil

	case models.BindingSubjectService:
		return "aaa/service", binding.SubjectID, nil

	default:
		return "", "", fmt.Errorf("unknown subject type: %s", binding.SubjectType)
	}
}

// compileRoleBinding compiles a role binding into tuples
func (tc *TupleCompiler) compileRoleBinding(ctx context.Context, binding *models.Binding,
	subjectType, subjectID string) ([]*CompiledTuple, error) {

	// Load the role with permissions
	var role models.Role
	if err := tc.db.WithContext(ctx).Preload("Permissions").
		First(&role, "id = ?", *binding.RoleID).Error; err != nil {
		return nil, fmt.Errorf("failed to load role: %w", err)
	}

	// Map role to resource-specific relations
	relations := tc.mapRoleToRelations(&role, binding.ResourceType)

	tuples := []*CompiledTuple{}
	for _, relation := range relations {
		tuple := &CompiledTuple{
			ResourceType: binding.ResourceType,
			ResourceID:   tc.getResourceID(binding),
			Relation:     relation,
			SubjectType:  subjectType,
			SubjectID:    subjectID,
		}

		// Add caveats if present
		if binding.Caveat != nil {
			caveat, err := tc.compileCaveat(binding.Caveat)
			if err != nil {
				tc.logger.Warn("Failed to compile caveat",
					zap.Error(err),
					zap.String("binding_id", binding.ID))
			} else {
				tuple.Caveat = caveat
			}
		}

		tuples = append(tuples, tuple)
	}

	return tuples, nil
}

// compilePermissionBinding compiles a permission binding into a tuple
func (tc *TupleCompiler) compilePermissionBinding(ctx context.Context, binding *models.Binding,
	subjectType, subjectID string) (*CompiledTuple, error) {

	// Load the permission
	var permission models.Permission
	if err := tc.db.WithContext(ctx).Preload("Action").
		First(&permission, "id = ?", *binding.PermissionID).Error; err != nil {
		return nil, fmt.Errorf("failed to load permission: %w", err)
	}

	// Create the tuple
	tuple := &CompiledTuple{
		ResourceType: binding.ResourceType,
		ResourceID:   tc.getResourceID(binding),
		Relation:     permission.Action.Name, // Use action name as relation
		SubjectType:  subjectType,
		SubjectID:    subjectID,
	}

	// Add caveats if present
	if binding.Caveat != nil {
		caveat, err := tc.compileCaveat(binding.Caveat)
		if err != nil {
			tc.logger.Warn("Failed to compile caveat",
				zap.Error(err),
				zap.String("binding_id", binding.ID))
		} else {
			tuple.Caveat = caveat
		}
	}

	return tuple, nil
}

// mapRoleToRelations maps a role to SpiceDB relations based on conventions
func (tc *TupleCompiler) mapRoleToRelations(role *models.Role, resourceType string) []string {
	// Convention-based mapping
	relations := []string{}

	switch role.Name {
	case "admin", "administrator", "owner":
		relations = append(relations, "role_admin", "role_editor", "role_viewer")
	case "editor", "writer", "contributor":
		relations = append(relations, "role_editor", "role_viewer")
	case "viewer", "reader", "guest":
		relations = append(relations, "role_viewer")
	default:
		// Use role name as relation with prefix
		relations = append(relations, fmt.Sprintf("role_%s", role.Name))
	}

	// For table resources, use different relations
	if resourceType == "aaa/table" {
		tableRelations := []string{}
		for _, rel := range relations {
			switch rel {
			case "role_admin":
				tableRelations = append(tableRelations, "role_admin")
			case "role_editor":
				tableRelations = append(tableRelations, "role_writer")
			case "role_viewer":
				tableRelations = append(tableRelations, "role_reader")
			}
		}
		return tableRelations
	}

	return relations
}

// getResourceID returns the resource ID or a wildcard
func (tc *TupleCompiler) getResourceID(binding *models.Binding) string {
	if binding.ResourceID != nil && *binding.ResourceID != "" {
		return *binding.ResourceID
	}
	return "*" // Wildcard for type-level bindings
}

// compileCaveat compiles a binding caveat into SpiceDB caveat
func (tc *TupleCompiler) compileCaveat(caveat *models.Caveat) (*CompiledCaveat, error) {
	if caveat == nil {
		return nil, nil
	}

	compiled := &CompiledCaveat{
		Context: make(map[string]interface{}),
	}

	// Check for time-based caveats
	if startsAt, ok := (*caveat)["starts_at"]; ok {
		if endsAt, ok := (*caveat)["ends_at"]; ok {
			compiled.Name = "within_time"
			compiled.Context["starts_at"] = startsAt
			compiled.Context["ends_at"] = endsAt
			compiled.Context["current_time"] = time.Now().Unix()
			return compiled, nil
		}
	}

	// Check for attribute-based caveats
	if attrs, ok := (*caveat)["required_attributes"]; ok {
		compiled.Name = "has_attributes"
		compiled.Context["required"] = attrs
		compiled.Context["attrs"] = map[string]interface{}{} // Will be filled at check time
		return compiled, nil
	}

	// Check for column-based caveats
	if columnGroups, ok := (*caveat)["column_groups"]; ok {
		compiled.Name = "has_columns"
		compiled.Context["allowed_groups"] = columnGroups
		compiled.Context["column_groups"] = []string{} // Will be filled at check time
		return compiled, nil
	}

	// If no specific caveat type matches, return a generic caveat
	compiled.Name = "generic_caveat"
	for k, v := range *caveat {
		compiled.Context[k] = v
	}

	return compiled, nil
}

// WriteTuple writes a compiled tuple to PostgreSQL
// This function is no longer needed as we're using PostgreSQL RBAC directly
func (tc *TupleCompiler) WriteTuple(ctx context.Context, tuple *CompiledTuple) error {
	// This function is no longer needed as we're using PostgreSQL RBAC directly
	// The logic for handling tuple writing in PostgreSQL would go here.
	// For now, we'll just log the action.
	tc.logger.Info("Skipping WriteTuple as we're using PostgreSQL RBAC directly",
		zap.String("resource_type", tuple.ResourceType),
		zap.String("resource_id", tuple.ResourceID))
	return nil
}

// WriteTuples writes multiple compiled tuples to PostgreSQL
// This function is no longer needed as we're using PostgreSQL RBAC directly
func (tc *TupleCompiler) WriteTuples(ctx context.Context, tuples []*CompiledTuple) error {
	// This function is no longer needed as we're using PostgreSQL RBAC directly
	// The logic for handling tuple writing in PostgreSQL would go here.
	// For now, we'll just log the action.
	tc.logger.Info("Skipping WriteTuples as we're using PostgreSQL RBAC directly",
		zap.Int("tuple_count", len(tuples)))
	return nil
}

// DeleteTuplesForBinding deletes tuples associated with a binding
// This function is no longer needed as we're using PostgreSQL RBAC directly
func (tc *TupleCompiler) DeleteTuplesForBinding(ctx context.Context, binding *models.Binding) error {
	// This function is no longer needed as we're using PostgreSQL RBAC directly
	// The logic for handling tuple deletion in PostgreSQL would go here.
	// For now, we'll just log the action.
	tc.logger.Info("Skipping DeleteTuplesForBinding as we're using PostgreSQL RBAC directly",
		zap.String("binding_id", binding.ID))
	return nil
}
