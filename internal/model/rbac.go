package model

import "time"

type Role struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Code      string    `gorm:"size:32;uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Permission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Code      string    `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	Type      string    `gorm:"size:16;not null;default:api" json:"type"` // api/button/menu
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Menu struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ParentID       uint      `gorm:"index;default:0" json:"parent_id"`
	Path           string    `gorm:"size:128;not null" json:"path"`
	Name           string    `gorm:"size:64" json:"name"`
	Component      string    `gorm:"size:255" json:"component"`
	Redirect       string    `gorm:"size:128" json:"redirect"`
	Title          string    `gorm:"size:64;not null" json:"title"`
	Icon           string    `gorm:"size:64" json:"icon"`
	Badge          string    `gorm:"size:32" json:"badge"`
	PermissionCode string    `gorm:"size:64" json:"permission_code"`
	AlwaysShow     bool      `json:"always_show"`
	Affix          bool      `json:"affix"`
	NoKeepAlive    bool      `json:"no_keep_alive"`
	Hidden         bool      `json:"hidden"`
	Sort           int       `gorm:"default:0" json:"sort"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UserRole struct {
	UserID    uint      `gorm:"primaryKey;autoIncrement:false" json:"user_id"`
	RoleID    uint      `gorm:"primaryKey;autoIncrement:false" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

type RolePermission struct {
	RoleID       uint      `gorm:"primaryKey;autoIncrement:false" json:"role_id"`
	PermissionID uint      `gorm:"primaryKey;autoIncrement:false" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type RoleMenu struct {
	RoleID    uint      `gorm:"primaryKey;autoIncrement:false" json:"role_id"`
	MenuID    uint      `gorm:"primaryKey;autoIncrement:false" json:"menu_id"`
	CreatedAt time.Time `json:"created_at"`
}
