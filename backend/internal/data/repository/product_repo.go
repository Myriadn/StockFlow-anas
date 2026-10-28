package repository

import (
	"context"
	"stackflow_anas/internal/data"
	"stackflow_anas/internal/data/model"
)

type ProductRepo struct {
	pool data.Querier
}

type ProductForOrder struct {
	SKU   string
	Price int
	Stock int
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

func (r *ProductRepo) GetByID(ctx context.Context, q data.Querier, id int) (*model.Product, error) {
	const sql = `SELECT id, sku, name, price, stock, is_active, created_at FROM products WHERE id=$1`
	var p model.Product
	err := q.QueryRow(ctx, sql, id).Scan(&p.ID, &p.SKU, &p.Name, &p.Price, &p.Stock, &p.IsActive, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) Create(ctx context.Context, q data.Querier, sku, name string, price int) (*model.Product, error) {
	const sql = `INSERT INTO products (sku, name, price) VALUES ($1,$2,$3)
	             RETURNING id, sku, name, price, stock, is_active, created_at`
	var p model.Product
	err := q.QueryRow(ctx, sql, sku, name, price).Scan(&p.ID, &p.SKU, &p.Name, &p.Price, &p.Stock, &p.IsActive, &p.CreatedAt)
	return &p, err
}

func (r *ProductRepo) Update(ctx context.Context, q data.Querier, id int, sku *string, name *string, price *int, isActive *bool) (*model.Product, error) {
	const sql = `UPDATE products SET
					sku = COALESCE($2, sku),
	                name = COALESCE($3, name),
	                price = COALESCE($4, price),
	                is_active = COALESCE($5, is_active)
	              WHERE id=$1
	              RETURNING id, sku, name, price, stock, is_active, created_at`
	var p model.Product
	err := q.QueryRow(ctx, sql, id, sku, name, price, isActive).Scan(&p.ID, &p.SKU, &p.Name, &p.Price, &p.Stock, &p.IsActive, &p.CreatedAt)
	return &p, err
}

func (r *ProductRepo) Delete(ctx context.Context, q data.Querier, id int) (int64, error) {
	const sql = `DELETE FROM products WHERE id=$1
	             AND NOT EXISTS (SELECT 1 FROM stock_movements WHERE product_id=$1)`
	tag, err := q.Exec(ctx, sql, id)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *ProductRepo) HasMovements(ctx context.Context, q data.Querier, id int) (bool, error) {
	var exists bool
	err := q.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM stock_movements WHERE product_id=$1)`, id).Scan(&exists)
	return exists, err
}

// AdjustStock: guard atomic on WHERE
func (r *ProductRepo) AdjustStock(ctx context.Context, q data.Querier, id, delta int) (int, error) {
	const sql = `UPDATE products SET stock = stock + $1
	             WHERE id = $2 AND stock + $1 >= 0
	             RETURNING stock`
	var newStock int
	err := q.QueryRow(ctx, sql, delta, id).Scan(&newStock)
	return newStock, err
}

func (r *ProductRepo) GetStock(ctx context.Context, q data.Querier, id int) (int, error) {
	var s int
	err := q.QueryRow(ctx, `SELECT stock FROM products WHERE id=$1`, id).Scan(&s)
	return s, err
}

func (r *ProductRepo) InsertMovement(ctx context.Context, q data.Querier, productID int, typ string, qty int, note string) error {
	_, err := q.Exec(ctx,
		`INSERT INTO stock_movements (product_id, type, qty, note) VALUES ($1,$2,$3,$4)`,
		productID, typ, qty, note)
	return err
}

func (r *ProductRepo) GetBySKUs(ctx context.Context, q data.Querier, skus []string) (map[string]ProductForOrder, error) {
	const sql = `SELECT sku, price, stock FROM products WHERE sku = ANY($1)`
	rows, err := q.Query(ctx, sql, skus)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[string]ProductForOrder)
	for rows.Next() {
		var p ProductForOrder
		if err := rows.Scan(&p.SKU, &p.Price, &p.Stock); err != nil {
			return nil, err
		}
		m[p.SKU] = p
	}
	return m, rows.Err()
}

// DecrementStockBySKU
// guard atomic in WHERE, return price from the same row
func (r *ProductRepo) DecrementStockBySKU(ctx context.Context, q data.Querier, sku string, qty int) (int, int, error) {
	const sql = `UPDATE products SET stock = stock - $1
	             WHERE sku = $2 AND stock >= $1
	             RETURNING stock, price`
	var newStock, price int
	err := q.QueryRow(ctx, sql, qty, sku).Scan(&newStock, &price)
	return newStock, price, err
}

func (r *ProductRepo) IncrementStockBySKU(ctx context.Context, q data.Querier, sku string, qty int) error {
	_, err := q.Exec(ctx, `UPDATE products SET stock = stock + $1 WHERE sku=$2`, qty, sku)
	return err
}

func (r *ProductRepo) GetStockBySKU(ctx context.Context, q data.Querier, sku string) (int, error) {
	var s int
	err := q.QueryRow(ctx, `SELECT stock FROM products WHERE sku=$1`, sku).Scan(&s)
	return s, err
}

func (r *ProductRepo) InsertMovementBySKU(ctx context.Context, q data.Querier, sku, typ string, qty int, note string) error {
	_, err := q.Exec(ctx,
		`INSERT INTO stock_movements (product_id, type, qty, note)
		 SELECT id, $1, $2, $3 FROM products WHERE sku=$4`,
		typ, qty, note, sku)
	return err
}
