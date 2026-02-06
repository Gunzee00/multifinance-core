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

type mockConsumerRepo struct {
	createFn func(ctx context.Context, tx *sql.Tx, c *entity.Consumer) (uint64, error)
}

func (m *mockConsumerRepo) Create(ctx context.Context, tx *sql.Tx, c *entity.Consumer) (uint64, error) {
	if m.createFn != nil {
		return m.createFn(ctx, tx, c)
	}
	return 0, nil
}

type mockAuthRepoForRegister struct {
	createFn func(ctx context.Context, tx *sql.Tx, u *entity.AuthUser) error
}

func (m *mockAuthRepoForRegister) Create(ctx context.Context, tx *sql.Tx, u *entity.AuthUser) error {
	if m.createFn != nil {
		return m.createFn(ctx, tx, u)
	}
	return nil
}
func (m *mockAuthRepoForRegister) FindByEmail(ctx context.Context, email string) (*entity.AuthUser, error) {
	return nil, sql.ErrNoRows
}

func TestRegister_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Expect transaction begin and commit, and 4 inserts into consumer_limits
	mock.ExpectBegin()
	// we expect 4 insert execs for consumer_limits
	for i := 0; i < 4; i++ {
		mock.ExpectExec("INSERT INTO consumer_limits").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	created := uint64(77)
	consumerRepo := &mockConsumerRepo{
		createFn: func(ctx context.Context, tx *sql.Tx, c *entity.Consumer) (uint64, error) {
			require.NotNil(t, tx)
			require.Equal(t, "08123", c.NIK)
			return created, nil
		},
	}

	authRepo := &mockAuthRepoForRegister{
		createFn: func(ctx context.Context, tx *sql.Tx, u *entity.AuthUser) error {
			require.NotNil(t, tx)
			require.Equal(t, created, u.ConsumerID)
			require.NotEmpty(t, u.Password)
			return nil
		},
	}

	u := NewAuthUsecase(db, consumerRepo, authRepo)

	req := RegisterRequest{
		NIK:         "08123",
		FullName:    "Test User",
		LegalName:   "Test User",
		BirthPlace:  "City",
		BirthDate:   "1990-01-01",
		Salary:      1000,
		Email:       "t@example.com",
		KTPPhoto:    "ktp.jpg",
		SelfiePhoto: "selfie.jpg",
		Password:    "secret",
	}

	err = u.Register(context.Background(), req)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRegister_ConsumerCreateError_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	consumerRepo := &mockConsumerRepo{
		createFn: func(ctx context.Context, tx *sql.Tx, c *entity.Consumer) (uint64, error) {
			return 0, errors.New("create failed")
		},
	}
	authRepo := &mockAuthRepoForRegister{}

	u := NewAuthUsecase(db, consumerRepo, authRepo)
	req := RegisterRequest{NIK: "x", FullName: "x", LegalName: "x", BirthPlace: "p", BirthDate: "d", Salary: 1, Email: "e", KTPPhoto: "k", SelfiePhoto: "s", Password: "p"}

	err = u.Register(context.Background(), req)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
