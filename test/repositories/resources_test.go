package repositories

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupResourceRepository(t *testing.T) (*repositories.ResourceRepository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	return &repositories.ResourceRepository{DB: gormDB}, mock
}

func TestCheckIfResourceExists(t *testing.T) {
	repo, mock := setupResourceRepository(t)
	resourceName := "test_resource"

	t.Run("Resource exists", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
			AddRow("1", resourceName, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "resources" WHERE name = $1 ORDER BY "resources"."id" LIMIT $2`,
		)).
			WithArgs(resourceName, 1).
			WillReturnRows(rows)

		err := repo.CheckIfResourceExists(resourceName)

		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusConflict, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "already exists")
		}
	})

	t.Run("Resource does not exist", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "resources" WHERE name = $1 ORDER BY "resources"."id" LIMIT $2`,
		)).
			WithArgs("nonexistent_resource", 1).
			WillReturnError(gorm.ErrRecordNotFound)

		err := repo.CheckIfResourceExists("nonexistent_resource")
		assert.NoError(t, err)
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "resources" WHERE name = $1 ORDER BY "resources"."id" LIMIT $2`,
		)).
			WithArgs(resourceName, 1).
			WillReturnError(errors.New("db error"))

		err := repo.CheckIfResourceExists(resourceName)

		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "database error")
		}
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateResource(t *testing.T) {
	repo, mock := setupResourceRepository(t)
	newResource := &model.Resource{
		Base: model.Base{
			ID:        "1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name: "new_resource",
	}

	t.Run("Successfully create resource", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`INSERT INTO "resources" ("id","created_at","updated_at","name") VALUES ($1,$2,$3,$4)`,
		)).
			WithArgs(newResource.ID, newResource.CreatedAt, newResource.UpdatedAt, newResource.Name).
			WillReturnResult(sqlmock.NewResult(1, 1)) // lastInsertID, rowsAffected
		mock.ExpectCommit()

		err := repo.CreateResource(newResource)
		assert.NoError(t, err)
	})

	t.Run("Database error on create", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`INSERT INTO "resources" ("id","created_at","updated_at","name") VALUES ($1,$2,$3,$4)`,
		)).
			WithArgs(newResource.ID, newResource.CreatedAt, newResource.UpdatedAt, newResource.Name).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.CreateResource(newResource)

		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "failed to create resource")
		}
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindResourceByID(t *testing.T) {
	repo, mock := setupResourceRepository(t)
	resourceID := "1"
	expectedResource := &model.Resource{
		Base: model.Base{
			ID:        resourceID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name: "test_resource",
	}

	t.Run("Successfully find resource", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(expectedResource.ID, expectedResource.CreatedAt, expectedResource.UpdatedAt, expectedResource.Name)

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "resources" WHERE id = $1 ORDER BY "resources"."id" LIMIT $2`,
		)).
			WithArgs(resourceID, 1).
			WillReturnRows(rows)

		result, err := repo.FindResourceByID(resourceID)
		assert.NoError(t, err)
		assert.Equal(t, expectedResource, result)
	})

	t.Run("Resource not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "resources" WHERE id = $1 ORDER BY "resources"."id" LIMIT $2`,
		)).
			WithArgs("nonexistent_id", 1).
			WillReturnError(gorm.ErrRecordNotFound)

		result, err := repo.FindResourceByID("nonexistent_id")

		assert.Nil(t, result)
		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusNotFound, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "not found")
		}
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "resources" WHERE id = $1 ORDER BY "resources"."id" LIMIT $2`,
		)).
			WithArgs(resourceID, 1).
			WillReturnError(errors.New("db error"))

		result, err := repo.FindResourceByID(resourceID)

		assert.Nil(t, result)
		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "failed to query resource")
		}
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindResources(t *testing.T) {
	repo, mock := setupResourceRepository(t)
	now := time.Now()
	expectedResources := []model.Resource{
		{
			Base: model.Base{
				ID:        "1",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name: "resource1",
		},
		{
			Base: model.Base{
				ID:        "2",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name: "resource2",
		},
	}

	t.Run("Successfully find resources without filter", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(expectedResources[0].ID, expectedResources[0].CreatedAt, expectedResources[0].UpdatedAt, expectedResources[0].Name).
			AddRow(expectedResources[1].ID, expectedResources[1].CreatedAt, expectedResources[1].UpdatedAt, expectedResources[1].Name)

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "resources"`,
		)).
			WillReturnRows(rows)

		result, err := repo.FindResources(map[string]interface{}{}, 0, 0)
		assert.NoError(t, err)
		assert.Equal(t, expectedResources, result)
	})

	t.Run("Successfully find resources with filter and pagination", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(expectedResources[0].ID, expectedResources[0].CreatedAt, expectedResources[0].UpdatedAt, expectedResources[0].Name)

		// Updated to match the actual query being executed
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "resources" WHERE name ILIKE $1 LIMIT $2`,
		)).
			WithArgs("%test%", 10).
			WillReturnRows(rows)

		result, err := repo.FindResources(map[string]interface{}{"name": "test"}, 1, 10)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "resources"`,
		)).
			WillReturnError(errors.New("db error"))

		result, err := repo.FindResources(map[string]interface{}{}, 0, 0)

		assert.Nil(t, result)
		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "failed to retrieve resources")
		}
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestUpdateResource(t *testing.T) {
	repo, mock := setupResourceRepository(t)
	resourceID := "1"

	// Create a fixed time reference for consistent testing
	fixedTime := time.Now()
	updatedResource := model.Resource{
		Base: model.Base{
			UpdatedAt: fixedTime,
		},
		Name: "updated_name",
	}

	t.Run("Successfully update resource", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`UPDATE "resources" SET "updated_at"=$1,"name"=$2 WHERE id = $3`,
		)).
			// Use AnyArg for time to avoid exact match issues
			WithArgs(AnyTime{}, updatedResource.Name, resourceID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateResource(resourceID, updatedResource)
		assert.NoError(t, err)
	})

	t.Run("Resource not found", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`UPDATE "resources" SET "updated_at"=$1,"name"=$2 WHERE id = $3`,
		)).
			WithArgs(AnyTime{}, updatedResource.Name, "nonexistent_id").
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit() // Changed to ExpectCommit to match GORM behavior

		err := repo.UpdateResource("nonexistent_id", updatedResource)

		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusNotFound, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "resource with ID nonexistent_id not found")
		}
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`UPDATE "resources" SET "updated_at"=$1,"name"=$2 WHERE id = $3`,
		)).
			WithArgs(AnyTime{}, updatedResource.Name, resourceID).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.UpdateResource(resourceID, updatedResource)

		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "failed to update resource")
		}
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

// AnyTime is a custom matcher for time arguments
type AnyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func TestDeleteResource(t *testing.T) {
	repo, mock := setupResourceRepository(t)
	resourceID := "1"

	t.Run("Successfully delete resource", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`DELETE FROM "resources" WHERE id = $1`,
		)).
			WithArgs(resourceID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.DeleteResource(resourceID)
		assert.NoError(t, err)
	})

	t.Run("Resource not found", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "resources" WHERE id = $1`)).
			WithArgs("nonexistent_id").
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit() // Changed from ExpectRollback to ExpectCommit

		err := repo.DeleteResource("nonexistent_id")

		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusNotFound, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "resource with ID nonexistent_id not found")
		}
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`DELETE FROM "resources" WHERE id = $1`,
		)).
			WithArgs(resourceID).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.DeleteResource(resourceID)

		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "failed to delete resource")
		}
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
