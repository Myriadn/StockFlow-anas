package usecase_test

import (
	"context"
	"log"
	"os"
	"stackflow_anas/internal/data"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://stockflow:stockflow@localhost:5432/db_stockflow?sslmode=disable"
	}
	pool, err := data.NewPool(context.Background(), url)
	if err != nil {
		log.Fatalf("test db: %v", err)
	}
	testPool = pool
	code := m.Run()
	pool.Close()
	os.Exit(code)
}

func resetDB(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(),
		"TRUNCATE sales_order_items, sales_orders, stock_movements, products RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}
