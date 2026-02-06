package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"multifinance-core/internal/domain/entity"
)

func setupConsumerTransactionMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, ConsumerTransactionRepository, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock db: %v", err)
	}

	repo := NewConsumerTransactionRepo(db)

	cleanup := func() { db.Close() }
	return db, mock, repo, cleanup
}

func TestConsumerTransactionRepo_Create_Success(t *testing.T) {
	db, mock, repo, cleanup := setupConsumerTransactionMockDB(t)
	defer cleanup()

	ctx := context.Background()
	mock.ExpectBegin()
	tx, _ := db.Begin()

	tr := &entity.Transaction{
		ContractNo:      "C-1-1",
		ConsumerID:      1,
		ConsumerLimitID: 2,
		AssetID:         3,
		TenorMonth:      3,
		OTR:             1000,
		AdminFee:        50,
		JumlahBunga:     20,
		JumlahCicilan:   340,
		Status:          "SUCCESS",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO consumer_transactions (contract_no, consumer_id, consumer_limit_id, asset_id, tenor_month, otr, admin_fee, jumlah_bunga, jumlah_cicilan, status, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs(tr.ContractNo, tr.ConsumerID, tr.ConsumerLimitID, tr.AssetID, tr.TenorMonth, tr.OTR, tr.AdminFee, tr.JumlahBunga, tr.JumlahCicilan, tr.Status, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(5, 1))

	id, err := repo.Create(ctx, tx, tr)
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), id)
	mock.ExpectCommit()
	tx.Commit()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConsumerTransactionRepo_Create_ExecError(t *testing.T) {
	db, mock, repo, cleanup := setupConsumerTransactionMockDB(t)
	defer cleanup()

	ctx := context.Background()
	mock.ExpectBegin()
	tx, _ := db.Begin()

	tr := &entity.Transaction{ContractNo: "C-1-2", ConsumerID: 1}

	mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO consumer_transactions (contract_no, consumer_id, consumer_limit_id, asset_id, tenor_month, otr, admin_fee, jumlah_bunga, jumlah_cicilan, status, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	id, err := repo.Create(ctx, tx, tr)
	assert.Error(t, err)
	assert.Equal(t, uint64(0), id)
	mock.ExpectRollback()
	tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConsumerTransactionRepo_Create_LastInsertError(t *testing.T) {
	db, mock, repo, cleanup := setupConsumerTransactionMockDB(t)
	defer cleanup()

	ctx := context.Background()
	mock.ExpectBegin()
	tx, _ := db.Begin()

	tr := &entity.Transaction{ContractNo: "C-1-3", ConsumerID: 1}

	mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO consumer_transactions (contract_no, consumer_id, consumer_limit_id, asset_id, tenor_month, otr, admin_fee, jumlah_bunga, jumlah_cicilan, status, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))

	id, err := repo.Create(ctx, tx, tr)
	assert.Error(t, err)
	assert.Equal(t, uint64(0), id)
	mock.ExpectRollback()
	tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConsumerTransactionRepo_ListByConsumer_Success(t *testing.T) {
	_, mock, repo, cleanup := setupConsumerTransactionMockDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "contract_no", "consumer_id", "consumer_limit_id", "asset_id", "tenor_month", "otr", "admin_fee", "jumlah_bunga", "jumlah_cicilan", "status", "created_at"}).
		AddRow(1, "C-1-1", 1, 2, 3, 3, 1000, 50, 20, 340, "SUCCESS", now).
		AddRow(2, "C-1-2", 1, 2, 4, 6, 1500, 75, 30, 435, "SUCCESS", now)

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, contract_no, consumer_id, consumer_limit_id, asset_id, tenor_month, otr, admin_fee, jumlah_bunga, jumlah_cicilan, status, created_at
        FROM consumer_transactions WHERE consumer_id = ? ORDER BY created_at DESC`)).
		WithArgs(uint64(1)).
		WillReturnRows(rows)

	res, err := repo.ListByConsumer(ctx, 1)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, uint64(1), res[0].ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConsumerTransactionRepo_ListByConsumer_QueryError(t *testing.T) {
	_, mock, repo, cleanup := setupConsumerTransactionMockDB(t)
	defer cleanup()

	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, contract_no, consumer_id, consumer_limit_id, asset_id, tenor_month, otr, admin_fee, jumlah_bunga, jumlah_cicilan, status, created_at
        FROM consumer_transactions WHERE consumer_id = ? ORDER BY created_at DESC`)).
		WithArgs(uint64(99)).
		WillReturnError(sql.ErrConnDone)

	res, err := repo.ListByConsumer(ctx, 99)
	assert.Error(t, err)
	assert.Nil(t, res)

	assert.NoError(t, mock.ExpectationsWereMet())
}
