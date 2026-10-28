# Stockflow - Inventory & Sales

Mini ERP module for inventory & sales: manage products, adjust stock, create sales orders that consume stock.

**Stack:** React + TypeScript (Vite) · Go + PostgreSQL · frontend consumes the real backend API.

## Repository Layout

```
stockflow/
├── backend/            Go service (Gin, pgx)
├── frontend/           React + TypeScript (Vite)
├── db/                 .sql (auto-applied by Postgres on init)
├── docker-compose.yml  Postgres 16 with schema + seed
└── README.md
```

## How to Run

**Prerequisites:** Docker, Go ≥ 1.22, Node ≥ 18.

```bash
# 1. Database (Postgres 16, schema + seed auto-applied on first run)
docker compose up -d
# wait for "ready to accept connections", then Ctrl+C
docker compose logs -f postgres

# 2. Backend (http://localhost:9002)
cd backend
go run .

# 3. Frontend (http://localhost:5173)
cd frontend
npm install
npm run dev
```

Open http://localhost:5173.

### Environment

Backend reads `.env` (see `backend/.env.example`):

```
DATABASE_URL=postgres://stockflow:stockflow@localhost:5432/db_stockflow?sslmode=disable
TEST_DATABASE_URL=postgres://stockflow:stockflow@localhost:5432/db_stockflow?sslmode=disable
PORT=9002
```

Frontend uses Vite proxy: `/api` → `http://localhost:9002`. No CORS middleware needed.

### Reset Database

Schema and seed live in `db/` and are applied by Postgres's `docker-entrypoint-initdb.d` **only when the data volume is empty**. To re-apply:

```bash
docker compose down -v && docker compose up -d
```

> `go test` truncates the dev database. Run the reset command above after running tests.

## API

All errors use the envelope `{"error": {"code": "...", ...extras}}`.

```bash
# Products
curl 'localhost:9002/api/products?search=kopi&page=1&limit=10'

curl -X POST localhost:9002/api/products \
  -H 'Content-Type: application/json' \
  -d '{"sku":"TEST-1","name":"Test","price":10000}'

curl -X PUT localhost:9002/api/products/1 \
  -H 'Content-Type: application/json' -d '{"name":"Renamed"}'

# 409 if movement history exists
curl -X DELETE localhost:9002/api/products/1

# Stock
# 409 INSUFFICIENT_STOCK with available
curl -X POST localhost:9002/api/products/1/stock \
  -H 'Content-Type: application/json' \
  -d '{"type":"out","qty":5,"note":"test"}'

# Movements
curl localhost:9002/api/products/1/movements

# Orders
curl -X POST localhost:9002/api/orders \
  -H 'Content-Type: application/json' \
  -d '{"customer":"Toko Maju","items":[{"sku":"KOPI-ARB-1K","qty":2}]}'

curl localhost:9002/api/orders/1

# restores stock; 409 if already cancelled
curl -X POST localhost:9002/api/orders/1/cancel

# Health
curl localhost:9002/healthz
```

## Tests

```bash
cd backend
go test ./... -race -count=1
```

Covers:

- Product CRUD: duplicate SKU, not found, delete with history returns 409
- Stock adjust: happy path writes movement + updates stock; **insufficient stock returns 409 and nothing is written**; not found
- Order create: happy path writes movements + decrements all items atomically; **insufficient stock returns 409 and nothing is written** (order count, item count, movement count, and stock all unchanged); unknown SKU; duplicate SKU
- Order cancel: restores stock; second cancel returns 409; not found

## Design Decisions

- **`last_movement_at`**: derived via subquery `MAX(created_at)` from `stock_movements` instead of maintaining a column on `products`. Simpler, no drift risk, fine at this scale.
- **Transaction boundary lives in the usecase**: Repository methods take a `data.Querier` interface (satisfied by both `*pgxpool.Pool` and `pgx.Tx`), so a usecase can `Begin()` a tx and pass it through every repository call. This is what makes multi-item order creation atomic.
- **Concurrency safety on stock**: `UPDATE products SET stock = stock - $1 WHERE sku = $2 AND stock >= $1`. Never `SELECT`-then-`UPDATE`. `rowsAffected == 0` → 409 `INSUFFICIENT_STOCK` with `available` fetched inside the same transaction.
- **Order total**: computed server-side from `RETURNING price` on the same `UPDATE`. Client never supplies prices.
- **Cancel**: restores stock via a `in` movement (`note: "Cancel order #<id>"`). Guarded by `WHERE status='CREATED'` so double-cancel is safe; returns 409 `ORDER_ALREADY_CANCELLED`.
- **Delete product**: checks `stock_movements` first; returns 409 `PRODUCT_HAS_HISTORY` if any. History is never deleted.
- **Frontend ↔ backend**: Vite proxy (`/api` → `localhost:9002`). Frontend uses relative paths only; one API client module owns all fetch calls and error envelope parsing.
- **Migrations**: `docker-entrypoint-initdb.d` (alphabetical order). Not a real migration tool (no incremental changes on an existing volume), but sufficient for a 1-day test.

## What I Would Improve

- **Idempotency-Key** on `POST /api/orders` (bonus item, skipped to focus on concurrency correctness).
- **Separate test DB** via `TEST_DATABASE_URL` so `go test` doesn't truncate the dev database.
- **Concurrent order test**, spin up N goroutines ordering the same SKU and assert stock never goes negative and `-race` is clean.
- **Frontend**, create-order form, order detail + cancel UI, filter (All/Active/Inactive/Low stock), summary cards, skeleton loading, debounced search, Vitest for pure logic.
- **Migration tool** (golang-migrate) for incremental schema changes.
- **Structured logging** and request IDs.

## Hours Spent

Approximately **6-7 hours** of focused work.
