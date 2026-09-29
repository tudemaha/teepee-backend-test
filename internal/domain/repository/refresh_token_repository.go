package repository

import (
	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
)

type RefreshTokenRepository interface {
	Create(token *entity.RefreshToken) error
	FindByToken(tokenStr string) (*entity.RefreshToken, error)
	DeleteByToken(tokenStr string) error
	DeleteByUserID(userID uuid.UUID) error
}
