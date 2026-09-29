package repository

import (
	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
)

type CategoryRepository interface {
	Create(category *entity.Category) error
	FindAll() ([]entity.Category, error)
	FindByID(id uuid.UUID) (*entity.Category, error)
	FindBySlug(slug string) (*entity.Category, error)
	Delete(id uuid.UUID) error
}
