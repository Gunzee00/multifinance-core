package usecase

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "math"
    "time"

    "multifinance-core/internal/domain/entity"
    "multifinance-core/internal/repository"
)

var ErrInsufficientLimit = errors.New("insufficient limit")
var ErrInvalidTenor = errors.New("invalid tenor")

type ConsumerTransactionUsecase struct {
    db       *sql.DB
    assetRepo repository.AssetRepository
    limitRepo repository.ConsumerLimitRepository
    txRepo    repository.ConsumerTransactionRepository
}

func NewConsumerTransactionUsecase(db *sql.DB, a repository.AssetRepository, l repository.ConsumerLimitRepository, t repository.ConsumerTransactionRepository) *ConsumerTransactionUsecase {
    return &ConsumerTransactionUsecase{db, a, l, t}
}

func allowedTenor(t uint8) bool {
    return t == 1 || t == 2 || t == 3 || t == 6
}

// Purchase executes a purchase: checks limit, writes transaction and updates used_limit atomically.
func (u *ConsumerTransactionUsecase) Purchase(ctx context.Context, consumerID uint64, assetID uint64, tenor uint8) (*entity.Transaction, error) {
    if !allowedTenor(tenor) {
        return nil, ErrInvalidTenor
    }

    tx, err := u.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    // get asset
    asset, err := u.assetRepo.GetByID(ctx, assetID)
    if err != nil {
        return nil, err
    }

    // lock consumer_limit row FOR UPDATE
    row := tx.QueryRowContext(ctx, `SELECT id, consumer_id, tenor_month, max_limit, used_limit FROM consumer_limits WHERE consumer_id = ? AND tenor_month = ? FOR UPDATE`, consumerID, tenor)
    var clID uint64
    var cID uint64
    var tMonth uint8
    var maxLimit float64
    var usedLimit float64
    if err := row.Scan(&clID, &cID, &tMonth, &maxLimit, &usedLimit); err != nil {
        return nil, err
    }

    available := maxLimit - usedLimit
    price := asset.PriceProduct
    if price > available {
        // record failed transaction for audit
        tr := &entity.Transaction{
            ContractNo: fmt.Sprintf("C-%d-%d", consumerID, time.Now().UTC().UnixNano()),
            ConsumerID: consumerID,
            ConsumerLimitID: clID,
            AssetID: assetID,
            TenorMonth: tenor,
            OTR: int64(math.Round(price)),
            AdminFee: int64(math.Round(price * 0.05)),
            JumlahBunga: int64(math.Round(price * 0.02 * float64(tenor))),
            JumlahCicilan: 0,
            Status: "FAILED",
            CreatedAt: time.Now().UTC(),
        }
        if _, err := u.txRepo.Create(ctx, tx, tr); err != nil {
            return nil, err
        }
        _ = tx.Commit()
        return nil, ErrInsufficientLimit
    }

    // compute amounts
    otr := int64(math.Round(price))
    admin := int64(math.Round(price * 0.05))
    bunga := int64(math.Round(price * 0.02 * float64(tenor)))
    total := float64(otr + admin + bunga)
    cicilan := int64(math.Round(total / float64(tenor)))

    // create transaction record
    tr := &entity.Transaction{
        ContractNo: fmt.Sprintf("C-%d-%d", consumerID, time.Now().UTC().UnixNano()),
        ConsumerID: consumerID,
        ConsumerLimitID: clID,
        AssetID: assetID,
        TenorMonth: tenor,
        OTR: otr,
        AdminFee: admin,
        JumlahBunga: bunga,
        JumlahCicilan: cicilan,
        Status: "SUCCESS",
        CreatedAt: time.Now().UTC(),
    }

    if _, err := u.txRepo.Create(ctx, tx, tr); err != nil {
        return nil, err
    }

    // update used_limit (increase by OTR)
    newUsed := usedLimit + float64(otr)
    if _, err := tx.ExecContext(ctx, `UPDATE consumer_limits SET used_limit = ?, updated_at = ? WHERE id = ?`, newUsed, time.Now().UTC(), clID); err != nil {
        return nil, err
    }

    if err := tx.Commit(); err != nil {
        return nil, err
    }
    return tr, nil
}

func (u *ConsumerTransactionUsecase) ListByConsumer(ctx context.Context, consumerID uint64) ([]*entity.Transaction, error) {
    return u.txRepo.ListByConsumer(ctx, consumerID)
}