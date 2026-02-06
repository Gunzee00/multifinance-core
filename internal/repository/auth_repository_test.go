package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"multifinance-core/internal/domain/entity"
)

func setupAuthMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, AuthRepository, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}

	repo := NewAuthRepo(db)

	cleanup := func() {
		db.Close()
	}

	return db, mock, repo, cleanup
}

func TestAuthRepo_Create_Success(t *testing.T) {
	db, mock, repo, cleanup := setupAuthMockDB(t)
	defer cleanup()

	ctx := context.Background()

	user := &entity.AuthUser{
		ConsumerID: 1,
		Email:      "budi@mail.com",
		Password:   "hashedpassword",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO auth_users (consumer_id, email, password)
		VALUES (?, ?, ?)`)).
		WithArgs(user.ConsumerID, user.Email, user.Password).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tx, _ := db.Begin()

	err := repo.Create(ctx, tx, user)
	assert.NoError(t, err)

	tx.Commit()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepo_Create_Error(t *testing.T) {
	db, mock, repo, cleanup := setupAuthMockDB(t)
	defer cleanup()

	ctx := context.Background()

	user := &entity.AuthUser{
		ConsumerID: 1,
		Email:      "duplicate@mail.com",
		Password:   "hashedpassword",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO auth_users (consumer_id, email, password)
		VALUES (?, ?, ?)`)).
		WithArgs(user.ConsumerID, user.Email, user.Password).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	tx, _ := db.Begin()

	err := repo.Create(ctx, tx, user)
	assert.Error(t, err)

	tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepo_FindByEmail_Success(t *testing.T) {
	_, mock, repo, cleanup := setupAuthMockDB(t)
	defer cleanup()

	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "consumer_id", "email", "password",
	}).AddRow(1, 10, "budi@mail.com", "hashedpassword")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, consumer_id, email, password
		FROM auth_users WHERE email = ?`)).
		WithArgs("budi@mail.com").
		WillReturnRows(rows)

	user, err := repo.FindByEmail(ctx, "budi@mail.com")

	assert.NoError(t, err)
	assert.Equal(t, uint64(1), user.ID)
	assert.Equal(t, uint64(10), user.ConsumerID)
	assert.Equal(t, "budi@mail.com", user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepo_FindByEmail_NotFound(t *testing.T) {
	_, mock, repo, cleanup := setupAuthMockDB(t)
	defer cleanup()

	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, consumer_id, email, password
		FROM auth_users WHERE email = ?`)).
		WithArgs("notfound@mail.com").
		WillReturnError(sql.ErrNoRows)

	user, err := repo.FindByEmail(ctx, "notfound@mail.com")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
