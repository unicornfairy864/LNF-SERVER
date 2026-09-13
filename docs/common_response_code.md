# LNF-SERVER 统一消息码设计文档

> 基于 `model/mysql` 下 `items.sql`、`claims.sql`、`tags.sql`、`item_tags.sql` 等建表文件设计。
> 编码规则：**五位数字，前两位为业务域，后三位为具体错误**。
> 业务错误统一返回 `HTTP 200`，由前端根据 `code` 差异化处理；传输层错误（参数、认证、权限）返回对应 HTTP 状态码。

---

## 一、业务域分段总览

| 码段 | 业务域 | 对应表 / 模块 |
|---|---|---|
| `0xxxx` | 通用 / 系统 | 全局 |
| `1xxxx` | 用户与认证 | `users` / JWT |
| `2xxxx` | 物品 | `items.sql` |
| `3xxxx` | 认领 | `claims.sql` |
| `4xxxx` | 标签 & 关联 | `tags.sql` / `item_tags.sql` |
| `5xxxx` | 积分 | `credit_logs` |
| `6xxxx` | 通知 | `notifications` |
| `7xxxx` | 举报 | `reports` |
| `8xxxx` | 地点 | `locations` |

---

## 二、通用 / 系统（0xxxx）

| 错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景 |
|---|---|---|---|---|
| 00000 | `CodeSuccess` | Success | 操作成功 | 接口正常返回 |
| 00001 | `CodeParamError` | Invalid Parameter | 请求参数错误 | 参数缺失、类型错误、校验失败 |
| 00002 | `CodeUnauthorized` | Unauthorized | 未登录或 Token 无效 | Token 缺失、伪造、解析失败 |
| 00003 | `CodeForbidden` | Forbidden | 无权限操作 | 当前用户无权访问该资源 |
| 00004 | `CodeNotFound` | Resource Not Found | 资源不存在 | 通用资源查询为空 |
| 00005 | `CodeServerError` | Internal Server Error | 服务器内部错误 | 未捕获的 panic、未知异常 |
| 00006 | `CodeDatabaseError` | Database Operation Failed | 数据库操作失败 | SQL 执行异常、连接失败 |
| 00007 | `CodeTooManyRequests` | Too Many Requests | 请求过于频繁 | 触发限流、防刷策略 |
| 00008 | `CodeOperationFailed` | Operation Failed | 操作失败 | 通用兜底失败提示 |

---

## 三、用户与认证（1xxxx）

| 错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景 |
|---|---|---|---|---|
| 10001 | `CodeUserNotFound` | User Not Found | 用户不存在 | 按 `user_id` 查询无记录 |
| 10002 | `CodeUsernameOccupied` | Username Already Taken | 用户名已被占用 | 注册 / 修改时重复 |
| 10003 | `CodePasswordWrong` | Incorrect Password | 密码错误 | 登录密码校验失败 |
| 10004 | `CodeTokenExpired` | Token Expired | Token 已过期 | JWT 超过有效期 |
| 10005 | `CodeUserDisabled` | Account Disabled | 账户已被禁用 | 管理员禁用该用户 |
| 10006 | `CodeOldPasswordWrong` | Old Password Incorrect | 原密码错误 | 修改密码时原密码校验失败 |
| 10007 | `CodeEmailOccupied` | Email Already Taken | 邮箱已被占用 | 注册 / 绑定邮箱时重复 |
| 10008 | `CodePhoneOccupied` | Phone Already Taken | 手机号已被占用 | 注册 / 绑定手机号时重复 |
| 10009 | `CodeCaptchaWrong` | Captcha Incorrect | 验证码错误 | 图形 / 短信验证码校验失败 |
| 10010 | `CodeUserNotActive` | User Not Activated | 用户未激活 | 账号未完成邮箱 / 手机激活 |

---

## 四、物品模块（2xxxx）

> 对应 `items.sql`：`type`(0丢失 1拾到)、`status`(0待审核 1已发布 2已认领 3已关闭)、`claim_user_id`、`credit_reward`、`lost_found_time`。

| 错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景 |
|---|---|---|---|---|
| 20001 | `CodeItemNotFound` | Item Not Found | 物品不存在 | 按 `id` 查询无记录 |
| 20002 | `CodeItemClosed` | Item Already Closed | 物品已关闭 | `status=3` 时尝试认领 / 评论 |
| 20003 | `CodeItemAlreadyClaimed` | Item Already Claimed | 物品已被认领 | `status=2` 或 `claim_user_id` 已有值 |
| 20004 | `CodeItemPendingAudit` | Item Pending Audit | 物品待审核 | `status=0` 时用户尝试认领 |
| 20005 | `CodeItemNoPermission` | No Permission On Item | 无权操作该物品 | 非发布者尝试编辑 / 关闭 |
| 20006 | `CodeItemTypeInvalid` | Invalid Item Type | 物品类型非法 | `type` 不属于 0 或 1 |
| 20007 | `CodeItemCreditNegative` | Credit Reward Cannot Be Negative | 积分奖励不能为负数 | `credit_reward < 0` |
| 20008 | `CodeItemTimeEmpty` | Lost/Found Time Required | 丢失/拾到时间不能为空 | `lost_found_time` 为空 |
| 20009 | `CodeItemTitleEmpty` | Item Title Required | 物品标题不能为空 | `title` 为空 |
| 20010 | `CodeItemLocationInvalid` | Invalid Location | 地点无效 | `location_id` 不存在或已禁用 |
| 20011 | `CodeItemAlreadyPublished` | Item Already Published | 物品已发布 | 重复发布同一物品 |
| 20012 | `CodeItemCannotClaimSelf` | Cannot Claim Self Item | 不能认领自己发布的物品 | `claimer_id == item.user_id` |

---

## 五、认领模块（3xxxx）

> 对应 `claims.sql`：`status`(0待审核 1已通过 2已拒绝 3已取消)。

| 错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景 |
|---|---|---|---|---|
| 30001 | `CodeClaimNotFound` | Claim Not Found | 认领记录不存在 | 按 `id` 查询无记录 |
| 30002 | `CodeClaimAlreadyExists` | Claim Already Exists | 该物品已有待审核认领申请 | 同一 `item_id` 存在 `status=0` 记录 |
| 30003 | `CodeClaimAlreadyHandled` | Claim Already Handled | 认领申请已被处理 | `status` 不为 0（非待审核） |
| 30004 | `CodeClaimCancelled` | Claim Already Cancelled | 认领申请已取消 | `status=3` 时尝试审核 |
| 30005 | `CodeClaimNoPermission` | No Permission On Claim | 无权审核该认领申请 | 审核人非物品发布者或管理员 |
| 30006 | `CodeClaimSelfItem` | Cannot Claim Own Item | 不能认领自己发布的物品 | `claimer_id == item.user_id` |
| 30007 | `CodeClaimDescriptionTooLong` | Claim Description Too Long | 认领描述过长 | `description` 超出业务限制 |
| 30008 | `CodeClaimDuplicate` | Duplicate Claim | 重复提交认领申请 | 同一用户对同一物品重复提交 |

---

## 六、标签模块（4xxxx）

> 对应 `tags.sql`：`name` 唯一约束 `uk_name`、`status`(0禁用 1启用)。

| 错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景 |
|---|---|---|---|---|
| 40001 | `CodeTagNotFound` | Tag Not Found | 标签不存在 | 按 `id` 查询无记录 |
| 40002 | `CodeTagDuplicate` | Tag Name Already Exists | 标签名称已存在 | 违反 `uk_name` 唯一约束 |
| 40003 | `CodeTagDisabled` | Tag Disabled | 标签已被禁用 | `status=0` 时尝试关联物品 |
| 40004 | `CodeTagNameInvalid` | Invalid Tag Name | 标签名称无效 | `name` 为空或超过 50 字符 |
| 40005 | `CodeTagInUse` | Tag In Use | 标签正在被使用 | 删除标签时仍被物品关联 |
| 40010 | `CodeItemTagExists` | Item Tag Already Exists | 物品已关联该标签 | 违反 `uk_item_tag` 唯一约束 |
| 40011 | `CodeItemTagNotFound` | Item Tag Relation Not Found | 标签关联记录不存在 | 取消关联时无对应记录 |

---

## 七、积分模块（5xxxx）

> 对应 `credit_logs` / 用户积分余额。

| 错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景 |
|---|---|---|---|---|
| 50001 | `CodeCreditInsufficient` | Insufficient Credit Balance | 积分余额不足 | 可用积分 < 操作所需积分 |
| 50002 | `CodeCreditLogNotFound` | Credit Log Not Found | 积分流水记录不存在 | 按 `id` 查询无记录 |
| 50003 | `CodeCreditTypeInvalid` | Invalid Credit Type | 积分操作类型非法 | 流水类型不在允许枚举中 |
| 50004 | `CodeCreditAmountInvalid` | Invalid Credit Amount | 积分数量非法 | 数量为 0 或负数（除扣减场景） |
| 50005 | `CodeCreditAlreadyRewarded` | Credit Already Rewarded | 积分已发放 | 同一物品重复发放积分 |

---

## 八、通知模块（6xxxx）

> 对应 `notifications`。

| 错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景 |
|---|---|---|---|---|
| 60001 | `CodeNotificationNotFound` | Notification Not Found | 通知不存在 | 按 `id` 查询无记录 |
| 60002 | `CodeNotificationNoPermission` | No Permission On Notification | 无权查看该通知 | 通知不属于当前用户 |
| 60003 | `CodeNotificationAlreadyRead` | Notification Already Read | 通知已读 | 重复标记已读 |

---

## 九、举报模块（7xxxx）

> 对应 `reports`。

| 错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景 |
|---|---|---|---|---|
| 70001 | `CodeReportNotFound` | Report Not Found | 举报记录不存在 | 按 `id` 查询无记录 |
| 70002 | `CodeReportDuplicate` | Report Already Exists | 已举报过该内容 | 同一用户对同一目标重复举报 |
| 70003 | `CodeReportSelfContent` | Cannot Report Own Content | 不能举报自己的内容 | 举报人 == 内容所有者 |
| 70004 | `CodeReportAlreadyHandled` | Report Already Handled | 举报已被处理 | 重复审核已处理的举报 |

---

## 十、地点模块（8xxxx）

> 对应 `locations`（`items.sql` 通过 `location_id` 关联）。

| 错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景 |
|---|---|---|---|---|
| 80001 | `CodeLocationNotFound` | Location Not Found | 地点不存在 | 按 `id` 查询无记录 |
| 80002 | `CodeLocationDuplicate` | Location Already Exists | 地点名称已存在 | 违反地点名称唯一约束 |
| 80003 | `CodeLocationDisabled` | Location Disabled | 地点已被禁用 | `status=0` 时尝试关联物品 |
| 80004 | `CodeLocationInUse` | Location In Use | 地点正在被使用 | 删除地点时仍被物品关联 |
