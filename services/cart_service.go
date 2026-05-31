package services

import (
	"errors"

	"github.com/dwiilhammaulana/gin-firebase-backend/models"
	"github.com/dwiilhammaulana/gin-firebase-backend/repositories"
	"gorm.io/gorm"
)

var (
	ErrProductNotFound  = errors.New("produk tidak ditemukan")
	ErrCartItemNotFound = errors.New("item keranjang tidak ditemukan")
	ErrStockNotEnough   = errors.New("stock tidak cukup")
)

type CartService struct {
	cartRepo *repositories.CartRepository
}

func NewCartService() *CartService {
	return &CartService{cartRepo: repositories.NewCartRepository()}
}

func (s *CartService) GetCart(userID uint) (*models.CartResponse, error) {
	items, err := s.cartRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	return buildCartResponse(items), nil
}

func (s *CartService) AddToCart(userID uint, req *models.AddToCartRequest) error {
	product, err := s.cartRepo.FindProductByID(req.ProductID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrProductNotFound
	}
	if err != nil {
		return err
	}

	existingItem, err := s.cartRepo.FindByUserAndProduct(userID, req.ProductID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	newQuantity := req.Quantity
	if existingItem != nil && existingItem.ID != 0 {
		newQuantity += existingItem.Quantity
	}
	if product.Stock < newQuantity {
		return ErrStockNotEnough
	}

	if existingItem != nil && existingItem.ID != 0 {
		existingItem.Quantity = newQuantity
		return s.cartRepo.Update(existingItem)
	}

	item := &models.CartItem{
		UserID:    userID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	}
	return s.cartRepo.Create(item)
}

func (s *CartService) UpdateItem(userID, itemID uint, req *models.UpdateCartItemRequest) error {
	item, err := s.cartRepo.FindByIDAndUserID(itemID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrCartItemNotFound
	}
	if err != nil {
		return err
	}
	if item.Product.Stock < req.Quantity {
		return ErrStockNotEnough
	}

	item.Quantity = req.Quantity
	return s.cartRepo.Update(item)
}

func (s *CartService) RemoveItem(userID, itemID uint) error {
	if _, err := s.cartRepo.FindByIDAndUserID(itemID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCartItemNotFound
		}
		return err
	}
	return s.cartRepo.Delete(itemID, userID)
}

func (s *CartService) ClearCart(userID uint) error {
	return s.cartRepo.ClearByUserID(userID)
}

func buildCartResponse(items []models.CartItem) *models.CartResponse {
	response := &models.CartResponse{
		Items: []models.CartItemResponse{},
	}

	for _, item := range items {
		subtotal := item.Product.Price * float64(item.Quantity)
		response.Items = append(response.Items, models.CartItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Product: models.CartProductResponse{
				ID:       item.Product.ID,
				Name:     item.Product.Name,
				Price:    item.Product.Price,
				ImageURL: item.Product.ImageURL,
				Category: item.Product.Category,
			},
			Quantity: item.Quantity,
			Subtotal: subtotal,
		})
		response.Total += subtotal
		response.ItemCount += item.Quantity
	}

	return response
}
