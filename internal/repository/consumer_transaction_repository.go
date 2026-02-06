package repository

import (
	"context"
	"database/sql"
	"time"

	"multifinance-core/internal/domain/entity"
)

type ConsumerTransactionRepository interface {
	Create(ctx context.Context, tx *sql.Tx, t *entity.Transaction) (uint64, error)
	ListByConsumer(ctx context.Context, consumerID uint64) ([]*entity.Transaction, error)
}

type consumerTransactionRepo struct {
	db *sql.DB
}

func NewConsumerTransactionRepo(db *sql.DB) ConsumerTransactionRepository {
	return &consumerTransactionRepo{db}
}

func (r *consumerTransactionRepo) Create(ctx context.Context, tx *sql.Tx, t *entity.Transaction) (uint64, error) {
	now := time.Now().UTC()
	res, err := tx.ExecContext(ctx, `
        INSERT INTO consumer_transactions (contract_no, consumer_id, consumer_limit_id, asset_id, tenor_month, otr, admin_fee, jumlah_bunga, jumlah_cicilan, status, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ContractNo, t.ConsumerID, t.ConsumerLimitID, t.AssetID, t.TenorMonth, t.OTR, t.AdminFee, t.JumlahBunga, t.JumlahCicilan, t.Status, now,
	)
	if err != nil {
		return 0, err
	}
	last, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(last), nil
}

func (r *consumerTransactionRepo) ListByConsumer(ctx context.Context, consumerID uint64) ([]*entity.Transaction, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, contract_no, consumer_id, consumer_limit_id, asset_id, tenor_month, otr, admin_fee, jumlah_bunga, jumlah_cicilan, status, created_at
        FROM consumer_transactions WHERE consumer_id = ? ORDER BY created_at DESC`, consumerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*entity.Transaction
	for rows.Next() {
		var t entity.Transaction
		if err := rows.Scan(&t.ID, &t.ContractNo, &t.ConsumerID, &t.ConsumerLimitID, &t.AssetID, &t.TenorMonth, &t.OTR, &t.AdminFee, &t.JumlahBunga, &t.JumlahCicilan, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, &t)
	}
	return res, nil
}
