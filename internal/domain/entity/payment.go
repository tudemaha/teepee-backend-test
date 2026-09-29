package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentMethod string
type PaymentStatus string

const (
	PaymentMethodCOD          PaymentMethod = "cod"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodQRIS         PaymentMethod = "qris"

	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusFailed  PaymentStatus = "failed"
)

type Payment struct {
	ID        uuid.UUID     `gorm:"type:uuid;primaryKey"`
	OrderID   uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex"`
	Amount    float64       `gorm:"not null"`
	Method    PaymentMethod `gorm:"type:varchar(30)"`
	Status    PaymentStatus `gorm:"type:varchar(20);default:'pending'"`
	PaidAt    *time.Time    ``
	CreatedAt time.Time     ``
	Order     Order         `gorm:"foreignKey:OrderID"`
}

func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		p.ID = id
	}
	return nil
}
