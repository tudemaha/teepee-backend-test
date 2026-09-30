package usecase

import (
	"errors"
	"fmt"

	"github.com/tudemaha/marketplace-be/pkg/apperror"

	"context"

	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
	"golang.org/x/sync/errgroup"
)

type ReviewUseCase interface {
	CreateReview(userID uuid.UUID, req *dto.CreateReviewRequest) (*dto.ReviewResponse, error)
	GetProductReviews(productID uuid.UUID) (*dto.ProductReviewsResponse, error)
	DeleteReview(userID uuid.UUID, reviewID uuid.UUID) error
}

type reviewUseCase struct {
	reviewRepo  repository.ReviewRepository
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
}

func NewReviewUseCase(reviewRepo repository.ReviewRepository, orderRepo repository.OrderRepository, productRepo repository.ProductRepository) ReviewUseCase {
	return &reviewUseCase{
		reviewRepo:  reviewRepo,
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

func (u *reviewUseCase) CreateReview(userID uuid.UUID, req *dto.CreateReviewRequest) (*dto.ReviewResponse, error) {

	var hasCompleted bool

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		var err error
		_, err = u.productRepo.FindByID(req.ProductID)
		if err != nil {
			return fmt.Errorf("%w: %s", apperror.ErrNotFound, "product not found")
		}
		return nil
	})

	g.Go(func() error {
		var err error
		hasCompleted, err = u.orderRepo.HasCompletedOrderWithProduct(userID, req.ProductID)
		if err != nil || !hasCompleted {
			return fmt.Errorf("%w: %s", apperror.ErrForbidden, "user has no completed order for this product")
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	review := &entity.Review{
		UserID:    userID,
		ProductID: req.ProductID,
		Rating:    req.Rating,
		Comment:   req.Comment,
	}

	if err := u.reviewRepo.Create(review); err != nil {
		return nil, fmt.Errorf("%w: %s", apperror.ErrConflict, "failed to create review, user may have already reviewed this product")
	}

	return &dto.ReviewResponse{
		ID:        review.ID,
		ProductID: review.ProductID,
		UserID:    review.UserID,
		Rating:    review.Rating,
		Comment:   review.Comment,
		CreatedAt: review.CreatedAt,
	}, nil
}

func (u *reviewUseCase) GetProductReviews(productID uuid.UUID) (*dto.ProductReviewsResponse, error) {
	reviews, err := u.reviewRepo.FindByProductID(productID)
	if err != nil {
		return nil, errors.New("failed to fetch reviews")
	}

	var totalRating int
	var res []dto.ReviewResponse

	for _, r := range reviews {
		totalRating += r.Rating
		res = append(res, dto.ReviewResponse{
			ID:        r.ID,
			ProductID: r.ProductID,
			UserID:    r.UserID,
			UserName:  r.User.Name,
			Rating:    r.Rating,
			Comment:   r.Comment,
			CreatedAt: r.CreatedAt,
		})
	}

	if res == nil {
		res = []dto.ReviewResponse{}
	}

	var avg float64
	if len(reviews) > 0 {
		avg = float64(totalRating) / float64(len(reviews))
	}

	return &dto.ProductReviewsResponse{
		AverageRating: avg,
		TotalReviews:  len(reviews),
		Reviews:       res,
	}, nil
}

func (u *reviewUseCase) DeleteReview(userID uuid.UUID, reviewID uuid.UUID) error {
	review, err := u.reviewRepo.FindByID(reviewID)
	if err != nil {
		return errors.New("review not found")
	}

	if review.UserID != userID {
		return fmt.Errorf("%w: %s", apperror.ErrForbidden, "unauthorized to delete this review")
	}

	return u.reviewRepo.Delete(reviewID)
}
