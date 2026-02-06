package usecase

import (
	"context"
	"database/sql"
	"multifinance-core/internal/domain/entity"
	"multifinance-core/internal/repository"
)

type CreateAssetRequest struct {
	ProductName  string  `json:"product_name" binding:"required"`
	PriceProduct float64 `json:"price_product" binding:"required"`
	Seller       string  `json:"seller" binding:"required"`
}

type UpdateAssetRequest struct {
	ProductName  string  `json:"product_name" binding:"required"`
	PriceProduct float64 `json:"price_product" binding:"required"`
	Seller       string  `json:"seller" binding:"required"`
}

type AssetUsecase struct {
	db   *sql.DB
	repo repository.AssetRepository
}

func NewAssetUsecase(db *sql.DB, r repository.AssetRepository) *AssetUsecase {
	return &AssetUsecase{db: db, repo: r}
}

func (u *AssetUsecase) Create(ctx context.Context, req CreateAssetRequest) (uint64, error) {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	a := &entity.Asset{
		ProductName:  req.ProductName,
		PriceProduct: req.PriceProduct,
		Seller:       req.Seller,
	}

	id, err := u.repo.Create(ctx, tx, a)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (u *AssetUsecase) GetByID(ctx context.Context, id uint64) (*entity.Asset, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *AssetUsecase) List(ctx context.Context) ([]*entity.Asset, error) {
	return u.repo.List(ctx)
}

func (u *AssetUsecase) Update(ctx context.Context, id uint64, req UpdateAssetRequest) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	a := &entity.Asset{
		ID:           id,
		ProductName:  req.ProductName,
		PriceProduct: req.PriceProduct,
		Seller:       req.Seller,
	}

	if err := u.repo.Update(ctx, tx, a); err != nil {
		return err
	}
	return tx.Commit()
}

func (u *AssetUsecase) Delete(ctx context.Context, id uint64) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := u.repo.Delete(ctx, tx, id); err != nil {
		return err
	}
	return tx.Commit()
}
