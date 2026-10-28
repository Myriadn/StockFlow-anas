package adaptor

import (
	"net/http"
	"stackflow_anas/internal/dto"
	"stackflow_anas/internal/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListProducts(uc *usecase.ProductUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		resp, err := uc.List(c.Request.Context(), dto.ListProductsQuery{
			Search: c.Query("search"),
			Page:   page,
			Limit:  limit,
		})
		if err != nil {
			Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", nil)
			return
		}
		Success(c, http.StatusOK, resp)
	}
}
