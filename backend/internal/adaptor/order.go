package adaptor

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"stackflow_anas/internal/dto"
	"stackflow_anas/internal/usecase"
)

func CreateOrder(uc *usecase.OrderUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.CreateOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Error(c, http.StatusBadRequest, "VALIDATION_ERROR", gin.H{"message": err.Error()})
			return
		}
		resp, err := uc.Create(c.Request.Context(), req)
		if err != nil {
			respondErr(c, err)
			return
		}
		Success(c, http.StatusCreated, resp)
	}
}

func GetOrder(uc *usecase.OrderUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil || id <= 0 {
			Error(c, http.StatusBadRequest, "INVALID_ID", nil)
			return
		}
		resp, err := uc.Get(c.Request.Context(), id)
		if err != nil {
			respondErr(c, err)
			return
		}
		Success(c, http.StatusOK, resp)
	}
}

func CancelOrder(uc *usecase.OrderUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil || id <= 0 {
			Error(c, http.StatusBadRequest, "INVALID_ID", nil)
			return
		}
		resp, err := uc.Cancel(c.Request.Context(), id)
		if err != nil {
			respondErr(c, err)
			return
		}
		Success(c, http.StatusOK, resp)
	}
}
