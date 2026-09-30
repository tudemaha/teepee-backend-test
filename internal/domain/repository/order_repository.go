package repository

import (
	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
)

type OrderRepository interface {
	Create(order *entity.Order, cartIDs []uuid.UUID) error
	FindByBuyerID(buyerID uuid.UUID) ([]entity.Order, error)
	FindByID(id uuid.UUID) (*entity.Order, error)
	Update(order *entity.Order) error
	CancelOrderAndRollbackStock(order *entity.Order) error
}
