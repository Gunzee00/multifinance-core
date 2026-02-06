package usecase

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"multifinance-core/internal/domain/entity"
	"multifinance-core/internal/repository"
	"multifinance-core/internal/utils"
)

type RegisterRequest struct {
	NIK         string  `json:"nik" binding:"required"`
	FullName    string  `json:"full_name" binding:"required"`
	LegalName   string  `json:"legal_name" binding:"required"`
	BirthPlace  string  `json:"birth_place" binding:"required"`
	BirthDate   string  `json:"birth_date" binding:"required"`
	Salary      float64 `json:"salary" binding:"required"`
	Email       string  `json:"email" binding:"required,email"`
	KTPPhoto    string  `json:"ktp_photo" binding:"required"`
	SelfiePhoto string  `json:"selfie_photo" binding:"required"`
	Password    string  `json:"password" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthUsecase struct {
	db           *sql.DB
	consumerRepo repository.ConsumerRepository
	authRepo     repository.AuthRepository
}

func NewAuthUsecase(db *sql.DB, c repository.ConsumerRepository, a repository.AuthRepository) *AuthUsecase {
	return &AuthUsecase{db, c, a}
}

func (u *AuthUsecase) Register(ctx context.Context, req RegisterRequest) error {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	consumer := &entity.Consumer{
		NIK:         req.NIK,
		FullName:    req.FullName,
		LegalName:   req.LegalName,
		BirthPlace:  req.BirthPlace,
		BirthDate:   req.BirthDate,
		KTPPhoto:    req.KTPPhoto,
		SelfiePhoto: req.SelfiePhoto,
		Salary:      req.Salary,
	}

	consumerID, err := u.consumerRepo.Create(ctx, tx, consumer)
	if err != nil {
		return err
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}

	err = u.authRepo.Create(ctx, tx, &entity.AuthUser{
		ConsumerID: consumerID,
		Email:      req.Email,
		Password:   hash,
	})
	if err != nil {
		return err
	}

	tenors := []uint8{1, 2, 3, 6}
	now := time.Now().UTC()
	for _, t := range tenors {
		maxLimit := req.Salary * 0.4 * float64(t)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO consumer_limits (consumer_id, tenor_month, max_limit, used_limit, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			consumerID, t, maxLimit, 0.0, now, now,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (u *AuthUsecase) Login(ctx context.Context, req LoginRequest) (string, error) {
	user, err := u.authRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return "", err
	}

	if err := utils.ComparePassword(user.Password, req.Password); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := req.Email + ":" + time.Now().UTC().Format(time.RFC3339)
	return token, nil
}
