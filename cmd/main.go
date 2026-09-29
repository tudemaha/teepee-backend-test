package main

import (
	"log"

	"github.com/labstack/echo/v5"
	"github.com/tudemaha/marketplace-be/config"
	"github.com/tudemaha/marketplace-be/internal/delivery/http"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/handler"
	"github.com/tudemaha/marketplace-be/internal/infrastructure/database"
	"github.com/tudemaha/marketplace-be/internal/repository/postgres"
	"github.com/tudemaha/marketplace-be/internal/usecase"
	"github.com/tudemaha/marketplace-be/pkg/validator"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Connect to Database & AutoMigrate
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Initialize Repositories
	userRepo := postgres.NewUserRepository(db)
	rtRepo := postgres.NewRefreshTokenRepository(db)

	// 4. Initialize UseCases
	authUseCase := usecase.NewAuthUseCase(userRepo, rtRepo, cfg.JWT.SecretKey)

	// 5. Initialize Handlers
	authHandler := handler.NewAuthHandler(authUseCase)

	// 6. Initialize Echo Framework
	e := echo.New()
	e.Validator = validator.New()

	// Simple health check route
	e.GET("/health", func(c *echo.Context) error {
		return c.String(200, "OK")
	})

	// 7. Register Routes
	http.RegisterRoutes(e, cfg.JWT.SecretKey, authHandler)

	// 8. Start Server
	log.Printf("Server starting on port %s in %s mode...", cfg.App.Port, cfg.App.Env)
	if err := e.Start(":" + cfg.App.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
