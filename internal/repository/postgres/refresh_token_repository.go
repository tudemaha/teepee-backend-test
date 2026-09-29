package postgres

import (
	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
	"gorm.io/gorm"
)

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) repository.RefreshTokenRepository {
	return &refreshTokenRepository{db}
}

func (r *refreshTokenRepository) Create(token *entity.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *refreshTokenRepository) FindByToken(tokenStr string) (*entity.RefreshToken, error) {
	var rt entity.RefreshToken
	err := r.db.Where("token = ?", tokenStr).First(&rt).Error
	return &rt, err
}

func (r *refreshTokenRepository) DeleteByToken(tokenStr string) error {
	return r.db.Where("token = ?", tokenStr).Delete(&entity.RefreshToken{}).Error
}

func (r *refreshTokenRepository) DeleteByUserID(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&entity.RefreshToken{}).Error
}
