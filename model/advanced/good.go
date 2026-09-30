package model

import "time"

// Good 积分商城商品模型
// 结构体取单数 Good：gorm 默认命名策略（snake_case + 复数）即映射 goods 表；
// 仍按项目惯例显式提供 TableName() 兜底
type Good struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Description *string   `gorm:"column:description;type:text" json:"description,omitempty"`
	ImageURL    *string   `gorm:"column:image_url;type:varchar(500)" json:"image_url,omitempty"`
	Price       int64     `gorm:"column:price;type:int;not null" json:"price"`
	Stock       int64     `gorm:"column:stock;type:int;not null;default:0" json:"stock"`
	SortOrder   int       `gorm:"column:sort_order;type:int;not null;default:0" json:"sort_order"`
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	// 逻辑删除（下架）
	IsDeleted int8 `gorm:"column:is_deleted;type:tinyint;not null;default:0" json:"-"`
}

// TableName 指定表名
func (Good) TableName() string {
	return "goods"
}

// CreateGoodRequest 创建商品请求体（管理员）
// price 不做 binding 必填，缺失/非法由 service 返回 11004，错误码更精确
type CreateGoodRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	Price       int64   `json:"price"`
	Stock       *int64  `json:"stock,omitempty"`      // 缺省 0
	SortOrder   *int    `json:"sort_order,omitempty"` // 缺省 0
}

// UpdateGoodRequest 更新商品请求体（id 必填，其余字段增量更新，非整体替换）
type UpdateGoodRequest struct {
	ID          int64   `json:"id" binding:"required"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	Price       *int64  `json:"price,omitempty"`
	Stock       *int64  `json:"stock,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
}

// DeleteGoodRequest 删除（下架）商品请求体
type DeleteGoodRequest struct {
	ID int64 `json:"id" binding:"required"`
}

// ListGoodQuery 商品列表查询条件（GET 参数）
// min_price/max_price：积分闭区间筛选（含边界），0 或不传表示该侧不限；
// 负数由 binding 拦截，min>max 由 service 校验返回参数错误
type ListGoodQuery struct {
	Keyword  string `form:"keyword,omitempty"`
	MinPrice int64  `form:"min_price,omitempty" binding:"omitempty,min=0"`
	MaxPrice int64  `form:"max_price,omitempty" binding:"omitempty,min=0"`
	Page     int    `form:"page,omitempty" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size,omitempty" binding:"omitempty,min=1,max=100"`
}

// GoodListResponse 商品分页列表响应
type GoodListResponse struct {
	Total    int64  `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Items    []Good `json:"items"`
}
