package common

type Code int

const (
	// ==================== 通用 / 系统 0xxxx ====================
	CodeSuccess         Code = 0
	CodeParamError      Code = 1
	CodeUnauthorized    Code = 2
	CodeForbidden       Code = 3
	CodeNotFound        Code = 4
	CodeServerError     Code = 5
	CodeDatabaseError   Code = 6
	CodeTooManyRequests Code = 7
	CodeOperationFailed Code = 8

	// ==================== 用户与认证 1xxxx ====================
	CodeUserNotFound     Code = 10001
	CodeUsernameOccupied Code = 10002
	CodePasswordWrong    Code = 10003
	CodeTokenExpired     Code = 10004
	CodeUserDisabled     Code = 10005
	CodeOldPasswordWrong Code = 10006
	CodeEmailOccupied    Code = 10007
	CodePhoneOccupied    Code = 10008
	CodeCaptchaWrong     Code = 10009
	CodeUserNotActive    Code = 10010

	// ==================== 物品 2xxxx ====================
	CodeItemNotFound         Code = 20001
	CodeItemClosed           Code = 20002
	CodeItemAlreadyClaimed   Code = 20003
	CodeItemPendingAudit     Code = 20004
	CodeItemNoPermission     Code = 20005
	CodeItemTypeInvalid      Code = 20006
	CodeItemCreditNegative   Code = 20007
	CodeItemTimeEmpty        Code = 20008
	CodeItemTitleEmpty       Code = 20009
	CodeItemLocationInvalid  Code = 20010
	CodeItemAlreadyPublished Code = 20011
	CodeItemCannotClaimSelf  Code = 20012

	// ==================== 认领 3xxxx ====================
	CodeClaimNotFound           Code = 30001
	CodeClaimAlreadyExists      Code = 30002
	CodeClaimAlreadyHandled     Code = 30003
	CodeClaimCancelled          Code = 30004
	CodeClaimNoPermission       Code = 30005
	CodeClaimSelfItem           Code = 30006
	CodeClaimDescriptionTooLong Code = 30007
	CodeClaimDuplicate          Code = 30008

	// ==================== 标签 & 关联 4xxxx ====================
	CodeTagNotFound     Code = 40001
	CodeTagDuplicate    Code = 40002
	CodeTagDisabled     Code = 40003
	CodeTagNameInvalid  Code = 40004
	CodeTagInUse        Code = 40005
	CodeItemTagExists   Code = 40010
	CodeItemTagNotFound Code = 40011

	// ==================== 积分 5xxxx ====================
	CodeCreditInsufficient    Code = 50001
	CodeCreditLogNotFound     Code = 50002
	CodeCreditTypeInvalid     Code = 50003
	CodeCreditAmountInvalid   Code = 50004
	CodeCreditAlreadyRewarded Code = 50005

	// ==================== 通知 6xxxx ====================
	CodeNotificationNotFound     Code = 60001
	CodeNotificationNoPermission Code = 60002
	CodeNotificationAlreadyRead  Code = 60003

	// ==================== 举报 7xxxx ====================
	CodeReportNotFound       Code = 70001
	CodeReportDuplicate      Code = 70002
	CodeReportSelfContent    Code = 70003
	CodeReportAlreadyHandled Code = 70004

	// ==================== 地点 8xxxx ====================
	CodeLocationNotFound  Code = 80001
	CodeLocationDuplicate Code = 80002
	CodeLocationDisabled  Code = 80003
	CodeLocationInUse     Code = 80004
)

var Msg = map[Code]string{
	CodeSuccess:         "ok",
	CodeParamError:      "请求参数错误",
	CodeUnauthorized:    "未登录或Token无效",
	CodeForbidden:       "无权限操作",
	CodeNotFound:        "资源不存在",
	CodeServerError:     "服务器内部错误",
	CodeDatabaseError:   "数据库操作失败",
	CodeTooManyRequests: "请求过于频繁",
	CodeOperationFailed: "操作失败",

	CodeUserNotFound:     "用户不存在",
	CodeUsernameOccupied: "用户名已被占用",
	CodePasswordWrong:    "密码错误",
	CodeTokenExpired:     "Token已过期",
	CodeUserDisabled:     "账户已被禁用",
	CodeOldPasswordWrong: "原密码错误",
	CodeEmailOccupied:    "邮箱已被占用",
	CodePhoneOccupied:    "手机号已被占用",
	CodeCaptchaWrong:     "验证码错误",
	CodeUserNotActive:    "用户未激活",

	CodeItemNotFound:         "物品不存在",
	CodeItemClosed:           "物品已关闭，不可操作",
	CodeItemAlreadyClaimed:   "物品已被认领",
	CodeItemPendingAudit:     "物品待审核，暂不可操作",
	CodeItemNoPermission:     "无权操作该物品",
	CodeItemTypeInvalid:      "物品类型非法",
	CodeItemCreditNegative:   "积分奖励不能为负数",
	CodeItemTimeEmpty:        "丢失/拾到时间不能为空",
	CodeItemTitleEmpty:       "物品标题不能为空",
	CodeItemLocationInvalid:  "地点无效",
	CodeItemAlreadyPublished: "物品已发布",
	CodeItemCannotClaimSelf:  "不能认领自己发布的物品",

	CodeClaimNotFound:           "认领记录不存在",
	CodeClaimAlreadyExists:      "该物品已有待审核的认领申请",
	CodeClaimAlreadyHandled:     "认领申请已被处理",
	CodeClaimCancelled:          "认领申请已取消",
	CodeClaimNoPermission:       "无权审核该认领申请",
	CodeClaimSelfItem:           "不能认领自己发布的物品",
	CodeClaimDescriptionTooLong: "认领描述过长",
	CodeClaimDuplicate:          "重复提交认领申请",

	CodeTagNotFound:     "标签不存在",
	CodeTagDuplicate:    "标签名称已存在",
	CodeTagDisabled:     "标签已被禁用",
	CodeTagNameInvalid:  "标签名称无效",
	CodeTagInUse:        "标签正在被使用",
	CodeItemTagExists:   "物品已关联该标签",
	CodeItemTagNotFound: "标签关联记录不存在",

	CodeCreditInsufficient:    "积分余额不足",
	CodeCreditLogNotFound:     "积分流水记录不存在",
	CodeCreditTypeInvalid:     "积分操作类型非法",
	CodeCreditAmountInvalid:   "积分数量非法",
	CodeCreditAlreadyRewarded: "积分已发放",

	CodeNotificationNotFound:     "通知不存在",
	CodeNotificationNoPermission: "无权查看该通知",
	CodeNotificationAlreadyRead:  "通知已读",

	CodeReportNotFound:       "举报记录不存在",
	CodeReportDuplicate:      "已举报过该内容",
	CodeReportSelfContent:    "不能举报自己的内容",
	CodeReportAlreadyHandled: "举报已被处理",

	CodeLocationNotFound:  "地点不存在",
	CodeLocationDuplicate: "地点名称已存在",
	CodeLocationDisabled:  "地点已被禁用",
	CodeLocationInUse:     "地点正在被使用",
}

type CommonHttpStatus struct {
}