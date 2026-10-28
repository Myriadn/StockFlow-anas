package wire

import (
	"stackflow_anas/internal/adaptor"
	"stackflow_anas/internal/data/repository"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	ProductRepo *repository.ProductRepo
}

func NewRouter(h Handlers) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/healthz", adaptor.Health(h.ProductRepo))
	return r
}
