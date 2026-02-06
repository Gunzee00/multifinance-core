package entity

import "time"

type Asset struct {
	ID           uint64
	ProductName  string
	PriceProduct float64
	Seller       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
