package postgres

import "context"

import (

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

func (r *orderRepository) Create(ctx context.Context, order *entity.Order) error {
	return ExtractDB(ctx, r.db).Create(order).Error
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

func (r *orderRepository) Update(ctx context.Context, order *entity.Order) error {
	return ExtractDB(ctx, r.db).Save(order).Error
}



func (r *orderRepository) HasCompletedOrderWithProduct(buyerID uuid.UUID, productID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&entity.Order{}).
		Joins("JOIN order_details ON order_details.order_id = orders.id").
		Where("orders.buyer_id = ? AND orders.status = ? AND order_details.product_id = ?", buyerID, entity.OrderStatusCompleted, productID).
		Count(&count).Error
	return count > 0, err
}
