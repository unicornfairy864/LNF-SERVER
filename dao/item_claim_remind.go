package dao

import (
	"time"

	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
)

// ListClaimsApproachingDeadline 查询「认领即将超时」的物品（用于超时前提醒发帖人）：
//   - status=1（已认领）、未删除
//   - claim_time ∈ (from, to]：即已认领但尚未达到自动关闭时限，且进入提醒窗口
//
// 说明：claim_time 由 ClaimItem 写入，自动关闭任务以 now - server.claim_auto_close 为截止线
func (itemGroup *ItemGroup) ListClaimsApproachingDeadline(from, to time.Time, limit int) ([]model.Item, error) {
	if limit <= 0 {
		limit = 200
	}
	var items []model.Item
	// status=1 与 service 层 itemStatusClaimed 一致（避免 dao → service 循环依赖，此处用字面量）
	err := global.LNF_DB.Where("is_deleted = 0 AND status = ? AND claim_time > ? AND claim_time <= ?",
		int8(1), from, to).Order("claim_time ASC").Limit(limit).Find(&items).Error
	return items, err
}
