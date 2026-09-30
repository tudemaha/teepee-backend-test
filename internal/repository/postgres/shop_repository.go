package postgres

import "context"

import (

	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
	"gorm.io/gorm"
)

type shopRepository struct {
	db *gorm.DB
}

func NewShopRepository(db *gorm.DB) repository.ShopRepository {
	return &shopRepository{db}
}

func (r *shopRepository) Create(ctx context.Context, shop *entity.Shop) error {
	return ExtractDB(ctx, r.db).Create(shop).Error
}

func (r *shopRepository) FindByID(id uuid.UUID) (*entity.Shop, error) {
	var s entity.Shop
	err := r.db.First(&s, "id = ?", id).Error
	return &s, err
}

func (r *shopRepository) FindByOwnerID(ownerID uuid.UUID) (*entity.Shop, error) {
	var s entity.Shop
	err := r.db.Where("owner_id = ?", ownerID).First(&s).Error
	return &s, err
}

func (r *shopRepository) Update(shop *entity.Shop) error {
	return r.db.Save(shop).Error
}
