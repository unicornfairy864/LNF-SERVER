package initialization

import (
	"time"

	"github.com/unicornfairy864/LNF-SERVER/service"
)

// claimAutoCloseScanInterval 认领超时自动关闭的扫描间隔
const claimAutoCloseScanInterval = time.Minute

// StartClaimAutoCloseScheduler 启动认领超时自动关闭后台任务
// 每分钟扫描一次认领时间超过 server.claim_auto_close 的物品，
// 按"确认由他人找回"语义自动关闭并发放积分
func StartClaimAutoCloseScheduler() {
	go func() {
		// 启动时先执行一次，弥补停机期间到期的物品
		service.ItemService.AutoCloseExpiredClaimsService()
		ticker := time.NewTicker(claimAutoCloseScanInterval)
		defer ticker.Stop()
		for range ticker.C {
			service.ItemService.AutoCloseExpiredClaimsService()
		}
	}()
}
