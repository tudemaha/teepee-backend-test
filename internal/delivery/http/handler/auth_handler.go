package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"github.com/tudemaha/marketplace-be/internal/usecase"
	"github.com/tudemaha/marketplace-be/pkg/response"
	"github.com/tudemaha/marketplace-be/pkg/validator"
)

type AuthHandler struct {
	authUC usecase.AuthUseCase
}

func NewAuthHandler(authUC usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUC: authUC,
	}
}

func (h *AuthHandler) Register(c *echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", []string{err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "validation failed", validator.FormatErrors(err))
	}

	res, err := h.authUC.Register(&req)
	if err != nil {
		if err.Error() == "email already in use" {
			return response.Error(c, http.StatusConflict, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, http.StatusCreated, "user registered successfully", res)
}

func (h *AuthHandler) Login(c *echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", []string{err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "validation failed", validator.FormatErrors(err))
	}

	res, err := h.authUC.Login(&req)
	if err != nil {
		return response.Error(c, http.StatusUnauthorized, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "login successful", res)
}

func (h *AuthHandler) Refresh(c *echo.Context) error {
	var req dto.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", []string{err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "validation failed", validator.FormatErrors(err))
	}

	res, err := h.authUC.Refresh(&req)
	if err != nil {
		return response.Error(c, http.StatusUnauthorized, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "token refreshed successfully", res)
}

func (h *AuthHandler) Logout(c *echo.Context) error {
	var req dto.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", []string{err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "validation failed", validator.FormatErrors(err))
	}

	if err := h.authUC.Logout(req.RefreshToken); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to logout", []string{err.Error()})
	}

	return response.Success(c, http.StatusOK, "logout successful", nil)
}

func (h *AuthHandler) Me(c *echo.Context) error {
	// Extract user_id from context (set by JWT middleware)
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.Error(c, http.StatusUnauthorized, "invalid user ID format", []string{err.Error()})
	}

	res, err := h.authUC.GetProfile(userID)
	if err != nil {
		return response.Error(c, http.StatusNotFound, err.Error(), nil)
	}

	return response.Success(c, http.StatusOK, "profile fetched successfully", res)
}
