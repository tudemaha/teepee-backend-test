package repository

import (

	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
)

type UserRepository interface {
	Create(user *entity.User) error
	FindByID(id uuid.UUID) (*entity.User, error)
	FindByEmail(email string) (*entity.User, error)
	Update(user *entity.User) error
	Delete(id uuid.UUID) error
}
