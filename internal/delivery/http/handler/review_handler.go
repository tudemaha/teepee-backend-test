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

type ReviewHandler struct {
	reviewUC usecase.ReviewUseCase
}

func NewReviewHandler(g *echo.Group, reviewUC usecase.ReviewUseCase, jwtSecret string) {
	h := &ReviewHandler{
		reviewUC: reviewUC,
	}

	g.GET("/products/:id/reviews", h.GetProductReviews)

	reviewGroup := g.Group("/reviews")
	reviewGroup.Use(middleware.Auth(jwtSecret))

	reviewGroup.POST("", h.CreateReview)
	reviewGroup.DELETE("/:id", h.DeleteReview)
}

func (h *ReviewHandler) CreateReview(c *echo.Context) error {
	userIDStr := c.Get("user_id").(string)
	userID, _ := uuid.Parse(userIDStr)

	var req dto.CreateReviewRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", []string{err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "validation failed", validator.FormatErrors(err))
	}

	res, err := h.reviewUC.CreateReview(userID, &req)
	if err != nil {
		if err.Error() == "product not found" {
			return response.Error(c, http.StatusNotFound, err.Error(), nil)
		}
		if err.Error() == "user has no completed order for this product" {
			return response.Error(c, http.StatusForbidden, err.Error(), nil)
		}
		if err.Error()[:23] == "failed to create review" {
			return response.Error(c, http.StatusConflict, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusCreated, "review submitted successfully", res)
}

func (h *ReviewHandler) GetProductReviews(c *echo.Context) error {
	idParam := c.Param("id")
	productID, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid product id format", nil)
	}

	res, err := h.reviewUC.GetProductReviews(productID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "reviews fetched successfully", res)
}

func (h *ReviewHandler) DeleteReview(c *echo.Context) error {
	userIDStr := c.Get("user_id").(string)
	userID, _ := uuid.Parse(userIDStr)

	idParam := c.Param("id")
	reviewID, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid review id format", nil)
	}

	if err := h.reviewUC.DeleteReview(userID, reviewID); err != nil {
		if err.Error() == "unauthorized to delete this review" {
			return response.Error(c, http.StatusForbidden, err.Error(), nil)
		}
		if err.Error() == "review not found" {
			return response.Error(c, http.StatusNotFound, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "review deleted successfully", nil)
}
