package usecase_test

import (
	"context"
	"errors"
	"testing"

	"stackflow_anas/internal/data/repository"
	"stackflow_anas/internal/dto"
	"stackflow_anas/internal/usecase"
)

func newProductUC() *usecase.ProductUsecase {
	return usecase.NewProductUsecase(repository.NewProductRepo(testPool), testPool)
}

func mustCreate(t *testing.T, uc *usecase.ProductUsecase, sku, name string, price int) *dto.ProductResponse {
	t.Helper()
	p, err := uc.Create(context.Background(), dto.CreateProductRequest{SKU: sku, Name: name, Price: &price})
	if err != nil {
		t.Fatalf("create %s: %v", sku, err)
	}
	return p
}

func TestCreate_DuplicateSKU(t *testing.T) {
	resetDB(t)
	uc := newProductUC()
	mustCreate(t, uc, "A", "Alpha", 1000)

	price := 1000
	_, err := uc.Create(context.Background(), dto.CreateProductRequest{SKU: "A", Name: "Alpha2", Price: &price})

	var appErr *usecase.AppError
	if !errors.As(err, &appErr) || appErr.Code != "SKU_ALREADY_EXISTS" {
		t.Fatalf("want SKU_ALREADY_EXISTS, got %v", err)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	resetDB(t)
	uc := newProductUC()
	name := "X"
	_, err := uc.Update(context.Background(), 999, dto.UpdateProductRequest{Name: &name})

	var appErr *usecase.AppError
	if !errors.As(err, &appErr) || appErr.Code != "PRODUCT_NOT_FOUND" {
		t.Fatalf("want PRODUCT_NOT_FOUND, got %v", err)
	}
}

func TestDelete_WithMovement_Returns409(t *testing.T) {
	resetDB(t)
	uc := newProductUC()
	p := mustCreate(t, uc, "B", "Beta", 2000)

	_, err := testPool.Exec(context.Background(),
		`INSERT INTO stock_movements (product_id, type, qty, note) VALUES ($1,'in',5,'seed')`, p.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = uc.Delete(context.Background(), p.ID)

	var appErr *usecase.AppError
	if !errors.As(err, &appErr) || appErr.Code != "PRODUCT_HAS_HISTORY" {
		t.Fatalf("want PRODUCT_HAS_HISTORY, got %v", err)
	}

	var n int
	_ = testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM products WHERE id=$1`, p.ID).Scan(&n)
	if n != 1 {
		t.Fatalf("product should still exist, got %d", n)
	}
}

func TestDelete_NotFound(t *testing.T) {
	resetDB(t)
	uc := newProductUC()
	err := uc.Delete(context.Background(), 4242)

	var appErr *usecase.AppError
	if !errors.As(err, &appErr) || appErr.Code != "PRODUCT_NOT_FOUND" {
		t.Fatalf("want PRODUCT_NOT_FOUND, got %v", err)
	}
}

func TestList_SearchAndPagination(t *testing.T) {
	resetDB(t)
	uc := newProductUC()
	mustCreate(t, uc, "KOPI-1", "Kopi A", 1000)
	mustCreate(t, uc, "KOPI-2", "Kopi B", 2000)
	mustCreate(t, uc, "TEH-1", "Teh A", 500)

	resp, err := uc.List(context.Background(), dto.ListProductsQuery{Search: "kopi", Page: 1, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 2 || len(resp.Data) != 2 {
		t.Fatalf("want 2 kopi, got total=%d len=%d", resp.Total, len(resp.Data))
	}
}
