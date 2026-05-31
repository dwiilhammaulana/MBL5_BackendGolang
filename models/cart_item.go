package models

import "time"

type CartItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	ProductID uint      `gorm:"not null;index" json:"product_id"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"product"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AddToCartRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

type CartProductResponse struct {
	ID       uint    `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	ImageURL string  `json:"image_url"`
	Category string  `json:"category"`
}

type CartItemResponse struct {
	ID        uint                `json:"id"`
	ProductID uint                `json:"product_id"`
	Product   CartProductResponse `json:"product"`
	Quantity  int                 `json:"quantity"`
	Subtotal  float64             `json:"subtotal"`
}

type CartResponse struct {
	Items     []CartItemResponse `json:"items"`
	Total     float64            `json:"total"`
	ItemCount int                `json:"item_count"`
}
