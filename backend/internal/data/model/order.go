package model

import "time"

type Order struct {
	ID          int
	Customer    string
	Status      string
	Total       int
	CreatedAt   time.Time
	CancelledAt *time.Time
}

type OrderItem struct {
	ID        int
	OrderID   int
	SKU       string
	Qty       int
	UnitPrice int
}
