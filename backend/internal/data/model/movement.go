package model

import "time"

type StockMovement struct {
	ID        int
	ProductID int
	Type      string
	Qty       int
	Note      *string
	CreatedAt time.Time
}
