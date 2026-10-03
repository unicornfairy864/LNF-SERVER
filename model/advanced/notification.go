package model

import "time"

// Notification 通知表
type Notification struct {
	ID      int64 `gorm:"column:id;primaryKey;autoIncrement"                        json:"id"         comment:"主键ID"`
	AdminID int64 `gorm:"column:admin_id;not null"                                  json:"admin_id"   comment:"发布管理员ID"`
	UserID  int64 `gorm:"column:user_id;not null"                                   json:"user_id"    comment:"接收用户ID"`
	// 通知类型触发点（2026-09-30 shop 兑换已接入，占位注释移除）：
	//  5 积分变动：shop 兑换扣分成功后写入（service/advanced/shop_service.go）
	//  6 商品兑换：兑换成功后写入发货提醒（同上）
	Type      int8       `gorm:"column:type;not null"                                      json:"type"       comment:"类型: 0系统通知 1物品匹配 2认领申请 3认领结果 4评论回复 5积分变动 6商品兑换"`
	Title     string     `gorm:"column:title;type:varchar(100);not null"                   json:"title"      comment:"通知标题"`
	Content   string     `gorm:"column:content;type:text"                                  json:"content"    comment:"通知内容（支持markdown格式）"`
	RelatedID *int64     `gorm:"column:related_id"                                         json:"related_id" comment:"关联ID（物品ID/认领ID/评论ID等）"`
	IsRead    int8       `gorm:"column:is_read;not null;default:0"                         json:"is_read"    comment:"是否已读: 0未读 1已读"`
	ReadAt    *time.Time `gorm:"column:read_at"                                            json:"read_at"    comment:"阅读时间"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"      json:"created_at" comment:"创建时间"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at" comment:"更新时间"`
	IsDeleted int8       `gorm:"column:is_deleted;not null;default:0"                      json:"is_deleted" comment:"逻辑删除: 0否 1是"`
}

func (Notification) TableName() string {
	return "notifications"
}

// NotificationListRequest 列表查询参数
type NotificationListRequest struct {
	Limit   int    `form:"limit"      json:"limit"`    // 每次取多少条
	Offset  int    `form:"offset"     json:"offset"`   // 忽略条数
	Type    *int8  `form:"type"       json:"type"`     // 按类型筛选，nil 不筛
	IsRead  *int8  `form:"is_read"    json:"is_read"`  // 按已读筛选，nil 不筛
	AdminID *int64 `form:"admin_id"   json:"admin_id"` // 按发布者筛选（区分我收到的/我发出的）
}

// NotificationSendRequest 发送通知请求
type NotificationSendRequest struct {
	UserIDs   []int64 `json:"user_ids"    binding:"required_without=SendToAll"` // 目标用户ID列表
	SendToAll bool    `json:"send_to_all"`                                      // 是否发给全体用户
	Type      int8    `json:"type"        binding:"required"`                   // 通知类型
	Title     string  `json:"title"       binding:"required,max=100"`           // 通知标题
	Content   string  `json:"content"     binding:"required"`                   // 通知内容
	RelatedID *int64  `json:"related_id"`                                       // 关联ID，可选
}

// NotificationIDsRequest 批量已读/删除请求
type NotificationIDsRequest struct {
	IDs []int64 `json:"ids" binding:"required,min=1"` // 目标记录ID列表
}

// NotificationIDsResponse 批量操作结果
type NotificationIDsResponse struct {
	Affected int `json:"affected"` // 实际影响条数
	Skipped  int `json:"skipped"`  // 跳过条数（如群发记录）
}

// NotificationItem 列表项（仅列表所需字段）
type NotificationItem struct {
	ID        int64     `json:"id"`
	AdminID   int64     `json:"admin_id"`
	Type      int8      `json:"type"`
	Title     string    `json:"title"`
	IsRead    int8      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// ToNotification 发送请求转模型
func ToNotification(r *NotificationSendRequest, adminID, userID int64) *Notification {
	return &Notification{
		AdminID:   adminID,
		UserID:    userID,
		Type:      r.Type,
		Title:     r.Title,
		Content:   r.Content,
		RelatedID: r.RelatedID,
	}
}

// ToNotificationResponse 模型转响应（隐藏敏感字段）
func ToNotificationResponse(n *Notification) map[string]any {
	return map[string]any{
		"id":         n.ID,
		"admin_id":   n.AdminID,
		"type":       n.Type,
		"title":      n.Title,
		"content":    n.Content,
		"related_id": n.RelatedID,
		"is_read":    n.IsRead,
		"read_at":    n.ReadAt,
		"created_at": n.CreatedAt,
	}
}

// ToNotificationItemList 批量转列表响应
func ToNotificationItemList(list []*Notification) []NotificationItem {
	res := make([]NotificationItem, 0, len(list))
	for _, n := range list {
		res = append(res, NotificationItem{
			ID:        n.ID,
			AdminID:   n.AdminID,
			Type:      n.Type,
			Title:     n.Title,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt,
		})
	}
	return res
}
