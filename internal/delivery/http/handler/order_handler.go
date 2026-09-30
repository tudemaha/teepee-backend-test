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

type OrderHandler struct {
	orderUC usecase.OrderUseCase
}

func NewOrderHandler(g *echo.Group, orderUC usecase.OrderUseCase, jwtSecret string) {
	h := &OrderHandler{
		orderUC: orderUC,
	}

	orderGroup := g.Group("/orders")

	orderGroup.Use(middleware.Auth(jwtSecret))
	
	orderGroup.POST("/checkout", h.Checkout)
	orderGroup.GET("", h.GetMyOrders)
	orderGroup.GET("/:id", h.GetByID)
	orderGroup.PATCH("/:id/status", h.UpdateStatus)
}

func (h *OrderHandler) Checkout(c *echo.Context) error {
	userIDStr := c.Get("user_id").(string)
	userID, _ := uuid.Parse(userIDStr)

	var req dto.CheckoutRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", []string{err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "validation failed", validator.FormatErrors(err))
	}

	res, err := h.orderUC.Checkout(c.Request().Context(), userID, &req)
	if err != nil {
		if errors.Is(err, apperror.ErrBadRequest) {
			return response.Error(c, http.StatusBadRequest, err.Error(), nil)
		}
		if errors.Is(err, apperror.ErrBadRequest) {
			return response.Error(c, http.StatusBadRequest, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusCreated, "checkout successful", res)
}

func (h *OrderHandler) GetMyOrders(c *echo.Context) error {
	userIDStr := c.Get("user_id").(string)
	userID, _ := uuid.Parse(userIDStr)

	res, err := h.orderUC.GetMyOrders(userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "orders fetched successfully", res)
}

func (h *OrderHandler) GetByID(c *echo.Context) error {
	userIDStr := c.Get("user_id").(string)
	userID, _ := uuid.Parse(userIDStr)

	idParam := c.Param("id")
	orderID, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid order id format", nil)
	}

	res, err := h.orderUC.GetByID(userID, orderID)
	if err != nil {
		if err.Error() == "unauthorized to view this order" {
			return response.Error(c, http.StatusForbidden, err.Error(), nil)
		}
		if errors.Is(err, apperror.ErrNotFound) {
			return response.Error(c, http.StatusNotFound, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "order fetched successfully", res)
}

func (h *OrderHandler) UpdateStatus(c *echo.Context) error {
	userIDStr := c.Get("user_id").(string)
	userID, _ := uuid.Parse(userIDStr)
	role := c.Get("role").(string)

	idParam := c.Param("id")
	orderID, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid order id format", nil)
	}

	var req dto.UpdateOrderStatusRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", []string{err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "validation failed", validator.FormatErrors(err))
	}

	if err := h.orderUC.UpdateStatus(c.Request().Context(), userID, role, orderID, &req); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return response.Error(c, http.StatusNotFound, err.Error(), nil)
		}
		if errors.Is(err, apperror.ErrForbidden) {
			return response.Error(c, http.StatusForbidden, err.Error(), nil)
		}
		if errors.Is(err, apperror.ErrBadRequest) {
			return response.Error(c, http.StatusBadRequest, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "order status updated successfully", nil)
}
