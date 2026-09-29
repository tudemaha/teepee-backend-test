package entity

import "github.com/google/uuid"

// ProductCategory is the join table for the Product ↔ Category many-to-many.
// Uses a composite PK (product_id, category_id) — no surrogate id.
type ProductCategory struct {
	ProductID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	CategoryID uuid.UUID `gorm:"type:uuid;primaryKey"`
}
