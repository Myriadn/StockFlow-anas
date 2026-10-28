package repository

import (
	"context"

	"stackflow_anas/internal/data"
	"stackflow_anas/internal/data/model"
)

type OrderRepo struct {
	pool data.Querier
}

func NewOrderRepo(pool data.Querier) *OrderRepo { return &OrderRepo{pool: pool} }

func (r *OrderRepo) Create(ctx context.Context, q data.Querier, customer string, total int) (*model.Order, error) {
	const sql = `INSERT INTO sales_orders (customer, total) VALUES ($1,$2)
	             RETURNING id, customer, status, total, created_at, cancelled_at`
	var o model.Order
	err := q.QueryRow(ctx, sql, customer, total).
		Scan(&o.ID, &o.Customer, &o.Status, &o.Total, &o.CreatedAt, &o.CancelledAt)
	return &o, err
}

func (r *OrderRepo) InsertItem(ctx context.Context, q data.Querier, orderID int, sku string, qty, unitPrice int) error {
	_, err := q.Exec(ctx,
		`INSERT INTO sales_order_items (order_id, sku, qty, unit_price) VALUES ($1,$2,$3,$4)`,
		orderID, sku, qty, unitPrice)
	return err
}

func (r *OrderRepo) GetByID(ctx context.Context, q data.Querier, id int) (*model.Order, error) {
	const sql = `SELECT id, customer, status, total, created_at, cancelled_at FROM sales_orders WHERE id=$1`
	var o model.Order
	err := q.QueryRow(ctx, sql, id).
		Scan(&o.ID, &o.Customer, &o.Status, &o.Total, &o.CreatedAt, &o.CancelledAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepo) GetItems(ctx context.Context, q data.Querier, orderID int) ([]model.OrderItem, error) {
	const sql = `SELECT id, order_id, sku, qty, unit_price FROM sales_order_items WHERE order_id=$1 ORDER BY id`
	rows, err := q.Query(ctx, sql, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []model.OrderItem
	for rows.Next() {
		var it model.OrderItem
		if err := rows.Scan(&it.ID, &it.OrderID, &it.SKU, &it.Qty, &it.UnitPrice); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *OrderRepo) Cancel(ctx context.Context, q data.Querier, id int) (int64, error) {
	const sql = `UPDATE sales_orders SET status='CANCELLED', cancelled_at=now()
	             WHERE id=$1 AND status='CREATED'`
	tag, err := q.Exec(ctx, sql, id)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
