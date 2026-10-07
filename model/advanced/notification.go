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
	Type    int8   `gorm:"column:type;not null"                                      json:"type"       comment:"类型: 0系统通知 1物品匹配 2认领申请 3认领结果 4评论回复 5积分变动 6商品兑换"`
	Title   string `gorm:"column:title;type:varchar(100);not null"                   json:"title"      comment:"通知标题"`
	Content string `gorm:"column:content;type:text"                                     json:"content"   comment:"通知内容（支持markdown格式）"`
	IsRead  int8   `gorm:"column:is_read;not null;default:0"                         json:"is_read"    comment:"是否已读: 0未读 1已读"`
	// read_at 未读时为 NULL：omitempty → 未读时整个键缺失（前端必须按可选处理）
	ReadAt    *time.Time `gorm:"column:read_at"                                            json:"read_at,omitempty" comment:"阅读时间（未读时键缺失）"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"      json:"created_at" comment:"创建时间"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at" comment:"更新时间"`
	// is_deleted 为内部逻辑删除位（查询恒过滤 =0）：沿用项目惯例 json:"-" 彻底不返回
	// （item/user/good/announcement 同款；不用 omitempty 是为了逻辑删除位=1 时也不外泄）
	IsDeleted int8 `gorm:"column:is_deleted;not null;default:0"                      json:"-"          comment:"逻辑删除: 0否 1是"`
}

func (Notification) TableName() string {
	return "notifications"
}

// NotificationListRequest 列表查询参数
type NotificationListRequest struct {
	Limit   int    `form:"limit"      json:"limit"`                                              // 每次取多少条，缺省 10，上限 100（service 侧兜底）
	Offset  int    `form:"offset"     json:"offset"`                                             // 忽略条数
	Type    *int8  `form:"type"       json:"type"       binding:"omitempty,oneof=0 1 2 3 4 5 6"` // 按类型筛选，nil 不筛
	IsRead  *int8  `form:"is_read"    json:"is_read"    binding:"omitempty,oneof=0 1"`           // 按已读筛选，nil 不筛
	AdminID *int64 `form:"admin_id"   json:"admin_id"`                                           // 按发布者筛选（区分我收到的/我发出的）
}

// NotificationSendRequest 发送通知请求
type NotificationSendRequest struct {
	UserIDs   []int64 `json:"user_ids"    binding:"required_without=SendToAll"`   // 目标用户ID列表，service 侧上限 1000；send_to_all=true 时必须为空
	SendToAll bool    `json:"send_to_all"`                                        // 是否发给全体用户
	Type      *int8   `json:"type"        binding:"required,oneof=0 1 2 3 4 5 6"` // 通知类型 0-6；指针 required：缺字段→1，0 正常读入
	Title     string  `json:"title"       binding:"required,max=100"`             // 通知标题
	Content   string  `json:"content"     binding:"required"`                     // 通知内容
}

// NotificationIDsRequest 批量已读/删除请求
type NotificationIDsRequest struct {
	IDs []int64 `json:"ids" binding:"required,min=1,max=200"` // 目标记录ID列表，单次 1-200 条
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
		AdminID: adminID,
		UserID:  userID,
		Type:    *r.Type, // binding required 保证非 nil（0 是合法值）
		Title:   r.Title,
		Content: r.Content,
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
