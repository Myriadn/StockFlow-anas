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

type CreateProductRequest struct {
	SKU   string `json:"sku" binding:"required"`
	Name  string `json:"name" binding:"required"`
	Price *int   `json:"price" binding:"required,gte=0"`
}

type UpdateProductRequest struct {
	SKU      *string `json:"sku"`
	Name     *string `json:"name"`
	Price    *int    `json:"price" binding:"omitempty,gte=0"`
	IsActive *bool   `json:"is_active"`
}

type AdjustStockRequest struct {
	Type string `json:"type" binding:"required,oneof=in out"`
	Qty  int    `json:"qty"  binding:"required,gt=0"`
	Note string `json:"note"`
}

type AdjustStockResponse struct {
	ProductID int    `json:"product_id"`
	Stock     int    `json:"stock"`
	Type      string `json:"type"`
	Qty       int    `json:"qty"`
}
