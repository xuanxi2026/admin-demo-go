package model

import "time"

type DictItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DictType  string    `gorm:"size:64;not null;uniqueIndex:idx_dict_type_value" json:"dict_type"`
	Label     string    `gorm:"size:128;not null" json:"label"`
	Value     string    `gorm:"size:128;not null;uniqueIndex:idx_dict_type_value" json:"value"`
	Status    string    `gorm:"size:16;not null;default:enabled" json:"status"`
	Sort      int       `gorm:"not null;default:0" json:"sort"`
	Remark    string    `gorm:"size:255" json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SystemConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ConfigKey   string    `gorm:"size:64;not null;uniqueIndex" json:"config_key"`
	ConfigValue string    `gorm:"type:text;not null" json:"config_value"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Group       string    `gorm:"size:64;not null;index" json:"group"`
	ValueType   string    `gorm:"size:32;not null;default:string" json:"value_type"`
	Remark      string    `gorm:"size:255" json:"remark"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type OperationLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Module    string    `gorm:"size:64;not null;index" json:"module"`
	Action    string    `gorm:"size:64;not null" json:"action"`
	Operator  string    `gorm:"size:64;index" json:"operator"`
	Target    string    `gorm:"size:128" json:"target"`
	RequestID string    `gorm:"size:64;index" json:"request_id"`
	IP        string    `gorm:"size:64" json:"ip"`
	Detail    string    `gorm:"size:255" json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

type Department struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ParentID  uint      `gorm:"index;default:0" json:"parent_id"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	Code      string    `gorm:"size:64;not null;uniqueIndex" json:"code"`
	Leader    string    `gorm:"size:64" json:"leader"`
	Phone     string    `gorm:"size:32" json:"phone"`
	Status    string    `gorm:"size:16;not null;default:enabled" json:"status"`
	Sort      int       `gorm:"not null;default:0" json:"sort"`
	Remark    string    `gorm:"size:255" json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
