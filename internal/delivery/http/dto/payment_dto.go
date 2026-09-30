package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreatePaymentRequest struct {
	OrderID uuid.UUID `json:"order_id" validate:"required"`
	Method  string    `json:"method" validate:"required,oneof=cod bank_transfer qris"`
}

type UpdatePaymentStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending paid failed"`
}

type PaymentResponse struct {
	ID        uuid.UUID  `json:"id"`
	OrderID   uuid.UUID  `json:"order_id"`
	Amount    int64      `json:"amount"`
	Method    string     `json:"method"`
	Status    string     `json:"status"`
	PaidAt    *time.Time `json:"paid_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
