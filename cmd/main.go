package main

import (
	"log"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/tudemaha/marketplace-be/config"
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
	txManager := postgres.NewTxManager(db)
	userRepo := postgres.NewUserRepository(db)
	rtRepo := postgres.NewRefreshTokenRepository(db)
	categoryRepo := postgres.NewCategoryRepository(db)
	shopRepo := postgres.NewShopRepository(db)
	productRepo := postgres.NewProductRepository(db)
	cartRepo := postgres.NewCartRepository(db)
	orderRepo := postgres.NewOrderRepository(db)
	paymentRepo := postgres.NewPaymentRepository(db)
	reviewRepo := postgres.NewReviewRepository(db)

	// 4. Initialize UseCases
	authUseCase := usecase.NewAuthUseCase(userRepo, rtRepo, cfg.JWT.SecretKey)
	categoryUseCase := usecase.NewCategoryUseCase(categoryRepo)
	shopUseCase := usecase.NewShopUseCase(txManager, shopRepo, userRepo)
	productUseCase := usecase.NewProductUseCase(productRepo, shopRepo, categoryRepo)
	cartUseCase := usecase.NewCartUseCase(cartRepo, productRepo)
	orderUseCase := usecase.NewOrderUseCase(txManager, orderRepo, cartRepo, productRepo, userRepo)
	paymentUseCase := usecase.NewPaymentUseCase(txManager, paymentRepo, orderRepo, userRepo)
	reviewUseCase := usecase.NewReviewUseCase(reviewRepo, orderRepo, productRepo)

	// 5. Initialize Echo Framework
	e := echo.New()
	e.Validator = validator.New()

	// Global Middleware
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.CORS.AllowOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	// Simple health check route
	e.GET("/health", func(c *echo.Context) error {
		return c.String(200, "OK")
	})

	// 6. Initialize Handlers & Routes
	v1 := e.Group("/api/v1")
	handler.NewAuthHandler(v1, authUseCase, cfg.JWT.SecretKey)
	handler.NewCategoryHandler(v1, categoryUseCase, cfg.JWT.SecretKey)
	handler.NewShopHandler(v1, shopUseCase, cfg.JWT.SecretKey)
	handler.NewProductHandler(v1, productUseCase, cfg.JWT.SecretKey)
	handler.NewCartHandler(v1, cartUseCase, cfg.JWT.SecretKey)
	handler.NewOrderHandler(v1, orderUseCase, cfg.JWT.SecretKey)
	handler.NewPaymentHandler(v1, paymentUseCase, cfg.JWT.SecretKey)
	handler.NewReviewHandler(v1, reviewUseCase, cfg.JWT.SecretKey)

	// 7. Start Server
	log.Printf("Server starting on port %s in %s mode...", cfg.App.Port, cfg.App.Env)
	if err := e.Start(":" + cfg.App.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
