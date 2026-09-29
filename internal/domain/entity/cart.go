package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Cart struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Quantity     int       `gorm:"not null;default:1"`
	IsCheckedOut bool      `gorm:"default:false"`
	ProductID    uuid.UUID `gorm:"type:uuid;not null"`
	UserID       uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt    time.Time ``
	UpdatedAt    time.Time ``
	Product      Product   `gorm:"foreignKey:ProductID"`
	User         User      `gorm:"foreignKey:UserID"`
}

func (c *Cart) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		c.ID = id
	}
	return nil
}
