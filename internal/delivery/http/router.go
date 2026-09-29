package http

import (
	"github.com/labstack/echo/v5"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/handler"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/middleware"
)

func RegisterRoutes(e *echo.Echo, jwtSecret string, authHandler *handler.AuthHandler) {
	v1 := e.Group("/api/v1")

	// --- Auth Routes ---
	auth := v1.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)
	auth.POST("/logout", authHandler.Logout)

	// Protected Auth Routes
	auth.GET("/me", authHandler.Me, middleware.Auth(jwtSecret))
}
