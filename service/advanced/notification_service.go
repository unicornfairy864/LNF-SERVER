package advanced

import (
	"fmt"
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

// Send 发送通知（支持全体，异步）
// 这里并发有风险，炸了优先查这里
func (s *NotificationServiceGroup) Send(req *model.NotificationSendRequest, adminID int64) response.Code {
	go s.sendAsync(req, adminID)
	return response.CodeSuccess
}

// sendAsync 异步执行发送
func (s *NotificationServiceGroup) sendAsync(req *model.NotificationSendRequest, adminID int64) {
	if req.SendToAll {
		var userIDs []int64
		if err := global.LNF_DB.Model(&modelbasic.User{}).Pluck("id", &userIDs).Error; err != nil {
			return
		}
		s.batchCreate(userIDs, req, adminID)
		return
	}
	s.batchCreate(req.UserIDs, req, adminID)
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
		return
	}
	for _, uid := range userIDs {
		_ = dao.RedisDao.DelKey(unreadKey(uid))
	}
}

// List 获取列表
func (s *NotificationServiceGroup) List(userID int64, req *model.NotificationListRequest) ([]model.NotificationItem, response.Code) {
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

// Create 内部方法，供认领/评论/积分等模块调用
func (s *NotificationServiceGroup) Create(adminID, userID int64, ntype int8, title, content string, relatedID *int64) error {
	n := &model.Notification{
		AdminID:   adminID,
		UserID:    userID,
		Type:      ntype,
		Title:     title,
		Content:   content,
		RelatedID: relatedID,
	}
	if err := dao.NotificationDao.Create(n); err != nil {
		return err
	}
	_ = dao.RedisDao.DelKey(unreadKey(userID))
	return nil
}
