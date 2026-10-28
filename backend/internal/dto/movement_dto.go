package dto

type MovementResponse struct {
	ID        int     `json:"id"`
	Type      string  `json:"type"`
	Qty       int     `json:"qty"`
	Note      *string `json:"note"`
	CreatedAt string  `json:"created_at"`
}
