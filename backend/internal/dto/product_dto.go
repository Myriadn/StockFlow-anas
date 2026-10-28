package dto

type ListProductsQuery struct {
	Search string
	Page   int
	Limit  int
}

type ProductResponse struct {
	ID             int     `json:"id"`
	SKU            string  `json:"sku"`
	Name           string  `json:"name"`
	Price          int     `json:"price"`
	Stock          int     `json:"stock"`
	IsActive       bool    `json:"is_active"`
	LastMovementAt *string `json:"last_movement_at"`
}
type ListProductsResponse struct {
	Data  []ProductResponse `json:"data"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
	Total int               `json:"total"`
}
