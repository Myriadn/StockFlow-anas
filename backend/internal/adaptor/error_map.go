package adaptor

import (
	"errors"
	"net/http"
	"stackflow_anas/internal/usecase"

	"github.com/gin-gonic/gin"
)

func respondErr(c *gin.Context, err error) {
	var appErr *usecase.AppError
	if errors.As(err, &appErr) {
		Error(c, appErr.Status, appErr.Code, gin.H(appErr.Extra))
		return
	}
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", nil)
}
