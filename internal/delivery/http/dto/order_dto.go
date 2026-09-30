package dto

import (
	"time"

	"github.com/google/uuid"
)

type CheckoutRequest struct {
	ShippingAddress string      `json:"shipping_address" validate:"required"`
	ProductID       *uuid.UUID  `json:"product_id" validate:"omitempty"`
	Quantity        *int        `json:"quantity" validate:"omitempty,gt=0"`
	CartItemIDs     []uuid.UUID `json:"cart_item_ids" validate:"omitempty"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending confirmed shipped delivered completed cancelled"`
}

type OrderDetailResponse struct {
	ID        uuid.UUID       `json:"id"`
	ProductID uuid.UUID       `json:"product_id"`
	Product   ProductResponse `json:"product"`
	Quantity  int             `json:"quantity"`
	UnitPrice int64           `json:"unit_price"`
	Subtotal  int64           `json:"subtotal"`
}

type OrderResponse struct {
	ID              uuid.UUID             `json:"id"`
	Status          string                `json:"status"`
	TotalAmount     int64                 `json:"total_amount"`
	ShippingAddress string                `json:"shipping_address"`
	CreatedAt       time.Time             `json:"created_at"`
	Details         []OrderDetailResponse `json:"details,omitempty"`
}
