package models

import "time"

type Order struct {
	ID              uint        `gorm:"primaryKey" json:"id"`
	UserID          uint        `gorm:"not null;index" json:"user_id"`
	TotalAmount     float64     `gorm:"not null" json:"total_amount"`
	Status          string      `gorm:"size:50;default:pending;index" json:"status"`
	ShippingAddress string      `gorm:"type:text;not null" json:"shipping_address"`
	Notes           string      `gorm:"type:text" json:"notes"`
	PaymentMethod   string      `gorm:"size:50;not null" json:"payment_method"`
	PaymentStatus   string      `gorm:"size:50;default:pending;index" json:"payment_status"`
	VANumber        *string     `gorm:"size:100" json:"va_number"`
	GopayDeeplink   *string     `gorm:"size:500" json:"gopay_deeplink"`
	PaidAt          *time.Time  `json:"paid_at"`
	Items           []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
	Payment         *Payment    `gorm:"foreignKey:OrderID" json:"payment,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type CheckoutRequest struct {
	ShippingAddress string `json:"shipping_address" binding:"required"`
	Notes           string `json:"notes"`
	PaymentMethod   string `json:"payment_method" binding:"required"`
}

type OrderResponse struct {
	ID              uint                `json:"id"`
	TotalAmount     float64             `json:"total_amount"`
	Status          string              `json:"status"`
	ShippingAddress string              `json:"shipping_address"`
	Notes           string              `json:"notes"`
	PaymentMethod   string              `json:"payment_method"`
	PaymentStatus   string              `json:"payment_status"`
	VANumber        *string             `json:"va_number"`
	GopayDeeplink   *string             `json:"gopay_deeplink"`
	PaidAt          *time.Time          `json:"paid_at"`
	Items           []OrderItemResponse `json:"items"`
	CreatedAt       time.Time           `json:"created_at"`
}

type OrdersListResponse struct {
	Items []OrderResponse `json:"items"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
	Total int64           `json:"total"`
}
