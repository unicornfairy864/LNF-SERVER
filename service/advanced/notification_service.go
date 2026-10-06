package advanced

import (
	"fmt"
	"log"
	"time"

	"github.com/unicornfairy864/LNF-SERVER/dao"
	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	modelbasic "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

type NotificationServiceGroup struct{}

// unreadKey 未读数缓存键
func unreadKey(userID int64) string {
	return fmt.Sprintf("notification:unread:%d", userID)
}

// notificationDefaultLimit List 未传 limit 时的缺省值
// （limit=0 会生成 LIMIT 0 恒空列表，负数会退化为无 LIMIT 全表返回）
const notificationDefaultLimit = 10

// Send 发送通知（支持全体，异步）
// 这里并发有风险，炸了优先查这里
func (s *NotificationServiceGroup) Send(req *model.NotificationSendRequest, adminID int64) response.Code {
	// 目标校验放同步段，错误可即时返回调用方：
	// SendToAll 与 UserIDs 互斥（同时传不再静默忽略显式列表）；
	// 显式列表过滤不存在/已删除用户，全部无效则拒绝，避免写出孤儿通知
	var targets []int64
	if req.SendToAll {
		if len(req.UserIDs) > 0 {
			return response.CodeParamError
		}
	} else {
		if len(req.UserIDs) == 0 {
			return response.CodeParamError
		}
		for _, u := range dao.UserDao.GetUserListByIDs(req.UserIDs) {
			if u.IsDeleted == 0 {
				targets = append(targets, u.ID)
			}
		}
		if len(targets) == 0 {
			return response.CodeParamError
		}
	}
	go s.sendAsync(req, targets, adminID)
	return response.CodeSuccess
}

// sendAsync 异步执行发送（targets 为同步段过滤后的显式目标；SendToAll 现场拉取全体未删除用户）
func (s *NotificationServiceGroup) sendAsync(req *model.NotificationSendRequest, targets []int64, adminID int64) {
	if req.SendToAll {
		var userIDs []int64
		if err := global.LNF_DB.Model(&modelbasic.User{}).
			Where("is_deleted = 0").
			Pluck("id", &userIDs).Error; err != nil {
			log.Printf("[notification] SendToAll 拉取全体用户失败: %v", err)
			return
		}
		s.batchCreate(userIDs, req, adminID)
		return
	}
	s.batchCreate(targets, req, adminID)
}

// batchCreate 构造并写入通知
func (s *NotificationServiceGroup) batchCreate(userIDs []int64, req *model.NotificationSendRequest, adminID int64) {
	if len(userIDs) == 0 {
		return
	}
	list := make([]*model.Notification, 0, len(userIDs))
	for _, uid := range userIDs {
		list = append(list, model.ToNotification(req, adminID, uid))
	}
	if err := dao.NotificationDao.BatchCreate(list); err != nil {
		log.Printf("[notification] batchCreate 写入失败 admin_id=%d type=%d 目标数=%d: %v", adminID, *req.Type, len(userIDs), err)
		return
	}
	for _, uid := range userIDs {
		_ = dao.RedisDao.DelKey(unreadKey(uid))
	}
}

// List 获取列表
func (s *NotificationServiceGroup) List(userID int64, req *model.NotificationListRequest) ([]model.NotificationItem, response.Code) {
	// limit/offset 兑底：未传或非法统一归位（0 生成 LIMIT 0 恒空，负数退化为全表返回）
	if req.Limit <= 0 {
		req.Limit = notificationDefaultLimit
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	list, err := dao.NotificationDao.List(userID, req)
	if err != nil {
		return nil, response.CodeNotificationQueryFailed
	}
	return model.ToNotificationItemList(list), response.CodeSuccess
}

// UnreadCount 未读数（Redis 缓存 5 分钟，删除重建）
func (s *NotificationServiceGroup) UnreadCount(userID int64) (int64, response.Code) {
	key := unreadKey(userID)
	if count, err := dao.RedisDao.GetValueInt64(key); err == nil && count >= 0 {
		return count, response.CodeSuccess
	}
	count, err := dao.NotificationDao.UnreadCount(userID)
	if err != nil {
		return 0, response.CodeNotificationQueryFailed
	}
	_ = dao.RedisDao.SetKey(key, count, 5*time.Minute)
	return count, response.CodeSuccess
}

// GetAndRead 详情自动标记已读
func (s *NotificationServiceGroup) GetAndRead(id, userID int64) (*model.Notification, response.Code) {
	n := dao.NotificationDao.GetByID(id, userID)
	if n == nil {
		return nil, response.CodeNotificationNotFound
	}
	if n.IsRead == 0 {
		if _, err := dao.NotificationDao.BatchMarkRead([]int64{id}, userID); err != nil {
			return nil, response.CodeNotificationUpdateFailed
		}
		_ = dao.RedisDao.DelKey(unreadKey(userID))
		n.IsRead = 1
	}
	return n, response.CodeSuccess
}

// BatchMarkRead 批量已读
func (s *NotificationServiceGroup) BatchMarkRead(ids []int64, userID int64) response.Code {
	if _, err := dao.NotificationDao.BatchMarkRead(ids, userID); err != nil {
		return response.CodeNotificationUpdateFailed
	}
	_ = dao.RedisDao.DelKey(unreadKey(userID))
	return response.CodeSuccess
}

// BatchDelete 批量删除（跳过自己发给自己的记录）
func (s *NotificationServiceGroup) BatchDelete(ids []int64, userID int64) response.Code {
	if _, _, err := dao.NotificationDao.BatchDelete(ids, userID); err != nil {
		return response.CodeNotificationDeleteFailed
	}
	_ = dao.RedisDao.DelKey(unreadKey(userID))
	return response.CodeSuccess
}

// Create 内部方法，供各模块系统触发点调用（adminID=0 表示系统触发；失败由调用方记日志，不影响主流程）。
// 已接入触发点（2026-09-30）：
//  1. item 认领关闭/确认/自行关闭 → type=3 认领结果（service/basic/item_service.go）
//  2. user QQ 绑定成功 → type=0 系统通知（service/basic/user_service.go）
//  3. shop 兑换事务提交成功 → type=5 积分变动 + type=6 商品兑换（本包 shop_service.go，
//     先积分变动后发货提醒；群内 at 提醒由 chensong 发送）
func (s *NotificationServiceGroup) Create(adminID, userID int64, ntype int8, title, content string, relatedID *int64) error {
	n := &model.Notification{
		AdminID: adminID,
		UserID:  userID,
		Type:    ntype,
		Title:   title,
		Content: content,
	}
	if err := dao.NotificationDao.Create(n); err != nil {
		return err
	}
	_ = dao.RedisDao.DelKey(unreadKey(userID))
	return nil
}
