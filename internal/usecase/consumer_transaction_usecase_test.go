package usecase

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"regexp"
	"testing"

	"multifinance-core/internal/domain/entity"

	"github.com/DATA-DOG/go-sqlmock"
)

type mockAssetRepoTx struct {
	getByIDFn func(ctx context.Context, id uint64) (*entity.Asset, error)
}

func (m *mockAssetRepoTx) GetByID(ctx context.Context, id uint64) (*entity.Asset, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockAssetRepoTx) Create(ctx context.Context, tx *sql.Tx, a *entity.Asset) (uint64, error) {
	return 0, nil
}
func (m *mockAssetRepoTx) List(ctx context.Context) ([]*entity.Asset, error)             { return nil, nil }
func (m *mockAssetRepoTx) Update(ctx context.Context, tx *sql.Tx, a *entity.Asset) error { return nil }
func (m *mockAssetRepoTx) Delete(ctx context.Context, tx *sql.Tx, id uint64) error       { return nil }

type mockTxRepoTx struct {
	createFn         func(ctx context.Context, tx *sql.Tx, t *entity.Transaction) (uint64, error)
	listByConsumerFn func(ctx context.Context, consumerID uint64) ([]*entity.Transaction, error)
}

func (m *mockTxRepoTx) Create(ctx context.Context, tx *sql.Tx, t *entity.Transaction) (uint64, error) {
	return m.createFn(ctx, tx, t)
}

func (m *mockTxRepoTx) ListByConsumer(ctx context.Context, consumerID uint64) ([]*entity.Transaction, error) {
	return m.listByConsumerFn(ctx, consumerID)
}
func TestPurchase_InvalidTenor(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	uc := NewConsumerTransactionUsecase(db, nil, nil, nil)

	_, err := uc.Purchase(context.Background(), 1, 1, 5)

	if !errors.Is(err, ErrInvalidTenor) {
		t.Fatal("expected invalid tenor error")
	}
}
func TestPurchase_InsufficientLimit(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectBegin()

	// SELECT FOR UPDATE
	rows := sqlmock.NewRows([]string{
		"id", "consumer_id", "tenor_month", "max_limit", "used_limit",
	}).AddRow(1, 1, 3, 1000000.0, 900000.0)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, consumer_id, tenor_month, max_limit, used_limit FROM consumer_limits WHERE consumer_id = ? AND tenor_month = ? FOR UPDATE`,
	)).WithArgs(1, 3).WillReturnRows(rows)

	mock.ExpectCommit()

	assetRepo := &mockAssetRepoTx{
		getByIDFn: func(ctx context.Context, id uint64) (*entity.Asset, error) {
			return &entity.Asset{
				ID:           1,
				PriceProduct: 500000,
			}, nil
		},
	}

	txRepo := &mockTxRepoTx{
		createFn: func(ctx context.Context, tx *sql.Tx, tr *entity.Transaction) (uint64, error) {
			if tr.Status != "FAILED" {
				t.Fatal("expected FAILED status")
			}
			return 1, nil
		},
	}

	uc := NewConsumerTransactionUsecase(db, assetRepo, nil, txRepo)

	_, err := uc.Purchase(context.Background(), 1, 1, 3)

	if !errors.Is(err, ErrInsufficientLimit) {
		t.Fatal("expected insufficient limit error")
	}
}
func TestPurchase_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{
		"id", "consumer_id", "tenor_month", "max_limit", "used_limit",
	}).AddRow(1, 1, 3, 5000000.0, 0.0)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, consumer_id, tenor_month, max_limit, used_limit FROM consumer_limits WHERE consumer_id = ? AND tenor_month = ? FOR UPDATE`,
	)).WithArgs(1, 3).WillReturnRows(rows)

	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE consumer_limits SET used_limit = ?, updated_at = ? WHERE id = ?`,
	)).WithArgs(1000000.0, sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	price := 1000000.0

	assetRepo := &mockAssetRepoTx{
		getByIDFn: func(ctx context.Context, id uint64) (*entity.Asset, error) {
			return &entity.Asset{
				ID:           1,
				PriceProduct: price,
			}, nil
		},
	}

	txRepo := &mockTxRepoTx{
		createFn: func(ctx context.Context, tx *sql.Tx, tr *entity.Transaction) (uint64, error) {

			expectedAdmin := int64(math.Round(price * 0.05))
			expectedBunga := int64(math.Round(price * 0.02 * 3))

			if tr.AdminFee != expectedAdmin {
				t.Fatal("admin fee mismatch")
			}
			if tr.JumlahBunga != expectedBunga {
				t.Fatal("bunga mismatch")
			}
			if tr.Status != "SUCCESS" {
				t.Fatal("status should be SUCCESS")
			}

			return 1, nil
		},
	}

	uc := NewConsumerTransactionUsecase(db, assetRepo, nil, txRepo)

	tr, err := uc.Purchase(context.Background(), 1, 1, 3)
	if err != nil {
		t.Fatal(err)
	}

	if tr.Status != "SUCCESS" {
		t.Fatal("transaction should success")
	}
}
func TestListByConsumer(t *testing.T) {
	expected := []*entity.Transaction{
		{ID: 1, ConsumerID: 1},
	}

	txRepo := &mockTxRepoTx{
		listByConsumerFn: func(ctx context.Context, consumerID uint64) ([]*entity.Transaction, error) {
			return expected, nil
		},
	}

	uc := NewConsumerTransactionUsecase(nil, nil, nil, txRepo)

	result, err := uc.ListByConsumer(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 1 {
		t.Fatal("should return 1 record")
	}
}
