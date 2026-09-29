package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/middleware"
	"github.com/tudemaha/marketplace-be/internal/usecase"
	"github.com/tudemaha/marketplace-be/pkg/response"
	"github.com/tudemaha/marketplace-be/pkg/validator"
)

type ShopHandler struct {
	shopUC usecase.ShopUseCase
}

func NewShopHandler(g *echo.Group, shopUC usecase.ShopUseCase, jwtSecret string) {
	h := &ShopHandler{
		shopUC: shopUC,
	}

	shopGroup := g.Group("/shops")

	shopGroup.GET("/:id", h.GetByID)

	shopGroup.POST("", h.Create, middleware.Auth(jwtSecret))
	shopGroup.PUT("/:id", h.Update, middleware.Auth(jwtSecret))
}

func (h *ShopHandler) Create(c *echo.Context) error {
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.Error(c, http.StatusUnauthorized, "invalid user ID", nil)
	}

	var req dto.CreateShopRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", []string{err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "validation failed", validator.FormatErrors(err))
	}

	res, err := h.shopUC.Create(userID, &req)
	if err != nil {
		if err.Error() == "user already has a shop" {
			return response.Error(c, http.StatusConflict, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusCreated, "shop created successfully", res)
}

func (h *ShopHandler) GetByID(c *echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid shop id format", nil)
	}

	res, err := h.shopUC.GetByID(id)
	if err != nil {
		return response.Error(c, http.StatusNotFound, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "shop fetched successfully", res)
}

func (h *ShopHandler) Update(c *echo.Context) error {
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.Error(c, http.StatusUnauthorized, "invalid user ID", nil)
	}

	shopIDParam := c.Param("id")
	shopID, err := uuid.Parse(shopIDParam)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid shop id format", nil)
	}

	var req dto.UpdateShopRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", []string{err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "validation failed", validator.FormatErrors(err))
	}

	res, err := h.shopUC.Update(userID, shopID, &req)
	if err != nil {
		if err.Error() == "shop not found" {
			return response.Error(c, http.StatusNotFound, err.Error(), nil)
		}
		if err.Error() == "unauthorized to update this shop" {
			return response.Error(c, http.StatusForbidden, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "shop updated successfully", res)
}
