package repository

import "context"

import (
	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
)

type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	FindByBuyerID(buyerID uuid.UUID) ([]entity.Order, error)
	FindByID(id uuid.UUID) (*entity.Order, error)
	Update(ctx context.Context, order *entity.Order) error
	HasCompletedOrderWithProduct(buyerID uuid.UUID, productID uuid.UUID) (bool, error)
}
