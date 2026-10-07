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

// notificationMaxLimit List 单页上限，防一次拉取过多（与 item 模块 page_size max=100 对齐）
const notificationMaxLimit = 100

// notificationMaxUserIDs Send 显式 user_ids 单次上限，防超大 IN 子句与事务体量失控
const notificationMaxUserIDs = 1000

// Send 发送通知（支持全体，异步）
//
// 并发风险说明（原“这里并发有风险，炸了优先查这里”占位注释的展开，2026-10-07）：
//  1. fire-and-forget：同步段只做目标校验，随即 go sendAsync 并立即返回成功；
//     异步写入失败或进程重启只会留下日志，不会重试（既定契约，见 agent.md/api_agent.md）。
//  2. 无界 goroutine：每次 Send 起一个 goroutine，没有队列、没有并发上限；
//     SendToAll 还会在 goroutine 内现场拉取全体用户，并一次性构造全量通知切片。
//     重复/并发调用群发时，内存峰值与 DB 连接池同时承压——这是原注释所指
//     “炸了”的最可能实体。已接受现状（不引入队列/闸门），靠调用方自律控制频率。
//  3. 原子性：批量写入已由 dao.BatchCreate 包在单事务内（修复前分批 INSERT
//     各自提交，中途失败会留下“半程投递”的部分收件人）；现在单次群发要么全成
//     功要么全失败，失败时重发整批即可。
//  4. sendAsync 已加 recover：异步段 panic 只记日志，不再带崩整个进程。
//  5. 已知竞态（接受）：同步段校验通过后、异步写入前用户被删 → 产生孤儿通知
//     （无害，仅占一行）；写入后逐个 DelKey 未读缓存与 UnreadCount 读穿之间
//     存在窗口 → 未读数最多陈旧一个 TTL（5 分钟）。
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
		if len(req.UserIDs) == 0 || len(req.UserIDs) > notificationMaxUserIDs {
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
// recover：此函数运行在独立 goroutine 中，任何 panic 若不捕获会直接打崩进程，
// 故统一在此收口记日志（并发风险的完整说明见 Send 的注释）
func (s *NotificationServiceGroup) sendAsync(req *model.NotificationSendRequest, targets []int64, adminID int64) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[notification] sendAsync panic 恢复 admin_id=%d send_to_all=%v 目标数=%d: %v",
				adminID, req.SendToAll, len(targets), r)
		}
	}()
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
	// limit/offset 兑底：未传或非法统一归位（0 生成 LIMIT 0 恒空，负数退化为全表返回），
	// 并对 limit 封顶防单次拉取过大
	if req.Limit <= 0 {
		req.Limit = notificationDefaultLimit
	} else if req.Limit > notificationMaxLimit {
		req.Limit = notificationMaxLimit
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
// 已知竞态（接受现状，见 Send 注释第 5 点）：读穿流程“查库后、SetKey 前”若并发
// 发生新通知写入 + DelKey，旧值仍会被写回缓存，未读数最多陈旧一个 TTL（5 分钟）；
// TTL 有界且未读数非关键数据，不引入额外一致性开销
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
func (s *NotificationServiceGroup) Create(adminID, userID int64, ntype int8, title, content string) error {
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
