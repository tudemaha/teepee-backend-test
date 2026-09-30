package dto

import "github.com/google/uuid"

type CreateProductRequest struct {
	Name        string      `json:"name" validate:"required,min=2"`
	Description string      `json:"description" validate:"omitempty"`
	Price       int64       `json:"price" validate:"required,gt=0"`
	Stock       int         `json:"stock" validate:"gte=0"`
	CategoryIDs []uuid.UUID `json:"category_ids" validate:"required,min=1"`
	ImageURLs   []string    `json:"image_urls" validate:"omitempty"`
}

type UpdateProductRequest struct {
	Name        string      `json:"name" validate:"required,min=2"`
	Description string      `json:"description" validate:"omitempty"`
	Price       int64       `json:"price" validate:"required,gt=0"`
	Stock       int         `json:"stock" validate:"gte=0"`
	CategoryIDs []uuid.UUID `json:"category_ids" validate:"required,min=1"`
}

type ProductImageResponse struct {
	ID        uuid.UUID `json:"id"`
	URL       string    `json:"url"`
	IsPrimary bool      `json:"is_primary"`
}

type ProductResponse struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Price       int64                  `json:"price"`
	Stock       int                    `json:"stock"`
	Status      string                 `json:"status"`
	ShopID      uuid.UUID              `json:"shop_id"`
	ShopName    string                 `json:"shop_name,omitempty"`
	Categories  []CategoryResponse     `json:"categories"`
	Images      []ProductImageResponse `json:"images"`
}

type ProductListResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Price    int64     `json:"price"`
	Stock    int       `json:"stock"`
	Status   string    `json:"status"`
	ShopName string    `json:"shop_name,omitempty"`
	Image    string    `json:"image,omitempty"`
}

type ProductListFilter struct {
	ShopID     uuid.UUID `query:"shop_id"`
	CategoryID uuid.UUID `query:"category_id"`
	Search     string    `query:"search"`
	Limit      int       `query:"limit"`
	Offset     int       `query:"offset"`
}

type UpdateProductStockRequest struct {
	Stock int `json:"stock" validate:"required,min=0"`
}

type UpdateProductAvailabilityRequest struct {
	Available *bool `json:"available" validate:"required"`
}
