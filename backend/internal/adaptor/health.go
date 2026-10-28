package adaptor

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

func Health(p Pinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := p.Ping(ctx); err != nil {
			Error(c, http.StatusServiceUnavailable, "DB_UNHEALTHY", nil)
			return
		}
		Success(c, http.StatusOK, gin.H{"status": "ok"})
	}
}
