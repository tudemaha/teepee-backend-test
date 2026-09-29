package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Review struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_product"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_product"`
	Rating    int       `gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Comment   string    ``
	CreatedAt time.Time ``
	User      User      `gorm:"foreignKey:UserID"`
	Product   Product   `gorm:"foreignKey:ProductID"`
}

func (r *Review) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		r.ID = id
	}
	return nil
}
