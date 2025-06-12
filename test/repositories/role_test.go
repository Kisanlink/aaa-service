package repositories

// import (
// 	"errors"
// 	"net/http"
// 	"regexp"
// 	"testing"
// 	"time"

// 	"github.com/DATA-DOG/go-sqlmock"
// 	"github.com/Kisanlink/aaa-service/helper"
// 	"github.com/Kisanlink/aaa-service/model"
// 	"github.com/Kisanlink/aaa-service/repositories"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// )

// func setupRoleRepository(t *testing.T) (*repositories.RoleRepository, sqlmock.Sqlmock) {
// 	db, mock, err := sqlmock.New()
// 	require.NoError(t, err)

// 	gormDB, err := gorm.Open(postgres.New(postgres.Config{
// 		Conn: db,
// 	}), &gorm.Config{})
// 	require.NoError(t, err)

// 	return &repositories.RoleRepository{DB: gormDB}, mock
// }

// func TestCheckIfRoleExists(t *testing.T) {
// 	repo, mock := setupRoleRepository(t)
// 	roleName := "test_role"

// 	t.Run("Role exists", func(t *testing.T) {
// 		rows := sqlmock.NewRows([]string{"id", "name"}).
// 			AddRow("1", roleName)

// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`SELECT * FROM "roles" WHERE name = $1 ORDER BY "roles"."id" LIMIT $2`,
// 		)).
// 			WithArgs(roleName, 1).
// 			WillReturnRows(rows)

// 		err := repo.CheckIfRoleExists(roleName)

// 		var appErr *helper.AppError
// 		if assert.ErrorAs(t, err, &appErr) {
// 			assert.Equal(t, http.StatusConflict, appErr.StatusCode)
// 			assert.Contains(t, appErr.Error(), "already exists")
// 		}
// 	})

// 	t.Run("Role does not exist", func(t *testing.T) {
// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`SELECT * FROM "roles" WHERE name = $1 ORDER BY "roles"."id" LIMIT $2`,
// 		)).
// 			WithArgs("nonexistent_role", 1).
// 			WillReturnError(gorm.ErrRecordNotFound)

// 		err := repo.CheckIfRoleExists("nonexistent_role")
// 		assert.NoError(t, err)
// 	})

// 	t.Run("Database error", func(t *testing.T) {
// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`SELECT * FROM "roles" WHERE name = $1 ORDER BY "roles"."id" LIMIT $2`,
// 		)).
// 			WithArgs(roleName, 1).
// 			WillReturnError(errors.New("db error"))

// 		err := repo.CheckIfRoleExists(roleName)

// 		var appErr *helper.AppError
// 		if assert.ErrorAs(t, err, &appErr) {
// 			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
// 			assert.Contains(t, appErr.Error(), "database error")
// 		}
// 	})

// 	assert.NoError(t, mock.ExpectationsWereMet())
// }

// func TestCreateRoleWithPermissions(t *testing.T) {
// 	repo, mock := setupRoleRepository(t)

// 	role := &model.Role{
// 		Base: model.Base{
// 			ID:        "1",
// 			CreatedAt: time.Now(),
// 			UpdatedAt: time.Now(),
// 		},
// 		Name:        "admin",
// 		Description: "Administrator role",
// 	}

// 	permissions := []model.Permission{
// 		{Resource: "users", Actions: []string{"create", "read", "update", "delete"}},
// 		{Resource: "roles", Actions: []string{"read", "update"}},
// 	}

// 	t.Run("Successfully create role with permissions", func(t *testing.T) {
// 		mock.ExpectBegin()

// 		// Role creation
// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`INSERT INTO "roles" ("id","created_at","updated_at","name","description") VALUES ($1,$2,$3,$4,$5) RETURNING "description","source"`,
// 		)).
// 			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), role.Name, role.Description).
// 			WillReturnRows(sqlmock.NewRows([]string{"description", "source"}).
// 				AddRow(role.Description, role.Source))

// 		// First permission creation - updated to match actual query
// 		mock.ExpectExec(regexp.QuoteMeta(
// 			`INSERT INTO "permissions" ("id","created_at","updated_at","role_id","resource","actions") VALUES ($1,$2,$3,$4,$5,$6)`,
// 		)).
// 			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), role.ID, permissions[0].Resource, permissions[0].Actions).
// 			WillReturnResult(sqlmock.NewResult(1, 1))

// 		// Second permission creation
// 		mock.ExpectExec(regexp.QuoteMeta(
// 			`INSERT INTO "permissions" ("id","created_at","updated_at","role_id","resource","actions") VALUES ($1,$2,$3,$4,$5,$6)`,
// 		)).
// 			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), role.ID, permissions[1].Resource, permissions[1].Actions).
// 			WillReturnResult(sqlmock.NewResult(1, 1))

// 		mock.ExpectCommit()

// 		err := repo.CreateRoleWithPermissions(role, permissions)
// 		assert.NoError(t, err)
// 		assert.NotEmpty(t, role.ID)
// 	})

// 	t.Run("Fail to create role", func(t *testing.T) {
// 		mock.ExpectBegin()
// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`INSERT INTO "roles" ("id","created_at","updated_at","name","description") VALUES ($1,$2,$3,$4,$5) RETURNING "description","source"`,
// 		)).
// 			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), role.Name, role.Description).
// 			WillReturnError(errors.New("create error"))
// 		mock.ExpectRollback()

// 		err := repo.CreateRoleWithPermissions(role, permissions)

// 		var appErr *helper.AppError
// 		if assert.ErrorAs(t, err, &appErr) {
// 			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
// 			assert.Contains(t, appErr.Error(), "failed to create role")
// 		}
// 	})

// 	t.Run("Fail to create permission", func(t *testing.T) {
// 		mock.ExpectBegin()
// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`INSERT INTO "roles" ("id","created_at","updated_at","name","description") VALUES ($1,$2,$3,$4,$5) RETURNING "description","source"`,
// 		)).
// 			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), role.Name, role.Description).
// 			WillReturnRows(sqlmock.NewRows([]string{"description", "source"}).
// 				AddRow(role.Description, role.Source))

// 		mock.ExpectExec(regexp.QuoteMeta(
// 			`INSERT INTO "permissions" ("id","created_at","updated_at","role_id","resource","actions") VALUES ($1,$2,$3,$4,$5,$6)`,
// 		)).
// 			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), role.ID, permissions[0].Resource, permissions[0].Actions).
// 			WillReturnError(errors.New("permission error"))
// 		mock.ExpectRollback()

// 		err := repo.CreateRoleWithPermissions(role, permissions)

// 		var appErr *helper.AppError
// 		if assert.ErrorAs(t, err, &appErr) {
// 			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
// 			assert.Contains(t, appErr.Error(), "failed to create permission")
// 		}
// 	})

// 	assert.NoError(t, mock.ExpectationsWereMet())
// }

// func TestGetRoleByName(t *testing.T) {
// 	repo, mock := setupRoleRepository(t)
// 	roleName := "admin"

// 	t.Run("Successfully get role with permissions", func(t *testing.T) {
// 		roleRows := sqlmock.NewRows([]string{"id", "name", "description"}).
// 			AddRow("1", roleName, "Admin role")

// 		// Create a driver.Value for the string array
// 		actions1 := []byte(`{"create","read","update","delete"}`) // PostgreSQL array format
// 		actions2 := []byte(`{"read","update"}`)

// 		permissionRows := sqlmock.NewRows([]string{"id", "role_id", "resource", "actions"}).
// 			AddRow("1", "1", "users", actions1).
// 			AddRow("2", "1", "roles", actions2)

// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`SELECT * FROM "roles" WHERE name = $1 ORDER BY "roles"."id" LIMIT $2`,
// 		)).
// 			WithArgs(roleName, 1).
// 			WillReturnRows(roleRows)

// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`SELECT * FROM "permissions" WHERE "permissions"."role_id" = $1`,
// 		)).
// 			WithArgs("1").
// 			WillReturnRows(permissionRows)

// 		role, err := repo.GetRoleByName(roleName)
// 		assert.NoError(t, err)
// 		assert.Equal(t, roleName, role.Name)
// 		assert.Len(t, role.Permissions, 2)
// 	})

// 	t.Run("Role not found", func(t *testing.T) {
// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`SELECT * FROM "roles" WHERE name = $1 ORDER BY "roles"."id" LIMIT $2`,
// 		)).
// 			WithArgs("nonexistent_role", 1).
// 			WillReturnError(gorm.ErrRecordNotFound)

// 		role, err := repo.GetRoleByName("nonexistent_role")

// 		var appErr *helper.AppError
// 		if assert.ErrorAs(t, err, &appErr) {
// 			assert.Equal(t, http.StatusNotFound, appErr.StatusCode)
// 			assert.Contains(t, appErr.Error(), "not found")
// 		}
// 		assert.Nil(t, role)
// 	})

// 	t.Run("Database error", func(t *testing.T) {
// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`SELECT * FROM "roles" WHERE name = $1 ORDER BY "roles"."id" LIMIT $2`,
// 		)).
// 			WithArgs(roleName, 1).
// 			WillReturnError(errors.New("db error"))

// 		role, err := repo.GetRoleByName(roleName)

// 		var appErr *helper.AppError
// 		if assert.ErrorAs(t, err, &appErr) {
// 			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
// 			assert.Contains(t, appErr.Error(), "failed to query role")
// 		}
// 		assert.Nil(t, role)
// 	})

// 	assert.NoError(t, mock.ExpectationsWereMet())
// }

// func TestFindRoleByID(t *testing.T) {
// 	repo, mock := setupRoleRepository(t)
// 	roleID := "1"

// 	t.Run("Successfully find role by ID", func(t *testing.T) {
// 		roleRows := sqlmock.NewRows([]string{"id", "name", "description"}).
// 			AddRow(roleID, "admin", "Admin role")

// 		// Create a driver.Value for the string array in PostgreSQL array format
// 		actions := []byte(`{"create","read","update","delete"}`)

// 		permissionRows := sqlmock.NewRows([]string{"id", "role_id", "resource", "actions"}).
// 			AddRow("1", roleID, "users", actions)

// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`SELECT * FROM "roles" WHERE id = $1 ORDER BY "roles"."id" LIMIT $2`,
// 		)).
// 			WithArgs(roleID, 1).
// 			WillReturnRows(roleRows)

// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`SELECT * FROM "permissions" WHERE "permissions"."role_id" = $1`,
// 		)).
// 			WithArgs(roleID).
// 			WillReturnRows(permissionRows)

// 		role, err := repo.FindRoleByID(roleID)
// 		assert.NoError(t, err)
// 		assert.Equal(t, roleID, role.ID)
// 		assert.Equal(t, "admin", role.Name)
// 		assert.Equal(t, "Admin role", role.Description)
// 		assert.Len(t, role.Permissions, 1)
// 		assert.Equal(t, "users", role.Permissions[0].Resource)

// 		// Convert pq.StringArray to []string for comparison
// 		actualActions := []string(role.Permissions[0].Actions)
// 		expectedActions := []string{"create", "read", "update", "delete"}
// 		assert.Equal(t, expectedActions, actualActions)
// 	})

// 	t.Run("Role not found", func(t *testing.T) {
// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`SELECT * FROM "roles" WHERE id = $1 ORDER BY "roles"."id" LIMIT $2`,
// 		)).
// 			WithArgs("999", 1).
// 			WillReturnError(gorm.ErrRecordNotFound)

// 		role, err := repo.FindRoleByID("999")

// 		var appErr *helper.AppError
// 		if assert.ErrorAs(t, err, &appErr) {
// 			assert.Equal(t, http.StatusNotFound, appErr.StatusCode)
// 			assert.Contains(t, appErr.Error(), "not found")
// 		}
// 		assert.Nil(t, role)
// 	})

// 	assert.NoError(t, mock.ExpectationsWereMet())
// }
// func TestFindRoles(t *testing.T) {
// 	repo, mock := setupRoleRepository(t)

// 	t.Run("Successfully find roles with filters", func(t *testing.T) {
// 		filter := map[string]interface{}{
// 			"name": "admin",
// 		}

// 		roleRows := sqlmock.NewRows([]string{"id", "name", "description"}).
// 			AddRow("1", "admin", "Admin role").
// 			AddRow("2", "admin-readonly", "Readonly admin")

// 		// Use PostgreSQL array format for actions
// 		actions1 := []byte(`{"create","read","update","delete"}`)
// 		actions2 := []byte(`{"read"}`)

// 		permissionRows := sqlmock.NewRows([]string{"id", "role_id", "resource", "actions"}).
// 			AddRow("1", "1", "users", actions1).
// 			AddRow("2", "2", "users", actions2)

// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`SELECT * FROM "roles" WHERE name ILIKE $1`,
// 		)).
// 			WithArgs("%admin%").
// 			WillReturnRows(roleRows)

// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`SELECT * FROM "permissions" WHERE "permissions"."role_id" IN ($1,$2)`,
// 		)).
// 			WithArgs("1", "2").
// 			WillReturnRows(permissionRows)

// 		roles, err := repo.FindRoles(filter, 0, 0)
// 		assert.NoError(t, err)
// 		assert.Len(t, roles, 2)
// 		assert.Len(t, roles[0].Permissions, 1)

// 		// Convert pq.StringArray to []string for comparison
// 		expectedActions1 := []string{"create", "read", "update", "delete"}
// 		assert.Equal(t, expectedActions1, []string(roles[0].Permissions[0].Actions))
// 	})

// 	t.Run("Successfully find roles with pagination", func(t *testing.T) {
// 		roleRows := sqlmock.NewRows([]string{"id", "name", "description"}).
// 			AddRow("1", "admin", "Admin role")

// 		actions := []byte(`{"create","read","update","delete"}`)
// 		permissionRows := sqlmock.NewRows([]string{"id", "role_id", "resource", "actions"}).
// 			AddRow("1", "1", "users", actions)

// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`SELECT * FROM "roles" LIMIT $1`,
// 		)).
// 			WithArgs(10).
// 			WillReturnRows(roleRows)

// 		// Change to expect the equality syntax
// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`SELECT * FROM "permissions" WHERE "permissions"."role_id" = $1`,
// 		)).
// 			WithArgs("1").
// 			WillReturnRows(permissionRows)

// 		roles, err := repo.FindRoles(nil, 1, 10)
// 		assert.NoError(t, err)
// 		assert.Len(t, roles, 1)
// 		assert.Equal(t, []string{"create", "read", "update", "delete"}, []string(roles[0].Permissions[0].Actions))
// 	})
// 	t.Run("Database error", func(t *testing.T) {
// 		mock.ExpectQuery(regexp.QuoteMeta(
// 			`SELECT * FROM "roles"`,
// 		)).
// 			WillReturnError(errors.New("db error"))

// 		roles, err := repo.FindRoles(nil, 0, 0)

// 		var appErr *helper.AppError
// 		if assert.ErrorAs(t, err, &appErr) {
// 			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
// 			assert.Contains(t, appErr.Error(), "failed to retrieve roles")
// 		}
// 		assert.Nil(t, roles)
// 	})

// 	assert.NoError(t, mock.ExpectationsWereMet())
// }
// func TestUpdateRoleWithPermissions(t *testing.T) {
// 	repo, mock := setupRoleRepository(t)
// 	roleID := "1"

// 	updatedRole := model.Role{
// 		Name:        "updated-admin",
// 		Description: "Updated admin role",
// 	}

// 	permissions := []model.Permission{
// 		{Resource: "users", Actions: []string{"create", "read", "update"}},
// 		{Resource: "roles", Actions: []string{"read"}},
// 	}

// 	t.Run("Successfully update role with permissions", func(t *testing.T) {
// 		mock.ExpectBegin()

// 		// Update role
// 		mock.ExpectExec(regexp.QuoteMeta(
// 			`UPDATE "roles" SET "updated_at"=$1,"name"=$2,"description"=$3 WHERE id = $4`,
// 		)).
// 			WithArgs(sqlmock.AnyArg(), updatedRole.Name, updatedRole.Description, roleID).
// 			WillReturnResult(sqlmock.NewResult(1, 1))

// 		// Delete existing permissions
// 		mock.ExpectExec(regexp.QuoteMeta(
// 			`DELETE FROM "permissions" WHERE role_id = $1`,
// 		)).
// 			WithArgs(roleID).
// 			WillReturnResult(sqlmock.NewResult(0, 2))

// 		// Create new permissions - updated to match actual query
// 		mock.ExpectExec(regexp.QuoteMeta(
// 			`INSERT INTO "permissions" ("id","created_at","updated_at","role_id","resource","actions") VALUES ($1,$2,$3,$4,$5,$6)`,
// 		)).
// 			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), roleID, permissions[0].Resource, permissions[0].Actions).
// 			WillReturnResult(sqlmock.NewResult(1, 1))

// 		mock.ExpectExec(regexp.QuoteMeta(
// 			`INSERT INTO "permissions" ("id","created_at","updated_at","role_id","resource","actions") VALUES ($1,$2,$3,$4,$5,$6)`,
// 		)).
// 			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), roleID, permissions[1].Resource, permissions[1].Actions).
// 			WillReturnResult(sqlmock.NewResult(1, 1))

// 		mock.ExpectCommit()

// 		err := repo.UpdateRoleWithPermissions(roleID, updatedRole, permissions)
// 		assert.NoError(t, err)
// 	})

// 	t.Run("Role not found", func(t *testing.T) {
// 		mock.ExpectBegin()

// 		// The implementation checks RowsAffected, not ErrRecordNotFound
// 		mock.ExpectExec(regexp.QuoteMeta(
// 			`UPDATE "roles" SET "updated_at"=$1,"name"=$2,"description"=$3 WHERE id = $4`,
// 		)).
// 			WithArgs(sqlmock.AnyArg(), updatedRole.Name, updatedRole.Description, "999").
// 			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

// 		mock.ExpectRollback()

// 		err := repo.UpdateRoleWithPermissions("999", updatedRole, permissions)

// 		var appErr *helper.AppError
// 		if assert.ErrorAs(t, err, &appErr) {
// 			assert.Equal(t, http.StatusNotFound, appErr.StatusCode)
// 			assert.Contains(t, appErr.Error(), "role with ID 999 not found")
// 		}
// 	})

// 	t.Run("Fail to delete permissions", func(t *testing.T) {
// 		mock.ExpectBegin()
// 		mock.ExpectExec(regexp.QuoteMeta(
// 			`UPDATE "roles" SET "updated_at"=$1,"name"=$2,"description"=$3 WHERE id = $4`,
// 		)).
// 			WithArgs(sqlmock.AnyArg(), updatedRole.Name, updatedRole.Description, roleID).
// 			WillReturnResult(sqlmock.NewResult(1, 1))

// 		mock.ExpectExec(regexp.QuoteMeta(
// 			`DELETE FROM "permissions" WHERE role_id = $1`,
// 		)).
// 			WithArgs(roleID).
// 			WillReturnError(errors.New("delete error"))
// 		mock.ExpectRollback()

// 		err := repo.UpdateRoleWithPermissions(roleID, updatedRole, permissions)

// 		var appErr *helper.AppError
// 		if assert.ErrorAs(t, err, &appErr) {
// 			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
// 			assert.Contains(t, appErr.Error(), "failed to delete old permissions")
// 		}
// 	})

// 	assert.NoError(t, mock.ExpectationsWereMet())
// }
// func TestDeleteRole(t *testing.T) {
// 	repo, mock := setupRoleRepository(t)
// 	roleID := "1"

// 	t.Run("Successfully delete role", func(t *testing.T) {
// 		mock.ExpectBegin()

// 		// Delete permissions
// 		mock.ExpectExec(regexp.QuoteMeta(
// 			`DELETE FROM "permissions" WHERE role_id = $1`,
// 		)).
// 			WithArgs(roleID).
// 			WillReturnResult(sqlmock.NewResult(0, 2))

// 		// Delete role
// 		mock.ExpectExec(regexp.QuoteMeta(
// 			`DELETE FROM "roles" WHERE id = $1`,
// 		)).
// 			WithArgs(roleID).
// 			WillReturnResult(sqlmock.NewResult(1, 1))

// 		mock.ExpectCommit()

// 		err := repo.DeleteRole(roleID)
// 		assert.NoError(t, err)
// 	})

// 	t.Run("Role not found", func(t *testing.T) {
// 		mock.ExpectBegin()
// 		mock.ExpectExec(regexp.QuoteMeta(
// 			`DELETE FROM "permissions" WHERE role_id = $1`,
// 		)).
// 			WithArgs("999").
// 			WillReturnResult(sqlmock.NewResult(0, 0))

// 		mock.ExpectExec(regexp.QuoteMeta(
// 			`DELETE FROM "roles" WHERE id = $1`,
// 		)).
// 			WithArgs("999").
// 			WillReturnResult(sqlmock.NewResult(0, 0))
// 		mock.ExpectRollback()

// 		err := repo.DeleteRole("999")

// 		var appErr *helper.AppError
// 		if assert.ErrorAs(t, err, &appErr) {
// 			assert.Equal(t, http.StatusNotFound, appErr.StatusCode)
// 			assert.Contains(t, appErr.Error(), "not found")
// 		}
// 	})

// 	t.Run("Fail to delete permissions", func(t *testing.T) {
// 		mock.ExpectBegin()
// 		mock.ExpectExec(regexp.QuoteMeta(
// 			`DELETE FROM "permissions" WHERE role_id = $1`,
// 		)).
// 			WithArgs(roleID).
// 			WillReturnError(errors.New("delete error"))
// 		mock.ExpectRollback()

// 		err := repo.DeleteRole(roleID)

// 		var appErr *helper.AppError
// 		if assert.ErrorAs(t, err, &appErr) {
// 			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
// 			assert.Contains(t, appErr.Error(), "failed to delete permissions")
// 		}
// 	})

// 	assert.NoError(t, mock.ExpectationsWereMet())
// }
