package dto

type CreateOrderItem struct {
	SKU string `json:"sku" binding:"required"`
	Qty int    `json:"qty" binding:"required,gt=0"`
}

type CreateOrderRequest struct {
	Customer string            `json:"customer" binding:"required"`
	Items    []CreateOrderItem `json:"items" binding:"required,min=1,dive"`
}

type OrderItemResponse struct {
	SKU       string `json:"sku"`
	Qty       int    `json:"qty"`
	UnitPrice int    `json:"unit_price"`
}

type OrderResponse struct {
	ID          int                 `json:"id"`
	Customer    string              `json:"customer"`
	Status      string              `json:"status"`
	Total       int                 `json:"total"`
	CreatedAt   string              `json:"created_at"`
	CancelledAt *string             `json:"cancelled_at"`
	Items       []OrderItemResponse `json:"items"`
}
