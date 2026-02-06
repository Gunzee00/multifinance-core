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

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, AssetRepository, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}

	repo := NewAssetRepo(db)

	cleanup := func() {
		db.Close()
	}

	return db, mock, repo, cleanup
}

func TestAssetRepo_Create(t *testing.T) {
	db, mock, repo, cleanup := setupMockDB(t)
	defer cleanup()

	ctx := context.Background()
	tx, _ := db.Begin()

	asset := &entity.Asset{
		ProductName:  "Motor Honda",
		PriceProduct: 15000000,
		Seller:       "Dealer A",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO assets (product_name, price_product, seller, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?)`)).
		WithArgs(asset.ProductName, asset.PriceProduct, asset.Seller, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	tx, _ = db.Begin()
	id, err := repo.Create(ctx, tx, asset)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), id)

	tx.Commit()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAssetRepo_GetByID(t *testing.T) {
	_, mock, repo, cleanup := setupMockDB(t)
	defer cleanup()

	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "product_name", "price_product", "seller", "created_at", "updated_at",
	}).AddRow(1, "Motor Yamaha", 17000000, "Dealer B", time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, product_name, price_product, seller, created_at, updated_at
        FROM assets WHERE id = ?`)).
		WithArgs(1).
		WillReturnRows(rows)

	asset, err := repo.GetByID(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, uint64(1), asset.ID)
	assert.Equal(t, "Motor Yamaha", asset.ProductName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAssetRepo_List(t *testing.T) {
	_, mock, repo, cleanup := setupMockDB(t)
	defer cleanup()

	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "product_name", "price_product", "seller", "created_at", "updated_at",
	}).
		AddRow(1, "TV Samsung", 5000000, "Seller A", time.Now(), time.Now()).
		AddRow(2, "Kulkas LG", 4000000, "Seller B", time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, product_name, price_product, seller, created_at, updated_at
        FROM assets`)).
		WillReturnRows(rows)

	list, err := repo.List(ctx)

	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "TV Samsung", list[0].ProductName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAssetRepo_Update(t *testing.T) {
	db, mock, repo, cleanup := setupMockDB(t)
	defer cleanup()

	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE assets SET product_name = ?, price_product = ?, seller = ?, updated_at = ? WHERE id = ?`)).
		WithArgs("Updated Name", float64(20000000), "Dealer C", sqlmock.AnyArg(), uint64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tx, _ := db.Begin()

	asset := &entity.Asset{
		ID:           1,
		ProductName:  "Updated Name",
		PriceProduct: 20000000.0,
		Seller:       "Dealer C",
	}

	err := repo.Update(ctx, tx, asset)
	assert.NoError(t, err)

	tx.Commit()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAssetRepo_Delete(t *testing.T) {
	db, mock, repo, cleanup := setupMockDB(t)
	defer cleanup()

	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM assets WHERE id = ?`)).
		WithArgs(uint64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tx, _ := db.Begin()
	err := repo.Delete(ctx, tx, 1)
	assert.NoError(t, err)

	tx.Commit()
	assert.NoError(t, mock.ExpectationsWereMet())
}
