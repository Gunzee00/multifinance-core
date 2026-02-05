package usecase

import (
	"context"
	"database/sql"
	"errors"

	"multifinance-core/internal/repository"
)

type ConsumerLimitUsecase struct {
	db   *sql.DB
	repo repository.ConsumerLimitRepository
}

func NewConsumerLimitUsecase(db *sql.DB, r repository.ConsumerLimitRepository) *ConsumerLimitUsecase {
	return &ConsumerLimitUsecase{db: db, repo: r}
}

// IncreaseUsedLimit safely increases used_limit; fails if it would exceed max_limit
func (u *ConsumerLimitUsecase) IncreaseUsedLimit(ctx context.Context, consumerID uint64, tenor uint8, amount float64) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	cl, err := u.repo.GetByConsumerAndTenor(ctx, consumerID, tenor)
	if err != nil {
		return err
	}

	if cl.UsedLimit+amount > cl.MaxLimit {
		return errors.New("used limit exceeds max limit")
	}

	newUsed := cl.UsedLimit + amount
	if err := u.repo.UpdateUsedLimit(ctx, tx, consumerID, tenor, newUsed); err != nil {
		return err
	}

	return tx.Commit()
}
