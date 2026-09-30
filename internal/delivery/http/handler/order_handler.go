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

	res, err := h.orderUC.Checkout(userID, &req)
	if err != nil {
		if err.Error() == "cart is empty" || err.Error() == "insufficient stock for one or more items during checkout" {
			return response.Error(c, http.StatusBadRequest, err.Error(), nil)
		}
		if len(err.Error()) > 18 && err.Error()[:18] == "insufficient stock" {
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
		if err.Error() == "order not found" {
			return response.Error(c, http.StatusNotFound, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "order fetched successfully", res)
}
