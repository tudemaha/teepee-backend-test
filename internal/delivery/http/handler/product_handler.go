package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/middleware"
	"github.com/tudemaha/marketplace-be/internal/usecase"
	"github.com/tudemaha/marketplace-be/pkg/response"
	"github.com/tudemaha/marketplace-be/pkg/validator"
)

type ProductHandler struct {
	productUC usecase.ProductUseCase
}

func NewProductHandler(g *echo.Group, productUC usecase.ProductUseCase, jwtSecret string) {
	h := &ProductHandler{
		productUC: productUC,
	}

	prodGroup := g.Group("/products")

	prodGroup.GET("", h.GetAll)
	prodGroup.GET("/:id", h.GetByID)

	prodGroup.POST("", h.Create, middleware.Auth(jwtSecret), middleware.RequireRole("seller", "admin"))
	prodGroup.PUT("/:id", h.Update, middleware.Auth(jwtSecret), middleware.RequireRole("seller", "admin"))
	prodGroup.DELETE("/:id", h.Delete, middleware.Auth(jwtSecret), middleware.RequireRole("seller", "admin"))
}

func (h *ProductHandler) Create(c *echo.Context) error {
	userIDStr := c.Get("user_id").(string)
	userID, _ := uuid.Parse(userIDStr)

	var req dto.CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", []string{err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "validation failed", validator.FormatErrors(err))
	}

	res, err := h.productUC.Create(userID, &req)
	if err != nil {
		if err.Error() == "shop not found for this seller" {
			return response.Error(c, http.StatusForbidden, err.Error(), nil)
		}
		if err.Error() == "one or more categories not found" {
			return response.Error(c, http.StatusBadRequest, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusCreated, "product created successfully", res)
}

func (h *ProductHandler) GetByID(c *echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid product id format", nil)
	}

	res, err := h.productUC.GetByID(id)
	if err != nil {
		return response.Error(c, http.StatusNotFound, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "product fetched successfully", res)
}

func (h *ProductHandler) GetAll(c *echo.Context) error {
	var filter dto.ProductListFilter

	if shopID := c.QueryParam("shop_id"); shopID != "" {
		filter.ShopID, _ = uuid.Parse(shopID)
	}
	if catID := c.QueryParam("category_id"); catID != "" {
		filter.CategoryID, _ = uuid.Parse(catID)
	}
	filter.Search = c.QueryParam("search")

	if limit := c.QueryParam("limit"); limit != "" {
		filter.Limit, _ = strconv.Atoi(limit)
	}
	if offset := c.QueryParam("offset"); offset != "" {
		filter.Offset, _ = strconv.Atoi(offset)
	}

	res, err := h.productUC.GetAll(filter)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "products fetched successfully", res)
}

func (h *ProductHandler) Update(c *echo.Context) error {
	userIDStr := c.Get("user_id").(string)
	userID, _ := uuid.Parse(userIDStr)

	idParam := c.Param("id")
	productID, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid product id format", nil)
	}

	var req dto.UpdateProductRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", []string{err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "validation failed", validator.FormatErrors(err))
	}

	res, err := h.productUC.Update(userID, productID, &req)
	if err != nil {
		if err.Error() == "product not found" {
			return response.Error(c, http.StatusNotFound, err.Error(), nil)
		}
		if err.Error() == "unauthorized to update this product" {
			return response.Error(c, http.StatusForbidden, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "product updated successfully", res)
}

func (h *ProductHandler) Delete(c *echo.Context) error {
	userIDStr := c.Get("user_id").(string)
	userID, _ := uuid.Parse(userIDStr)

	idParam := c.Param("id")
	productID, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid product id format", nil)
	}

	if err := h.productUC.Delete(userID, productID); err != nil {
		if err.Error() == "product not found" {
			return response.Error(c, http.StatusNotFound, err.Error(), nil)
		}
		if err.Error() == "unauthorized to delete this product" {
			return response.Error(c, http.StatusForbidden, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "product deleted successfully", nil)
}
