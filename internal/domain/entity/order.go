package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID              uuid.UUID     `gorm:"type:uuid;primaryKey"`
	BuyerID         uuid.UUID     `gorm:"type:uuid;not null"`
	Status          OrderStatus   `gorm:"type:varchar(20);default:'pending'"`
	TotalAmount     float64       `gorm:"not null"`
	ShippingAddress string        `gorm:"not null"`
	CreatedAt       time.Time     ``
	UpdatedAt       time.Time     ``
	Buyer           User          `gorm:"foreignKey:BuyerID"`
	Details         []OrderDetail `gorm:"foreignKey:OrderID"`
	Payment         *Payment      `gorm:"foreignKey:OrderID"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		o.ID = id
	}
	return nil
}
