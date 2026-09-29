package postgres

import (
	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
	"gorm.io/gorm"
)

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) repository.CategoryRepository {
	return &categoryRepository{db}
}

func (r *categoryRepository) Create(c *entity.Category) error {
	return r.db.Create(c).Error
}

func (r *categoryRepository) FindAll() ([]entity.Category, error) {
	var categories []entity.Category
	err := r.db.Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) FindByID(id uuid.UUID) (*entity.Category, error) {
	var c entity.Category
	err := r.db.First(&c, "id = ?", id).Error
	return &c, err
}

func (r *categoryRepository) FindBySlug(slug string) (*entity.Category, error) {
	var c entity.Category
	err := r.db.Where("slug = ?", slug).First(&c).Error
	return &c, err
}

func (r *categoryRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.Category{}, "id = ?", id).Error
}
