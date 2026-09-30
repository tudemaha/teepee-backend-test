package postgres

import (
	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
	"gorm.io/gorm"
)

type cartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) repository.CartRepository {
	return &cartRepository{db}
}

func (r *cartRepository) FindActiveByUserID(userID uuid.UUID) ([]entity.Cart, error) {
	var carts []entity.Cart
	err := r.db.Preload("Product").
		Preload("Product.Shop").
		Preload("Product.Images").
		Where("user_id = ? AND is_checked_out = ?", userID, false).
		Find(&carts).Error
	return carts, err
}

func (r *cartRepository) FindByUserAndProduct(userID uuid.UUID, productID uuid.UUID) (*entity.Cart, error) {
	var c entity.Cart
	err := r.db.Where("user_id = ? AND product_id = ? AND is_checked_out = ?", userID, productID, false).First(&c).Error
	return &c, err
}

func (r *cartRepository) FindByID(id uuid.UUID) (*entity.Cart, error) {
	var c entity.Cart
	err := r.db.Preload("Product").First(&c, "id = ?", id).Error
	return &c, err
}

func (r *cartRepository) Create(cart *entity.Cart) error {
	return r.db.Create(cart).Error
}

func (r *cartRepository) Update(cart *entity.Cart) error {
	return r.db.Save(cart).Error
}

func (r *cartRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.Cart{}, "id = ?", id).Error
}
