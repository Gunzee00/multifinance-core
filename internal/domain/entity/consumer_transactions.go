package entity

import "time"

type Transaction struct {
	ID              uint64
	ContractNo      string
	ConsumerID      uint64
	ConsumerLimitID uint64
	AssetID         uint64
	TenorMonth      uint8
	OTR             int64
	AdminFee        int64
	JumlahBunga     int64
	JumlahCicilan   int64
	Status          string
	CreatedAt       time.Time
}
