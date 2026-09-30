package repository

import (
	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
)

type PaymentRepository interface {
	Create(payment *entity.Payment) error
	FindByOrderID(orderID uuid.UUID) (*entity.Payment, error)
	FindByID(id uuid.UUID) (*entity.Payment, error)
	Update(payment *entity.Payment) error
	UpdatePaymentAndOrderStatus(payment *entity.Payment, order *entity.Order) error
}
