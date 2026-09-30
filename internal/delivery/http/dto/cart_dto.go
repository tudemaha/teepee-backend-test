package dto

import "github.com/google/uuid"

type AddToCartRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,gt=0"`
}

type UpdateCartRequest struct {
	Quantity int `json:"quantity" validate:"required,gt=0"`
}

type CartItemResponse struct {
	ID        uuid.UUID       `json:"id"`
	ProductID uuid.UUID       `json:"product_id"`
	Product   ProductResponse `json:"product"`
	Quantity  int             `json:"quantity"`
	Subtotal  int64           `json:"subtotal"`
}

type CartResponse struct {
	Items      []CartItemResponse `json:"items"`
	TotalPrice int64              `json:"total_price"`
}
