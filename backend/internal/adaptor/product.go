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
			respondErr(c, err)
			return
		}
		Success(c, http.StatusOK, resp)
	}
}

func CreateProduct(uc *usecase.ProductUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.CreateProductRequest
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

func UpdateProduct(uc *usecase.ProductUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			Error(c, http.StatusBadRequest, "INVALID_ID", nil)
			return
		}
		var req dto.UpdateProductRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Error(c, http.StatusBadRequest, "VALIDATION_ERROR", gin.H{"message": err.Error()})
			return
		}
		resp, err := uc.Update(c.Request.Context(), id, req)
		if err != nil {
			respondErr(c, err)
			return
		}
		Success(c, http.StatusOK, resp)
	}
}

func DeleteProduct(uc *usecase.ProductUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			Error(c, http.StatusBadRequest, "INVALID_ID", nil)
			return
		}
		if err := uc.Delete(c.Request.Context(), id); err != nil {
			respondErr(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func AdjustStock(uc *usecase.ProductUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return
		}
		var req dto.AdjustStockRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Error(c, http.StatusBadRequest, "VALIDATION_ERROR", gin.H{"message": err.Error()})
			return
		}
		resp, err := uc.AdjustStock(c.Request.Context(), id, req)
		if err != nil {
			respondErr(c, err)
			return
		}
		Success(c, http.StatusOK, resp)
	}
}
