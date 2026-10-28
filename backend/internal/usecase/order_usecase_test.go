package usecase_test

import (
	"context"
	"errors"
	"testing"

	"stackflow_anas/internal/data/repository"
	"stackflow_anas/internal/dto"
	"stackflow_anas/internal/usecase"
)

func newOrderUC() (*usecase.OrderUsecase, *usecase.ProductUsecase) {
	pr := repository.NewProductRepo(testPool)
	or := repository.NewOrderRepo(testPool)
	return usecase.NewOrderUsecase(or, pr, testPool), usecase.NewProductUsecase(pr, testPool)
}

// happy path writes movements + updates stock
func TestOrderCreate_HappyPath(t *testing.T) {
	resetDB(t)
	puc := newProductUC()
	ouc, _ := newOrderUC()

	p1 := mustCreate(t, puc, "A", "Alpha", 1000)
	p2 := mustCreate(t, puc, "B", "Beta", 2000)
	_, _ = puc.AdjustStock(context.Background(), p1.ID, dto.AdjustStockRequest{Type: "in", Qty: 10})
	_, _ = puc.AdjustStock(context.Background(), p2.ID, dto.AdjustStockRequest{Type: "in", Qty: 10})

	resp, err := ouc.Create(context.Background(), dto.CreateOrderRequest{
		Customer: "Cust",
		Items:    []dto.CreateOrderItem{{SKU: "A", Qty: 3}, {SKU: "B", Qty: 2}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 3*1000+2*2000 {
		t.Fatalf("total want 7000 got %d", resp.Total)
	}

	var sa, sb int
	_ = testPool.QueryRow(context.Background(), `SELECT stock FROM products WHERE sku='A'`).Scan(&sa)
	_ = testPool.QueryRow(context.Background(), `SELECT stock FROM products WHERE sku='B'`).Scan(&sb)
	if sa != 7 || sb != 8 {
		t.Fatalf("want stock A=7 B=8, got A=%d B=%d", sa, sb)
	}

	var movOut int
	_ = testPool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM stock_movements WHERE type='out'`).Scan(&movOut)
	if movOut != 2 {
		t.Fatalf("want 2 out movements, got %d", movOut)
	}
}

// insufficient stock => 409, NOTHING written
func TestOrderCreate_InsufficientStock_NothingWritten(t *testing.T) {
	resetDB(t)
	puc := newProductUC()
	ouc, _ := newOrderUC()

	p1 := mustCreate(t, puc, "A", "Alpha", 1000)
	p2 := mustCreate(t, puc, "B", "Beta", 2000)
	_, _ = puc.AdjustStock(context.Background(), p1.ID, dto.AdjustStockRequest{Type: "in", Qty: 5})
	_, _ = puc.AdjustStock(context.Background(), p2.ID, dto.AdjustStockRequest{Type: "in", Qty: 5})

	var ordersBefore, itemsBefore, movBefore int
	_ = testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM sales_orders`).Scan(&ordersBefore)
	_ = testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM sales_order_items`).Scan(&itemsBefore)
	_ = testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM stock_movements`).Scan(&movBefore)

	_, err := ouc.Create(context.Background(), dto.CreateOrderRequest{
		Customer: "Cust",
		Items:    []dto.CreateOrderItem{{SKU: "A", Qty: 3}, {SKU: "B", Qty: 999}},
	})

	var appErr *usecase.AppError
	if !errors.As(err, &appErr) || appErr.Code != "INSUFFICIENT_STOCK" {
		t.Fatalf("want INSUFFICIENT_STOCK, got %v", err)
	}
	if appErr.Extra["sku"] != "B" || appErr.Extra["available"] != 5 {
		t.Fatalf("want sku=B available=5, got %v", appErr.Extra)
	}

	// NOTHING written
	var ordersAfter, itemsAfter, movAfter int
	_ = testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM sales_orders`).Scan(&ordersAfter)
	_ = testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM sales_order_items`).Scan(&itemsAfter)
	_ = testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM stock_movements`).Scan(&movAfter)
	if ordersBefore != ordersAfter || itemsBefore != itemsAfter || movBefore != movAfter {
		t.Fatalf("nothing should be written: orders %d→%d items %d→%d mov %d→%d",
			ordersBefore, ordersAfter, itemsBefore, itemsAfter, movBefore, movAfter)
	}

	// stock A harus tetap 5 (tidak terlanjur decrement item pertama)
	var sa int
	_ = testPool.QueryRow(context.Background(), `SELECT stock FROM products WHERE sku='A'`).Scan(&sa)
	if sa != 5 {
		t.Fatalf("stock A must remain 5, got %d", sa)
	}
}

func TestOrderCreate_UnknownSKU(t *testing.T) {
	resetDB(t)
	ouc, _ := newOrderUC()
	_, err := ouc.Create(context.Background(), dto.CreateOrderRequest{
		Customer: "C", Items: []dto.CreateOrderItem{{SKU: "NOPE", Qty: 1}},
	})
	var appErr *usecase.AppError
	if !errors.As(err, &appErr) || appErr.Code != "UNKNOWN_SKU" {
		t.Fatalf("want UNKNOWN_SKU, got %v", err)
	}
}

func TestOrderCreate_DuplicateSKU(t *testing.T) {
	resetDB(t)
	ouc, _ := newOrderUC()
	_, err := ouc.Create(context.Background(), dto.CreateOrderRequest{
		Customer: "C",
		Items:    []dto.CreateOrderItem{{SKU: "A", Qty: 1}, {SKU: "A", Qty: 2}},
	})
	var appErr *usecase.AppError
	if !errors.As(err, &appErr) || appErr.Code != "DUPLICATE_SKU" {
		t.Fatalf("want DUPLICATE_SKU, got %v", err)
	}
}

func TestOrderCancel_RestoresStock(t *testing.T) {
	resetDB(t)
	puc := newProductUC()
	ouc, _ := newOrderUC()

	p1 := mustCreate(t, puc, "A", "Alpha", 1000)
	_, _ = puc.AdjustStock(context.Background(), p1.ID, dto.AdjustStockRequest{Type: "in", Qty: 10})

	order, _ := ouc.Create(context.Background(), dto.CreateOrderRequest{
		Customer: "C", Items: []dto.CreateOrderItem{{SKU: "A", Qty: 4}},
	})

	var s int
	_ = testPool.QueryRow(context.Background(), `SELECT stock FROM products WHERE sku='A'`).Scan(&s)
	if s != 6 {
		t.Fatalf("want 6 got %d", s)
	}

	resp, err := ouc.Cancel(context.Background(), order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "CANCELLED" {
		t.Fatalf("want CANCELLED got %s", resp.Status)
	}

	_ = testPool.QueryRow(context.Background(), `SELECT stock FROM products WHERE sku='A'`).Scan(&s)
	if s != 10 {
		t.Fatalf("want 10 got %d", s)
	}

	_, err = ouc.Cancel(context.Background(), order.ID)
	var appErr *usecase.AppError
	if !errors.As(err, &appErr) || appErr.Code != "ORDER_ALREADY_CANCELLED" {
		t.Fatalf("want ORDER_ALREADY_CANCELLED, got %v", err)
	}
}

func TestOrderCancel_NotFound(t *testing.T) {
	resetDB(t)
	ouc, _ := newOrderUC()
	_, err := ouc.Cancel(context.Background(), 999)
	var appErr *usecase.AppError
	if !errors.As(err, &appErr) || appErr.Code != "ORDER_NOT_FOUND" {
		t.Fatalf("want ORDER_NOT_FOUND, got %v", err)
	}
}
