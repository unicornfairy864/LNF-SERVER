package response

type Code int

const (
	// ==================== 通用 / 系统 0xxxx ==============
	CodeSuccess         Code = 0
	CodeParamError      Code = 1
	CodeUnauthorized    Code = 2
	CodeForbidden       Code = 3
	CodeNotFound        Code = 4
	CodeServerError     Code = 5
	CodeDatabaseError   Code = 6
	CodeOperationFailed Code = 7
	CodeUnknownError    Code = 8
	CodeChenSongError   Code = 9

	// ==================== 用户与认证 1xxxx ===============
	CodeUserOrPasswordError  Code = 10001
	CodeUsernameOccupied     Code = 10002
	CodeFormInvalid          Code = 10003
	CodeInvalidSigningMethod Code = 10004
	CodeTokenBanned          Code = 10005
	CodeUserNotFoundOrBanned Code = 10006
	CodeQQCodeAlreadyExists  Code = 10007
	CodeQQCodeError          Code = 10008

	// ==================== 物品 2xxxx ====================

	// ==================== 认领 3xxxx ====================

	// ==================== 标签 & 关联 4xxxx ==============

	// ==================== 积分 5xxxx ====================

	// ==================== 通知 6xxxx ====================

	// ==================== 举报 7xxxx ====================

	// ==================== 地点 8xxxx ====================

	// ==================== 公告 9xxxx ====================

	// ==================== 测试 -xxxx ====================
	CodeTest Code = -1
)

var Msg = map[Code]string{
	CodeSuccess:       "ok",
	CodeParamError:    "请求参数错误",
	CodeUnauthorized:  "未登录或Token无效",
	CodeForbidden:     "无权限操作",
	CodeNotFound:      "资源不存在",
	CodeServerError:   "服务器内部错误",
	CodeDatabaseError: "数据库操作失败",
	CodeUnknownError:  "未知错误",
	CodeChenSongError: "陈松出了故障!",

	CodeUserOrPasswordError:  "用户名或密码错误",
	CodeUsernameOccupied:     "用户名已被占用",
	CodeFormInvalid:          "用户名/昵称/密码不符合规则",
	CodeInvalidSigningMethod: "无效的加密方式",
	CodeTokenBanned:          "令牌被禁用",
	CodeUserNotFoundOrBanned: "用户不存在或被禁用",
	CodeQQCodeAlreadyExists:  "验证会话已经存在，请等失效后重试",
	CodeQQCodeError:          "会话的QQ验证码错误",

	CodeTest: "这是开发人员测试",
}
