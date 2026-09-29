package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductImage struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	URL       string    `gorm:"not null"`
	IsPrimary bool      `gorm:"default:false"`
	ProductID uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt time.Time ``
	UpdatedAt time.Time ``
}

func (pi *ProductImage) BeforeCreate(tx *gorm.DB) error {
	if pi.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		pi.ID = id
	}
	return nil
}
