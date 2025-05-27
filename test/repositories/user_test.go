package repositories

import (
	"errors"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/repositories"
	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func strPtr(s string) *string {
	return &s
}

func setupUserRepository(t *testing.T) (*repositories.UserRepository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	return &repositories.UserRepository{DB: gormDB}, mock
}

func TestCreateUser(t *testing.T) {
	repo, mock := setupUserRepository(t)

	// Define common test data
	now := time.Now()
	userId := uuid.New().String()

	// Define test cases
	testCases := []struct {
		name   string
		user   *model.User
		testFn func(*testing.T, *repositories.UserRepository, sqlmock.Sqlmock, *model.User)
	}{
		{
			name: "Success case",
			user: &model.User{
				Base: model.Base{
					ID:        userId,
					CreatedAt: now,
					UpdatedAt: now,
				},
				Username:     "testuser",
				Password:     "hashedpassword",
				MobileNumber: 9876543210,
				CountryCode:  strPtr("+91"),
				Tokens:       1000,
			},
			testFn: testCreateUserSuccess,
		},
		{
			name: "Duplicate username",
			user: &model.User{
				Username:     "duplicateuser",
				MobileNumber: 9876543210,
				CountryCode:  strPtr("+91"),
				Tokens:       1000,
			},
			testFn: testCreateUserDuplicateUsername,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.testFn(t, repo, mock, tc.user)
		})
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

func testCreateUserSuccess(t *testing.T, repo *repositories.UserRepository, mock sqlmock.Sqlmock, user *model.User) {
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO "users" ("id","created_at","updated_at","username","password","is_validated","aadhaar_number","status","name","care_of","date_of_birth","photo","email_hash","share_code","year_of_birth","mobile_number","country_code","message","address_id","tokens") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`)).
		WithArgs(
			user.ID,
			user.CreatedAt,
			user.UpdatedAt,
			user.Username,
			user.Password,
			user.IsValidated,
			user.AadhaarNumber,
			user.Status,
			user.Name,
			user.CareOf,
			user.DateOfBirth,
			user.Photo,
			user.EmailHash,
			user.ShareCode,
			user.YearOfBirth,
			user.MobileNumber,
			user.CountryCode,
			user.Message,
			user.AddressID,
			user.Tokens,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	result, err := repo.CreateUser(user)
	assert.NoError(t, err)
	assert.Equal(t, user, result)
}

func testCreateUserDuplicateUsername(t *testing.T, repo *repositories.UserRepository, mock sqlmock.Sqlmock, user *model.User) {
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO "users" ("id","created_at","updated_at","username","password","is_validated","aadhaar_number","status","name","care_of","date_of_birth","photo","email_hash","share_code","year_of_birth","mobile_number","country_code","message","address_id","tokens") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`)).
		WithArgs(
			sqlmock.AnyArg(), // ID
			sqlmock.AnyArg(), // CreatedAt
			sqlmock.AnyArg(), // UpdatedAt
			user.Username,
			sqlmock.AnyArg(), // Password
			sqlmock.AnyArg(), // IsValidated
			sqlmock.AnyArg(), // AadhaarNumber
			sqlmock.AnyArg(), // Status
			sqlmock.AnyArg(), // Name
			sqlmock.AnyArg(), // CareOf
			sqlmock.AnyArg(), // DateOfBirth
			sqlmock.AnyArg(), // Photo
			sqlmock.AnyArg(), // EmailHash
			sqlmock.AnyArg(), // ShareCode
			sqlmock.AnyArg(), // YearOfBirth
			user.MobileNumber,
			user.CountryCode,
			sqlmock.AnyArg(), // Message
			sqlmock.AnyArg(), // AddressID
			user.Tokens,
		).
		WillReturnError(errors.New("duplicate key value violates unique constraint \"users_username_key\""))
	mock.ExpectRollback()

	_, err := repo.CreateUser(user)
	assert.Error(t, err)
	assert.IsType(t, &helper.AppError{}, err)
	assert.Equal(t, http.StatusInternalServerError, err.(*helper.AppError).StatusCode)
}

func TestGetUserByID(t *testing.T) {
	repo, mock := setupUserRepository(t)
	userId := uuid.New().String()

	t.Run("Success", func(t *testing.T) {
		expectedUser := model.User{
			Base: model.Base{
				ID: userId,
			},
			Username:     "testuser",
			MobileNumber: 1234567890,
		}

		rows := sqlmock.NewRows([]string{
			"id", "username", "mobile_number", // include only fields you need to test
		}).AddRow(
			expectedUser.ID,
			expectedUser.Username,
			expectedUser.MobileNumber,
		)

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`,
		)).
			WithArgs(userId, 1).
			WillReturnRows(rows)

		// Only mock these if your GetUserByID actually loads them
		if false { // change to true if needed
			mock.ExpectQuery(regexp.QuoteMeta(
				`SELECT * FROM "addresses" WHERE "addresses"."id" = $1`,
			)).
				WithArgs(sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"id"}))

			mock.ExpectQuery(regexp.QuoteMeta(
				`SELECT * FROM "user_roles" WHERE "user_roles"."user_id" = $1`,
			)).
				WithArgs(userId).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}))
		}

		user, err := repo.GetUserByID(userId)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Username, user.Username)
	})

	t.Run("User not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`,
		)).
			WithArgs("nonexistent-id", 1).
			WillReturnError(gorm.ErrRecordNotFound)

		_, err := repo.GetUserByID("nonexistent-id")

		// Verify it returns an AppError with 404 status
		var appErr *helper.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusNotFound, appErr.StatusCode)
		}
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestGetUsers(t *testing.T) {
	repo, mock := setupUserRepository(t)

	t.Run("Success with pagination", func(t *testing.T) {
		users := []model.User{
			{Base: model.Base{ID: uuid.New().String()}, Username: "user1", MobileNumber: 1234567890},
			{Base: model.Base{ID: uuid.New().String()}, Username: "user2", MobileNumber: 9876543210},
		}

		rows := sqlmock.NewRows([]string{"id", "username", "mobile_number"}).
			AddRow(users[0].ID, users[0].Username, users[0].MobileNumber).
			AddRow(users[1].ID, users[1].Username, users[1].MobileNumber)

		// Match query with just LIMIT
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" LIMIT $1`)).
			WithArgs(10).
			WillReturnRows(rows)

		result, err := repo.GetUsers(1, 10)
		assert.NoError(t, err)
		if assert.Len(t, result, 2) {
			assert.Equal(t, users[0].ID, result[0].ID)
			assert.Equal(t, users[1].Username, result[1].Username)
		}
	})

	t.Run("No pagination", func(t *testing.T) {
		user := model.User{
			Base:         model.Base{ID: uuid.New().String()},
			Username:     "user3",
			MobileNumber: 5555555555,
		}

		rows := sqlmock.NewRows([]string{"id", "username", "mobile_number"}).
			AddRow(user.ID, user.Username, user.MobileNumber)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
			WillReturnRows(rows)

		result, err := repo.GetUsers(0, 0)
		assert.NoError(t, err)
		if assert.Len(t, result, 1) {
			assert.Equal(t, user.Username, result[0].Username)
		}
	})

	t.Run("Empty result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "mobile_number"})

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" LIMIT $1`)).
			WithArgs(10).
			WillReturnRows(rows)

		result, err := repo.GetUsers(1, 10)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" LIMIT $1`)).
			WithArgs(10).
			WillReturnError(errors.New("database error"))

		result, err := repo.GetUsers(1, 10)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "database error")
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestFindUserRoles(t *testing.T) {
	repo, mock := setupUserRepository(t)
	userId := uuid.New().String()
	roleId := "role1"

	t.Run("Success", func(t *testing.T) {
		// Mock the user roles query
		userRolesRows := sqlmock.NewRows([]string{"user_id", "role_id"}).
			AddRow(userId, roleId)

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "user_roles" WHERE user_id = $1`,
		)).
			WithArgs(userId).
			WillReturnRows(userRolesRows)

		// Mock the roles query that gets called for each role
		roleRows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(roleId, "Admin")

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "roles" WHERE "roles"."id" = $1`,
		)).
			WithArgs(roleId).
			WillReturnRows(roleRows)

		roles, err := repo.FindUserRoles(userId)
		assert.NoError(t, err)
		if assert.Len(t, roles, 1) {
			assert.Equal(t, userId, roles[0].UserID)
			assert.Equal(t, roleId, roles[0].RoleID)
			// Add additional assertions for role details if needed
		}
	})

	t.Run("No roles found", func(t *testing.T) {
		// Mock empty user roles result
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "user_roles" WHERE user_id = $1`,
		)).
			WithArgs(userId).
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "role_id"}))

		roles, err := repo.FindUserRoles(userId)
		assert.NoError(t, err)
		assert.Empty(t, roles)
	})

	t.Run("Error fetching user roles", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "user_roles" WHERE user_id = $1`,
		)).
			WithArgs(userId).
			WillReturnError(errors.New("database error"))

		roles, err := repo.FindUserRoles(userId)
		assert.Error(t, err)
		assert.Nil(t, roles)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUserRoles(t *testing.T) {
	repo, mock := setupUserRepository(t)
	userID := uuid.New().String()
	roleID := "role1"
	userRole := model.UserRole{
		UserID: userID,
		RoleID: roleID,
	}

	t.Run("Success", func(t *testing.T) {
		// Mock the duplicate check
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT count(*) FROM "user_roles" WHERE user_id = $1 AND role_id = $2`,
		)).
			WithArgs(userID, roleID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		// Mock the actual insert
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`INSERT INTO "user_roles" ("id","created_at","updated_at","user_id","role_id","is_active") VALUES ($1,$2,$3,$4,$5,$6)`,
		)).
			WithArgs(
				sqlmock.AnyArg(), // id
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				userID,
				roleID,
				false,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.CreateUserRoles(userRole)
		assert.NoError(t, err)
	})

	t.Run("Duplicate role", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT count(*) FROM "user_roles" WHERE user_id = $1 AND role_id = $2`,
		)).
			WithArgs(userID, roleID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		err := repo.CreateUserRoles(userRole)
		if assert.Error(t, err) {
			assert.Equal(t, http.StatusConflict, err.(*helper.AppError).StatusCode)
		}
	})

	t.Run("Database error on duplicate check", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT count(*) FROM "user_roles" WHERE user_id = $1 AND role_id = $2`,
		)).
			WithArgs(userID, roleID).
			WillReturnError(errors.New("db error"))

		err := repo.CreateUserRoles(userRole)
		assert.Error(t, err)
	})

	t.Run("Database error on insert", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT count(*) FROM "user_roles" WHERE user_id = $1 AND role_id = $2`,
		)).
			WithArgs(userID, roleID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`INSERT INTO "user_roles" ("id","created_at","updated_at","user_id","role_id","is_active") VALUES ($1,$2,$3,$4,$5,$6)`,
		)).
			WithArgs(
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				userID,
				roleID,
				false,
			).
			WillReturnError(errors.New("insert failed"))
		mock.ExpectRollback()

		err := repo.CreateUserRoles(userRole)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestUpdatePassword(t *testing.T) {
	repo, mock := setupUserRepository(t)
	userId := uuid.New().String()
	newPassword := "newhashedpassword"

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`UPDATE "users" SET "password"=$1 WHERE id = $2`,
		)).
			WithArgs(newPassword, userId).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdatePassword(userId, newPassword)
		assert.NoError(t, err)
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`UPDATE "users" SET "password"=$1 WHERE id = $2`,
		)).
			WithArgs(newPassword, userId).
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		err := repo.UpdatePassword(userId, newPassword)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreditDebitUserByID(t *testing.T) {
	repo, mock := setupUserRepository(t)
	userId := uuid.New().String()
	tokens := 100

	t.Run("Credit success", func(t *testing.T) {
		// Mock SELECT query
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`,
		)).
			WithArgs(userId, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "tokens"}).AddRow(userId, 500))

		// Mock transaction begin
		mock.ExpectBegin()

		// Mock UPDATE query
		mock.ExpectExec(regexp.QuoteMeta(
			`UPDATE "users" SET "created_at"=$1,"updated_at"=$2,"username"=$3,"password"=$4,"is_validated"=$5,"aadhaar_number"=$6,"status"=$7,"name"=$8,"care_of"=$9,"date_of_birth"=$10,"photo"=$11,"email_hash"=$12,"share_code"=$13,"year_of_birth"=$14,"mobile_number"=$15,"country_code"=$16,"message"=$17,"address_id"=$18,"tokens"=$19 WHERE "id" = $20`,
		)).
			WithArgs(
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				"",               // username
				"",               // password
				false,            // is_validated
				nil,              // aadhaar_number
				nil,              // status
				nil,              // name
				nil,              // care_of
				nil,              // date_of_birth
				nil,              // photo
				nil,              // email_hash
				nil,              // share_code
				nil,              // year_of_birth
				0,                // mobile_number
				nil,              // country_code
				nil,              // message
				nil,              // address_id
				600,              // tokens
				userId,           // id
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock transaction commit
		mock.ExpectCommit()

		user, err := repo.CreditUserByID(userId, tokens)
		require.NoError(t, err)
		assert.Equal(t, 600, user.Tokens)
	})

	t.Run("Debit success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`,
		)).
			WithArgs(userId, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "tokens"}).AddRow(userId, 500))

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`UPDATE "users" SET "created_at"=$1,"updated_at"=$2,"username"=$3,"password"=$4,"is_validated"=$5,"aadhaar_number"=$6,"status"=$7,"name"=$8,"care_of"=$9,"date_of_birth"=$10,"photo"=$11,"email_hash"=$12,"share_code"=$13,"year_of_birth"=$14,"mobile_number"=$15,"country_code"=$16,"message"=$17,"address_id"=$18,"tokens"=$19 WHERE "id" = $20`,
		)).
			WithArgs(
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				"",               // username
				"",               // password
				false,            // is_validated
				nil,              // aadhaar_number
				nil,              // status
				nil,              // name
				nil,              // care_of
				nil,              // date_of_birth
				nil,              // photo
				nil,              // email_hash
				nil,              // share_code
				nil,              // year_of_birth
				0,                // mobile_number
				nil,              // country_code
				nil,              // message
				nil,              // address_id
				400,              // tokens (500-100)
				userId,           // id
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		user, err := repo.DebitUserByID(userId, tokens)
		require.NoError(t, err)
		assert.Equal(t, 400, user.Tokens)
	})

	t.Run("User not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`,
		)).
			WithArgs("nonexistent-id", 1).
			WillReturnError(gorm.ErrRecordNotFound)

		_, err := repo.CreditUserByID("nonexistent-id", tokens)
		require.Error(t, err)
		assert.Equal(t, http.StatusNotFound, err.(*helper.AppError).StatusCode)
	})

	t.Run("Debit insufficient tokens", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`,
		)).
			WithArgs(userId, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "tokens"}).AddRow(userId, 50))

		_, err := repo.DebitUserByID(userId, 100)
		require.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, err.(*helper.AppError).StatusCode)
	})

	t.Run("Database error on select", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`,
		)).
			WithArgs(userId, 1).
			WillReturnError(errors.New("database error"))

		_, err := repo.CreditUserByID(userId, tokens)
		require.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, err.(*helper.AppError).StatusCode)
	})

	t.Run("Database error on update", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`,
		)).
			WithArgs(userId, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "tokens"}).AddRow(userId, 500))

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`UPDATE "users" SET "created_at"=$1,"updated_at"=$2,"username"=$3,"password"=$4,"is_validated"=$5,"aadhaar_number"=$6,"status"=$7,"name"=$8,"care_of"=$9,"date_of_birth"=$10,"photo"=$11,"email_hash"=$12,"share_code"=$13,"year_of_birth"=$14,"mobile_number"=$15,"country_code"=$16,"message"=$17,"address_id"=$18,"tokens"=$19 WHERE "id" = $20`,
		)).
			WillReturnError(errors.New("update failed"))
		mock.ExpectRollback()

		_, err := repo.CreditUserByID(userId, tokens)
		require.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
