package repository

import (
	"context"
	"database/sql"

	"multifinance-core/internal/domain/entity"
)

type ConsumerRepository interface {
	Create(ctx context.Context, tx *sql.Tx, c *entity.Consumer) (uint64, error)
}

type consumerRepo struct{}

func NewConsumerRepo() ConsumerRepository {
	return &consumerRepo{}
}
func (r *consumerRepo) Create(
	ctx context.Context,
	tx *sql.Tx,
	c *entity.Consumer,
) (uint64, error) {

	res, err := tx.ExecContext(ctx, `
		INSERT INTO consumers
		(nik, full_name, legal_name, birth_place, birth_date,ktp_photo, selfie_photo, salary)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		c.NIK,
		c.FullName,
		c.LegalName,
		c.BirthPlace,
		c.BirthDate,
		c.KTPPhoto,
		c.SelfiePhoto,
		c.Salary,
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
