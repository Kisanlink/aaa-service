package migrations

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedCoreResourcesRolesPermissions ensures core resources exist, creates permissions
// for admin and super_admin roles, assigns them in SQL, and syncs role→permission
// relationships to SpiceDB so authorization works. Accepts a gorm.DB for direct use.
func SeedCoreResourcesRolesPermissions(
	ctx context.Context,
	gormDB *gorm.DB,
	spicedbAddr, spicedbToken string,
	logger *zap.Logger,
) error {
	if err := seedCoreResources(ctx, gormDB, logger); err != nil {
		return fmt.Errorf("seed resources: %w", err)
	}
	if err := seedAdminAndSuperAdminPermissions(ctx, gormDB, logger); err != nil {
		return fmt.Errorf("seed role permissions: %w", err)
	}
	// SpiceDB sync removed - using PostgreSQL RBAC
	return nil
}

func seedCoreResources(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	core := []struct {
		Name string
		Type string
		Desc string
	}{
		{"user", models.ResourceTypeUser, "User resource"},
		{"role", models.ResourceTypeRole, "Role resource"},
		{"permission", models.ResourceTypePermission, "Permission resource"},
		{"audit_log", models.ResourceTypeAuditLog, "Audit log resource"},
		{"system", models.ResourceTypeSystem, "System resource"},
		{"api_endpoint", models.ResourceTypeAPIEndpoint, "API endpoint resource"},
		{"resource", models.ResourceTypeResource, "Generic AAA resource"},
	}
	for _, r := range core {
		var existing models.Resource
		err := db.WithContext(ctx).Where("name = ?", r.Name).First(&existing).Error
		if err == nil {
			continue
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		res := models.NewResource(r.Name, r.Type, r.Desc)
		if err := db.WithContext(ctx).Create(res).Error; err != nil {
			return err
		}
		if logger != nil {
			logger.Info("Created core resource", zap.String("name", r.Name))
		}
	}
	return nil
}

// minimal action fetch cache
type actionIndex struct{ byName map[string]models.Action }

func buildActionIndex(ctx context.Context, db *gorm.DB) (*actionIndex, error) {
	var actions []models.Action
	if err := db.WithContext(ctx).Find(&actions).Error; err != nil {
		return nil, err
	}
	idx := &actionIndex{byName: make(map[string]models.Action, len(actions))}
	for _, a := range actions {
		idx.byName[a.Name] = a
	}
	return idx, nil
}

func seedAdminAndSuperAdminPermissions(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	// Ensure all default roles exist
	defaultRoles := []struct {
		name        string
		description string
		scope       models.RoleScope
	}{
		{"super_admin", "Super Administrator with global access", models.RoleScopeGlobal},
		{"admin", "Administrator with organization-level access", models.RoleScopeOrg},
		{"user", "Regular user with basic access", models.RoleScopeOrg},
		{"viewer", "Read-only access user", models.RoleScopeOrg},
		{"aaa_admin", "AAA service administrator", models.RoleScopeGlobal},
		{"module_admin", "Module administrator for service management", models.RoleScopeOrg},
	}

	// Create roles if they don't exist
	createdRoles := make(map[string]*models.Role)
	for _, roleData := range defaultRoles {
		var role models.Role
		if err := db.WithContext(ctx).Where("name = ?", roleData.name).First(&role).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("error checking role %s: %w", roleData.name, err)
			}
			// Create new role
			if roleData.scope == models.RoleScopeGlobal {
				role = *models.NewGlobalRole(roleData.name, roleData.description)
			} else {
				role = *models.NewOrgRole(roleData.name, roleData.description, "")
			}
			if err := db.WithContext(ctx).Create(&role).Error; err != nil {
				return fmt.Errorf("error creating role %s: %w", roleData.name, err)
			}
			if logger != nil {
				logger.Info("Created default role", zap.String("name", roleData.name))
			}
		}
		createdRoles[roleData.name] = &role
	}

	// Load resources
	var resources []models.Resource
	if err := db.WithContext(ctx).Find(&resources).Error; err != nil {
		return err
	}
	resByName := map[string]*models.Resource{}
	for i := range resources {
		resByName[resources[i].Name] = &resources[i]
	}

	// Build actions index (SeedStaticActions should have created these)
	actIdx, err := buildActionIndex(ctx, db)
	if err != nil {
		return err
	}

	// Define permissions per role
	// Super admin: manage and full CRUD across core resources
	superAdminMatrix := map[string][]string{
		"user":         {"manage", "create", "read", "update", "delete", "assign"},
		"role":         {"manage", "create", "read", "update", "delete", "assign"},
		"permission":   {"manage", "create", "read", "update", "delete", "assign"},
		"audit_log":    {"view", "export"},
		"system":       {"backup", "restore", "manage"},
		"api_endpoint": {"call"},
		"resource":     {"manage", "read", "update"},
	}

	// Admin: a narrower subset
	adminMatrix := map[string][]string{
		"user":       {"read", "update"},
		"role":       {"read", "assign"},
		"permission": {"read"},
		"audit_log":  {"view"},
	}

	// User: basic access
	userMatrix := map[string][]string{
		"user":     {"read", "update"},
		"resource": {"read"},
	}

	// Viewer: read-only access
	viewerMatrix := map[string][]string{
		"user":     {"read"},
		"resource": {"read"},
	}

	// AAA Admin: AAA service management
	aaaAdminMatrix := map[string][]string{
		"user":         {"manage", "create", "read", "update", "delete"},
		"role":         {"manage", "create", "read", "update", "delete"},
		"permission":   {"manage", "create", "read", "update", "delete"},
		"audit_log":    {"read", "export"},
		"system":       {"manage"},
		"api_endpoint": {"call"},
		"resource":     {"manage", "read", "update"},
	}

	// Module Admin: module management
	moduleAdminMatrix := map[string][]string{
		"user":       {"read", "update"},
		"role":       {"read", "assign"},
		"permission": {"read"},
		"resource":   {"read", "update"},
	}

	// Helper to upsert permission and attach to role
	upsertAndAttach := func(role *models.Role, resourceName, actionName string) error {
		res := resByName[resourceName]
		act, ok := actIdx.byName[actionName]
		if res == nil || !ok {
			return nil
		}

		permName := fmt.Sprintf("%s:%s", resourceName, actionName)
		var perm models.Permission
		if err := db.WithContext(ctx).Where("name = ?", permName).First(&perm).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			// Create new permission only if it doesn't exist
			newPerm := models.NewPermissionWithResourceAndAction(permName, fmt.Sprintf("%s on %s", actionName, resourceName), res.ID, act.ID)
			if err := db.WithContext(ctx).Create(newPerm).Error; err != nil {
				return err
			}
			perm = *newPerm
		}
		// Check if role-permission relationship already exists
		var existingRPs []models.RolePermission
		if err := db.WithContext(ctx).Where("role_id = ? AND permission_id = ? AND is_active = ?", role.ID, perm.ID, true).Find(&existingRPs).Error; err != nil {
			return err
		}

		// Create role-permission relationship if it doesn't exist
		if len(existingRPs) == 0 {
			rp := models.NewRolePermission(role.ID, perm.ID)
			if err := db.WithContext(ctx).Create(rp).Error; err != nil {
				return err
			}
		}
		return nil
	}

	// Seed core users for super_admin and admin with default passwords
	// Create user helper
	createUserIfMissing := func(username, phone, country, plainPassword string) (*models.User, error) {
		var existing models.User
		if err := db.WithContext(ctx).Where("phone_number = ? AND country_code = ?", phone, country).First(&existing).Error; err == nil {
			return &existing, nil
		}
		// hash password using same algo as service
		hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		user := models.NewUserWithUsername(phone, country, username, string(hashed))
		status := "active"
		user.Status = &status
		user.IsValidated = true
		if err := db.WithContext(ctx).Create(user).Error; err != nil {
			return nil, err
		}
		return user, nil
	}

	superAdminUser, err := createUserIfMissing("superadmin", "9999999999", "+91", "SuperAdmin@123")
	if err != nil {
		return err
	}
	adminUser, err := createUserIfMissing("admin", "8888888888", "+91", "Admin@123")
	if err != nil {
		return err
	}

	// Attach roles to seeded users if not already
	attachRole := func(userID, roleID string) error {
		var count int64
		if err := db.WithContext(ctx).Model(&models.UserRole{}).Where("user_id = ? AND role_id = ? AND is_active = ?", userID, roleID, true).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		ur := models.NewUserRole(userID, roleID)
		return db.WithContext(ctx).Create(ur).Error
	}

	if err := attachRole(superAdminUser.ID, createdRoles["super_admin"].ID); err != nil {
		return err
	}
	if err := attachRole(adminUser.ID, createdRoles["admin"].ID); err != nil {
		return err
	}

	// Seed for super_admin
	superAdmin := createdRoles["super_admin"]
	for resName, actions := range superAdminMatrix {
		for _, act := range actions {
			if err := upsertAndAttach(superAdmin, resName, act); err != nil {
				return err
			}
		}
	}
	// Seed for admin
	admin := createdRoles["admin"]
	for resName, actions := range adminMatrix {
		for _, act := range actions {
			if err := upsertAndAttach(admin, resName, act); err != nil {
				return err
			}
		}
	}

	// Seed for user
	user := createdRoles["user"]
	for resName, actions := range userMatrix {
		for _, act := range actions {
			if err := upsertAndAttach(user, resName, act); err != nil {
				return err
			}
		}
	}

	// Seed for viewer
	viewer := createdRoles["viewer"]
	for resName, actions := range viewerMatrix {
		for _, act := range actions {
			if err := upsertAndAttach(viewer, resName, act); err != nil {
				return err
			}
		}
	}

	// Seed for aaa_admin
	aaaAdmin := createdRoles["aaa_admin"]
	for resName, actions := range aaaAdminMatrix {
		for _, act := range actions {
			if err := upsertAndAttach(aaaAdmin, resName, act); err != nil {
				return err
			}
		}
	}

	// Seed for module_admin
	moduleAdmin := createdRoles["module_admin"]
	for resName, actions := range moduleAdminMatrix {
		for _, act := range actions {
			if err := upsertAndAttach(moduleAdmin, resName, act); err != nil {
				return err
			}
		}
	}

	if logger != nil {
		logger.Info("Seeded core role permissions for all default roles")
	}
	return nil
}

// syncRolePermissionsToSpiceDB is deprecated - now using PostgreSQL RBAC
/*
func syncRolePermissionsToSpiceDB(ctx context.Context, db *gorm.DB, addr, token string, logger *zap.Logger) error {
	if addr == "" || token == "" {
		if logger != nil {
			logger.Warn("Skipping SpiceDB sync; addr/token not set")
		}
		return nil
	}
	client, err := authzed.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(token),
	)
	if err != nil {
		return err
	}

	// Load roles with attached permissions
	var roles []models.Role
	if err := db.WithContext(ctx).Preload("Permissions").Find(&roles).Error; err != nil {
		return err
	}

	// For each role-permission, write a relationship: role -perms-> permission
	for _, role := range roles {
		for _, perm := range role.Permissions {
			// FIX: Write role#perms@aaa/perm relationships to match schema
			update := &authzedpb.RelationshipUpdate{
				Operation: authzedpb.RelationshipUpdate_OPERATION_CREATE,
				Relationship: &authzedpb.Relationship{
					Resource: &authzedpb.ObjectReference{ObjectType: "aaa/role", ObjectId: role.ID},
					Relation: "perms",
					Subject:  &authzedpb.SubjectReference{Object: &authzedpb.ObjectReference{ObjectType: "aaa/perm", ObjectId: perm.ID}},
				},
			}
			req := &authzedpb.WriteRelationshipsRequest{Updates: []*authzedpb.RelationshipUpdate{update}}
			// best-effort write with short timeout
			wctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			_, werr := client.WriteRelationships(wctx, req)
			cancel()
			if werr != nil && logger != nil {
				logger.Warn("Failed to sync role-permission to SpiceDB", zap.String("role", role.Name), zap.String("permission", perm.Name), zap.Error(werr))
			}
		}
	}
	// FIX: Also seed aaa/users relationships to connect roles to the users resource
	// This allows permission checks on /api/v1/users endpoints to work
	if err := seedUsersResourceRelationships(ctx, client, db, logger); err != nil {
		if logger != nil {
			logger.Warn("Failed to seed users resource relationships", zap.Error(err))
		}
	}

	if logger != nil {
		logger.Info("Synchronized role→permission relationships to SpiceDB")
	}
	return nil
}
*/

// seedUsersResourceRelationships is deprecated - now using PostgreSQL RBAC
/*
func seedUsersResourceRelationships(ctx context.Context, client *authzed.Client, db *gorm.DB, logger *zap.Logger) error {
	// Load users with super_admin and admin roles
	var userRoles []models.UserRole
	if err := db.WithContext(ctx).
		Preload("User").
		Preload("Role").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.name IN ? AND user_roles.is_active = ?", []string{"super_admin", "admin"}, true).
		Find(&userRoles).Error; err != nil {
		return fmt.Errorf("failed to load user roles: %w", err)
	}

	// Create a users resource instance (using "default" as the ID)
	usersResourceID := "default"

	for _, ur := range userRoles {
		// Check if preloaded data exists
		if ur.User.ID == "" || ur.Role.ID == "" {
			continue
		}

		// Determine which relations to create based on role
		relations := []string{}
		if ur.Role.Name == "super_admin" {
			// Super admin gets all permissions
			relations = []string{"viewer", "creator", "editor", "deleter", "manager"}
		} else if ur.Role.Name == "admin" {
			// Admin gets view and edit permissions
			relations = []string{"viewer", "editor"}
		}

		for _, relation := range relations {
			// Create relationship: users#relation@user:user_id
			update := &authzedpb.RelationshipUpdate{
				Operation: authzedpb.RelationshipUpdate_OPERATION_CREATE,
				Relationship: &authzedpb.Relationship{
					Resource: &authzedpb.ObjectReference{ObjectType: "aaa/users", ObjectId: usersResourceID},
					Relation: relation,
					Subject:  &authzedpb.SubjectReference{Object: &authzedpb.ObjectReference{ObjectType: "aaa/user", ObjectId: ur.User.ID}},
				},
			}
			req := &authzedpb.WriteRelationshipsRequest{Updates: []*authzedpb.RelationshipUpdate{update}}
			wctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			_, werr := client.WriteRelationships(wctx, req)
			cancel()
			if werr != nil && logger != nil {
				// Handle Username being a pointer
				username := "unknown"
				if ur.User.Username != nil {
					username = *ur.User.Username
				}
				logger.Warn("Failed to create users resource relationship",
					zap.String("user", username),
					zap.String("relation", relation),
					zap.Error(werr))
			}
		}

		if logger != nil {
			// Handle Username being a pointer
			username := "unknown"
			if ur.User.Username != nil {
				username = *ur.User.Username
				}
			logger.Info("Created users resource relationships",
				zap.String("user", username),
				zap.String("role", ur.Role.Name))
		}
	}

	return nil
}
*/

// SeedCoreResourcesRolesPermissionsWithDBManager is a helper that obtains the GORM DB
// from kisanlink-db DatabaseManager and calls SeedCoreResourcesRolesPermissions.
func SeedCoreResourcesRolesPermissionsWithDBManager(
	ctx context.Context,
	dm *db.DatabaseManager,
	logger *zap.Logger,
) error {
	if dm == nil {
		return fmt.Errorf("database manager is nil")
	}
	pm := dm.GetPostgresManager()
	if pm == nil {
		return fmt.Errorf("postgres manager not available")
	}
	gormDB, err := pm.GetDB(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get postgres DB: %w", err)
	}
	// Use primary DBManager for consistent ID generation; GORM DB only for associations
	primary := dm.GetManager(db.BackendGorm)
	if primary == nil {
		return fmt.Errorf("primary DB manager (gorm) not available")
	}
	if err := seedCoreResourcesDM(ctx, primary, logger); err != nil {
		return fmt.Errorf("seed resources: %w", err)
	}
	if err := seedAdminAndSuperAdminPermissionsDM(ctx, primary, gormDB, logger); err != nil {
		return fmt.Errorf("seed role permissions: %w", err)
	}
	// PostgreSQL RBAC implementation - no external authorization service needed
	return nil
}

// DM-based seeding helpers
func seedCoreResourcesDM(ctx context.Context, primary db.DBManager, logger *zap.Logger) error {
	core := []struct {
		Name string
		Type string
		Desc string
	}{
		{"user", models.ResourceTypeUser, "User resource"},
		{"role", models.ResourceTypeRole, "Role resource"},
		{"permission", models.ResourceTypePermission, "Permission resource"},
		{"audit_log", models.ResourceTypeAuditLog, "Audit log resource"},
		{"system", models.ResourceTypeSystem, "System resource"},
		{"api_endpoint", models.ResourceTypeAPIEndpoint, "API endpoint resource"},
		{"resource", models.ResourceTypeResource, "Generic AAA resource"},
	}
	for _, r := range core {
		var existing []models.Resource
		filters := []base.FilterCondition{{Field: "name", Operator: base.OpEqual, Value: r.Name}}
		filter := &base.Filter{
			Group: base.FilterGroup{
				Conditions: filters,
				Logic:      base.LogicAnd,
			},
		}
		if err := primary.List(ctx, filter, &existing); err != nil {
			return err
		}
		if len(existing) > 0 {
			continue
		}
		res := models.NewResource(r.Name, r.Type, r.Desc)
		if err := primary.Create(ctx, res); err != nil {
			return err
		}
		if logger != nil {
			logger.Info("Created core resource", zap.String("name", r.Name))
		}
	}
	return nil
}

func buildActionIndexDM(ctx context.Context, primary db.DBManager) (*actionIndex, error) {
	var actions []models.Action
	emptyFilter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{},
			Logic:      base.LogicAnd,
		},
	}
	if err := primary.List(ctx, emptyFilter, &actions); err != nil {
		return nil, err
	}
	idx := &actionIndex{byName: make(map[string]models.Action, len(actions))}
	for _, a := range actions {
		idx.byName[a.Name] = a
	}
	return idx, nil
}

func seedAdminAndSuperAdminPermissionsDM(ctx context.Context, primary db.DBManager, gormDB *gorm.DB, logger *zap.Logger) error {
	// Ensure all default roles exist
	defaultRoles := []struct {
		name        string
		description string
		scope       models.RoleScope
	}{
		{"super_admin", "Super Administrator with global access", models.RoleScopeGlobal},
		{"admin", "Administrator with organization-level access", models.RoleScopeOrg},
		{"user", "Regular user with basic access", models.RoleScopeOrg},
		{"viewer", "Read-only access user", models.RoleScopeOrg},
		{"aaa_admin", "AAA service administrator", models.RoleScopeGlobal},
		{"module_admin", "Module administrator for service management", models.RoleScopeOrg},
	}

	// Create roles if they don't exist
	createdRoles := make(map[string]*models.Role)
	for _, roleData := range defaultRoles {
		var roles []models.Role
		if err := primary.List(ctx, &base.Filter{Group: base.FilterGroup{Conditions: []base.FilterCondition{{Field: "name", Operator: base.OpEqual, Value: roleData.name}}}}, &roles); err != nil {
			return fmt.Errorf("error checking role %s: %w", roleData.name, err)
		}

		var role models.Role
		if len(roles) > 0 {
			// Role exists, use it
			role = roles[0]
		} else {
			// Create new role
			if roleData.scope == models.RoleScopeGlobal {
				role = *models.NewGlobalRole(roleData.name, roleData.description)
			} else {
				role = *models.NewOrgRole(roleData.name, roleData.description, "")
			}
			if err := primary.Create(ctx, &role); err != nil {
				return fmt.Errorf("error creating role %s: %w", roleData.name, err)
			}
			if logger != nil {
				logger.Info("Created default role", zap.String("name", roleData.name))
			}
		}
		createdRoles[roleData.name] = &role
	}

	// Load resources
	var resources []models.Resource
	emptyFilter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{},
			Logic:      base.LogicAnd,
		},
	}
	if err := primary.List(ctx, emptyFilter, &resources); err != nil {
		return err
	}
	resByName := map[string]*models.Resource{}
	for i := range resources {
		resByName[resources[i].Name] = &resources[i]
	}

	// Build actions index
	actIdx, err := buildActionIndexDM(ctx, primary)
	if err != nil {
		return err
	}

	// Permission matrices
	superAdminMatrix := map[string][]string{
		"user":         {"manage", "create", "read", "update", "delete", "assign"},
		"role":         {"manage", "create", "read", "update", "delete", "assign"},
		"permission":   {"manage", "create", "read", "update", "delete", "assign"},
		"audit_log":    {"read", "export"},
		"system":       {"backup", "restore", "manage"},
		"api_endpoint": {"call"},
		"resource":     {"manage", "read", "update"},
	}
	adminMatrix := map[string][]string{
		"user":       {"read", "update"},
		"role":       {"read", "assign"},
		"permission": {"read"},
		"audit_log":  {"read"},
	}

	// User: basic access
	userMatrix := map[string][]string{
		"user":     {"read", "update"},
		"resource": {"read"},
	}

	// Viewer: read-only access
	viewerMatrix := map[string][]string{
		"user":     {"read"},
		"resource": {"read"},
	}

	// AAA Admin: AAA service management
	aaaAdminMatrix := map[string][]string{
		"user":         {"manage", "create", "read", "update", "delete"},
		"role":         {"manage", "create", "read", "update", "delete"},
		"permission":   {"manage", "create", "read", "update", "delete"},
		"audit_log":    {"read", "export"},
		"system":       {"manage"},
		"api_endpoint": {"call"},
		"resource":     {"manage", "read", "update"},
	}

	// Module Admin: module management
	moduleAdminMatrix := map[string][]string{
		"user":       {"read", "update"},
		"role":       {"read", "assign"},
		"permission": {"read"},
		"resource":   {"read", "update"},
	}

	// Upsert permission using DM; attach using GORM association
	upsertAndAttach := func(role *models.Role, resourceName, actionName string) error {
		res := resByName[resourceName]
		act, ok := actIdx.byName[actionName]
		if res == nil || !ok {
			return nil
		}
		permName := fmt.Sprintf("%s:%s", resourceName, actionName)
		var perms []models.Permission
		permFilter := &base.Filter{
			Group: base.FilterGroup{
				Conditions: []base.FilterCondition{{Field: "name", Operator: base.OpEqual, Value: permName}},
				Logic:      base.LogicAnd,
			},
		}
		if err := primary.List(ctx, permFilter, &perms); err != nil {
			return err
		}
		var perm models.Permission
		if len(perms) == 0 {
			// Create new permission only if it doesn't exist
			newPerm := models.NewPermissionWithResourceAndAction(permName, fmt.Sprintf("%s on %s", actionName, resourceName), res.ID, act.ID)
			if err := primary.Create(ctx, newPerm); err != nil {
				return err
			}
			perm = *newPerm
		} else {
			// Use existing permission
			perm = perms[0]
		}

		// Check if role-permission relationship already exists
		var existingRPs []models.RolePermission
		filters := []base.FilterCondition{
			{Field: "role_id", Operator: base.OpEqual, Value: role.ID},
			{Field: "permission_id", Operator: base.OpEqual, Value: perm.ID},
			{Field: "is_active", Operator: base.OpEqual, Value: true},
		}
		filter := &base.Filter{
			Group: base.FilterGroup{
				Conditions: filters,
				Logic:      base.LogicAnd,
			},
		}
		if err := primary.List(ctx, filter, &existingRPs); err != nil {
			return err
		}

		// Create role-permission relationship if it doesn't exist
		if len(existingRPs) == 0 {
			rp := models.NewRolePermission(role.ID, perm.ID)
			if err := primary.Create(ctx, rp); err != nil {
				return err
			}
		}

		return nil
	}

	// Seed core users and attach roles
	createUserIfMissing := func(username, phone, country, plainPassword string) (*models.User, error) {
		var users []models.User
		filters := []base.FilterCondition{{Field: "phone_number", Operator: base.OpEqual, Value: phone}, {Field: "country_code", Operator: base.OpEqual, Value: country}}
		filter := &base.Filter{
			Group: base.FilterGroup{
				Conditions: filters,
				Logic:      base.LogicAnd,
			},
		}
		if err := primary.List(ctx, filter, &users); err == nil && len(users) > 0 {
			return &users[0], nil
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		user := models.NewUserWithUsername(phone, country, username, string(hashed))
		status := "active"
		user.Status = &status
		user.IsValidated = true
		if err := primary.Create(ctx, user); err != nil {
			return nil, err
		}
		return user, nil
	}

	su, err := createUserIfMissing("superadmin", "9999999999", "+91", "SuperAdmin@123")
	if err != nil {
		return err
	}
	ad, err := createUserIfMissing("admin", "8888888888", "+91", "Admin@123")
	if err != nil {
		return err
	}

	attachRole := func(userID, roleID string) error {
		var urs []models.UserRole
		filters := []base.FilterCondition{{Field: "user_id", Operator: base.OpEqual, Value: userID}, {Field: "role_id", Operator: base.OpEqual, Value: roleID}, {Field: "is_active", Operator: base.OpEqual, Value: true}}
		filter := &base.Filter{
			Group: base.FilterGroup{
				Conditions: filters,
				Logic:      base.LogicAnd,
			},
		}
		if err := primary.List(ctx, filter, &urs); err != nil {
			return err
		}
		if len(urs) > 0 {
			return nil
		}
		ur := models.NewUserRole(userID, roleID)
		return primary.Create(ctx, ur)
	}
	if err := attachRole(su.ID, createdRoles["super_admin"].ID); err != nil {
		return err
	}
	if err := attachRole(ad.ID, createdRoles["admin"].ID); err != nil {
		return err
	}

	// Seed permissions for all roles
	superAdmin := createdRoles["super_admin"]
	for resName, actions := range superAdminMatrix {
		for _, act := range actions {
			if err := upsertAndAttach(superAdmin, resName, act); err != nil {
				return err
			}
		}
	}
	admin := createdRoles["admin"]
	for resName, actions := range adminMatrix {
		for _, act := range actions {
			if err := upsertAndAttach(admin, resName, act); err != nil {
				return err
			}
		}
	}

	// Seed for user
	user := createdRoles["user"]
	for resName, actions := range userMatrix {
		for _, act := range actions {
			if err := upsertAndAttach(user, resName, act); err != nil {
				return err
			}
		}
	}

	// Seed for viewer
	viewer := createdRoles["viewer"]
	for resName, actions := range viewerMatrix {
		for _, act := range actions {
			if err := upsertAndAttach(viewer, resName, act); err != nil {
				return err
			}
		}
	}

	// Seed for aaa_admin
	aaaAdmin := createdRoles["aaa_admin"]
	for resName, actions := range aaaAdminMatrix {
		for _, act := range actions {
			if err := upsertAndAttach(aaaAdmin, resName, act); err != nil {
				return err
			}
		}
	}

	// Seed for module_admin
	moduleAdmin := createdRoles["module_admin"]
	for resName, actions := range moduleAdminMatrix {
		for _, act := range actions {
			if err := upsertAndAttach(moduleAdmin, resName, act); err != nil {
				return err
			}
		}
	}

	if logger != nil {
		logger.Info("Seeded core role permissions for all default roles")
	}
	return nil
}
