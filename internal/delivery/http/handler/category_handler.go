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

type CategoryHandler struct {
	categoryUC usecase.CategoryUseCase
}

func NewCategoryHandler(g *echo.Group, categoryUC usecase.CategoryUseCase, jwtSecret string) {
	h := &CategoryHandler{
		categoryUC: categoryUC,
	}

	cat := g.Group("/categories")

	cat.GET("", h.GetAll)

	cat.POST("", h.Create, middleware.Auth(jwtSecret), middleware.RequireRole("admin"))
	cat.DELETE("/:id", h.Delete, middleware.Auth(jwtSecret), middleware.RequireRole("admin"))
}

func (h *CategoryHandler) GetAll(c *echo.Context) error {
	res, err := h.categoryUC.GetAll()
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}
	return response.Success(c, http.StatusOK, "categories fetched successfully", res)
}

func (h *CategoryHandler) Create(c *echo.Context) error {
	var req dto.CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", []string{err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "validation failed", validator.FormatErrors(err))
	}

	res, err := h.categoryUC.Create(&req)
	if err != nil {
		if err.Error() == "category with similar name already exists" {
			return response.Error(c, http.StatusConflict, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusCreated, "category created successfully", res)
}

func (h *CategoryHandler) Delete(c *echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid category id format", nil)
	}

	if err := h.categoryUC.Delete(id); err != nil {
		if err.Error() == "category not found" {
			return response.Error(c, http.StatusNotFound, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "category deleted successfully", nil)
}
