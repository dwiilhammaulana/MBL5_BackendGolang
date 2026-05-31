package repositories

import (
	"github.com/dwiilhammaulana/gin-firebase-backend/config"
	"github.com/dwiilhammaulana/gin-firebase-backend/models"
	"gorm.io/gorm"
)

type CartRepository struct{}

func NewCartRepository() *CartRepository {
	return &CartRepository{}
}

func (r *CartRepository) FindByUserID(userID uint) ([]models.CartItem, error) {
	var items []models.CartItem
	err := config.DB.Preload("Product").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&items).Error
	return items, err
}

func (r *CartRepository) FindByUserAndProduct(userID, productID uint) (*models.CartItem, error) {
	var item models.CartItem
	err := config.DB.Where("user_id = ? AND product_id = ?", userID, productID).First(&item).Error
	return &item, err
}

func (r *CartRepository) FindByIDAndUserID(id, userID uint) (*models.CartItem, error) {
	var item models.CartItem
	err := config.DB.Preload("Product").
		Where("id = ? AND user_id = ?", id, userID).
		First(&item).Error
	return &item, err
}

func (r *CartRepository) Create(item *models.CartItem) error {
	return config.DB.Create(item).Error
}

func (r *CartRepository) Update(item *models.CartItem) error {
	return config.DB.Save(item).Error
}

func (r *CartRepository) Delete(id, userID uint) error {
	return config.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.CartItem{}).Error
}

func (r *CartRepository) ClearByUserID(userID uint) error {
	return config.DB.Where("user_id = ?", userID).Delete(&models.CartItem{}).Error
}

func (r *CartRepository) FindProductByID(productID uint) (*models.Product, error) {
	var product models.Product
	err := config.DB.Where("id = ? AND is_active = ?", productID, true).First(&product).Error
	return &product, err
}

func (r *CartRepository) WithTransaction(tx *gorm.DB) *CartRepositoryTx {
	return &CartRepositoryTx{tx: tx}
}

type CartRepositoryTx struct {
	tx *gorm.DB
}

func (r *CartRepositoryTx) FindByUserID(userID uint) ([]models.CartItem, error) {
	var items []models.CartItem
	err := r.tx.Preload("Product").
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&items).Error
	return items, err
}

func (r *CartRepositoryTx) ClearByUserID(userID uint) error {
	return r.tx.Where("user_id = ?", userID).Delete(&models.CartItem{}).Error
}
