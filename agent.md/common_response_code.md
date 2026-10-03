# LNF-SERVER 统一消息码设计文档

基于 model/mysql 下 items.sql、claims.sql、tags.sql、item_tags.sql 等建表文件设计。

编码规则：五位数字，前两位为业务域，后三位为具体错误。

业务错误统一返回 HTTP 200，由前端根据 code 差异化处理；传输层错误（参数、认证、权限）返回对应 HTTP 状态码。

## 一、业务域分段总览

码段 | 业务域 | 对应表 / 模块
---|---|---
0xxxx | 通用 / 系统 | 全局
1xxxx | 用户与认证 | users / JWT
2xxxx | 物品 | items.sql
3xxxx | 认领 | claims.sql
4xxxx | 标签 & 关联 | tags.sql / item_tags.sql
5xxxx | 积分 | credit_logs
6xxxx | 通知 | notifications
7xxxx | 举报 | reports
8xxxx | 地点 | locations
9xxxx | 公告 | announcements
01xxx | 评论 | comments

## 二、通用 / 系统（0xxxx）

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
00000 | CodeSuccess | Success | 操作成功 | 接口正常返回
00001 | CodeParamError | Invalid Parameter | 请求参数错误 | 参数缺失、类型错误、校验失败
00002 | CodeUnauthorized | Unauthorized | 未登录或 Token 无效 | Token 缺失、伪造、解析失败
00003 | CodeForbidden | Forbidden | 无权限操作 | 当前用户无权访问该资源
00004 | CodeNotFound | Resource Not Found | 资源不存在 | 通用资源查询为空
00005 | CodeServerError | Internal Server Error | 服务器内部错误 | 未捕获的 panic、未知异常
00006 | CodeDatabaseError | Database Operation Failed | 数据库操作失败 | SQL 执行异常、连接失败
00007 | CodeUnknownError | Unknown Error | 未知错误 | 未分类的兜底错误
00008 | CodeChenSongError | ChenSong Error | 陈松出了故障! | 开发人员自定义错误
00009 | CodeUploadFileTooLarge | Upload File Too Large | 上传文件过大 | 文件大小超出限制
00010 | CodeUploadFileTypeInvalid | Upload File Type Invalid | 上传文件类型不支持 | 文件 MIME 类型不在允许列表
00011 | CodeUploadFailed | Upload Failed | 上传文件保存失败 | 文件写入磁盘失败

## 三、用户与认证（1xxxx）

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
10001 | CodeUserOrPasswordError | User Or Password Error | 用户名或密码错误 | 登录校验失败
10002 | CodeUsernameOccupied | Username Already Taken | 用户名已被占用 | 注册 / 修改时重复
10003 | CodeFormInvalid | Form Invalid | 用户名/昵称/密码不符合规则 | 格式校验失败
10004 | CodeInvalidSigningMethod | Invalid Signing Method | 无效的加密方式 | JWT 签名算法不匹配
10005 | CodeTokenBanned | Token Banned | 令牌被禁用 | Token 被加入黑名单
10006 | CodeUserNotFoundOrBanned | User Not Found Or Banned | 用户不存在或被禁用 | 按 user_id 查询无记录或已被禁用
10007 | CodeQQSessionAlreadyExists | QQ Session Already Exists | 同一个QQ的验证会话已经存在，请等失效后重试 | 重复发起 QQ 绑定
10008 | CodeQQCodeError | QQ Code Error | 会话的QQ验证码错误 | 验证码校验失败
10009 | CodeQQUserNotInGroup | QQ User Not In Group | 用户不在QQ群内 | 群成员校验失败
10010 | CodeQQNumberError | QQ Number Error | 需要验证的QQ号不一致 | 提交的 QQ 号与验证会话不匹配
10011 | CodeQQTooManyRequests | QQ Too Many Requests | QQ绑定请求过多，请等冷却后再试 | 限流触发
10012 | CodeQQSessionNotExist | QQ Session Not Exist | QQ绑定会话不存在或已经失效 | 按 session_id 查询无记录
10013 | CodeQQAlreadyRegistered | QQ Already Registered | QQ已被注册或账号已绑定QQ | 重复绑定
10014 | CodeCreditNotEnough | Credit Not Enough | 积分不够 | 积分余额不足

## 四、物品模块（2xxxx）

对应 items.sql：type(0丢失 1拾到)、status(0待审核 1已发布 2已认领 3已关闭)、claim_user_id、credit_reward、lost_found_time。

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
20001 | CodeItemNotFound | Item Not Found | 物品不存在 | 按 id 查询无记录
20002 | CodeItemClosed | Item Already Closed | 物品已关闭 | status=3 时尝试认领 / 评论
20003 | CodeItemAlreadyClaimed | Item Already Claimed | 物品已被认领 | status=2 或 claim_user_id 已有值
20004 | CodeItemPendingAudit | Item Pending Audit | 物品待审核 | status=0 时用户尝试认领
20005 | CodeItemNoPermission | No Permission On Item | 无权操作该物品 | 非发布者尝试编辑 / 关闭
20006 | CodeItemTypeInvalid | Invalid Item Type | 物品类型非法 | type 不属于 0 或 1
20007 | CodeItemCreditNegative | Credit Reward Cannot Be Negative | 积分奖励不能为负数 | credit_reward < 0
20008 | CodeItemTimeEmpty | Lost/Found Time Required | 丢失/拾到时间不能为空 | lost_found_time 为空
20009 | CodeItemTitleEmpty | Item Title Required | 物品标题不能为空 | title 为空
20010 | CodeItemLocationInvalid | Invalid Location | 地点无效 | location_id 不存在或已禁用
20011 | CodeItemAlreadyPublished | Item Already Published | 物品已发布 | 重复发布同一物品
20012 | CodeItemCannotClaimSelf | Cannot Claim Self Item | 不能认领自己发布的物品 | claimer_id == item.user_id
20013 | CodeItemImageTooMany | Item Image Too Many | 物品图片最多3张 | 上传图片数量超出限制
20014 | CodeItemImageInvalid | Item Image Invalid | 物品图片参数错误 | 图片参数校验失败

## 五、认领模块（3xxxx）

对应 claims.sql：status(0待审核 1已通过 2已拒绝 3已取消)。

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
30001 | CodeItemAlreadyClaimed | Item Already Claimed | 物品已被认领 | status=2 或 claim_user_id 已有值
30002 | CodeClaimNotFound | Claim Not Found | 认领记录不存在 | 按 id 查询无记录
30003 | CodeClaimNoPermission | No Permission On Claim | 无权审核该认领申请 | 审核人非物品发布者或管理员
30004 | CodeClaimSelfItem | Cannot Claim Own Item | 不能认领自己发布的物品 | claimer_id == item.user_id
30005 | CodeItemClosed | Item Already Closed | 物品已关闭 | status=3 时尝试认领 / 评论
30006 | CodeClaimQQRequired | Claim QQ Required | 认领前请先绑定QQ | 用户未绑定 QQ 时尝试认领
30007 | CodeClaimDescriptionTooLong | Claim Description Too Long | 认领描述过长 | description 超出业务限制
30008 | CodeClaimDuplicate | Duplicate Claim | 重复提交认领申请 | 同一用户对同一物品重复提交

## 六、标签模块（4xxxx）

对应 tags.sql：name 唯一约束 uk_name、status(0禁用 1启用)。

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
40001 | CodeTagNotFound | Tag Not Found | 标签不存在 | 按 id 查询无记录
40002 | CodeTagDuplicate | Tag Name Already Exists | 标签名称已存在 | 违反 uk_name 唯一约束
40003 | CodeTagDisabled | Tag Disabled | 标签已被禁用 | status=0 时尝试关联物品
40004 | CodeTagNameInvalid | Invalid Tag Name | 标签名称无效 | name 为空或超过 50 字符
40005 | CodeTagInUse | Tag In Use | 标签正在被使用 | 删除标签时仍被物品关联
40010 | CodeItemTagExists | Item Tag Already Exists | 物品已关联该标签 | 违反 uk_item_tag 唯一约束
40011 | CodeItemTagNotFound | Item Tag Relation Not Found | 标签关联记录不存在 | 取消关联时无对应记录

## 七、积分模块（5xxxx）

对应 credit_logs / 用户积分余额。

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
50001 | CodeCreditInsufficient | Insufficient Credit Balance | 积分余额不足 | 可用积分 < 操作所需积分
50002 | CodeCreditLogNotFound | Credit Log Not Found | 积分流水记录不存在 | 按 id 查询无记录
50003 | CodeCreditTypeInvalid | Invalid Credit Type | 积分操作类型非法 | 流水类型不在允许枚举中
50004 | CodeCreditAmountInvalid | Invalid Credit Amount | 积分数量非法 | 数量为 0 或负数（除扣减场景）
50005 | CodeCreditAlreadyRewarded | Credit Already Rewarded | 积分已发放 | 同一物品重复发放积分

## 八、通知模块（6xxxx）

对应 notifications。

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
60001 | CodeNotificationNotFound | Notification Not Found | 通知不存在 | 按 id 查询无记录
60002 | CodeNotificationNoPermission | No Permission On Notification | 无权查看该通知 | 通知不属于当前用户
60003 | CodeNotificationAlreadyRead | Notification Already Read | 通知已读 | 重复标记已读
60004 | CodeNotificationQueryFailed | Notification Query Failed | 通知查询失败 | 列表查询、未读数统计等数据库读取异常
60005 | CodeNotificationCreateFailed | Notification Create Failed | 通知创建失败 | 单条或批量插入通知记录时数据库写入异常
60006 | CodeNotificationUpdateFailed | Notification Update Failed | 通知更新失败 | 标记已读、批量已读等更新操作数据库异常
60007 | CodeNotificationDeleteFailed | Notification Delete Failed | 通知删除失败 | 批量软删除通知时数据库操作异常或事务回滚
60008 | CodeNotificationSendFailed | Notification Send Failed | 通知发送失败 | 发送通知（含全体发送）时查询用户失败或批量插入失败

## 九、举报模块（7xxxx）

对应 reports。

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
70001 | CodeReportNotFound | Report Not Found | 举报记录不存在 | 按 id 查询无记录
70002 | CodeReportDuplicate | Report Already Exists | 已举报过该内容 | 同一用户对同一目标重复举报
70003 | CodeReportSelfContent | Cannot Report Own Content | 不能举报自己的内容 | 举报人 == 内容所有者
70004 | CodeReportAlreadyHandled | Report Already Handled | 举报已被处理 | 重复审核已处理的举报

## 十、地点模块（8xxxx）

对应 locations（items.sql 通过 location_id 关联）。

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
80001 | CodeLocationNotFound | Location Not Found | 地点不存在 | 按 id 查询无记录
80002 | CodeLocationDuplicate | Location Already Exists | 地点名称已存在 | 违反地点名称唯一约束
80003 | CodeLocationDisabled | Location Disabled | 地点已被禁用 | status=0 时尝试关联物品
80004 | CodeLocationInUse | Location In Use | 地点正在被使用 | 删除地点时仍被物品关联
80005 | CodeLocationHasChildren | Location Has Children | 存在子地点，无法删除 | 删除地点时仍有子地点关联

## 十一、公告模块（9xxxx）

对应 announcements：admin_id、title、content、type(0系统公告 1活动公告 2维护通知 3其他)、status(0草稿 1已发布 2已下架)、is_top、published_at、is_deleted。

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
90001 | CodeAnnouncementNotFound | Announcement Not Found | 公告不存在 | 按 id 查询无记录或 is_deleted=1
90002 | CodeAnnouncementNoPermission | No Permision On Announcement | 无权操作该公告 | 当前用户非管理员或无权操作该公告
90003 | CodeAnnouncementInvalid | Invalid Announcement | 公告参数或状态错误 | 标题/内容为空、类型/状态非法、标题过长、发布时间无效、公告已发布/已下架/已删除/长度超过业务限制等

## 十二、评论模块（10xxxx）

对应 comments

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
| 110001 | CodeCommentNotFound | 评论不存在或已删除 | 按 id 查询无记录、is_deleted=1、对已删除评论执行操作 |
| 110002 | CodeCommentContentInvalid | 评论内容非法 | 内容为空、超过长度限制、命中敏感词 |
| 110003 | CodeCommentNoPermission | 无权操作评论 | 非作者编辑/删除、非管理员审核/隐藏/恢复 |
| 110004 | CodeCommentOperationFailed | 评论操作失败 | 创建/更新/删除/恢复/批量删除时数据库异常或事务回滚 |
| 110005 | CodeCommentItemUnavailable | 物品不可评论 | 物品不存在、已关闭、未发布、已删除、被禁用 |
| 110006 | CodeCommentReplyInvalid | 回复无效 | 父评论不存在/已删除、与物品不匹配、回复自己、层级超限 |
| 110007 | CodeCommentAuditInvalid | 审核状态异常 | 待审核时编辑/删除、已隐藏查看/互动、已拒绝、重复审核、无权审核 |
| 110008 | CodeCommentInteractionFailed | 互动失败 | 重复点赞、点赞不存在、点赞自己、重复举报、点赞/取消点赞写入异常 |
| 110009 | CodeCommentQueryInvalid | 查询参数或查询失败 | 分页/排序/筛选参数非法、列表查询、计数统计数据库异常 |
| 110010 | CodeCommentLimitExceeded | 评论受限 | 请求过多、重复评论、用户被禁评、无评论权限、每日评论超限 | 

