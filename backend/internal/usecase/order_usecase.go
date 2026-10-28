package usecase

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"stackflow_anas/internal/data/model"
	"stackflow_anas/internal/data/repository"
	"stackflow_anas/internal/dto"
)

type OrderUsecase struct {
	orderRepo   *repository.OrderRepo
	productRepo *repository.ProductRepo
	pool        *pgxpool.Pool
}

func NewOrderUsecase(o *repository.OrderRepo, p *repository.ProductRepo, pool *pgxpool.Pool) *OrderUsecase {
	return &OrderUsecase{orderRepo: o, productRepo: p, pool: pool}
}

func (u *OrderUsecase) Create(ctx context.Context, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	seen := make(map[string]bool)
	var dup []string
	for _, it := range req.Items {
		if seen[it.SKU] {
			dup = append(dup, it.SKU)
		}
		seen[it.SKU] = true
	}
	if len(dup) > 0 {
		return nil, NewAppError(400, "DUPLICATE_SKU", map[string]any{"skus": dup})
	}

	skus := make([]string, 0, len(req.Items))
	for _, it := range req.Items {
		skus = append(skus, it.SKU)
	}
	known, err := u.productRepo.GetBySKUs(ctx, u.pool, skus)
	if err != nil {
		return nil, err
	}
	var unknown []string
	for _, s := range skus {
		if _, ok := known[s]; !ok {
			unknown = append(unknown, s)
		}
	}
	if len(unknown) > 0 {
		return nil, NewAppError(400, "UNKNOWN_SKU", map[string]any{"skus": unknown})
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	total := 0
	prices := make(map[string]int, len(req.Items))
	for _, it := range req.Items {
		_, price, err := u.productRepo.DecrementStockBySKU(ctx, tx, it.SKU, it.Qty)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				available, err2 := u.productRepo.GetStockBySKU(ctx, tx, it.SKU)
				if errors.Is(err2, pgx.ErrNoRows) {
					return nil, NewAppError(400, "UNKNOWN_SKU", map[string]any{"skus": []string{it.SKU}})
				}
				if err2 != nil {
					return nil, err2
				}
				return nil, NewAppError(409, "INSUFFICIENT_STOCK", map[string]any{
					"sku": it.SKU, "available": available,
				})
			}
			return nil, err
		}
		prices[it.SKU] = price
		total += price * it.Qty
	}

	order, err := u.orderRepo.Create(ctx, tx, req.Customer, total)
	if err != nil {
		return nil, err
	}

	for _, it := range req.Items {
		price := prices[it.SKU]
		if err := u.orderRepo.InsertItem(ctx, tx, order.ID, it.SKU, it.Qty, price); err != nil {
			return nil, err
		}
		note := "Order #" + strconv.Itoa(order.ID)
		if err := u.productRepo.InsertMovementBySKU(ctx, tx, it.SKU, "out", it.Qty, note); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	resp := &dto.OrderResponse{
		ID: order.ID, Customer: order.Customer, Status: order.Status,
		Total: order.Total, CreatedAt: order.CreatedAt.UTC().Format(time.RFC3339),
	}
	for _, it := range req.Items {
		resp.Items = append(resp.Items, dto.OrderItemResponse{
			SKU: it.SKU, Qty: it.Qty, UnitPrice: prices[it.SKU],
		})
	}
	return resp, nil
}

func (u *OrderUsecase) Get(ctx context.Context, id int) (*dto.OrderResponse, error) {
	o, err := u.orderRepo.GetByID(ctx, u.pool, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, NewAppError(404, "ORDER_NOT_FOUND", nil)
		}
		return nil, err
	}
	items, err := u.orderRepo.GetItems(ctx, u.pool, id)
	if err != nil {
		return nil, err
	}
	return toOrderResponse(*o, items), nil
}

func (u *OrderUsecase) Cancel(ctx context.Context, id int) (*dto.OrderResponse, error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rows, err := u.orderRepo.Cancel(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		if _, err := u.orderRepo.GetByID(ctx, tx, id); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, NewAppError(404, "ORDER_NOT_FOUND", nil)
			}
			return nil, err
		}
		return nil, NewAppError(409, "ORDER_ALREADY_CANCELLED", nil)
	}

	items, err := u.orderRepo.GetItems(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	for _, it := range items {
		if err := u.productRepo.IncrementStockBySKU(ctx, tx, it.SKU, it.Qty); err != nil {
			return nil, err
		}
		note := "Cancel order #" + strconv.Itoa(id)
		if err := u.productRepo.InsertMovementBySKU(ctx, tx, it.SKU, "in", it.Qty, note); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return u.Get(ctx, id)
}

func toOrderResponse(o model.Order, items []model.OrderItem) *dto.OrderResponse {
	resp := &dto.OrderResponse{
		ID: o.ID, Customer: o.Customer, Status: o.Status, Total: o.Total,
		CreatedAt: o.CreatedAt.UTC().Format(time.RFC3339),
	}
	if o.CancelledAt != nil {
		s := o.CancelledAt.UTC().Format(time.RFC3339)
		resp.CancelledAt = &s
	}
	for _, it := range items {
		resp.Items = append(resp.Items, dto.OrderItemResponse{
			SKU: it.SKU, Qty: it.Qty, UnitPrice: it.UnitPrice,
		})
	}
	return resp
}
