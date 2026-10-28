package usecase_test

import (
	"context"
	"errors"
	"testing"

	"stackflow_anas/internal/dto"
	"stackflow_anas/internal/usecase"
)

func TestAdjustStock_In_HappyPath(t *testing.T) {
	resetDB(t)
	uc := newProductUC()
	p := mustCreate(t, uc, "S1", "Stock Item", 1000)

	resp, err := uc.AdjustStock(context.Background(), p.ID, dto.AdjustStockRequest{Type: "in", Qty: 10, Note: "restock"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Stock != 10 {
		t.Fatalf("want stock 10, got %d", resp.Stock)
	}

	var stock, movCount int
	_ = testPool.QueryRow(context.Background(), `SELECT stock FROM products WHERE id=$1`, p.ID).Scan(&stock)
	_ = testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM stock_movements WHERE product_id=$1`, p.ID).Scan(&movCount)
	if stock != 10 || movCount != 1 {
		t.Fatalf("want stock=10 mov=1, got stock=%d mov=%d", stock, movCount)
	}
}

// insufficient stock denied, NOTHING written.
func TestAdjustStock_Out_Insufficient_NothingWritten(t *testing.T) {
	resetDB(t)
	uc := newProductUC()
	p := mustCreate(t, uc, "S2", "Item", 1000)

	// in stock 5
	if _, err := uc.AdjustStock(context.Background(), p.ID, dto.AdjustStockRequest{Type: "in", Qty: 5}); err != nil {
		t.Fatal(err)
	}

	var stockBefore, movBefore int
	_ = testPool.QueryRow(context.Background(), `SELECT stock FROM products WHERE id=$1`, p.ID).Scan(&stockBefore)
	_ = testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM stock_movements WHERE product_id=$1`, p.ID).Scan(&movBefore)

	// out 999 => should 409
	_, err := uc.AdjustStock(context.Background(), p.ID, dto.AdjustStockRequest{Type: "out", Qty: 999})

	var appErr *usecase.AppError
	if !errors.As(err, &appErr) || appErr.Code != "INSUFFICIENT_STOCK" {
		t.Fatalf("want INSUFFICIENT_STOCK, got %v", err)
	}
	if got, _ := appErr.Extra["available"].(int); got != stockBefore {
		t.Fatalf("want available=%d, got %v", stockBefore, appErr.Extra["available"])
	}

	// NOTHING written: stock & movement count identic
	var stockAfter, movAfter int
	_ = testPool.QueryRow(context.Background(), `SELECT stock FROM products WHERE id=$1`, p.ID).Scan(&stockAfter)
	_ = testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM stock_movements WHERE product_id=$1`, p.ID).Scan(&movAfter)
	if stockAfter != stockBefore || movAfter != movBefore {
		t.Fatalf("nothing should be written: stock %d→%d, mov %d→%d",
			stockBefore, stockAfter, movBefore, movAfter)
	}
}

func TestAdjustStock_NotFound(t *testing.T) {
	resetDB(t)
	uc := newProductUC()
	_, err := uc.AdjustStock(context.Background(), 999, dto.AdjustStockRequest{Type: "in", Qty: 1})

	var appErr *usecase.AppError
	if !errors.As(err, &appErr) || appErr.Code != "PRODUCT_NOT_FOUND" {
		t.Fatalf("want PRODUCT_NOT_FOUND, got %v", err)
	}
}
