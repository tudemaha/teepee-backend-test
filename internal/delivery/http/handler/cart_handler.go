package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/middleware"
	"github.com/tudemaha/marketplace-be/internal/usecase"
	"github.com/tudemaha/marketplace-be/pkg/apperror"
	"github.com/tudemaha/marketplace-be/pkg/response"
	"github.com/tudemaha/marketplace-be/pkg/validator"
)

type CartHandler struct {
	cartUC usecase.CartUseCase
}

func NewCartHandler(g *echo.Group, cartUC usecase.CartUseCase, jwtSecret string) {
	h := &CartHandler{
		cartUC: cartUC,
	}

	cartGroup := g.Group("/carts")

	cartGroup.Use(middleware.Auth(jwtSecret))

	cartGroup.GET("", h.GetMyCart)
	cartGroup.POST("", h.AddToCart)
	cartGroup.PUT("/:id", h.UpdateCartItem)
	cartGroup.DELETE("/:id", h.RemoveFromCart)
}

func (h *CartHandler) GetMyCart(c *echo.Context) error {
	userIDStr := c.Get("user_id").(string)
	userID, _ := uuid.Parse(userIDStr)

	res, err := h.cartUC.GetMyCart(userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "cart fetched successfully", res)
}

func (h *CartHandler) AddToCart(c *echo.Context) error {
	userIDStr := c.Get("user_id").(string)
	userID, _ := uuid.Parse(userIDStr)

	var req dto.AddToCartRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", []string{err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "validation failed", validator.FormatErrors(err))
	}

	if err := h.cartUC.AddToCart(userID, &req); err != nil {
		if errors.Is(err, apperror.ErrBadRequest) {
			return response.Error(c, http.StatusBadRequest, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusCreated, "item added to cart successfully", nil)
}

func (h *CartHandler) UpdateCartItem(c *echo.Context) error {
	userIDStr := c.Get("user_id").(string)
	userID, _ := uuid.Parse(userIDStr)

	idParam := c.Param("id")
	cartID, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid cart item id", nil)
	}

	var req dto.UpdateCartRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", []string{err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "validation failed", validator.FormatErrors(err))
	}

	if err := h.cartUC.UpdateCartItem(userID, cartID, &req); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return response.Error(c, http.StatusNotFound, err.Error(), nil)
		}
		if errors.Is(err, apperror.ErrBadRequest) {
			return response.Error(c, http.StatusBadRequest, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "cart item updated successfully", nil)
}

func (h *CartHandler) RemoveFromCart(c *echo.Context) error {
	userIDStr := c.Get("user_id").(string)
	userID, _ := uuid.Parse(userIDStr)

	idParam := c.Param("id")
	cartID, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid cart item id", nil)
	}

	if err := h.cartUC.RemoveFromCart(userID, cartID); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return response.Error(c, http.StatusNotFound, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "cart item removed successfully", nil)
}
