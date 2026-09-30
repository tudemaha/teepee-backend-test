package repository

import (
	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
)

type ReviewRepository interface {
	Create(review *entity.Review) error
	FindByProductID(productID uuid.UUID) ([]entity.Review, error)
	FindByID(id uuid.UUID) (*entity.Review, error)
	Delete(id uuid.UUID) error
}
