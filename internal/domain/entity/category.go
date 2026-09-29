package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Category is flat (no parent_id). Nesting can be added later if needed.
type Category struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name string    `gorm:"not null"`
	Slug string    `gorm:"uniqueIndex;not null"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		c.ID = id
	}
	return nil
}
