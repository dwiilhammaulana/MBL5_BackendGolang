package models

import "time"

type OrderItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	OrderID     uint      `gorm:"not null;index" json:"order_id"`
	ProductID   uint      `gorm:"not null;index" json:"product_id"`
	ProductName string    `gorm:"size:200;not null" json:"product_name"`
	Price       float64   `gorm:"not null" json:"price"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	Subtotal    float64   `gorm:"not null" json:"subtotal"`
	CreatedAt   time.Time `json:"created_at"`
}

type OrderItemResponse struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}
