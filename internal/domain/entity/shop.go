package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Shop struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Name           string         `gorm:"not null"`
	Address        string         ``
	ProfilePicture string         ``
	OwnerID        uuid.UUID      `gorm:"type:uuid;not null"`
	CreatedAt      time.Time      ``
	UpdatedAt      time.Time      ``
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	Owner          User           `gorm:"foreignKey:OwnerID"`
	Products       []Product      `gorm:"foreignKey:ShopID"`
}

func (s *Shop) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		s.ID = id
	}
	return nil
}
