package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db}
}

func (r *userRepository) Create(u *entity.User) error {
	return r.db.Create(u).Error
}

func (r *userRepository) FindByID(id uuid.UUID) (*entity.User, error) {
	var u entity.User
	err := r.db.First(&u, "id = ?", id).Error
	return &u, err
}

func (r *userRepository) FindByEmail(email string) (*entity.User, error) {
	var u entity.User
	err := r.db.Where("email = ?", email).First(&u).Error
	return &u, err
}

func (r *userRepository) Update(ctx context.Context, u *entity.User) error {
	return ExtractDB(ctx, r.db).Save(u).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.User{}, "id = ?", id).Error
}
