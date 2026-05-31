package models

import "time"

type Payment struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	OrderID       uint       `gorm:"not null;index" json:"order_id"`
	PaymentMethod string     `gorm:"size:50;not null" json:"payment_method"`
	PaymentStatus string     `gorm:"size:50;default:pending;index" json:"payment_status"`
	Amount        float64    `gorm:"not null" json:"amount"`
	VANumber      *string    `gorm:"size:100" json:"va_number"`
	GopayDeeplink *string    `gorm:"size:500" json:"gopay_deeplink"`
	PaidAt        *time.Time `json:"paid_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
