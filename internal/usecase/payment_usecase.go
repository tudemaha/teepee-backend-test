package usecase

import (
	"errors"
	"fmt"
	"github.com/tudemaha/marketplace-be/pkg/apperror"
	"time"

	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
)

type PaymentUseCase interface {
	CreatePayment(buyerID uuid.UUID, req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error)
	GetPaymentByOrderID(buyerID uuid.UUID, orderID uuid.UUID) (*dto.PaymentResponse, error)
	UpdatePaymentStatus(adminID uuid.UUID, paymentID uuid.UUID, req *dto.UpdatePaymentStatusRequest) error
}

type paymentUseCase struct {
	paymentRepo repository.PaymentRepository
	orderRepo   repository.OrderRepository
	userRepo    repository.UserRepository
}

func NewPaymentUseCase(paymentRepo repository.PaymentRepository, orderRepo repository.OrderRepository, userRepo repository.UserRepository) PaymentUseCase {
	return &paymentUseCase{
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
		userRepo:    userRepo,
	}
}

func (u *paymentUseCase) CreatePayment(buyerID uuid.UUID, req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error) {
	order, err := u.orderRepo.FindByID(req.OrderID)
	if err != nil || order.BuyerID != buyerID {
		return nil, fmt.Errorf("%w: %s", apperror.ErrNotFound, "order not found or unauthorized")
	}

	if order.Status != entity.OrderStatusPending {
		return nil, fmt.Errorf("%w: %s", apperror.ErrBadRequest, "order is no longer pending")
	}

	if existing, err := u.paymentRepo.FindByOrderID(req.OrderID); err == nil {
		return mapToPaymentResponse(existing), nil
	}

	payment := &entity.Payment{
		OrderID: req.OrderID,
		Amount:  order.TotalAmount,
		Method:  entity.PaymentMethod(req.Method),
		Status:  entity.PaymentStatusPending,
	}

	if err := u.paymentRepo.Create(payment); err != nil {
		return nil, errors.New("failed to create payment")
	}

	return mapToPaymentResponse(payment), nil
}

func (u *paymentUseCase) GetPaymentByOrderID(buyerID uuid.UUID, orderID uuid.UUID) (*dto.PaymentResponse, error) {
	order, err := u.orderRepo.FindByID(orderID)
	if err != nil || order.BuyerID != buyerID {
		return nil, fmt.Errorf("%w: %s", apperror.ErrNotFound, "order not found or unauthorized")
	}

	payment, err := u.paymentRepo.FindByOrderID(orderID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", apperror.ErrNotFound, "payment not found")
	}

	return mapToPaymentResponse(payment), nil
}

func (u *paymentUseCase) UpdatePaymentStatus(adminID uuid.UUID, paymentID uuid.UUID, req *dto.UpdatePaymentStatusRequest) error {
	adminUser, err := u.userRepo.FindByID(adminID)
	if err != nil || adminUser.Role != entity.RoleAdmin {
		return fmt.Errorf("%w: %s", apperror.ErrForbidden, "unauthorized")
	}

	payment, err := u.paymentRepo.FindByID(paymentID)
	if err != nil {
		return fmt.Errorf("%w: %s", apperror.ErrNotFound, "payment not found")
	}

	order, err := u.orderRepo.FindByID(payment.OrderID)
	if err != nil {
		return fmt.Errorf("%w: %s", apperror.ErrNotFound, "order not found")
	}

	newStatus := entity.PaymentStatus(req.Status)
	if payment.Status == newStatus {
		return nil
	}

	payment.Status = newStatus
	switch newStatus {
	case entity.PaymentStatusPaid:
		now := time.Now()
		payment.PaidAt = &now
		order.Status = entity.OrderStatusConfirmed
	}

	return u.paymentRepo.UpdatePaymentAndOrderStatus(payment, order)
}

func mapToPaymentResponse(p *entity.Payment) *dto.PaymentResponse {
	return &dto.PaymentResponse{
		ID:        p.ID,
		OrderID:   p.OrderID,
		Amount:    p.Amount,
		Method:    string(p.Method),
		Status:    string(p.Status),
		PaidAt:    p.PaidAt,
		CreatedAt: p.CreatedAt,
	}
}
