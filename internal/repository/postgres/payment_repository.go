package postgres

import "context"

import (
	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
	"gorm.io/gorm"
)

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) repository.PaymentRepository {
	return &paymentRepository{db}
}

func (r *paymentRepository) Create(payment *entity.Payment) error {
	return r.db.Create(payment).Error
}

func (r *paymentRepository) FindByOrderID(orderID uuid.UUID) (*entity.Payment, error) {
	var p entity.Payment
	err := r.db.Where("order_id = ?", orderID).First(&p).Error
	return &p, err
}

func (r *paymentRepository) FindByID(id uuid.UUID) (*entity.Payment, error) {
	var p entity.Payment
	err := r.db.First(&p, "id = ?", id).Error
	return &p, err
}

func (r *paymentRepository) Update(ctx context.Context, payment *entity.Payment) error {
	return ExtractDB(ctx, r.db).Save(payment).Error
}


