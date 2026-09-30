package model

import "time"

// Order 积分兑换订单模型（快照设计：goods_name/price/qq/nickname 落库，
// 商品后续改名/删除不影响历史记录与群消息发送；无状态字段、无软删）
type Order struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderNo   string    `gorm:"column:order_no;type:varchar(32);not null;uniqueIndex:uk_order_no" json:"order_no"`
	UserID    int64     `gorm:"column:user_id;type:bigint;not null;index:idx_user_created,priority:1" json:"user_id"`
	GoodsID   int64     `gorm:"column:goods_id;type:bigint;not null;index:idx_goods_created,priority:1" json:"goods_id"`
	GoodsName string    `gorm:"column:goods_name;type:varchar(100);not null" json:"goods_name"`
	Price     int64     `gorm:"column:price;type:int;not null" json:"price"`
	QQ        string    `gorm:"column:qq;type:varchar(50);not null" json:"qq"`
	Nickname  string    `gorm:"column:nickname;type:varchar(50);not null" json:"nickname"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;index:idx_user_created,priority:2;index:idx_goods_created,priority:2" json:"created_at"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// OrderListQuery 我的兑换列表查询条件（GET 参数）
type OrderListQuery struct {
	Page     int `form:"page,omitempty" binding:"omitempty,min=1"`
	PageSize int `form:"page_size,omitempty" binding:"omitempty,min=1,max=100"`
}

// OrderListResponse 兑换记录分页列表响应
type OrderListResponse struct {
	Total    int64   `json:"total"`
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
	Orders   []Order `json:"orders"`
}

// RedeemGoodsResponse 兑换成功响应
type RedeemGoodsResponse struct {
	OrderNo   string    `json:"order_no"`
	GoodsID   int64     `json:"goods_id"`
	GoodsName string    `json:"goods_name"`
	Price     int64     `json:"price"`
	Credit    int64     `json:"credit"` // 兑换后剩余积分
	CreatedAt time.Time `json:"created_at"`
}
