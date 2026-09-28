package response

type Code int

// ==================== 通用 / 系统 0xxxx ====================
const (
	CodeSuccess               Code = 0  // 操作成功
	CodeParamError            Code = 1  // 请求参数错误
	CodeUnauthorized          Code = 2  // 未登录或 Token 无效
	CodeForbidden             Code = 3  // 无权限操作
	CodeNotFound              Code = 4  // 资源不存在
	CodeServerError           Code = 5  // 服务器内部错误
	CodeDatabaseError         Code = 6  // 数据库操作失败
	CodeUnknownError          Code = 7  // 未知错误
	CodeChenSongError         Code = 8  // 陈松出了故障!
	CodeUploadFileTooLarge    Code = 9  // 上传文件过大
	CodeUploadFileTypeInvalid Code = 10 // 上传文件类型不支持
	CodeUploadFailed          Code = 11 // 上传文件保存失败
)

// ==================== 用户与认证 1xxxx ====================
const (
	CodeUserOrPasswordError    Code = 10001 // 用户名或密码错误
	CodeUsernameOccupied       Code = 10002 // 用户名已被占用
	CodeFormInvalid            Code = 10003 // 用户名/昵称/密码不符合规则
	CodeInvalidSigningMethod   Code = 10004 // 无效的加密方式
	CodeTokenBanned            Code = 10005 // 令牌被禁用
	CodeUserNotFoundOrBanned   Code = 10006 // 用户不存在或被禁用
	CodeQQSessionAlreadyExists Code = 10007 // 同一个QQ的验证会话已经存在，请等失效后重试
	CodeQQCodeError            Code = 10008 // 会话的QQ验证码错误
	CodeQQUserNotInGroup       Code = 10009 // 用户不在QQ群内
	CodeQQNumberError          Code = 10010 // 需要验证的QQ号不一致
	CodeQQTooManyRequests      Code = 10011 // QQ绑定请求过多，请等冷却后再试
	CodeQQSessionNotExist      Code = 10012 // QQ绑定会话不存在或已经失效
	CodeQQAlreadyRegistered    Code = 10013 // QQ已被注册或账号已绑定QQ
	CodeCreditNotEnough        Code = 10014 // 积分不够
)

// ==================== 物品 2xxxx ====================
const (
	CodeItemNotFound         Code = 20001 // 物品不存在
	CodeItemClosed           Code = 20002 // 物品已关闭
	CodeItemAlreadyClaimed   Code = 20003 // 物品已被认领
	CodeItemPendingAudit     Code = 20004 // 物品待审核
	CodeItemNoPermission     Code = 20005 // 无权操作该物品
	CodeItemTypeInvalid      Code = 20006 // 物品类型非法
	CodeItemCreditNegative   Code = 20007 // 积分奖励不能为负数
	CodeItemTimeEmpty        Code = 20008 // 丢失/拾到时间不能为空
	CodeItemTitleEmpty       Code = 20009 // 物品标题不能为空
	CodeItemLocationInvalid  Code = 20010 // 地点无效
	CodeItemAlreadyPublished Code = 20011 // 物品已发布
	CodeItemCannotClaimSelf  Code = 20012 // 不能认领自己发布的物品
	CodeItemImageTooMany     Code = 20013 // 物品图片最多3张
	CodeItemImageInvalid     Code = 20014 // 物品图片参数错误
)

// ==================== 认领 3xxxx ====================
// 注：30001 / 30005 的语义与 2xxxx 中的 20003 / 20002 相同，
// 但编码值不同，Go 常量名不可重复，故加 Claim 前缀区分。
const (
	CodeClaimItemAlreadyClaimed Code = 30001 // 物品已被认领
	CodeClaimNotFound           Code = 30002 // 认领记录不存在
	CodeClaimNoPermission       Code = 30003 // 无权审核该认领申请
	CodeClaimSelfItem           Code = 30004 // 不能认领自己发布的物品
	CodeClaimItemClosed         Code = 30005 // 物品已关闭
	CodeClaimQQRequired         Code = 30006 // 认领前请先绑定QQ
	CodeClaimDescriptionTooLong Code = 30007 // 认领描述过长
	CodeClaimDuplicate          Code = 30008 // 重复提交认领申请
)

// ==================== 标签 & 关联 4xxxx ====================
const (
	CodeTagNotFound     Code = 40001 // 标签不存在
	CodeTagDuplicate    Code = 40002 // 标签名称已存在
	CodeTagDisabled     Code = 40003 // 标签已被禁用
	CodeTagNameInvalid  Code = 40004 // 标签名称无效
	CodeTagInUse        Code = 40005 // 标签正在被使用
	CodeItemTagExists   Code = 40010 // 物品已关联该标签
	CodeItemTagNotFound Code = 40011 // 标签关联记录不存在
)

// ==================== 积分 5xxxx ====================
const (
	CodeCreditInsufficient    Code = 50001 // 积分余额不足
	CodeCreditLogNotFound     Code = 50002 // 积分流水记录不存在
	CodeCreditTypeInvalid     Code = 50003 // 积分操作类型非法
	CodeCreditAmountInvalid   Code = 50004 // 积分数量非法
	CodeCreditAlreadyRewarded Code = 50005 // 积分已发放
)

// ==================== 通知 6xxxx ====================
const (
	CodeNotificationNotFound     Code = 60001 // 通知不存在
	CodeNotificationNoPermission Code = 60002 // 无权查看该通知
	CodeNotificationAlreadyRead  Code = 60003 // 通知已读
	CodeNotificationQueryFailed  Code = 60004 // 通知查询失败
	CodeNotificationCreateFailed Code = 60005 // 通知创建失败
	CodeNotificationUpdateFailed Code = 60006 // 通知更新失败
	CodeNotificationDeleteFailed Code = 60007 // 通知删除失败
	CodeNotificationSendFailed   Code = 60008 // 通知发送失败
)

// ==================== 举报 7xxxx ====================
const (
	CodeReportNotFound       Code = 70001 // 举报记录不存在
	CodeReportDuplicate      Code = 70002 // 已举报过该内容
	CodeReportSelfContent    Code = 70003 // 不能举报自己的内容
	CodeReportAlreadyHandled Code = 70004 // 举报已被处理
)

// ==================== 地点 8xxxx ====================
const (
	CodeLocationNotFound    Code = 80001 // 地点不存在
	CodeLocationDuplicate   Code = 80002 // 地点名称已存在
	CodeLocationDisabled    Code = 80003 // 地点已被禁用
	CodeLocationInUse       Code = 80004 // 地点正在被使用
	CodeLocationHasChildren Code = 80005 // 存在子地点，无法删除
)

// ==================== 公告 9xxxx ====================
const (
	CodeAnnouncementNotFound     Code = 90001 // 公告不存在
	CodeAnnouncementNoPermission Code = 90002 // 无权操作该公告
	CodeAnnouncementInvalid      Code = 90003 // 公告参数或状态错误
)

// ==================== 测试 -xxxx ====================
const (
	CodeTest Code = -1
)

// Msg 全局消息映射
var Msg = map[Code]string{
	// 通用 / 系统
	CodeSuccess:               "操作成功",
	CodeParamError:            "请求参数错误",
	CodeUnauthorized:          "未登录或 Token 无效",
	CodeForbidden:             "无权限操作",
	CodeNotFound:              "资源不存在",
	CodeServerError:           "服务器内部错误",
	CodeDatabaseError:         "数据库操作失败",
	CodeUnknownError:          "未知错误",
	CodeChenSongError:         "陈松出了故障!",
	CodeUploadFileTooLarge:    "上传文件过大",
	CodeUploadFileTypeInvalid: "上传文件类型不支持",
	CodeUploadFailed:          "上传文件保存失败",

	// 用户与认证
	CodeUserOrPasswordError:    "用户名或密码错误",
	CodeUsernameOccupied:       "用户名已被占用",
	CodeFormInvalid:            "用户名/昵称/密码不符合规则",
	CodeInvalidSigningMethod:   "无效的加密方式",
	CodeTokenBanned:            "令牌被禁用",
	CodeUserNotFoundOrBanned:   "用户不存在或被禁用",
	CodeQQSessionAlreadyExists: "同一个QQ的验证会话已经存在，请等失效后重试",
	CodeQQCodeError:            "会话的QQ验证码错误",
	CodeQQUserNotInGroup:       "用户不在QQ群内",
	CodeQQNumberError:          "需要验证的QQ号不一致",
	CodeQQTooManyRequests:      "QQ绑定请求过多，请等冷却后再试",
	CodeQQSessionNotExist:      "QQ绑定会话不存在或已经失效",
	CodeQQAlreadyRegistered:    "QQ已被注册或账号已绑定QQ",
	CodeCreditNotEnough:        "积分不够",

	// 物品
	CodeItemNotFound:         "物品不存在",
	CodeItemClosed:           "物品已关闭",
	CodeItemAlreadyClaimed:   "物品已被认领",
	CodeItemPendingAudit:     "物品待审核",
	CodeItemNoPermission:     "无权操作该物品",
	CodeItemTypeInvalid:      "物品类型非法",
	CodeItemCreditNegative:   "积分奖励不能为负数",
	CodeItemTimeEmpty:        "丢失/拾到时间不能为空",
	CodeItemTitleEmpty:       "物品标题不能为空",
	CodeItemLocationInvalid:  "地点无效",
	CodeItemAlreadyPublished: "物品已发布",
	CodeItemCannotClaimSelf:  "不能认领自己发布的物品",
	CodeItemImageTooMany:     "物品图片最多3张",
	CodeItemImageInvalid:     "物品图片参数错误",

	// 认领
	CodeClaimItemAlreadyClaimed: "物品已被认领",
	CodeClaimNotFound:           "认领记录不存在",
	CodeClaimNoPermission:       "无权审核该认领申请",
	CodeClaimSelfItem:           "不能认领自己发布的物品",
	CodeClaimItemClosed:         "物品已关闭",
	CodeClaimQQRequired:         "认领前请先绑定QQ",
	CodeClaimDescriptionTooLong: "认领描述过长",
	CodeClaimDuplicate:          "重复提交认领申请",

	// 标签 & 关联
	CodeTagNotFound:     "标签不存在",
	CodeTagDuplicate:    "标签名称已存在",
	CodeTagDisabled:     "标签已被禁用",
	CodeTagNameInvalid:  "标签名称无效",
	CodeTagInUse:        "标签正在被使用",
	CodeItemTagExists:   "物品已关联该标签",
	CodeItemTagNotFound: "标签关联记录不存在",

	// 积分
	CodeCreditInsufficient:    "积分余额不足",
	CodeCreditLogNotFound:     "积分流水记录不存在",
	CodeCreditTypeInvalid:     "积分操作类型非法",
	CodeCreditAmountInvalid:   "积分数量非法",
	CodeCreditAlreadyRewarded: "积分已发放",

	// 通知
	CodeNotificationNotFound:     "通知不存在",
	CodeNotificationNoPermission: "无权查看该通知",
	CodeNotificationAlreadyRead:  "通知已读",
	CodeNotificationQueryFailed:  "通知查询失败",
	CodeNotificationCreateFailed: "通知创建失败",
	CodeNotificationUpdateFailed: "通知更新失败",
	CodeNotificationDeleteFailed: "通知删除失败",
	CodeNotificationSendFailed:   "通知发送失败",

	// 举报
	CodeReportNotFound:       "举报记录不存在",
	CodeReportDuplicate:      "已举报过该内容",
	CodeReportSelfContent:    "不能举报自己的内容",
	CodeReportAlreadyHandled: "举报已被处理",

	// 地点
	CodeLocationNotFound:    "地点不存在",
	CodeLocationDuplicate:   "地点名称已存在",
	CodeLocationDisabled:    "地点已被禁用",
	CodeLocationInUse:       "地点正在被使用",
	CodeLocationHasChildren: "存在子地点，无法删除",

	// 公告
	CodeAnnouncementNotFound:     "公告不存在",
	CodeAnnouncementNoPermission: "无权操作该公告",
	CodeAnnouncementInvalid:      "公告参数或状态错误",
}
