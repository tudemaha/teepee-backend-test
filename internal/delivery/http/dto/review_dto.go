package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateReviewRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Rating    int       `json:"rating" validate:"required,min=1,max=5"`
	Comment   string    `json:"comment" validate:"omitempty,max=1000"`
}

type ReviewResponse struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	UserID    uuid.UUID `json:"user_id"`
	UserName  string    `json:"user_name"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type ProductReviewsResponse struct {
	AverageRating float64          `json:"average_rating"`
	TotalReviews  int              `json:"total_reviews"`
	Reviews       []ReviewResponse `json:"reviews"`
}
