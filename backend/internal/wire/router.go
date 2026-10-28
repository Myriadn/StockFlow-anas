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
}

func NewRouter(h Handlers) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/healthz", adaptor.Health(h.ProductRepo))
	r.GET("/api/products", adaptor.ListProducts(h.ProductUsecase))
	return r
}
