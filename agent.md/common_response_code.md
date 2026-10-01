# LNF-SERVER 统一消息码设计文档

依据后端代码现状整理（2026-10-01），与 `response/response_code.go` 逐码对齐；建表文件参考 `model/mysql`。

编码规则：五位数字以内，前两位（或业务段）为业务域，后三位为具体错误；负数为内部测试专用。

**所有接口（含参数错误、未登录、权限不足等一切传输层错误）HTTP 状态码恒为 200**（见 `response/response.go` 的 `Result`，硬编码 `c.JSON(200, ...)`），前端统一以响应体 `code` 区分成败：`code=0` 成功，`code!=0` 失败且 `data` 为 `{}`。

## 一、业务域分段总览

码段 | 业务域 | 对应表 / 模块 | 实现状态
---|---|---|---
0xxxx 以下（0-11） | 通用 / 系统 | 全局 | ✅ 已实现
1xxxx | 用户与认证 | users / JWT | ✅ 已实现
2xxxx | 物品 | items.sql / item_images | ✅ 已实现
3xxxx | 认领 | items.status + claim_user_id（claims.sql 未使用） | ✅ 已实现
4xxxx | 标签 & 关联 | tags.sql / item_tags.sql | ✅ 已实现
5xxxx | 积分 | credit_logs | ✅ 已实现（仅 50001）
6xxxx | 通知 | notifications | ✅ 已实现
7xxxx | 举报 | reports | ❌ 未实现（仅错误码占位）
8xxxx | 地点 | locations | ✅ 已实现
9xxxx | 公告 | announcements | ✅ 已实现（见第十一节）
10xxxx（11001-11005） | 积分商城 | goods / orders | ✅ 已实现
100001 | 机器人 Agent | chensong | ❌ 未接线（占位）
-x（-1/-2/-3） | 内部测试 | chensong 内部 | 内部使用

## 二、通用 / 系统（0-11）

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
0 | CodeSuccess | Success | 操作成功 | 接口正常返回
1 | CodeParamError | Invalid Parameter | 请求参数错误 | 参数缺失、类型错误、binding 校验失败、业务参数非法（标题空/超长、QQ 号越界、page_size>100 等）
2 | CodeUnauthorized | Unauthorized | 未登录或 Token 无效 | Token 缺失/伪造/解析失败；**权限不足同码**（role 不满足管理中间件要求、admin/add-credit 的 operator_id 与登录身份不符）
3 | CodeForbidden | Forbidden | 无权限操作 | 预留（当前无业务路径返回，权限问题统一走 `2`）
4 | CodeNotFound | Resource Not Found | 资源不存在 | 预留（当前无业务路径返回，各模块用自己的 not found 码）
5 | CodeServerError | Internal Server Error | 服务器内部错误 | 用户创建落库失败、JWT 生成失败等
6 | CodeDatabaseError | Database Operation Failed | 数据库操作失败 | SQL 执行异常、连接失败、Redis 读写失败；`/user/update` 字段超长落库失败也返回此码
7 | CodeUnknownError | Unknown Error | 未知错误 | 未分类的兜底错误（当前仅 chensong 内部使用）
8 | CodeChenSongError | ChenSong Error | 陈松出了故障! | QQ 机器人接口调用失败（拉群成员列表失败、发消息失败或状态非 ok）
9 | CodeUploadFileTooLarge | Upload File Too Large | 上传文件过大 | 文件大小超出 storage.max_size（当前 5MB）
10 | CodeUploadFileTypeInvalid | Upload File Type Invalid | 上传文件类型不支持 | 文件头嗅探结果不在 jpg/png/webp 白名单
11 | CodeUploadFailed | Upload Failed | 上传文件保存失败 | 文件读取或落盘失败

## 三、用户与认证（1xxxx）

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
10001 | CodeUserOrPasswordError | User Or Password Error | 用户名或密码错误 | 登录时用户不存在或密码校验失败
10002 | CodeUsernameOccupied | Username Already Taken | 用户名已被占用 | 注册时用户名重复
10003 | CodeFormInvalid | Form Invalid | 用户名/昵称/密码不符合规则 | 登录表单格式不符；`/admin/change-role` role 非 0/1/2；`/admin/change-status` status 非 0/1
10004 | CodeInvalidSigningMethod | Invalid Signing Method | 无效的加密方式 | JWT 签名算法非 HMAC
10005 | CodeTokenBanned | Token Banned | 令牌被禁用 | token 进入黑名单（登出/续期换新后旧 token）；JWT 版本号不匹配（logout_all / 禁用用户）
10006 | CodeUserNotFoundOrBanned | User Not Found Or Banned | 用户不存在或被禁用 | `/user/me`、登录、QQ 绑定、认领物品、商城兑换时用户不存在或 status=0
10007 | CodeQQSessionAlreadyExists | QQ Session Already Exists | 同一个QQ的验证会话已经存在，请等失效后重试 | 同一 jti 会话未失效，或同一 QQ 号存在未失效会话时重复申请验证码
10008 | CodeQQCodeError | QQ Code Error | 会话的QQ验证码错误 | 绑定时验证码比对失败
10009 | CodeQQUserNotInGroup | QQ User Not In Group | 用户不在QQ群内 | 拉取群成员列表后未找到该 QQ
10010 | CodeQQNumberError | QQ Number Error | 需要验证的QQ号不一致 | 绑定时提交的 QQ 号与会话记录不一致
10011 | CodeQQTooManyRequests | QQ Too Many Requests | QQ绑定请求过多，请等冷却后再试 | 会话内验证码尝试次数超过 bind_max_tries（当前 3，即第 5 次提交时触发）
10012 | CodeQQSessionNotExist | QQ Session Not Exist | QQ绑定会话不存在或已经失效 | 绑定时按 jti 查不到验证码会话（未申请或已过期）
10013 | CodeQQAlreadyRegistered | QQ Already Registered | QQ已被注册或账号已绑定QQ | 该 QQ 号已绑定其他账号；或本账号已有 QQ 再绑
10014 | CodeCreditNotEnough | Credit Not Enough | 积分不够 | `/admin/add-credit` 扣减后余额 < 0

## 四、物品模块（2xxxx）

对应 items.sql：type(0丢失 1拾到)、status(0已发布 1已认领 2已关闭)、claim_user_id、credit_reward、lost_found_time。

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
20001 | CodeItemNotFound | Item Not Found | 物品不存在 | 按 id 查询无记录或已删除（详情/编辑/删除/图片/认领/撤回/确认/关闭/地点链）
20002 | CodeItemClosed | Item Already Closed | 物品已关闭 | status=2 时尝试 claim / cancel / confirm / close
20003 | CodeItemAlreadyClaimed | Item Already Claimed | 物品已被认领 | status=1 时尝试 claim（含并发冲突重查后）
20004 | CodeItemPendingAudit | Item Pending Audit | 物品待审核 | 预留（无审核环节，创建即发布）
20005 | CodeItemNoPermission | No Permission On Item | 无权操作该物品 | 非发布者尝试 update / delete / images / confirm / close
20006 | CodeItemTypeInvalid | Invalid Item Type | 物品类型非法 | create 时 type 不属于 0 或 1
20007 | CodeItemCreditNegative | Credit Reward Cannot Be Negative | 积分奖励不能为负数 | 预留（create/update 负 credit_reward 实际返回 `1`）
20008 | CodeItemTimeEmpty | Lost/Found Time Required | 丢失/拾到时间不能为空 | 预留（lost_found_time 必填由 binding 保证，缺失返回 `1`）
20009 | CodeItemTitleEmpty | Item Title Required | 物品标题不能为空 | 预留（标题空/超长实际返回 `1`）
20010 | CodeItemLocationInvalid | Invalid Location | 地点无效 | create/update 时 location_id（>0）不存在
20011 | CodeItemAlreadyPublished | Item Already Published | 物品已发布 | 预留
20012 | CodeItemCannotClaimSelf | Cannot Claim Self Item | 不能认领自己发布的物品 | 预留（实际使用 3xxxx 段的 `30004`）
20013 | CodeItemImageTooMany | Item Image Too Many | 物品图片最多3张 | 设置图片超过 3 张
20014 | CodeItemImageInvalid | Item Image Invalid | 物品图片参数错误 | sort_order 非 1-3 / 重复、image_url 空或超 500 字符

## 五、认领模块（3xxxx）

对应 items.status + claim_user_id / claim_time（claims.sql 建表存在但代码未使用，认领由 items 表字段直接表达）。
注：30001 / 30005 的语义与 2xxxx 中的 20003 / 20002 相同，但编码值不同，Go 常量名不可重复，故常量名加 Claim 前缀。

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
30001 | CodeClaimItemAlreadyClaimed | Item Already Claimed | 物品已被认领 | 预留（claim 已被认领实际返回 `20003`）
30002 | CodeClaimNotFound | Claim Not Found | 认领记录不存在 | cancel / confirm 时物品非 status=1 或无 claim_user_id
30003 | CodeClaimNoPermission | No Permission On Claim | 无权审核该认领申请 | cancel 时操作者既非认领者也非发布者
30004 | CodeClaimSelfItem | Cannot Claim Own Item | 不能认领自己发布的物品 | claim 时 claimer_id == item.user_id
30005 | CodeClaimItemClosed | Item Already Closed | 物品已关闭 | 预留（已关闭统一返回 `20002`）
30006 | CodeClaimQQRequired | Claim QQ Required | 认领前请先绑定QQ | server.claim_qq_required=true（当前配置）且用户未绑定 QQ 时 claim
30007 | CodeClaimDescriptionTooLong | Claim Description Too Long | 认领描述过长 | 预留（认领无描述字段）
30008 | CodeClaimDuplicate | Duplicate Claim | 重复提交认领申请 | 预留（status=1 时再 claim 走 `20003`）

## 六、标签 & 关联模块（4xxxx）

对应 tags.sql：name 唯一约束 uk_name；item_tags.sql：唯一约束 uk_item_tag（关联由代码去重保证，不触发唯一冲突）。

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
40001 | CodeTagNotFound | Tag Not Found | 标签不存在 | item create/update 的 tag_ids 含不存在 ID；tag update/delete 按 id 无记录
40002 | CodeTagDuplicate | Tag Name Already Exists | 标签名称已存在 | tag create 重名；update 改名与其他标签重名（排除自身）
40003 | CodeTagDisabled | Tag Disabled | 标签已被禁用 | 预留（tags 无 status 字段）
40004 | CodeTagNameInvalid | Invalid Tag Name | 标签名称无效 | tag create/update 名称为空或超过 50 字符
40005 | CodeTagInUse | Tag In Use | 标签正在被使用 | 删除标签时仍被 item_tags 关联
40010 | CodeItemTagExists | Item Tag Already Exists | 物品已关联该标签 | 预留（代码入关联前自动去重）
40011 | CodeItemTagNotFound | Item Tag Relation Not Found | 标签关联记录不存在 | 预留（取消关联为整体替换语义，无单条查询）

## 七、积分模块（5xxxx）

对应 credit_logs（type：0拾金不昧奖励 1认领成功奖励 2违规扣分 3系统调整 4积分兑换）/ users.credit。

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
50001 | CodeCreditInsufficient | Insufficient Credit Balance | 积分余额不足 | 商城兑换时余额 < 商品价格（快速失败或事务内行锁判定）
50002 | CodeCreditLogNotFound | Credit Log Not Found | 积分流水记录不存在 | 预留（无流水查询接口）
50003 | CodeCreditTypeInvalid | Invalid Credit Type | 积分操作类型非法 | 预留（type 不校验直接落库）
50004 | CodeCreditAmountInvalid | Invalid Credit Amount | 积分数量非法 | 预留（数量边界由各调用方自行校验）
50005 | CodeCreditAlreadyRewarded | Credit Already Rewarded | 积分已发放 | 预留（认领发分由条件更新保证幂等）

## 八、通知模块（6xxxx）

对应 notifications（type：0系统通知 1物品匹配 2认领申请 3认领结果 4评论回复 5积分变动 6商品兑换）。
系统触发点（2026-09-30 接入）：QQ 绑定成功(type=0)、认领确认/超时/自行关闭(type=3)、商城兑换(type=5、6)。

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
60001 | CodeNotificationNotFound | Notification Not Found | 通知不存在 | 详情查询无记录或不属于当前用户
60002 | CodeNotificationNoPermission | No Permission On Notification | 无权查看该通知 | 预留（越权查询已合并进 `60001`）
60003 | CodeNotificationAlreadyRead | Notification Already Read | 通知已读 | 预留（重复已读静默成功，只更新未读记录）
60004 | CodeNotificationQueryFailed | Notification Query Failed | 通知查询失败 | 列表查询、未读数统计数据库读取异常
60005 | CodeNotificationCreateFailed | Notification Create Failed | 通知创建失败 | 预留（系统触发点创建失败仅记日志，不影响主流程）
60006 | CodeNotificationUpdateFailed | Notification Update Failed | 通知更新失败 | 详情自动已读 / 批量已读更新异常
60007 | CodeNotificationDeleteFailed | Notification Delete Failed | 通知删除失败 | 批量软删除异常
60008 | CodeNotificationSendFailed | Notification Send Failed | 通知发送失败 | 预留（管理侧发送为异步执行，失败仅记日志）

## 九、举报模块（7xxxx）

对应 reports.sql（建表存在，业务代码未实现）。

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
70001 | CodeReportNotFound | Report Not Found | 举报记录不存在 | 预留（模块未实现）
70002 | CodeReportDuplicate | Report Already Exists | 已举报过该内容 | 预留（模块未实现）
70003 | CodeReportSelfContent | Cannot Report Own Content | 不能举报自己的内容 | 预留（模块未实现）
70004 | CodeReportAlreadyHandled | Report Already Handled | 举报已被处理 | 预留（模块未实现）

## 十、地点模块（8xxxx）

对应 locations（树形：parent_id / level，硬删除，无 is_deleted、无 status）。

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
80001 | CodeLocationNotFound | Location Not Found | 地点不存在 | create 时父级不存在；update/delete 时按 id 无记录
80002 | CodeLocationDuplicate | Location Already Exists | 地点名称已存在 | create 时同父级下重名；update 改名或移动父级后与目标父级下已有地点重名（2026-10-01 补齐 update 校验）
80003 | CodeLocationDisabled | Location Disabled | 地点已被禁用 | 预留（locations 无 status 字段）
80004 | CodeLocationInUse | Location In Use | 地点正在被使用 | 删除地点时仍被未删除物品的 location_id 引用
80005 | CodeLocationHasChildren | Location Has Children | 存在子地点，无法删除 | 删除地点时仍有子地点

## 十一、公告模块（9xxxx）

对应 announcements：admin_id、title、content、type(0系统公告 1活动公告 2维护通知 3其他)、status(1已发布 2已下架，0已废弃不迁移——历史 status=0 垃圾行在任何列表/详情中均不可见)、is_top、view_count、published_at、is_deleted。

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
90001 | CodeAnnouncementNotFound | Announcement Not Found | 公告不存在 | 按 id 查不到未删除公告，或（公开详情时）公告非已发布状态：`/admin/announcement/update`、`/admin/announcement/{id}` DELETE、公开详情 `GET /announcement/{id}`（已下架/已删除同样返回此码）
90002 | CodeAnnouncementNoPermission | No Permission On Announcement | 无权操作该公告 | 预留（管理接口权限由中间件返回 `2` 拦截）
90003 | CodeAnnouncementInvalid | Invalid Announcement | 公告参数或状态错误 | create/update 业务校验：title trim 后为空或超 100 字节、content trim 后为空、type 不在 0-3、status 不在 1-2（下架/重新上架以外的值）、is_top 不在 0/1

## 十二、积分商城模块（11xxxx）

对应 goods（is_deleted 软删即下架）/ orders（快照设计，无软删）。

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
11001 | CodeGoodsNotFound | Goods Not Found | 商品不存在 | 详情/更新/下架/兑换时商品不存在或已下架
11002 | CodeGoodsStockNotEnough | Goods Stock Not Enough | 商品库存不足 | 兑换时库存 ≤ 0（含事务内并发扣减失败）
11003 | CodeGoodsNameInvalid | Invalid Goods Name | 商品名称无效 | 创建/改名时名称空或超 100 字符；**未下架集合内重名也返回此码**（无独立重名码）
11004 | CodeGoodsPriceInvalid | Invalid Goods Price | 商品价格无效 | price 不在 1~1000000 区间
11005 | CodeShopQQRequired | Shop QQ Required | 使用商城功能需先绑定QQ | 兑换时用户未绑定 QQ 或 QQ 号格式非法（非 5~11 位纯数字）

## 十三、机器人（100001）与内部测试（-x）

错误码 | 英文常量 | 英文含义 | 中文含义 | 触发场景
---|---|---|---|---
100001 | CodeOpenAIError | Agent Error | Agent寄了 | 预留（chensong AI 翻译未接线，当前无路径返回）
-1 | CodeTest | Test | 测试码 | `GET /user/jwt-test` 的响应 code（注意：非 0，前端按 code=0 判成功会视为失败）；chensong 内部
-2 | CodeNoNeed | No Need | 无需处理 | chensong 内部
-3 | CodeNotMessage | Not Message | 不是message类型 | chensong 内部
