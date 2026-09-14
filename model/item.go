package model

import "time"

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
	CreditReward   int        `gorm:"column:credit_reward;type:int;default:0" json:"credit_reward"`
	ViewCount      int        `gorm:"column:view_count;type:int;not null;default:0" json:"view_count"`
	ClaimUserID    *int64     `gorm:"column:claim_user_id;type:bigint" json:"claim_user_id,omitempty"`
	ClaimTime      *time.Time `gorm:"column:claim_time;type:datetime" json:"claim_time,omitempty"`
	CreatedAt      time.Time  `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
	IsDeleted      int8       `gorm:"column:is_deleted;type:tinyint;not null;default:0" json:"-"`
}

// TableName 指定表名
func (Item) TableName() string {
	return "items"
}

// ItemResponse 物品响应模型
type ItemResponse struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Type           int8       `json:"type"`
	Status         int8       `json:"status"`
	LocationID     *int64     `json:"location_id,omitempty"`
	LocationDetail *string    `json:"location_detail,omitempty"`
	LostFoundTime  time.Time  `json:"lost_found_time"`
	Contact        *string    `json:"contact,omitempty"`
	CreditReward   int        `json:"credit_reward"`
	ViewCount      int        `json:"view_count"`
	ClaimUserID    *int64     `json:"claim_user_id,omitempty"`
	ClaimTime      *time.Time `json:"claim_time,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ItemCreateRequest 物品创建请求
type ItemCreateRequest struct {
	Title          string    `json:"title" binding:"required"`
	Description    string    `json:"description" binding:"required"`
	Type           int8      `json:"type" binding:"required"`
	LocationID     *int64    `json:"location_id,omitempty"`
	LocationDetail *string   `json:"location_detail,omitempty"`
	LostFoundTime  time.Time `json:"lost_found_time" binding:"required"`
	Contact        *string   `json:"contact,omitempty"`
	CreditReward   int       `json:"credit_reward,omitempty"`
}

// ItemUpdateRequest 物品更新请求
type ItemUpdateRequest struct {
	Title          *string    `json:"title,omitempty"`
	Description    *string    `json:"description,omitempty"`
	Status         *int8      `json:"status,omitempty"`
	LocationID     *int64     `json:"location_id,omitempty"`
	LocationDetail *string    `json:"location_detail,omitempty"`
	LostFoundTime  *time.Time `json:"lost_found_time,omitempty"`
	Contact        *string    `json:"contact,omitempty"`
	CreditReward   *int       `json:"credit_reward,omitempty"`
	ClaimUserID    *int64     `json:"claim_user_id,omitempty"`
	ClaimTime      *time.Time `json:"claim_time,omitempty"`
}

// ItemDeleteRequest 物品删除请求
type ItemDeleteRequest struct {
	ID int64 `json:"id" binding:"required"`
}

// ItemChangeStatusRequest 物品状态变更请求
type ItemChangeStatusRequest struct {
	ID     int64 `json:"id" binding:"required"`
	Status int8  `json:"status" binding:"required"`
}

// ItemToResponse 将Item模型转换为ItemResponse
func ItemToResponse(item *Item) ItemResponse {
	if item == nil {
		return ItemResponse{}
	}

	return ItemResponse{
		ID:             item.ID,
		UserID:         item.UserID,
		Title:          item.Title,
		Description:    item.Description,
		Type:           item.Type,
		Status:         item.Status,
		LocationID:     item.LocationID,
		LocationDetail: item.LocationDetail,
		LostFoundTime:  item.LostFoundTime,
		Contact:        item.Contact,
		CreditReward:   item.CreditReward,
		ViewCount:      item.ViewCount,
		ClaimUserID:    item.ClaimUserID,
		ClaimTime:      item.ClaimTime,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}
