package model

import "time"

// Location 地点模型
type Location struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(100);not null" json:"name"`
	ParentID  int64     `gorm:"column:parent_id;type:bigint;not null" json:"parent_id"`
	Level     int       `gorm:"column:level;type:int;not null" json:"level"`
	Address   *string   `gorm:"column:address;type:varchar(255)" json:"address,omitempty"`
	SortOrder int       `gorm:"column:sort_order;type:int;not null;default:0" json:"sort_order"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	// 硬删除
}

// TableName 指定表名
func (Location) TableName() string {
	return "locations"
}

// CreateLocationRequest 创建地点请求体
type CreateLocationRequest struct {
	Name      string  `json:"name" binding:"required"`
	ParentID  int64   `json:"parent_id" binding:"required"`
	Address   *string `json:"address,omitempty"`
	SortOrder int     `json:"sort_order" binding:"required"`
}
