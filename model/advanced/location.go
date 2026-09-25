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
// parent_id 允许为 0（根节点），sort_order 允许为 0；level 不由请求方指定，创建时按父级计算
// parent_id / sort_order 可省略（omitempty）：省略时按默认值处理，parent_id=0 视为根节点、sort_order=0
type CreateLocationRequest struct {
	Name      string  `json:"name" binding:"required"`
	ParentID  *int64  `json:"parent_id" binding:"omitempty,gte=0"`
	Address   *string `json:"address,omitempty"`
	SortOrder *int    `json:"sort_order" binding:"omitempty,gte=0"`
}

// UpdateLocationRequest 更新地点请求体（id 必填，其余字段增量更新，非整体替换）
type UpdateLocationRequest struct {
	ID        int64   `json:"id" binding:"required"`
	Name      *string `json:"name,omitempty"`
	ParentID  *int64  `json:"parent_id,omitempty" binding:"omitempty,gte=0"`
	Address   *string `json:"address,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty"`
}

// ListLocationRequest 地点列表查询条件（GET 参数，均可选）
type ListLocationRequest struct {
	ParentID *int64 `form:"parent_id,omitempty"`
	Level    *int   `form:"level,omitempty"`
}
