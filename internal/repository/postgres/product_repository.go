package postgres

import (
	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
	"gorm.io/gorm"
)

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) repository.ProductRepository {
	return &productRepository{db}
}

func (r *productRepository) Create(product *entity.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) FindByID(id uuid.UUID) (*entity.Product, error) {
	var p entity.Product
	err := r.db.Preload("Categories").
		Preload("Images").
		Preload("Shop").
		First(&p, "id = ?", id).Error
	return &p, err
}

func (r *productRepository) FindAll(filter dto.ProductListFilter) ([]entity.Product, error) {
	var products []entity.Product
	query := r.db.Model(&entity.Product{}).
		Preload("Categories").
		Preload("Images").
		Preload("Shop").
		Where("available = ?", true)

	if filter.ShopID != uuid.Nil {
		query = query.Where("shop_id = ?", filter.ShopID)
	}

	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
	}

	if filter.CategoryID != uuid.Nil {
		query = query.Joins("JOIN product_categories pc ON pc.product_id = products.id").
			Where("pc.category_id = ?", filter.CategoryID)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err := query.Find(&products).Error
	return products, err
}

func (r *productRepository) Update(product *entity.Product, newCategoryIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(product).Error; err != nil {
			return err
		}

		if newCategoryIDs != nil {
			var newCategories []entity.Category
			if err := tx.Where("id IN ?", newCategoryIDs).Find(&newCategories).Error; err != nil {
				return err
			}
			if err := tx.Model(product).Association("Categories").Replace(newCategories); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *productRepository) SoftDelete(id uuid.UUID) error {
	return r.db.Model(&entity.Product{}).Where("id = ?", id).Update("available", false).Error
}
