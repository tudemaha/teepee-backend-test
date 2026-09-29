package response

import "github.com/labstack/echo/v5"

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Success sends a standardized JSON success response
func Success(c *echo.Context, statusCode int, message string, data any) error {
	return c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error sends a standardized JSON error response
func Error(c *echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
	})
}
