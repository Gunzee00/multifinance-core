package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupConsumerLimitMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, ConsumerLimitRepository, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}

	repo := NewConsumerLimitRepo(db)

	cleanup := func() {
		db.Close()
	}

	return db, mock, repo, cleanup
}

func TestConsumerLimitRepo_GetByConsumerAndTenor_Success(t *testing.T) {
	_, mock, repo, cleanup := setupConsumerLimitMockDB(t)
	defer cleanup()

	ctx := context.Background()

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "consumer_id", "tenor_month", "max_limit", "used_limit", "created_at", "updated_at",
	}).AddRow(1, 10, 3, 10000000.0, 2000000.0, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, consumer_id, tenor_month, max_limit, used_limit, created_at, updated_at
        FROM consumer_limits WHERE consumer_id = ? AND tenor_month = ?`)).
		WithArgs(uint64(10), uint8(3)).
		WillReturnRows(rows)

	cl, err := repo.GetByConsumerAndTenor(ctx, 10, 3)

	assert.NoError(t, err)
	assert.Equal(t, uint64(1), cl.ID)
	assert.Equal(t, uint64(10), cl.ConsumerID)
	assert.Equal(t, uint8(3), cl.TenorMonth)
	assert.Equal(t, 10000000.0, cl.MaxLimit)
	assert.Equal(t, 2000000.0, cl.UsedLimit)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConsumerLimitRepo_GetByConsumerAndTenor_NotFound(t *testing.T) {
	_, mock, repo, cleanup := setupConsumerLimitMockDB(t)
	defer cleanup()

	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, consumer_id, tenor_month, max_limit, used_limit, created_at, updated_at
        FROM consumer_limits WHERE consumer_id = ? AND tenor_month = ?`)).
		WithArgs(uint64(99), uint8(6)).
		WillReturnError(sql.ErrNoRows)

	cl, err := repo.GetByConsumerAndTenor(ctx, 99, 6)

	assert.Error(t, err)
	assert.Nil(t, cl)
	assert.Equal(t, sql.ErrNoRows, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConsumerLimitRepo_UpdateUsedLimit_Success(t *testing.T) {
	db, mock, repo, cleanup := setupConsumerLimitMockDB(t)
	defer cleanup()

	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE consumer_limits SET used_limit = ?, updated_at = ? WHERE consumer_id = ? AND tenor_month = ?`,
	)).
		WithArgs(3000000.0, sqlmock.AnyArg(), uint64(10), uint8(3)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tx, _ := db.Begin()

	err := repo.UpdateUsedLimit(ctx, tx, 10, 3, 3000000.0)
	assert.NoError(t, err)

	tx.Commit()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConsumerLimitRepo_UpdateUsedLimit_Error(t *testing.T) {
	db, mock, repo, cleanup := setupConsumerLimitMockDB(t)
	defer cleanup()

	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE consumer_limits SET used_limit = ?, updated_at = ? WHERE consumer_id = ? AND tenor_month = ?`,
	)).
		WithArgs(3000000.0, sqlmock.AnyArg(), uint64(10), uint8(3)).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	tx, _ := db.Begin()

	err := repo.UpdateUsedLimit(ctx, tx, 10, 3, 3000000.0)
	assert.Error(t, err)

	tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}
