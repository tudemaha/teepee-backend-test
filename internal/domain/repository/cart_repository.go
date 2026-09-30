package repository

import (
	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
)

type CartRepository interface {
	FindActiveByUserID(userID uuid.UUID) ([]entity.Cart, error)
	FindByUserAndProduct(userID uuid.UUID, productID uuid.UUID) (*entity.Cart, error)
	FindByID(id uuid.UUID) (*entity.Cart, error)
	Create(cart *entity.Cart) error
	Update(cart *entity.Cart) error
	Delete(id uuid.UUID) error
}
