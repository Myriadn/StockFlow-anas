package repository

import (
	"context"
	"stackflow_anas/internal/data"
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
