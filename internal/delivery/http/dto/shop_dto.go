package dto

import "github.com/google/uuid"

type CreateShopRequest struct {
	Name           string `json:"name" validate:"required,min=2"`
	Address        string `json:"address" validate:"required"`
	ProfilePicture string `json:"profile_picture" validate:"omitempty,url"`
}

type UpdateShopRequest struct {
	Name           string `json:"name" validate:"required,min=2"`
	Address        string `json:"address" validate:"required"`
	ProfilePicture string `json:"profile_picture" validate:"omitempty,url"`
}

type ShopResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Address        string    `json:"address"`
	ProfilePicture string    `json:"profile_picture"`
	OwnerID        uuid.UUID `json:"owner_id"`
}
