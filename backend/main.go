package main

import (
	"context"
	"log"
	"stackflow_anas/cmd"
	"stackflow_anas/internal/config"
	"stackflow_anas/internal/data"
	"stackflow_anas/internal/data/repository"
	"stackflow_anas/internal/usecase"
	"stackflow_anas/internal/wire"
)

func main() {
	cfg := config.Load()

	pool, err := data.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db pool: %v", err)
	}
	defer pool.Close()

	productRepo := repository.NewProductRepo(pool)
	orderRepo := repository.NewOrderRepo(pool)

	productUC := usecase.NewProductUsecase(productRepo, pool)
	orderUC := usecase.NewOrderUsecase(orderRepo, productRepo, pool)

	router := wire.NewRouter(wire.Handlers{
		ProductUsecase: productUC,
		ProductRepo:    productRepo,
		OrderUsecase:   orderUC,
	})

	cmd.StartServer(router, cfg.Port)
}
