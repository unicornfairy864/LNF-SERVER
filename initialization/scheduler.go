package initialization

import (
	"time"

	"github.com/unicornfairy864/LNF-SERVER/service"
)

// claimAutoCloseScanInterval 认领超时自动关闭的扫描间隔
const claimAutoCloseScanInterval = 5 * time.Minute

// StartClaimAutoCloseScheduler 启动认领超时自动关闭后台任务
// 每 5 分钟扫描一次：① 先提醒「即将超时」的发帖人（server.claim_auto_close 前 2h，可用
// server.claim_auto_close_remind_before 调整）；② 再把认领时间超过 server.claim_auto_close
// 的物品按“确认由他人找回”语义自动关闭并发放积分
func StartClaimAutoCloseScheduler() {
	go func() {
		// 启动时先执行一次，弥补停机期间到期的物品
		service.ItemService.RemindExpiringClaimsService()
		service.ItemService.AutoCloseExpiredClaimsService()
		ticker := time.NewTicker(claimAutoCloseScanInterval)
		defer ticker.Stop()
		for range ticker.C {
			service.ItemService.RemindExpiringClaimsService()
			service.ItemService.AutoCloseExpiredClaimsService()
		}
	}()
}
