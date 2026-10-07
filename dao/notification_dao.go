package dao

import (
	"time"

	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"gorm.io/gorm"
)

// NotificationGroup 通知表数据访问
type NotificationGroup struct{}

// Create 单条插入
func (g *NotificationGroup) Create(n *model.Notification) error {
	return global.LNF_DB.Create(n).Error
}

// BatchCreate 批量插入，分批写入防止 SQL 过长；整体单事务：
// 中途任一批次失败则全部回滚，避免群发“半程投递”（代价：SendToAll 会形成
// 覆盖全部收件人的大事务，校园规模下可接受）
func (g *NotificationGroup) BatchCreate(list []*model.Notification) error {
	if len(list) == 0 {
		return nil
	}
	return global.LNF_DB.Transaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(list, 500).Error
	})
}

// GetByID 按 id + user_id 查询，防止越权
func (g *NotificationGroup) GetByID(id, userID int64) *model.Notification {
	var n model.Notification
	err := global.LNF_DB.
		Where("id = ? AND user_id = ? AND is_deleted = 0", id, userID).
		First(&n).Error
	if err != nil {
		return nil
	}
	return &n
}

// List 按条件分页查询
func (g *NotificationGroup) List(userID int64, req *model.NotificationListRequest) ([]*model.Notification, error) {
	var list []*model.Notification
	db := global.LNF_DB.Model(&model.Notification{}).
		Where("user_id = ? AND is_deleted = 0", userID)
	if req.Type != nil {
		db = db.Where("type = ?", *req.Type)
	}
	if req.IsRead != nil {
		db = db.Where("is_read = ?", *req.IsRead)
	}
	if req.AdminID != nil {
		db = db.Where("admin_id = ?", *req.AdminID)
	}
	// created_at 为秒级 DATETIME，群发记录同秒时间戳多，必须以 id 兜底保证排序稳定，
	// 否则 OFFSET 翻页会出现重复/丢行（与 item/good/order 模块一致）
	err := db.Order("created_at DESC, id DESC").
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&list).Error
	return list, err
}

// UnreadCount 未读数统计
func (g *NotificationGroup) UnreadCount(userID int64) (int64, error) {
	var count int64
	err := global.LNF_DB.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = 0 AND is_deleted = 0", userID).
		Count(&count).Error
	return count, err
}

// BatchMarkRead 批量标记已读，只操作自己的未读记录
func (g *NotificationGroup) BatchMarkRead(ids []int64, userID int64) (int64, error) {
	now := time.Now()
	res := global.LNF_DB.Model(&model.Notification{}).
		Where("id IN ? AND user_id = ? AND is_read = 0 AND is_deleted = 0", ids, userID).
		Updates(map[string]any{"is_read": 1, "read_at": now})
	return res.RowsAffected, res.Error
}

// BatchDelete 批量软删除，跳过自己发给自己的记录，返回(已删, 跳过)
func (g *NotificationGroup) BatchDelete(ids []int64, userID int64) (int64, int64, error) {
	// 统计不可删条数（user_id = admin_id 即群发给自己那条）
	var skipped int64
	err := global.LNF_DB.Model(&model.Notification{}).
		Where("id IN ? AND user_id = ? AND is_deleted = 0 AND user_id = admin_id", ids, userID).
		Count(&skipped).Error
	if err != nil {
		return 0, 0, err
	}

	// 删除其余记录
	res := global.LNF_DB.Model(&model.Notification{}).
		Where("id IN ? AND user_id = ? AND is_deleted = 0 AND user_id <> admin_id", ids, userID).
		Update("is_deleted", 1)
	return res.RowsAffected, skipped, res.Error
}
