package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductStatus string

const (
	ProductStatusAvailable    ProductStatus = "available"
	ProductStatusOutOfStock   ProductStatus = "out_of_stock"
	ProductStatusNotAvailable ProductStatus = "not_available"
)

type Product struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Name        string         `gorm:"not null"`
	Description string         ``
	Price       int64          `gorm:"not null;check:price > 0"`
	Stock       int            `gorm:"not null;default:0;check:stock >= 0"`
	Available   bool           `gorm:"default:true"`
	ShopID      uuid.UUID      `gorm:"type:uuid;not null"`
	CreatedAt   time.Time      ``
	UpdatedAt   time.Time      ``
	Status      ProductStatus  `gorm:"-"`
	Shop        Shop           `gorm:"foreignKey:ShopID"`
	Images      []ProductImage `gorm:"foreignKey:ProductID"`
	Categories  []Category     `gorm:"many2many:product_categories;"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		p.ID = id
	}
	return nil
}

func (p *Product) AfterFind(tx *gorm.DB) error {
	p.Status = p.computeStatus()
	return nil
}

func (p *Product) computeStatus() ProductStatus {
	if !p.Available {
		return ProductStatusNotAvailable
	}
	if p.Stock == 0 {
		return ProductStatusOutOfStock
	}
	return ProductStatusAvailable
}
