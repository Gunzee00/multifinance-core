package repository

import (
	"context"
	"database/sql"
	"time"

	"multifinance-core/internal/domain/entity"
)

type AssetRepository interface {
	Create(ctx context.Context, tx *sql.Tx, a *entity.Asset) (uint64, error)
	GetByID(ctx context.Context, id uint64) (*entity.Asset, error)
	List(ctx context.Context) ([]*entity.Asset, error)
	Update(ctx context.Context, tx *sql.Tx, a *entity.Asset) error
	Delete(ctx context.Context, tx *sql.Tx, id uint64) error
}

type assetRepo struct {
	db *sql.DB
}

func NewAssetRepo(db *sql.DB) AssetRepository {
	return &assetRepo{db}
}

func (r *assetRepo) Create(ctx context.Context, tx *sql.Tx, a *entity.Asset) (uint64, error) {
	now := time.Now().UTC()
	res, err := tx.ExecContext(ctx, `
        INSERT INTO assets (product_name, price_product, seller, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?)`,
		a.ProductName, a.PriceProduct, a.Seller, now, now,
	)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(lastID), nil
}

func (r *assetRepo) GetByID(ctx context.Context, id uint64) (*entity.Asset, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, product_name, price_product, seller, created_at, updated_at
        FROM assets WHERE id = ?`, id)

	var a entity.Asset
	err := row.Scan(&a.ID, &a.ProductName, &a.PriceProduct, &a.Seller, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *assetRepo) List(ctx context.Context) ([]*entity.Asset, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, product_name, price_product, seller, created_at, updated_at
        FROM assets`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*entity.Asset
	for rows.Next() {
		var a entity.Asset
		if err := rows.Scan(&a.ID, &a.ProductName, &a.PriceProduct, &a.Seller, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, &a)
	}
	return res, nil
}

func (r *assetRepo) Update(ctx context.Context, tx *sql.Tx, a *entity.Asset) error {
	now := time.Now().UTC()
	_, err := tx.ExecContext(ctx, `
        UPDATE assets SET product_name = ?, price_product = ?, seller = ?, updated_at = ? WHERE id = ?`,
		a.ProductName, a.PriceProduct, a.Seller, now, a.ID,
	)
	return err
}

func (r *assetRepo) Delete(ctx context.Context, tx *sql.Tx, id uint64) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM assets WHERE id = ?`, id)
	return err
}
