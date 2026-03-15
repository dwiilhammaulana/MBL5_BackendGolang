package models

import "gorm.io/gorm"

// User adalah model yang mapping ke tabel "users" di MySQL
// GORM otomatis plural nama struct -> nama tabel: User -> users
type User struct {
	gorm.Model // Embed: ID, CreatedAt, UpdatedAt, DeletedAt (soft delete)
	FirebaseUID   string `gorm:"uniqueIndex;size:128;not null" json:"firebase_uid"`
	Email         string `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Name          string `gorm:"size:100" json:"name"`
	Role          string `gorm:"size:20;default:user" json:"role"`
	EmailVerified bool   `gorm:"default:false" json:"email_verified"`
	LastLoginAt   *int64 `gorm:"index" json:"last_login_at,omitempty"`
}