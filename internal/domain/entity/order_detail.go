package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderDetail struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Quantity  int       `gorm:"not null"`
	UnitPrice int64     `gorm:"not null"`
	Subtotal  int64     `gorm:"not null"`
	OrderID   uuid.UUID `gorm:"type:uuid;not null"`
	ProductID uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt time.Time ``
	UpdatedAt time.Time ``
	Product   Product   `gorm:"foreignKey:ProductID"`
}

func (od *OrderDetail) BeforeCreate(tx *gorm.DB) error {
	if od.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		od.ID = id
	}
	return nil
}
