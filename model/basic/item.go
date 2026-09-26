package model

import (
	"time"

	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
)

// Item 物品实体模型
type Item struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID         int64      `gorm:"column:user_id;type:bigint;not null;index:idx_user_id" json:"user_id"`
	Title          string     `gorm:"column:title;type:varchar(100);not null" json:"title"`
	Description    string     `gorm:"column:description;type:text;not null" json:"description"`
	Type           int8       `gorm:"column:type;type:tinyint;not null;index:idx_type_status,priority:1" json:"type"`
	Status         int8       `gorm:"column:status;type:tinyint;not null;default:0;index:idx_type_status,priority:2" json:"status"`
	LocationID     *int64     `gorm:"column:location_id;type:bigint;index:idx_location_id" json:"location_id,omitempty"`
	LocationDetail *string    `gorm:"column:location_detail;type:varchar(200)" json:"location_detail,omitempty"`
	LostFoundTime  time.Time  `gorm:"column:lost_found_time;type:datetime;not null" json:"lost_found_time"`
	Contact        *string    `gorm:"column:contact;type:varchar(100)" json:"contact,omitempty"`
	CreditReward   int64      `gorm:"column:credit_reward;type:int;default:0" json:"credit_reward"`
	ViewCount      int64      `gorm:"column:view_count;type:int;not null;default:0" json:"view_count"`
	ClaimUserID    *int64     `gorm:"column:claim_user_id;type:bigint" json:"claim_user_id,omitempty"`
	ClaimTime      *time.Time `gorm:"column:claim_time;type:datetime" json:"claim_time,omitempty"`
	CreatedAt      time.Time  `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	IsDeleted      int8       `gorm:"column:is_deleted;type:tinyint;not null;default:0" json:"-"`
}

type ItemImage struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	ItemID    int64     `gorm:"column:item_id;type:bigint;not null;index:idx_item_images_item;comment:物品ID" json:"item_id"`
	ImageURL  string    `gorm:"column:image_url;type:varchar(500);not null;comment:图片链接" json:"image_url"`
	SortOrder int8      `gorm:"column:sort_order;type:tinyint;not null;default:1;comment:展示顺序:1封面 2第二张 3第三张" json:"sort_order"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 指定表名
func (ItemImage) TableName() string {
	return "item_images"
}

// TableName 指定表名
func (Item) TableName() string {
	return "items"
}

// ItemResponse 物品响应模型
type ItemResponse struct {
	ID             int64            `json:"id"`
	UserID         int64            `json:"user_id"`
	Title          string           `json:"title"`
	Description    string           `json:"description"`
	Type           int8             `json:"type"`
	Status         int8             `json:"status"`
	Locations      []model.Location `json:"locations"`
	LocationDetail *string          `json:"location_detail,omitempty"`
	Images         []ItemImage      `json:"images"`
	Tags           []model.Tag      `json:"tags"`
	LostFoundTime  time.Time        `json:"lost_found_time"`
	Contact        *string          `json:"contact,omitempty"`
	CreditReward   int64            `json:"credit_reward"`
	ViewCount      int64            `json:"view_count"`
	ClaimUserID    *int64           `json:"claim_user_id,omitempty"`
	ClaimTime      *time.Time       `json:"claim_time,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// CreateItemRequest 物品创建请求
type CreateItemRequest struct {
	Title          string    `json:"title" binding:"required"`
	Description    string    `json:"description" binding:"required"`
	Type           *int8     `json:"type" binding:"required"`
	LocationID     *int64    `json:"location_id,omitempty"`
	LocationDetail *string   `json:"location_detail,omitempty"`
	LostFoundTime  time.Time `json:"lost_found_time" binding:"required"`
	Contact        *string   `json:"contact,omitempty"`
	CreditReward   int       `json:"credit_reward,omitempty"`
	TagIDs         []int64   `json:"tag_ids,omitempty"`
}

// UpdateItemRequest 物品更新请求
type UpdateItemRequest struct {
	ID             int64      `json:"id" binding:"required"`
	Title          *string    `json:"title,omitempty"`
	Description    *string    `json:"description,omitempty"`
	Status         *int8      `json:"status,omitempty"`
	LocationID     *int64     `json:"location_id,omitempty"`
	LocationDetail *string    `json:"location_detail,omitempty"`
	LostFoundTime  *time.Time `json:"lost_found_time,omitempty"`
	Contact        *string    `json:"contact,omitempty"`
	CreditReward   *int       `json:"credit_reward,omitempty"`
	// TagIDs 非 nil 时整体替换物品标签关联
	TagIDs *[]int64 `json:"tag_ids,omitempty"`
}

// DeleteItemRequest 物品删除请求
type DeleteItemRequest struct {
	ID int64 `json:"id" binding:"required"`
}

// ListItemQuery 物品列表查询条件（GET 参数）
type ListItemQuery struct {
	Type       *int8  `form:"type,omitempty" binding:"omitempty,oneof=0 1"`
	Status     *int8  `form:"status,omitempty" binding:"omitempty,oneof=0 1 2"`
	LocationID *int64 `form:"location_id,omitempty"`
	TagID      *int64 `form:"tag_id,omitempty"`
	Keyword    string `form:"keyword,omitempty"`
	Page       int    `form:"page,omitempty" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size,omitempty" binding:"omitempty,min=1,max=100"`
}

// ItemListResponse 物品分页列表响应
type ItemListResponse struct {
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Items    []ItemResponse `json:"items"`
}

// ItemImageInput 单张图片输入
type ItemImageInput struct {
	ImageURL  string `json:"image_url" binding:"required"`
	SortOrder int8   `json:"sort_order" binding:"required"`
}

// SetItemImagesRequest 覆盖式设置物品图片请求
type SetItemImagesRequest struct {
	Images []ItemImageInput `json:"images" binding:"required"`
}
