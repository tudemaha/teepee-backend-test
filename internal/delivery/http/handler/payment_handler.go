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

type PaymentHandler struct {
	paymentUC usecase.PaymentUseCase
}

func NewPaymentHandler(g *echo.Group, paymentUC usecase.PaymentUseCase, jwtSecret string) {
	h := &PaymentHandler{
		paymentUC: paymentUC,
	}

	paymentGroup := g.Group("/payments")
	paymentGroup.Use(middleware.Auth(jwtSecret))

	paymentGroup.POST("", h.CreatePayment)
	paymentGroup.GET("/:orderId", h.GetPaymentByOrderID)
	paymentGroup.PATCH("/:id/status", h.UpdatePaymentStatus)
}

func (h *PaymentHandler) CreatePayment(c *echo.Context) error {
	buyerIDStr := c.Get("user_id").(string)
	buyerID, _ := uuid.Parse(buyerIDStr)

	var req dto.CreatePaymentRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", []string{err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "validation failed", validator.FormatErrors(err))
	}

	res, err := h.paymentUC.CreatePayment(buyerID, &req)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return response.Error(c, http.StatusNotFound, err.Error(), nil)
		}
		if errors.Is(err, apperror.ErrBadRequest) {
			return response.Error(c, http.StatusBadRequest, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusCreated, "payment created successfully", res)
}

func (h *PaymentHandler) GetPaymentByOrderID(c *echo.Context) error {
	buyerIDStr := c.Get("user_id").(string)
	buyerID, _ := uuid.Parse(buyerIDStr)

	idParam := c.Param("orderId")
	orderID, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid order id format", nil)
	}

	res, err := h.paymentUC.GetPaymentByOrderID(buyerID, orderID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return response.Error(c, http.StatusNotFound, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "payment fetched successfully", res)
}

func (h *PaymentHandler) UpdatePaymentStatus(c *echo.Context) error {
	adminIDStr := c.Get("user_id").(string)
	adminID, _ := uuid.Parse(adminIDStr)

	idParam := c.Param("id")
	paymentID, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid payment id format", nil)
	}

	var req dto.UpdatePaymentStatusRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", []string{err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "validation failed", validator.FormatErrors(err))
	}

	if err := h.paymentUC.UpdatePaymentStatus(adminID, paymentID, &req); err != nil {
		if errors.Is(err, apperror.ErrForbidden) {
			return response.Error(c, http.StatusForbidden, err.Error(), nil)
		}
		if errors.Is(err, apperror.ErrNotFound) {
			return response.Error(c, http.StatusNotFound, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "payment status updated successfully", nil)
}
