package middleware

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v5"
	"github.com/labstack/echo/v5"
	jwtpkg "github.com/tudemaha/marketplace-be/pkg/jwt"
	"github.com/tudemaha/marketplace-be/pkg/response"
)

// Auth uses the official echo-jwt middleware to validate tokens
func Auth(secret string) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(secret),
		// Tell echo-jwt to parse into our CustomClaims struct
		NewClaimsFunc: func(c *echo.Context) jwt.Claims {
			return new(jwtpkg.CustomClaims)
		},
		// Custom error response to match our API standards
		ErrorHandler: func(c *echo.Context, err error) error {
			return response.Error(c, http.StatusUnauthorized, "invalid or missing token")
		},
		// After successful validation, populate the context for downstream handlers
		SuccessHandler: func(c *echo.Context) error {
			userToken, ok := c.Get("user").(*jwt.Token)
			if ok {
				if claims, ok := userToken.Claims.(*jwtpkg.CustomClaims); ok {
					c.Set("user_id", claims.UserID.String())
					c.Set("role", claims.Role)
				}
			}
			return nil
		},
	})
}

// RequireRole is a middleware that restricts access to specific user roles.
// It must be used after the Auth middleware.
func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			userRole, ok := c.Get("role").(string)
			if !ok {
				return response.Error(c, http.StatusForbidden, "role information missing")
			}

			for _, role := range roles {
				if role == userRole {
					return next(c)
				}
			}

			return response.Error(c, http.StatusForbidden, "you do not have permission to access this resource")
		}
	}
}
