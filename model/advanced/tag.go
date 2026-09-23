package model

import "time"

type Tag struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(50);not null;uniqueIndex:uk_name;" json:"name"`
	Color     *string   `gorm:"column:color;type:varchar(20);" json:"color,omitempty"`
	SortOrder int       `gorm:"column:sort_order;not null;default:0;index:idx_sort_order;" json:"sort_order"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;autoCreateTime;" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;autoUpdateTime;" json:"updated_at"`
}

func (Tag) TableName() string {
	return "tags"
}

// ItemTag 物品标签关联模型
type ItemTag struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	ItemID    int64     `gorm:"column:item_id;type:bigint;not null;uniqueIndex:uk_item_tag;comment:物品ID" json:"item_id"`
	TagID     int64     `gorm:"column:tag_id;type:bigint;not null;uniqueIndex:uk_item_tag;index:idx_tag_item;comment:标签ID" json:"tag_id"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
}

// TableName 指定表名
func (ItemTag) TableName() string {
	return "item_tags"
}

// CreateTagRequest 创建标签请求体
type CreateTagRequest struct {
	Name      string  `json:"name" binding:"required"`
	Color     *string `json:"color,omitempty"`
	SortOrder int     `json:"sort_order,omitempty"`
}

// UpdateTagRequest 更新标签请求体（id 必填，其余字段增量更新，非整体替换）
type UpdateTagRequest struct {
	ID        int64   `json:"id" binding:"required"`
	Name      *string `json:"name,omitempty"`
	Color     *string `json:"color,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty"`
}
