package tests

import (
	"log"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/tudemaha/marketplace-be/config"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/handler"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/repository/postgres"
	"github.com/tudemaha/marketplace-be/internal/usecase"
	"github.com/tudemaha/marketplace-be/pkg/validator"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetupTestServer() (*echo.Echo, *gorm.DB) {
	// Force environment to testing
	os.Setenv("APP_ENV", "testing")
	os.Setenv("JWT_SECRET", "super_secret_test_key_123")
	os.Setenv("DB_PASSWORD", "postgres")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Use SQLite in-memory for fast, isolated integration tests
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to sqlite in-memory db: %v", err)
	}

	// AutoMigrate all entities
	err = db.AutoMigrate(
		&entity.User{},
		&entity.RefreshToken{},
		&entity.Category{},
		&entity.Shop{},
		&entity.Product{},
		&entity.ProductImage{},
		&entity.Cart{},
		&entity.Order{},
		&entity.OrderDetail{},
		&entity.Payment{},
		&entity.Review{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate test DB: %v", err)
	}

	// Initialize Repositories
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

	// Initialize UseCases
	authUseCase := usecase.NewAuthUseCase(userRepo, rtRepo, cfg.JWT.SecretKey)
	categoryUseCase := usecase.NewCategoryUseCase(categoryRepo)
	shopUseCase := usecase.NewShopUseCase(txManager, shopRepo, userRepo)
	productUseCase := usecase.NewProductUseCase(productRepo, shopRepo, categoryRepo)
	cartUseCase := usecase.NewCartUseCase(cartRepo, productRepo)
	orderUseCase := usecase.NewOrderUseCase(txManager, orderRepo, cartRepo, productRepo, userRepo)
	paymentUseCase := usecase.NewPaymentUseCase(txManager, paymentRepo, orderRepo, userRepo)
	reviewUseCase := usecase.NewReviewUseCase(reviewRepo, orderRepo, productRepo)

	// Initialize Echo Framework
	e := echo.New()
	e.Validator = validator.New()

	// Global Middleware
	e.Use(middleware.Recover())

	// Initialize Handlers & Routes
	v1 := e.Group("/api/v1")
	handler.NewAuthHandler(v1, authUseCase, cfg.JWT.SecretKey)
	handler.NewCategoryHandler(v1, categoryUseCase, cfg.JWT.SecretKey)
	handler.NewShopHandler(v1, shopUseCase, cfg.JWT.SecretKey)
	handler.NewProductHandler(v1, productUseCase, cfg.JWT.SecretKey)
	handler.NewCartHandler(v1, cartUseCase, cfg.JWT.SecretKey)
	handler.NewOrderHandler(v1, orderUseCase, cfg.JWT.SecretKey)
	handler.NewPaymentHandler(v1, paymentUseCase, cfg.JWT.SecretKey)
	handler.NewReviewHandler(v1, reviewUseCase, cfg.JWT.SecretKey)

	return e, db
}
