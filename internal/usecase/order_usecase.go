package usecase

import "context"

import (
	"errors"
	"fmt"

	"github.com/tudemaha/marketplace-be/pkg/apperror"

	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
	"github.com/tudemaha/marketplace-be/pkg/utils"
)

type OrderUseCase interface {
	Checkout(ctx context.Context, buyerID uuid.UUID, req *dto.CheckoutRequest) (*dto.OrderResponse, error)
	GetMyOrders(buyerID uuid.UUID) ([]dto.OrderResponse, error)
	GetByID(buyerID uuid.UUID, orderID uuid.UUID) (*dto.OrderResponse, error)
	UpdateStatus(ctx context.Context, userID uuid.UUID, role string, orderID uuid.UUID, req *dto.UpdateOrderStatusRequest) error
}

type orderUseCase struct {
	txManager   repository.TxManager
	orderRepo   repository.OrderRepository
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
	userRepo    repository.UserRepository
}

func NewOrderUseCase(txManager repository.TxManager, orderRepo repository.OrderRepository, cartRepo repository.CartRepository, productRepo repository.ProductRepository, userRepo repository.UserRepository) OrderUseCase {
	return &orderUseCase{
		txManager:   txManager,
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
		userRepo:    userRepo,
	}
}

func (u *orderUseCase) Checkout(ctx context.Context, buyerID uuid.UUID, req *dto.CheckoutRequest) (*dto.OrderResponse, error) {
	var totalAmount int64
	var orderDetails []entity.OrderDetail
	var cartIDs []uuid.UUID

	// directt checkout
	if req.ProductID != nil && req.Quantity != nil {
		product, err := u.productRepo.FindByID(*req.ProductID)
		if err != nil || !product.Available {
			return nil, fmt.Errorf("%w: %s", apperror.ErrNotFound, "product not found or unavailable")
		}

		if *req.Quantity > product.Stock {
			return nil, fmt.Errorf("%w: %s%s", apperror.ErrBadRequest, "insufficient stock for ", product.Name)
		}

		subtotal := int64(*req.Quantity) * product.Price
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
			return nil, fmt.Errorf("%w: %s", apperror.ErrBadRequest, "cart is empty")
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
				return nil, fmt.Errorf("%w: %s", apperror.ErrBadRequest, "selected cart items are invalid or already checked out")
			}
			carts = filtered
		}

		for _, c := range carts {
			if c.Quantity > c.Product.Stock {
				return nil, fmt.Errorf("%w: %s%s", apperror.ErrBadRequest, "insufficient stock for ", c.Product.Name)
			}

			subtotal := int64(c.Quantity) * c.Product.Price
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

	err := u.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.orderRepo.Create(txCtx, order); err != nil {
			return err
		}

		for _, detail := range order.Details {
			if err := u.productRepo.ReduceStock(txCtx, detail.ProductID, detail.Quantity); err != nil {
				return err
			}
		}

		if len(cartIDs) > 0 {
			if err := u.cartRepo.MarkCheckedOut(txCtx, cartIDs); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		if err.Error() == "insufficient stock" {
			return nil, fmt.Errorf("%w: %s", apperror.ErrBadRequest, "insufficient stock for one or more items during checkout")
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
		return nil, fmt.Errorf("%w: %s", apperror.ErrNotFound, "order not found")
	}

	if order.BuyerID != buyerID {
		return nil, fmt.Errorf("%w: %s", apperror.ErrForbidden, "unauthorized to view this order")
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

func (u *orderUseCase) UpdateStatus(ctx context.Context, userID uuid.UUID, role string, orderID uuid.UUID, req *dto.UpdateOrderStatusRequest) error {
	order, err := u.orderRepo.FindByID(orderID)
	if err != nil {
		return fmt.Errorf("%w: %s", apperror.ErrNotFound, "order not found")
	}

	newStatus := entity.OrderStatus(req.Status)
	if order.Status == newStatus {
		return nil
	}

	switch newStatus {
	case entity.OrderStatusCancelled:
		if role == string(entity.RoleBuyer) {
			if order.BuyerID != userID {
				return fmt.Errorf("%w: %s", apperror.ErrForbidden, "unauthorized to cancel this order")
			}
			if order.Status != entity.OrderStatusPending {
				return fmt.Errorf("%w: %s", apperror.ErrBadRequest, "only pending orders can be cancelled")
			}
		}
	case entity.OrderStatusConfirmed, entity.OrderStatusShipped, entity.OrderStatusDelivered:
		if role != string(entity.RoleAdmin) && role != string(entity.RoleSeller) {
			return fmt.Errorf("%w: %s", apperror.ErrForbidden, "unauthorized to update fulfillment status")
		}
	case entity.OrderStatusCompleted:
		if role == string(entity.RoleBuyer) && order.BuyerID != userID {
			return fmt.Errorf("%w: %s", apperror.ErrForbidden, "unauthorized to complete this order")
		}
	}

	if newStatus == entity.OrderStatusCancelled && order.Status != entity.OrderStatusCancelled {
		return u.txManager.RunInTx(ctx, func(txCtx context.Context) error {
			order.Status = newStatus
			if err := u.orderRepo.Update(txCtx, order); err != nil {
				return err
			}
			for _, detail := range order.Details {
				if err := u.productRepo.IncreaseStock(txCtx, detail.ProductID, detail.Quantity); err != nil {
					return err
				}
			}
			return nil
		})
	}

	order.Status = newStatus
	return u.orderRepo.Update(ctx, order)
}
