package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"multifinance-core/internal/domain/entity"
	"multifinance-core/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

type mockAssetRepo struct {
	createFn func(ctx context.Context, tx *sql.Tx, a *entity.Asset) (uint64, error)
	getFn    func(ctx context.Context, id uint64) (*entity.Asset, error)
	listFn   func(ctx context.Context) ([]*entity.Asset, error)
	updateFn func(ctx context.Context, tx *sql.Tx, a *entity.Asset) error
	deleteFn func(ctx context.Context, tx *sql.Tx, id uint64) error
}

func (m *mockAssetRepo) Create(ctx context.Context, tx *sql.Tx, a *entity.Asset) (uint64, error) {
	if m.createFn != nil {
		return m.createFn(ctx, tx, a)
	}
	return 0, nil
}
func (m *mockAssetRepo) GetByID(ctx context.Context, id uint64) (*entity.Asset, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return nil, sql.ErrNoRows
}
func (m *mockAssetRepo) List(ctx context.Context) ([]*entity.Asset, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}
func (m *mockAssetRepo) Update(ctx context.Context, tx *sql.Tx, a *entity.Asset) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, tx, a)
	}
	return nil
}
func (m *mockAssetRepo) Delete(ctx context.Context, tx *sql.Tx, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, tx, id)
	}
	return nil
}

func newAssetUsecaseWithDBAndRepo(db *sql.DB, r repository.AssetRepository) *AssetUsecase {
	return NewAssetUsecase(db, r)
}

func TestCreate_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	repo := &mockAssetRepo{
		createFn: func(ctx context.Context, tx *sql.Tx, a *entity.Asset) (uint64, error) {
			require.NotNil(t, tx)
			require.Equal(t, "phone", a.ProductName)
			return 42, nil
		},
	}

	u := newAssetUsecaseWithDBAndRepo(db, repo)
	id, err := u.Create(context.Background(), CreateAssetRequest{ProductName: "phone", PriceProduct: 100.0, Seller: "shop"})
	require.NoError(t, err)
	require.Equal(t, uint64(42), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_RepoError_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	repo := &mockAssetRepo{
		createFn: func(ctx context.Context, tx *sql.Tx, a *entity.Asset) (uint64, error) {
			return 0, errors.New("db error")
		},
	}

	u := newAssetUsecaseWithDBAndRepo(db, repo)
	_, err = u.Create(context.Background(), CreateAssetRequest{ProductName: "x", PriceProduct: 1.0, Seller: "s"})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_List_Update_Delete(t *testing.T) {
	// GetByID
	repo := &mockAssetRepo{
		getFn: func(ctx context.Context, id uint64) (*entity.Asset, error) {
			return &entity.Asset{ID: id, ProductName: "tv", PriceProduct: 200.0, Seller: "store"}, nil
		},
		listFn: func(ctx context.Context) ([]*entity.Asset, error) {
			return []*entity.Asset{{ID: 1, ProductName: "a"}}, nil
		},
	}

	u := newAssetUsecaseWithDBAndRepo(nil, repo)
	a, err := u.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), a.ID)

	list, err := u.List(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 1)

	// Update success
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectCommit()

	repo.updateFn = func(ctx context.Context, tx *sql.Tx, a *entity.Asset) error {
		require.NotNil(t, tx)
		require.Equal(t, uint64(2), a.ID)
		return nil
	}

	u2 := newAssetUsecaseWithDBAndRepo(db, repo)
	err = u2.Update(context.Background(), 2, UpdateAssetRequest{ProductName: "b", PriceProduct: 10, Seller: "s"})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())

	// Delete success
	db2, mock2, err := sqlmock.New()
	require.NoError(t, err)
	defer db2.Close()
	mock2.ExpectBegin()
	mock2.ExpectCommit()

	repo.deleteFn = func(ctx context.Context, tx *sql.Tx, id uint64) error {
		require.Equal(t, uint64(3), id)
		return nil
	}
	u3 := newAssetUsecaseWithDBAndRepo(db2, repo)
	err = u3.Delete(context.Background(), 3)
	require.NoError(t, err)
	require.NoError(t, mock2.ExpectationsWereMet())
}
