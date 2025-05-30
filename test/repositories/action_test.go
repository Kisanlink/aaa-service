package repositories

import (
	"errors"
	"net/http"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupActionRepository(t *testing.T) (*repositories.ActionRepository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	return &repositories.ActionRepository{DB: gormDB}, mock
}

func TestCheckIfActionExists(t *testing.T) {
	repo, mock := setupActionRepository(t)
	actionName := "test_action"

	t.Run("Action exists", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow("1", actionName)

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "actions" WHERE name = $1 ORDER BY "actions"."id" LIMIT $2`,
		)).
			WithArgs(actionName, 1).
			WillReturnRows(rows)

		err := repo.CheckIfActionExists(actionName)

		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusConflict, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "already exists")
		}
	})

	t.Run("Action does not exist", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "actions" WHERE name = $1 ORDER BY "actions"."id" LIMIT $2`,
		)).
			WithArgs("nonexistent_action", 1).
			WillReturnError(gorm.ErrRecordNotFound)

		err := repo.CheckIfActionExists("nonexistent_action")
		assert.NoError(t, err)
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "actions" WHERE name = $1 ORDER BY "actions"."id" LIMIT $2`,
		)).
			WithArgs(actionName, 1).
			WillReturnError(errors.New("db error"))

		err := repo.CheckIfActionExists(actionName)

		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "database error")
		}
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAction(t *testing.T) {
	repo, mock := setupActionRepository(t)
	newAction := &model.Action{
		Name: "new_action",
	}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`INSERT INTO "actions" ("id","created_at","updated_at","name") VALUES ($1,$2,$3,$4)`,
		)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), newAction.Name).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.CreateAction(newAction)
		assert.NoError(t, err)
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`INSERT INTO "actions" ("id","created_at","updated_at","name") VALUES ($1,$2,$3,$4)`,
		)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), newAction.Name).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.CreateAction(newAction)

		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "failed to create action")
		}
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindActionByID(t *testing.T) {
	repo, mock := setupActionRepository(t)
	actionID := "1"
	actionName := "test_action"

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(actionID, actionName)

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "actions" WHERE id = $1 ORDER BY "actions"."id" LIMIT $2`,
		)).
			WithArgs(actionID, 1).
			WillReturnRows(rows)

		action, err := repo.FindActionByID(actionID)
		assert.NoError(t, err)
		assert.Equal(t, actionID, action.ID)
		assert.Equal(t, actionName, action.Name)
	})

	t.Run("Action not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "actions" WHERE id = $1 ORDER BY "actions"."id" LIMIT $2`,
		)).
			WithArgs("nonexistent_id", 1).
			WillReturnError(gorm.ErrRecordNotFound)

		_, err := repo.FindActionByID("nonexistent_id")

		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusNotFound, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "not found")
		}
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "actions" WHERE id = $1 ORDER BY "actions"."id" LIMIT $2`,
		)).
			WithArgs(actionID, 1).
			WillReturnError(errors.New("db error"))

		_, err := repo.FindActionByID(actionID)

		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "failed to query action")
		}
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindActions(t *testing.T) {
	repo, mock := setupActionRepository(t)

	t.Run("Success without filters", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow("1", "action1").
			AddRow("2", "action2")

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "actions"`,
		)).
			WillReturnRows(rows)

		actions, err := repo.FindActions(nil, 0, 0)
		assert.NoError(t, err)
		assert.Len(t, actions, 2)
	})

	t.Run("Success with filters", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow("1", "test_action")

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "actions" WHERE name ILIKE $1`,
		)).
			WithArgs("test%").
			WillReturnRows(rows)

		filter := map[string]interface{}{
			"name": "test%",
		}
		actions, err := repo.FindActions(filter, 0, 0)
		assert.NoError(t, err)
		assert.Len(t, actions, 1)
	})

	t.Run("Success with pagination", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow("1", "action1")

		// Expect only LIMIT without OFFSET
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "actions" LIMIT $1`,
		)).
			WithArgs(1).
			WillReturnRows(rows)

		// Second parameter is limit, third is page (but code only uses limit)
		actions, err := repo.FindActions(nil, 1, 1)
		assert.NoError(t, err)
		assert.Len(t, actions, 1)
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "actions"`,
		)).
			WillReturnError(errors.New("db error"))

		_, err := repo.FindActions(nil, 0, 0)

		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "failed to retrieve actions")
		}
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAction(t *testing.T) {
	repo, mock := setupActionRepository(t)
	actionID := "1"
	updatedAction := model.Action{
		Name: "updated_action",
	}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`UPDATE "actions" SET "updated_at"=$1,"name"=$2 WHERE id = $3`,
		)).
			WithArgs(sqlmock.AnyArg(), updatedAction.Name, actionID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateAction(actionID, updatedAction)
		assert.NoError(t, err)
	})

	t.Run("Action not found", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`UPDATE "actions" SET "updated_at"=$1,"name"=$2 WHERE id = $3`,
		)).
			WithArgs(sqlmock.AnyArg(), updatedAction.Name, "nonexistent_id").
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.UpdateAction("nonexistent_id", updatedAction)

		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusNotFound, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "not found")
		}
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`UPDATE "actions" SET "updated_at"=$1,"name"=$2 WHERE id = $3`,
		)).
			WithArgs(sqlmock.AnyArg(), updatedAction.Name, actionID).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.UpdateAction(actionID, updatedAction)

		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "failed to update action")
		}
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteAction(t *testing.T) {
	repo, mock := setupActionRepository(t)
	actionID := "1"

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`DELETE FROM "actions" WHERE id = $1`,
		)).
			WithArgs(actionID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.DeleteAction(actionID)
		assert.NoError(t, err)
	})

	t.Run("Action not found", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`DELETE FROM "actions" WHERE id = $1`,
		)).
			WithArgs("nonexistent_id").
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.DeleteAction("nonexistent_id")

		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusNotFound, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "not found")
		}
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`DELETE FROM "actions" WHERE id = $1`,
		)).
			WithArgs(actionID).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.DeleteAction(actionID)

		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
			assert.Contains(t, appErr.Error(), "failed to delete action")
		}
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
