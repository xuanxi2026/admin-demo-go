package model

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:32;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Phone        string    `gorm:"size:20" json:"phone"`
	Email        string    `gorm:"size:128" json:"email"`
	Nickname     string    `gorm:"size:64" json:"nickname"`
	Avatar       string    `gorm:"size:255" json:"avatar"`
	Bio          string    `gorm:"size:255" json:"bio"`
	Role         string    `gorm:"size:16;not null;default:editor" json:"role"`
	GoogleSecret string    `gorm:"size:64" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
