package postgres

import (
	"errors"

	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) repository.OrderRepository {
	return &orderRepository{db}
}

func (r *orderRepository) Create(order *entity.Order, cartIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		for _, detail := range order.Details {
			result := tx.Model(&entity.Product{}).
				Where("id = ? AND stock >= ?", detail.ProductID, detail.Quantity).
				UpdateColumn("stock", gorm.Expr("stock - ?", detail.Quantity))

			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errors.New("insufficient stock for one or more items during checkout")
			}
		}

		if len(cartIDs) > 0 {
			if err := tx.Model(&entity.Cart{}).
				Where("id IN ?", cartIDs).
				Update("is_checked_out", true).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *orderRepository) FindByBuyerID(buyerID uuid.UUID) ([]entity.Order, error) {
	var orders []entity.Order
	err := r.db.Where("buyer_id = ?", buyerID).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

func (r *orderRepository) FindByID(id uuid.UUID) (*entity.Order, error) {
	var o entity.Order
	err := r.db.Preload("Details").
		Preload("Details.Product").
		Preload("Details.Product.Shop").
		Preload("Details.Product.Images").
		First(&o, "id = ?", id).Error
	return &o, err
}

func (r *orderRepository) Update(order *entity.Order) error {
	return r.db.Save(order).Error
}

func (r *orderRepository) CancelOrderAndRollbackStock(order *entity.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(order).Update("status", entity.OrderStatusCancelled).Error; err != nil {
			return err
		}

		for _, detail := range order.Details {
			if err := tx.Model(&entity.Product{}).
				Where("id = ?", detail.ProductID).
				UpdateColumn("stock", gorm.Expr("stock + ?", detail.Quantity)).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *orderRepository) HasCompletedOrderWithProduct(buyerID uuid.UUID, productID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&entity.Order{}).
		Joins("JOIN order_details ON order_details.order_id = orders.id").
		Where("orders.buyer_id = ? AND orders.status = ? AND order_details.product_id = ?", buyerID, entity.OrderStatusCompleted, productID).
		Count(&count).Error
	return count > 0, err
}
