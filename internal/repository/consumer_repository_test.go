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

func setupConsumerMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, ConsumerRepository, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock db: %v", err)
	}

	repo := NewConsumerRepo()

	cleanup := func() {
		db.Close()
	}

	return db, mock, repo, cleanup
}

func TestConsumerRepo_Create_Success(t *testing.T) {
	db, mock, repo, cleanup := setupConsumerMockDB(t)
	defer cleanup()

	ctx := context.Background()

	now := time.Now()

	consumer := &entity.Consumer{
		NIK:         "3173010101010001",
		FullName:    "Budi Santoso",
		LegalName:   "BUDI SANTOSO",
		BirthPlace:  "Jakarta",
		BirthDate:   now.Format(time.RFC3339),
		KTPPhoto:    "ktp.jpg",
		SelfiePhoto: "selfie.jpg",
		Salary:      5000000,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO consumers
		(nik, full_name, legal_name, birth_place, birth_date,ktp_photo, selfie_photo, salary)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs(
			consumer.NIK,
			consumer.FullName,
			consumer.LegalName,
			consumer.BirthPlace,
			consumer.BirthDate,
			consumer.KTPPhoto,
			consumer.SelfiePhoto,
			consumer.Salary,
		).
		WillReturnResult(sqlmock.NewResult(10, 1)) // ID = 10
	mock.ExpectCommit()

	tx, _ := db.Begin()

	id, err := repo.Create(ctx, tx, consumer)

	assert.NoError(t, err)
	assert.Equal(t, uint64(10), id)

	tx.Commit()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConsumerRepo_Create_ExecError(t *testing.T) {
	db, mock, repo, cleanup := setupConsumerMockDB(t)
	defer cleanup()

	ctx := context.Background()

	consumer := &entity.Consumer{
		NIK:         "duplicate-nik",
		FullName:    "Test",
		LegalName:   "TEST",
		BirthPlace:  "Bandung",
		BirthDate:   time.Now().Format(time.RFC3339),
		KTPPhoto:    "ktp.jpg",
		SelfiePhoto: "selfie.jpg",
		Salary:      4000000,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO consumers
		(nik, full_name, legal_name, birth_place, birth_date,ktp_photo, selfie_photo, salary)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs(
			consumer.NIK,
			consumer.FullName,
			consumer.LegalName,
			consumer.BirthPlace,
			consumer.BirthDate,
			consumer.KTPPhoto,
			consumer.SelfiePhoto,
			consumer.Salary,
		).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	tx, _ := db.Begin()

	id, err := repo.Create(ctx, tx, consumer)

	assert.Error(t, err)
	assert.Equal(t, uint64(0), id)

	tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConsumerRepo_Create_LastInsertIdError(t *testing.T) {
	db, mock, repo, cleanup := setupConsumerMockDB(t)
	defer cleanup()

	ctx := context.Background()

	consumer := &entity.Consumer{
		NIK:         "3173010101010002",
		FullName:    "Andi",
		LegalName:   "ANDI",
		BirthPlace:  "Surabaya",
		BirthDate:   time.Now().Format(time.RFC3339),
		KTPPhoto:    "ktp2.jpg",
		SelfiePhoto: "selfie2.jpg",
		Salary:      6000000,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO consumers
		(nik, full_name, legal_name, birth_place, birth_date,ktp_photo, selfie_photo, salary)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs(
			consumer.NIK,
			consumer.FullName,
			consumer.LegalName,
			consumer.BirthPlace,
			consumer.BirthDate,
			consumer.KTPPhoto,
			consumer.SelfiePhoto,
			consumer.Salary,
		).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	mock.ExpectRollback()

	tx, _ := db.Begin()

	id, err := repo.Create(ctx, tx, consumer)

	assert.Error(t, err)
	assert.Equal(t, uint64(0), id)

	tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}
