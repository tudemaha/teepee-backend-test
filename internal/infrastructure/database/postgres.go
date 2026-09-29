package database

import (
	"log"

	"github.com/tudemaha/marketplace-be/config"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	log.Println("Connected to PostgreSQL, running auto-migrations...")

	err = db.AutoMigrate(
		&entity.User{},
		&entity.RefreshToken{},
		&entity.Shop{},
		&entity.Category{},
		&entity.Product{},
		&entity.ProductImage{},
		&entity.Cart{},
		&entity.Order{},
		&entity.OrderDetail{},
		&entity.Payment{},
		&entity.Review{},
	)
	if err != nil {
		return nil, err
	}

	log.Println("Database auto-migration completed successfully.")
	return db, nil
}
