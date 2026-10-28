package usecase

import (
	"context"
	"stackflow_anas/internal/data"
	"stackflow_anas/internal/data/repository"
	"stackflow_anas/internal/dto"
	"time"
)

type ProductUsecase struct {
	repo *repository.ProductRepo
	pool data.Querier
}

func NewProductUsecase(repo *repository.ProductRepo, pool data.Querier) *ProductUsecase {
	return &ProductUsecase{repo: repo, pool: pool}
}

func (u *ProductUsecase) List(ctx context.Context, q dto.ListProductsQuery) (*dto.ListProductsResponse, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 100 {
		q.Limit = 100
	}
	offset := (q.Page - 1) * q.Limit

	products, err := u.repo.List(ctx, u.pool, q.Search, q.Limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := u.repo.Count(ctx, u.pool, q.Search)
	if err != nil {
		return nil, err
	}

	data := make([]dto.ProductResponse, 0, len(products))
	for _, p := range products {
		var lma *string
		if p.LastMovementAt != nil {
			s := p.LastMovementAt.UTC().Format(time.RFC3339)
			lma = &s
		}
		data = append(data, dto.ProductResponse{
			ID: p.ID, SKU: p.SKU, Name: p.Name, Price: p.Price,
			Stock: p.Stock, IsActive: p.IsActive, LastMovementAt: lma,
		})
	}

	return &dto.ListProductsResponse{Data: data, Page: q.Page, Limit: q.Limit, Total: total}, nil
}
