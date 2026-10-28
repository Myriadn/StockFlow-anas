package repository

import (
	"context"
	"stackflow_anas/internal/data"
	"stackflow_anas/internal/data/model"
)

type ProductRepo struct {
	pool data.Querier
}

func NewProductRepo(pool data.Querier) *ProductRepo {
	return &ProductRepo{pool: pool}
}

func (r *ProductRepo) Ping(ctx context.Context) error {
	return r.pool.QueryRow(ctx, "SELECT 1").Scan(new(int))
}

func (r *ProductRepo) List(ctx context.Context, q data.Querier, search string, limit, offset int) ([]model.Product, error) {
	const sql = `
		SELECT p.id, p.sku, p.name, p.price, p.stock, p.is_active, p.created_at,
		       (SELECT MAX(created_at) FROM stock_movements WHERE product_id = p.id) AS last_movement_at
		FROM products p
		WHERE ($1 = '' OR p.sku ILIKE '%'||$1||'%' OR p.name ILIKE '%'||$1||'%')
		ORDER BY p.id
		LIMIT $2 OFFSET $3`

	rows, err := q.Query(ctx, sql, search, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Price, &p.Stock, &p.IsActive, &p.CreatedAt, &p.LastMovementAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *ProductRepo) Count(ctx context.Context, q data.Querier, search string) (int, error) {
	const sql = `
		SELECT COUNT(*) FROM products
		WHERE ($1 = '' OR sku ILIKE '%'||$1||'%' OR name ILIKE '%'||$1||'%')`
	var n int
	err := q.QueryRow(ctx, sql, search).Scan(&n)
	return n, err
}
