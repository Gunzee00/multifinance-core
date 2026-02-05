package repository

import (
	"context"
	"database/sql"

	"multifinance-core/internal/domain/entity"
)

type AuthRepository interface {
	Create(ctx context.Context, tx *sql.Tx, u *entity.AuthUser) error
	FindByEmail(ctx context.Context, email string) (*entity.AuthUser, error)
}

type authRepo struct {
	db *sql.DB
}

func NewAuthRepo(db *sql.DB) AuthRepository {
	return &authRepo{db}
}

func (r *authRepo) Create(ctx context.Context, tx *sql.Tx, u *entity.AuthUser) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO auth_users (consumer_id, email, password)
		VALUES (?, ?, ?)`,
		u.ConsumerID, u.Email, u.Password,
	)
	return err
}

func (r *authRepo) FindByEmail(ctx context.Context, email string) (*entity.AuthUser, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, consumer_id, email, password
		FROM auth_users WHERE email = ?`, email)

	var u entity.AuthUser
	err := row.Scan(&u.ID, &u.ConsumerID, &u.Email, &u.Password)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
