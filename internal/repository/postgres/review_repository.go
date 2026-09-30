package postgres

import (
	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
	"gorm.io/gorm"
)

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) repository.ReviewRepository {
	return &reviewRepository{db}
}

func (r *reviewRepository) Create(review *entity.Review) error {
	return r.db.Create(review).Error
}

func (r *reviewRepository) FindByProductID(productID uuid.UUID) ([]entity.Review, error) {
	var reviews []entity.Review
	err := r.db.Preload("User").Where("product_id = ?", productID).Order("created_at DESC").Find(&reviews).Error
	return reviews, err
}

func (r *reviewRepository) FindByID(id uuid.UUID) (*entity.Review, error) {
	var review entity.Review
	err := r.db.First(&review, "id = ?", id).Error
	return &review, err
}

func (r *reviewRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.Review{}, "id = ?", id).Error
}
