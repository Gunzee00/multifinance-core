package repository

import (
	"context"
	"database/sql"
	"time"

	"multifinance-core/internal/domain/entity"
)

type ConsumerLimitRepository interface {
	GetByConsumerAndTenor(ctx context.Context, consumerID uint64, tenor uint8) (*entity.ConsumerLimit, error)
	UpdateUsedLimit(ctx context.Context, tx *sql.Tx, consumerID uint64, tenor uint8, newUsed float64) error
}

type consumerLimitRepo struct {
	db *sql.DB
}

func NewConsumerLimitRepo(db *sql.DB) ConsumerLimitRepository {
	return &consumerLimitRepo{db}
}

func (r *consumerLimitRepo) GetByConsumerAndTenor(ctx context.Context, consumerID uint64, tenor uint8) (*entity.ConsumerLimit, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, consumer_id, tenor_month, max_limit, used_limit, created_at, updated_at
        FROM consumer_limits WHERE consumer_id = ? AND tenor_month = ?`, consumerID, tenor)

	var cl entity.ConsumerLimit
	if err := row.Scan(&cl.ID, &cl.ConsumerID, &cl.TenorMonth, &cl.MaxLimit, &cl.UsedLimit, &cl.CreatedAt, &cl.UpdatedAt); err != nil {
		return nil, err
	}
	return &cl, nil
}

func (r *consumerLimitRepo) UpdateUsedLimit(ctx context.Context, tx *sql.Tx, consumerID uint64, tenor uint8, newUsed float64) error {
	_, err := tx.ExecContext(ctx, `UPDATE consumer_limits SET used_limit = ?, updated_at = ? WHERE consumer_id = ? AND tenor_month = ?`, newUsed, time.Now().UTC(), consumerID, tenor)
	return err
}
