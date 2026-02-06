package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"multifinance-core/internal/domain/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

type mockConsumerLimitRepo struct {
	getFn    func(ctx context.Context, consumerID uint64, tenor uint8) (*entity.ConsumerLimit, error)
	updateFn func(ctx context.Context, tx *sql.Tx, consumerID uint64, tenor uint8, newUsed float64) error
}

func (m *mockConsumerLimitRepo) GetByConsumerAndTenor(ctx context.Context, consumerID uint64, tenor uint8) (*entity.ConsumerLimit, error) {
	if m.getFn != nil {
		return m.getFn(ctx, consumerID, tenor)
	}
	return nil, sql.ErrNoRows
}

func (m *mockConsumerLimitRepo) UpdateUsedLimit(ctx context.Context, tx *sql.Tx, consumerID uint64, tenor uint8, newUsed float64) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, tx, consumerID, tenor, newUsed)
	}
	return nil
}

func TestIncreaseUsedLimit_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	repo := &mockConsumerLimitRepo{
		getFn: func(ctx context.Context, consumerID uint64, tenor uint8) (*entity.ConsumerLimit, error) {
			return &entity.ConsumerLimit{ConsumerID: consumerID, TenorMonth: tenor, MaxLimit: 100.0, UsedLimit: 10.0}, nil
		},
		updateFn: func(ctx context.Context, tx *sql.Tx, consumerID uint64, tenor uint8, newUsed float64) error {
			require.NotNil(t, tx)
			require.Equal(t, float64(60.0), newUsed)
			return nil
		},
	}

	u := NewConsumerLimitUsecase(db, repo)
	err = u.IncreaseUsedLimit(context.Background(), 1, 1, 50.0)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestIncreaseUsedLimit_ExceedLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	repo := &mockConsumerLimitRepo{
		getFn: func(ctx context.Context, consumerID uint64, tenor uint8) (*entity.ConsumerLimit, error) {
			return &entity.ConsumerLimit{ConsumerID: consumerID, TenorMonth: tenor, MaxLimit: 50.0, UsedLimit: 30.0}, nil
		},
	}

	u := NewConsumerLimitUsecase(db, repo)
	err = u.IncreaseUsedLimit(context.Background(), 1, 1, 25.0)
	require.Error(t, err)
	require.EqualError(t, err, "used limit exceeds max limit")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestIncreaseUsedLimit_GetError_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	repo := &mockConsumerLimitRepo{
		getFn: func(ctx context.Context, consumerID uint64, tenor uint8) (*entity.ConsumerLimit, error) {
			return nil, errors.New("not found")
		},
	}

	u := NewConsumerLimitUsecase(db, repo)
	err = u.IncreaseUsedLimit(context.Background(), 1, 1, 10.0)
	require.Error(t, err)
	require.EqualError(t, err, "not found")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestIncreaseUsedLimit_UpdateError_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	repo := &mockConsumerLimitRepo{
		getFn: func(ctx context.Context, consumerID uint64, tenor uint8) (*entity.ConsumerLimit, error) {
			return &entity.ConsumerLimit{ConsumerID: consumerID, TenorMonth: tenor, MaxLimit: 100.0, UsedLimit: 10.0}, nil
		},
		updateFn: func(ctx context.Context, tx *sql.Tx, consumerID uint64, tenor uint8, newUsed float64) error {
			return errors.New("update failed")
		},
	}

	u := NewConsumerLimitUsecase(db, repo)
	err = u.IncreaseUsedLimit(context.Background(), 1, 1, 20.0)
	require.Error(t, err)
	require.EqualError(t, err, "update failed")
	require.NoError(t, mock.ExpectationsWereMet())
}
