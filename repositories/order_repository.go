package repositories

import (
	"github.com/dwiilhammaulana/gin-firebase-backend/config"
	"github.com/dwiilhammaulana/gin-firebase-backend/models"
	"gorm.io/gorm"
)

type OrderRepository struct{}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{}
}

func (r *OrderRepository) FindByUserID(userID uint, page, limit int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := config.DB.Model(&models.Order{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Preload("Items").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&orders).Error

	return orders, total, err
}

func (r *OrderRepository) FindByIDAndUserID(id, userID uint) (*models.Order, error) {
	var order models.Order
	err := config.DB.Preload("Items").
		Where("id = ? AND user_id = ?", id, userID).
		First(&order).Error
	return &order, err
}

func (r *OrderRepository) WithTransaction(tx *gorm.DB) *OrderRepositoryTx {
	return &OrderRepositoryTx{tx: tx}
}

type OrderRepositoryTx struct {
	tx *gorm.DB
}

func (r *OrderRepositoryTx) Create(order *models.Order) error {
	return r.tx.Create(order).Error
}

func (r *OrderRepositoryTx) CreateOrderItems(items []models.OrderItem) error {
	return r.tx.Create(&items).Error
}

func (r *OrderRepositoryTx) CreatePayment(payment *models.Payment) error {
	return r.tx.Create(payment).Error
}

func (r *OrderRepositoryTx) DecreaseProductStock(productID uint, quantity int) (int64, error) {
	result := r.tx.Model(&models.Product{}).
		Where("id = ? AND stock >= ?", productID, quantity).
		UpdateColumn("stock", gorm.Expr("stock - ?", quantity))
	return result.RowsAffected, result.Error
}
