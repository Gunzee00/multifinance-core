package entity

import "time"

type ConsumerLimit struct {
	ID         uint64
	ConsumerID uint64
	TenorMonth uint8
	MaxLimit   float64
	UsedLimit  float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
