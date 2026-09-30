package usecase

import (
	"errors"

	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
	"github.com/tudemaha/marketplace-be/pkg/utils"
)

type OrderUseCase interface {
	Checkout(buyerID uuid.UUID, req *dto.CheckoutRequest) (*dto.OrderResponse, error)
	GetMyOrders(buyerID uuid.UUID) ([]dto.OrderResponse, error)
	GetByID(buyerID uuid.UUID, orderID uuid.UUID) (*dto.OrderResponse, error)
}

type orderUseCase struct {
	orderRepo   repository.OrderRepository
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

func NewOrderUseCase(orderRepo repository.OrderRepository, cartRepo repository.CartRepository, productRepo repository.ProductRepository) OrderUseCase {
	return &orderUseCase{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (u *orderUseCase) Checkout(buyerID uuid.UUID, req *dto.CheckoutRequest) (*dto.OrderResponse, error) {
	var totalAmount float64
	var orderDetails []entity.OrderDetail
	var cartIDs []uuid.UUID

	// directt checkout
	if req.ProductID != nil && req.Quantity != nil {
		product, err := u.productRepo.FindByID(*req.ProductID)
		if err != nil || !product.Available {
			return nil, errors.New("product not found or unavailable")
		}

		if *req.Quantity > product.Stock {
			return nil, errors.New("insufficient stock for " + product.Name)
		}

		subtotal := float64(*req.Quantity) * product.Price
		totalAmount = subtotal

		orderDetails = append(orderDetails, entity.OrderDetail{
			Quantity:  *req.Quantity,
			UnitPrice: product.Price,
			Subtotal:  subtotal,
			ProductID: product.ID,
			Product:   *product,
		})
	} else {
		// checkout from items in cart
		carts, err := u.cartRepo.FindActiveByUserID(buyerID)
		if err != nil || len(carts) == 0 {
			return nil, errors.New("cart is empty")
		}

		if len(req.CartItemIDs) > 0 {
			selectedMap := make(map[uuid.UUID]bool)
			for _, id := range req.CartItemIDs {
				selectedMap[id] = true
			}

			var filtered []entity.Cart
			for _, c := range carts {
				if selectedMap[c.ID] {
					filtered = append(filtered, c)
				}
			}

			if len(filtered) == 0 {
				return nil, errors.New("selected cart items are invalid or already checked out")
			}
			carts = filtered
		}

		for _, c := range carts {
			if c.Quantity > c.Product.Stock {
				return nil, errors.New("insufficient stock for " + c.Product.Name)
			}

			subtotal := float64(c.Quantity) * c.Product.Price
			totalAmount += subtotal

			orderDetails = append(orderDetails, entity.OrderDetail{
				Quantity:  c.Quantity,
				UnitPrice: c.Product.Price,
				Subtotal:  subtotal,
				ProductID: c.ProductID,
				Product:   c.Product,
			})

			cartIDs = append(cartIDs, c.ID)
		}
	}

	order := &entity.Order{
		BuyerID:         buyerID,
		Status:          entity.OrderStatusPending,
		TotalAmount:     totalAmount,
		ShippingAddress: req.ShippingAddress,
		Details:         orderDetails,
	}

	if err := u.orderRepo.Create(order, cartIDs); err != nil {
		if err.Error() == "insufficient stock for one or more items during checkout" {
			return nil, err
		}
		return nil, errors.New("failed to process checkout")
	}

	return mapToOrderResponse(order), nil
}

func (u *orderUseCase) GetMyOrders(buyerID uuid.UUID) ([]dto.OrderResponse, error) {
	orders, err := u.orderRepo.FindByBuyerID(buyerID)
	if err != nil {
		return nil, errors.New("failed to fetch orders")
	}

	var res []dto.OrderResponse
	for _, o := range orders {
		res = append(res, *mapToOrderSummaryResponse(&o))
	}

	if res == nil {
		res = []dto.OrderResponse{}
	}
	return res, nil
}

func (u *orderUseCase) GetByID(buyerID uuid.UUID, orderID uuid.UUID) (*dto.OrderResponse, error) {
	order, err := u.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	if order.BuyerID != buyerID {
		return nil, errors.New("unauthorized to view this order")
	}

	return mapToOrderResponse(order), nil
}

func mapToOrderResponse(o *entity.Order) *dto.OrderResponse {
	var details []dto.OrderDetailResponse
	for _, d := range o.Details {
		details = append(details, dto.OrderDetailResponse{
			ID:        d.ID,
			ProductID: d.ProductID,
			Product:   *utils.MapToProductResponse(&d.Product),
			Quantity:  d.Quantity,
			UnitPrice: d.UnitPrice,
			Subtotal:  d.Subtotal,
		})
	}

	return &dto.OrderResponse{
		ID:              o.ID,
		Status:          string(o.Status),
		TotalAmount:     o.TotalAmount,
		ShippingAddress: o.ShippingAddress,
		CreatedAt:       o.CreatedAt,
		Details:         details,
	}
}

func mapToOrderSummaryResponse(o *entity.Order) *dto.OrderResponse {
	return &dto.OrderResponse{
		ID:              o.ID,
		Status:          string(o.Status),
		TotalAmount:     o.TotalAmount,
		ShippingAddress: o.ShippingAddress,
		CreatedAt:       o.CreatedAt,
	}
}
