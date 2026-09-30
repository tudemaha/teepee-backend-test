package usecase

import (
	"errors"

	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
	"github.com/tudemaha/marketplace-be/pkg/utils"
)

type CartUseCase interface {
	GetMyCart(userID uuid.UUID) (*dto.CartResponse, error)
	AddToCart(userID uuid.UUID, req *dto.AddToCartRequest) error
	UpdateCartItem(userID uuid.UUID, cartID uuid.UUID, req *dto.UpdateCartRequest) error
	RemoveFromCart(userID uuid.UUID, cartID uuid.UUID) error
}

type cartUseCase struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

func NewCartUseCase(cartRepo repository.CartRepository, productRepo repository.ProductRepository) CartUseCase {
	return &cartUseCase{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (u *cartUseCase) GetMyCart(userID uuid.UUID) (*dto.CartResponse, error) {
	carts, err := u.cartRepo.FindActiveByUserID(userID)
	if err != nil {
		return nil, errors.New("failed to fetch cart")
	}

	var total float64
	var items []dto.CartItemResponse

	for _, c := range carts {
		subtotal := float64(c.Quantity) * c.Product.Price
		total += subtotal

		items = append(items, dto.CartItemResponse{
			ID:        c.ID,
			ProductID: c.ProductID,
			Product:   *utils.MapToProductResponse(&c.Product), // Reuse helper from product_usecase.go (needs to be public or in same package, wait! I will just inline a simplified version or map it here)
			Quantity:  c.Quantity,
			Subtotal:  subtotal,
		})
	}

	if items == nil {
		items = []dto.CartItemResponse{}
	}

	return &dto.CartResponse{
		Items:      items,
		TotalPrice: total,
	}, nil
}

func (u *cartUseCase) AddToCart(userID uuid.UUID, req *dto.AddToCartRequest) error {
	product, err := u.productRepo.FindByID(req.ProductID)
	if err != nil || !product.Available {
		return errors.New("product not found or unavailable")
	}

	existingCart, err := u.cartRepo.FindByUserAndProduct(userID, req.ProductID)
	if err == nil {
		newQty := existingCart.Quantity + req.Quantity
		if newQty > product.Stock {
			return errors.New("insufficient stock")
		}
		existingCart.Quantity = newQty
		return u.cartRepo.Update(existingCart)
	}

	if req.Quantity > product.Stock {
		return errors.New("insufficient stock")
	}

	newCart := &entity.Cart{
		Quantity:     req.Quantity,
		IsCheckedOut: false,
		ProductID:    req.ProductID,
		UserID:       userID,
	}

	return u.cartRepo.Create(newCart)
}

func (u *cartUseCase) UpdateCartItem(userID uuid.UUID, cartID uuid.UUID, req *dto.UpdateCartRequest) error {
	cart, err := u.cartRepo.FindByID(cartID)
	if err != nil || cart.UserID != userID || cart.IsCheckedOut {
		return errors.New("cart item not found")
	}

	product, err := u.productRepo.FindByID(cart.ProductID)
	if err != nil || !product.Available {
		return errors.New("product unavailable")
	}

	if req.Quantity > product.Stock {
		return errors.New("insufficient stock")
	}

	cart.Quantity = req.Quantity
	return u.cartRepo.Update(cart)
}

func (u *cartUseCase) RemoveFromCart(userID uuid.UUID, cartID uuid.UUID) error {
	cart, err := u.cartRepo.FindByID(cartID)
	if err != nil || cart.UserID != userID || cart.IsCheckedOut {
		return errors.New("cart item not found")
	}

	return u.cartRepo.Delete(cartID)
}
