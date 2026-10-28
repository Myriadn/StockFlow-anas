package wire

import (
	"stackflow_anas/internal/adaptor"
	"stackflow_anas/internal/data/repository"
	"stackflow_anas/internal/usecase"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	ProductUsecase *usecase.ProductUsecase
	ProductRepo    *repository.ProductRepo
	OrderUsecase   *usecase.OrderUsecase
}

func NewRouter(h Handlers) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/healthz", adaptor.Health(h.ProductRepo))

	r.GET("/api/products", adaptor.ListProducts(h.ProductUsecase))
	r.POST("/api/products", adaptor.CreateProduct(h.ProductUsecase))
	r.PUT("/api/products/:id", adaptor.UpdateProduct(h.ProductUsecase))
	r.DELETE("/api/products/:id", adaptor.DeleteProduct(h.ProductUsecase))
	r.POST("/api/products/:id/stock", adaptor.AdjustStock(h.ProductUsecase))
	r.GET("/api/products/:id/movements", adaptor.ListMovements(h.ProductUsecase))

	r.POST("/api/orders", adaptor.CreateOrder(h.OrderUsecase))
	r.GET("/api/orders/:id", adaptor.GetOrder(h.OrderUsecase))
	r.POST("/api/orders/:id/cancel", adaptor.CancelOrder(h.OrderUsecase))
	return r
}
