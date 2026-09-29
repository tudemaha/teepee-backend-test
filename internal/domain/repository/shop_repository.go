package repository

import (

	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
)

type ShopRepository interface {
	Create(shop *entity.Shop) error
	FindByID(id uuid.UUID) (*entity.Shop, error)
	FindByOwnerID(ownerID uuid.UUID) (*entity.Shop, error)
	Update(shop *entity.Shop) error
}
