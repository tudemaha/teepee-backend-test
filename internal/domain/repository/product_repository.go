package repository

import "context"

import (
	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
)

type ProductRepository interface {
	Create(product *entity.Product) error
	FindByID(id uuid.UUID) (*entity.Product, error)
	FindAll(filter dto.ProductListFilter) ([]entity.Product, error)
	Update(product *entity.Product, newCategoryIDs []uuid.UUID) error
	SoftDelete(id uuid.UUID) error
	ReduceStock(ctx context.Context, productID uuid.UUID, quantity int) error
	IncreaseStock(ctx context.Context, productID uuid.UUID, quantity int) error
}
