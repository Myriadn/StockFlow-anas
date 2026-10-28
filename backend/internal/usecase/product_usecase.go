package usecase

import (
	"context"
	"errors"
	"stackflow_anas/internal/data/model"
	"stackflow_anas/internal/data/repository"
	"stackflow_anas/internal/dto"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductUsecase struct {
	repo *repository.ProductRepo
	pool *pgxpool.Pool
}

func NewProductUsecase(repo *repository.ProductRepo, pool *pgxpool.Pool) *ProductUsecase {
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
		data = append(data, toProductResponse(p))
	}

	return &dto.ListProductsResponse{
		Data:  data,
		Page:  q.Page,
		Limit: q.Limit,
		Total: total,
	}, nil
}

func (u *ProductUsecase) Create(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	p, err := u.repo.Create(ctx, u.pool, req.SKU, req.Name, *req.Price)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, NewAppError(409, "SKU_ALREADY_EXISTS", map[string]any{"sku": req.SKU})
		}
		return nil, err
	}
	resp := toProductResponse(*p)
	return &resp, nil
}

func (u *ProductUsecase) Update(ctx context.Context, id int, req dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	p, err := u.repo.Update(ctx, u.pool, id, req.SKU, req.Name, req.Price, req.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, NewAppError(404, "PRODUCT_NOT_FOUND", nil)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, NewAppError(409, "SKU_ALREADY_EXISTS", map[string]any{"sku": *req.SKU})
		}
		return nil, err
	}
	resp := toProductResponse(*p)
	return &resp, nil
}

func (u *ProductUsecase) Delete(ctx context.Context, id int) error {
	rows, err := u.repo.Delete(ctx, u.pool, id)
	if err != nil {
		return err
	}
	if rows > 0 {
		return nil
	}
	_, err = u.repo.GetByID(ctx, u.pool, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return NewAppError(404, "PRODUCT_NOT_FOUND", nil)
	}
	if err != nil {
		return err
	}
	return NewAppError(409, "PRODUCT_HAS_HISTORY", nil)
}

func toProductResponse(p model.Product) dto.ProductResponse {
	var lma *string
	if p.LastMovementAt != nil {
		s := p.LastMovementAt.UTC().Format(time.RFC3339)
		lma = &s
	}
	return dto.ProductResponse{
		ID: p.ID, SKU: p.SKU, Name: p.Name, Price: p.Price,
		Stock: p.Stock, IsActive: p.IsActive, LastMovementAt: lma,
	}
}

func (u *ProductUsecase) AdjustStock(ctx context.Context, id int, req dto.AdjustStockRequest) (*dto.AdjustStockResponse, error) {
	delta := req.Qty
	if req.Type == "out" {
		delta = -req.Qty
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	newStock, err := u.repo.AdjustStock(ctx, tx, id, delta)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			available, err2 := u.repo.GetStock(ctx, tx, id)
			if errors.Is(err2, pgx.ErrNoRows) {
				return nil, NewAppError(404, "PRODUCT_NOT_FOUND", nil)
			}
			if err2 != nil {
				return nil, err2
			}
			return nil, NewAppError(409, "INSUFFICIENT_STOCK", map[string]any{"available": available})
		}
		return nil, err
	}

	if err := u.repo.InsertMovement(ctx, tx, id, req.Type, req.Qty, req.Note); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &dto.AdjustStockResponse{ProductID: id, Stock: newStock, Type: req.Type, Qty: req.Qty}, nil
}

func (u *ProductUsecase) ListMovements(ctx context.Context, productID int) ([]dto.MovementResponse, error) {
	if _, err := u.repo.GetByID(ctx, u.pool, productID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, NewAppError(404, "PRODUCT_NOT_FOUND", nil)
		}
		return nil, err
	}
	ms, err := u.repo.ListMovements(ctx, u.pool, productID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.MovementResponse, 0, len(ms))
	for _, m := range ms {
		out = append(out, dto.MovementResponse{
			ID: m.ID, Type: m.Type, Qty: m.Qty, Note: m.Note,
			CreatedAt: m.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return out, nil
}
