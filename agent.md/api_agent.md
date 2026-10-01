# LNF-SERVER 前端 AI 对接文档（api_agent）

> 读者：前端 AI / 代码生成代理。本文提供精确的接口契约、字段表、行为语义与陷阱清单，供直接生成对接代码。
> 读者：人类可读版见 `api_guide.md`。Swagger：`http://111.229.234.32:8080/api/v1/swagger/index.html`
> 范围：user、item（含 tag、location、image upload）、notification（站内通知）、shop（积分商城）、announcement（公告，见 2.6）。
> 依据：后端代码现状（2026-10-01），与 handler swagger 注释冲突处以本文实现语义为准。

---

## 0. 全局约定（必须先读）

| 约定 | 内容 |
|---|---|
| Base URL | `http://111.229.234.32:8080` |
| API 前缀 | `/api/v1` |
| 响应信封 | `{"code": int, "message": string, "data": any}` |
| HTTP 状态码 | **恒为 200**。成功判定：`code === 0`。失败：`code !== 0`，`data` 为 `{}` |
| 时间格式 | 请求与响应均为 RFC3339 字符串，如 `"2026-09-26T15:04:05+08:00"` |
| 鉴权 | 请求头 `Authorization: Bearer <token>` |
| token 获取 | `POST /user/login` 成功后从**响应头** `Authorization` 读取（`Bearer xxx`，去掉前缀存储） |
| token 续期 | 任意已鉴权请求的**响应头**可能出现新的 `Authorization`（旧 token 已被禁用）。**必须在响应拦截器中无条件捕获并覆盖本地 token**，否则出现间歇性 `10005` |
| token 有效期 | 24h（buffer_time=6h：签发 6h 后任意请求触发续期换新） |
| 角色数值 | `0` 普通 / `1` 服务管理员 / `2` 系统管理员 |
| 权限不足 | 返回 `code=2`（与未登录同码），不是 403 语义 |
| ⚠️ CORS | 后端**未配置 CORS**。浏览器跨源调用会被预检拦截；前端 dev server 必须配代理或同域部署。且因 token 走自定义响应头，跨域还需服务端 `Access-Control-Expose-Headers: Authorization`（当前缺失） |

### 通用失败码（任意接口可能返回）

`1` 参数错误、`5` 服务器错误、`6` 数据库错误、`2` 未登录/Token 无效/Token 被禁用/权限不足（区分：`10004` 签名方式无效、`10005` token 已禁用）、`7` 未知错误。

---

## 1. 数据模型（响应 data 内的对象结构）

### UserResponse（`/user/me`；`/user/login` 的 data 结构相同）

```json
{
  "id": 1, "username": "alice", "nickname": "小明",
  "realname": null, "gender": null, "qq": null, "avatar": null,
  "role": 0, "status": 1, "credit": 0,
  "last_login_at": "2026-09-26T10:00:00+08:00",
  "created_at": "...", "updated_at": "..."
}
```

- `realname/gender/qq/avatar` 为可空（null）。字段恒输出（无 omitempty）。

### PublicUserResponse（`/user/batch`）

```json
{"id": 1, "nickname": "小明", "gender": null, "avatar": null, "role": 0, "last_login_at": "...", "created_at": "..."}
```

- 不含 username/qq/credit/status。用于物品的发布者信息展示。

### ItemListResponse（`/item/list`、`/item/mine`、`/item/search`）

```json
{"total": 42, "page": 1, "page_size": 10, "items": [ /* ItemResponse[] */ ]}
```

### ItemResponse（列表项与详情同构）

```json
{
  "id": 10, "user_id": 1, "title": "丢失黑色钱包", "description": "..." ,
  "type": 0, "status": 0,
  "locations": [ /* Location[]，根→叶 */ ],
  "location_detail": "A栋201门口",
  "images": [ /* ItemImage[]，sort_order 升序 */ ],
  "tags": [ /* Tag[] */ ],
  "lost_found_time": "2026-09-25T18:00:00+08:00",
  "contact": "wx: xxx",
  "credit_reward": 0,
  "view_count": 3,
  "claim_user_id": null, "claim_time": null,
  "created_at": "...", "updated_at": "..."
}
```

- `locations`、`images`、`tags` 恒为数组（可为 `[]`）。
- `location_detail / contact / claim_user_id / claim_time` 为指针字段，**null 时整个键缺失（omitempty）**——取值需用可选链/默认值。

### ItemImage

```json
{"id": 1, "item_id": 10, "image_url": "/uploads/2026/09/26/uuid.jpg", "sort_order": 1, "created_at": "...", "updated_at": "..."}
```

### Tag

```json
{"id": 2, "name": "证件", "color": "#FF8800", "sort_order": 0, "created_at": "...", "updated_at": "..."}
```

- `color` omitempty：null 时键缺失。

### Location

```json
{"id": 3, "name": "教学楼A", "parent_id": 1, "level": 2, "address": null, "sort_order": 0, "created_at": "...", "updated_at": "..."}
```

- `address` omitempty。`parent_id=0` 为根节点；`level` 从 1 开始。

### UploadImageResponse（`/upload/image`）

```json
{"url": "/uploads/2026/09/26/uuid.jpg"}
```

- **相对 URL**。展示时拼 `${BASE_URL}${url}`；提交给 item images / avatar 时可直接存相对路径（展示时再拼）。

### Notification / NotificationItem（`/notifications`）

```json
// 列表项 NotificationItem
{"id": 1, "admin_id": 0, "type": 3, "title": "认领已确认", "is_read": 0, "created_at": "..."}

// 详情 Notification
{"id": 1, "admin_id": 0, "user_id": 2, "type": 3, "title": "...", "content": "...",
 "related_id": 10, "is_read": 1, "read_at": "...", "created_at": "..."}
```

- `type`：`0` 系统通知 / `1` 物品匹配 / `2` 认领申请 / `3` 认领结果 / `4` 评论回复 / `5` 积分变动 / `6` 商品兑换。
- `related_id` 可空：关联实体 ID（物品 ID / 订单 ID 等），可据此跳转。
- `admin_id=0` 表示系统触发（QQ 绑定、认领关闭、兑换等自动通知）。

### Good / GoodListResponse（shop）

```json
// Good
{"id": 1, "name": "校园卡挂绳", "description": "...", "image_url": "/uploads/...", 
 "price": 20, "stock": 5, "sort_order": 0, "created_at": "...", "updated_at": "..."}

// GoodListResponse
{"total": 10, "page": 1, "page_size": 10, "items": [ /* Good[] */ ]}
```

- `description` / `image_url` omitempty 可空；`image_url` 为相对路径，展示需拼 BASE_URL。

### Order / OrderListResponse / RedeemGoodsResponse（shop）

```json
// Order（快照设计：商品改名/下架不影响历史记录）
{"id": 1, "order_no": "20261001120000365123", "user_id": 2, "goods_id": 1,
 "goods_name": "校园卡挂绳", "price": 20, "qq": "12345678", "nickname": "小明", "created_at": "..."}

// OrderListResponse
{"total": 3, "page": 1, "page_size": 10, "orders": [ /* Order[] */ ]}

// RedeemGoodsResponse
{"order_no": "...", "goods_id": 1, "goods_name": "...", "price": 20, "credit": 80, "created_at": "..."}
```

### AnnouncementListResponse（`/announcement`、`/admin/announcement`）

```json
{"total": 42, "page": 1, "page_size": 10, "announcements": [ /* AnnouncementResponse[] */ ]}
```

### AnnouncementResponse（公告列表项与详情同构）

```json
{
  "id": 1, "admin_id": 2, "title": "系统公告", "content": "…（markdown 正文）",
  "type": 0, "status": 1, "is_top": 0, "view_count": 3,
  "published_at": "2026-10-01T12:00:00+08:00",
  "created_at": "...", "updated_at": "..."
}
```

- `published_at` 可空指针（omitempty）：正常创建即发布后恒有值，键仍可能缺失，类型定义用可选。
- 不含 `is_deleted` 字段；列表翻到底时 `announcements` 为 `[]`（非 null）。

### 枚举

| 字段 | 取值 |
|---|---|
| `item.type` | 0 丢失 / 1 拾到 |
| `item.status` | 0 已发布 / 1 已认领 / 2 已关闭 |
| `user.role` | 0 普通 / 1 服务管理员 / 2 系统管理员 |
| `user.status` | 0 禁用 / 1 正常 |
| `user.gender` | int8，后端不校验取值，语义由前端定义 |
| `add-credit.type` | 0 拾金不昧 / 1 认领成功 / 2 违规扣分 / 3 系统调整 |
| `notification.type` | 0 系统 / 1 物品匹配 / 2 认领申请 / 3 认领结果 / 4 评论回复 / 5 积分变动 / 6 商品兑换 |
| `announcement.type` | 0 系统公告 / 1 活动公告 / 2 维护通知 / 3 其他 |
| `announcement.status` | 1 已发布 / 2 已下架（0 已废弃：无草稿态，历史 0 值行不可见） |
| `goods.is_deleted`（下架） | 不可见即下架，前端无感知 |

---

## 2. 接口明细

### 2.1 user 模块

#### POST `/user/create`（public）注册

请求体：

| 字段 | 类型 | 必填 | 校验 |
|---|---|---|---|
| username | string | ✅ | UTF-8 字节长 2-32，唯一 |
| password | string | ✅ | 字节长 8-20 |
| nickname | string | ✅ | 字节长 2-32 |

- 成功：`data: {}`（不返回用户）。
- 错误：`1` 格式不符、`10002` 用户名占用、`5` 服务器错误。

#### POST `/user/login`（public）登录

请求体：`{"username": "...", "password": "..."}`（同上字节长校验）。

- 成功：`data` = UserResponse；**token 在响应头 Authorization**。
- 错误：`10003` 格式不符、`10001` 凭证错误、`10006` 被禁用、`6` 更新登录时间失败。

#### GET `/user/me`（private）本人信息

- 成功：`data` = UserResponse。错误：`10006`。
- `POST /user/me` 等价可用，响应头带 `Warning` 提示改用 GET。

#### POST `/user/logout`（private）登出

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| logout_all | int | ✅（必须显式传） | `1` 全部会话 / **其他任意值（含 0）= 仅当前会话** |

- 不传该字段 → `1`（pointer+required，`{}` 也不行）。
- 成功后本地 token 应丢弃。`logout_all=1` 时该用户全部 token 失效（`10005`，JWT 版本号 +1）。
#### POST `/user/update`（private）改资料（增量）

| 字段 | 类型 | 说明 |
|---|---|---|
| nickname | string? | 建议前端仍校验 2-32 字节（后端此处不校验） |
| realname | string? | ≤50 字符，超长 → `6` |
| gender | int? | 任意 int8 |
| avatar | string? | ≤255 字符，超长 → `6`；填上传返回的 url |

- **空对象 `{}` → `1`**（handler 判定全空请求非法）。至少传一个字段。
- 成功：`data: {}`。需刷新资料可再调 `/user/me`。

#### POST `/user/batch`（public）批量查公开用户

请求体：`{"ids": [1,2,3]}`（必填字段）。

- 成功：`data` = PublicUserResponse[]。
- 语义：**顺序不保证**、无效 id 被过滤；`ids: []` → `data: []`；传入的 id 全部未命中 → `data: null`（**前端必须兜底为 `[]`**）。

#### POST `/user/qq/get-code`（private）申请 QQ 验证码

请求体：`{"qq": 10000~99999999999}`（number，越界 → `1`）。

- 语义：机器人在指定 QQ 群 @该 QQ 发送 6 位验证码；会话 3 分钟（bind_timeout）。
- 错误：`10007` 已有未失效会话（同一 token 会话或同一 QQ 号）、`10009` 不在群内、`8` 机器人故障、`6` Redis 异常。

#### POST `/user/qq/bind`（private）提交验证码

请求体：`{"qq": number, "code": 100000~999999}`（均 number，越界 → `1`）。

- 语义：与会话中记录比对；尝试计数 ≤3 次（bind_max_tries=3，判定 `tries>3`，即**第 5 次提交才返回 `10011`**，前 4 次均可尝试比对）。
- 错误：`10012` 会话不存在/过期、`10010` qq 与申请时不一致、`10008` 码错、`10011` 次数超限、`10013` 该 QQ 已绑定其他账号或本账号已有 QQ、`10006`、`6`。
- 成功后 `/user/me` 的 `qq` 字段可查；会收到一条系统通知（type=0）。

#### GET `/user/jwt-test`（private，**仅测试用途**）

- 返回 **`code: -1`（CodeTest，非 0！）**，`data: {"jwtID": int, "jwtRole": int, "tokenFresh": false}`。`tokenFresh` 恒 false（占位）。
- **前端业务代码可不接入**；可用于联调期验证 token 有效性/逾期（失效 → `2`/`10005`）。注意以 `code === -1` 判定该接口成功。

#### 管理员（JWT + role=2）

| 接口 | 请求体 | 成功 | 专属错误 |
|---|---|---|---|
| POST `/admin/change-role` | `{"id": int64, "role": 0\|1\|2}` | `{}` | `10003` role 非法 |
| POST `/admin/change-status` | `{"id": int64, "status": 0\|1}` | `{}` | `10003` status 非法。禁用→该用户全部 token 失效；解禁→不影响存量 token |
| POST `/admin/add-credit` | `{"id": int64, "credit": int64, "type": 0\|1\|2\|3, "description"?: string, "operator_id": int64}` | `{}` | `10014` 积分不够扣。`credit=-99999` 清零；credit 可为负但结果下限 0。**`operator_id` 必填且必须等于当前登录管理员的 JWT 身份**：缺失/为 0 → `1`，与登录身份不符 → `2`（不执行） |

- 非 role=2 调用 → `2`。

### 2.2 item 模块

#### GET `/item/list`（public）公开列表

Query（全部可选）：

| 参数 | 类型 | 说明 |
|---|---|---|
| type | 0\|1 | 精确匹配（其他值 → `1`） |
| status | 0\|1\|2 | **缺省 = [0,1]（已发布+已认领）**；查已关闭必须显式 `status=2` |
| location_id | int64 | 精确匹配 |
| tag_id | int64 | 物品包含该标签 |
| keyword | string | title/description LIKE `%kw%` |
| page | int≥1 | 缺省 1（0/负数 → 1） |
| page_size | int | binding 1-100（>100 → `1`）；**0/缺省 → 10；51-100 → 后端静默按 10 返回** |

- 成功：`data` = ItemListResponse；排序 `created_at DESC, id DESC`。
- 绑定失败（如 page=0、page_size=200）→ `1`。
- ⚠️ 分页陷阱：`page_size` 传 51-100 不会报错但**实际按 10 条返回**（响应 `page_size` 字段回显你传的值，items 只有 ≤10 条）。前端固定传 1-50。

#### GET `/item/search`（public）多条件最小匹配检索

计分语义：`match_count = |物品标签 ∩ tag_ids| + (物品地点 ∈ location_ids 中 level=3 的 ? 1 : 0)`，返回 `match_count >= min_match` 的物品。

Query：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| tag_ids | int64[] | — | 逗号分隔 `tag_ids=1,2,3` 或重复参数；去重后不存在的 ID **自动剔除且不计入条件总数** |
| location_ids | int64[] | — | 同上格式；**仅 level=3 的地点生效**，非 level3/不存在的自动忽略且不计入条件总数 |
| min_match | int | ✅ | ≥1，且 ≤ 条件总数（= 有效 tag 数 + 有效 level3 地点数）；相等 = 全部条件必须满足；两组均为空 → `1` |
| type | 0\|1 | — | 精确匹配 |
| status | int8[] | ✅ | **必传可多选**，仅允许 0/1（传 2 → `1`），如 `status=0,1` |
| page | int≥1 | — | 缺省 1 |
| page_size | int | — | 同 `/item/list`：>100 → `1`；51-100 → 静默按 10 |

- 成功：`data` = ItemListResponse（排序 `created_at DESC, id DESC`，仅未删除物品）。
- 错误：`1`（min_match/status 非法、条件总数不足）、`6`。

#### GET `/item/:itemID`（public）详情

- 成功：`data` = ItemResponse；**每次调用 view_count +1**（返回值已含本次 +1）。
- 错误：`20001` 不存在/已删除、`1` 路径 id 非正整数。

#### POST `/item/create`（private）发布

请求体：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| title | string | ✅ | 非空（trim）且 ≤100 字符（按字节），否则 `1` |
| description | string | ✅ | text，不限长 |
| type | 0\|1 | ✅ | 其他值 → `20006` |
| lost_found_time | RFC3339 | ✅ | 缺失/格式错 → `1` |
| location_id | int64? | — | 必须存在 → 否则 `20010` |
| location_detail | string? | — | ≤200 字符，超长 → `1` |
| contact | string? | — | ≤100 字符，超长 → `1` |
| credit_reward | int? | — | ≥0，负数 → `1` |
| tag_ids | int64[]? | — | 任一不存在 → `40001`；自动去重 |

- 成功：`data: {}`，物品 status=0。**不返回新 id**——前端创建后应调 `GET /item/mine?page=1&page_size=1` 取最新一条。

#### POST `/item/update`（private，仅本人）增量编辑

请求体：

| 字段 | 类型 | 语义 |
|---|---|---|
| id | int64 | ✅ 必填 |
| title | string? | trim 非空 ≤100（按字节） |
| description | string? | |
| status | — | **不接受**：请求结构无此字段，传入被忽略；状态流转必须走 claim/confirm/close |
| location_id | int64? | **传 0 = 清除地点**；传正数需存在（否则 `20010`） |
| location_detail | string? | ≤200 字符，超长 → `1` |
| contact | string? | ≤100 字符，超长 → `1` |
| lost_found_time | RFC3339? | |
| credit_reward | int? | ≥0，负数 → `1` |
| tag_ids | int64[]? | **指针语义：字段出现即整体替换**（`[]` = 清空）；不出现 = 不动 |

- **`{}`（无任何有效字段）→ 成功但无操作**（前端应自行拦截空提交）。
- 错误：`20001`、`20005` 非本人、`1`、`20010`、`40001`、`6`。

#### POST `/item/delete`（private，仅本人）

请求体：`{"id": int64}`。软删除（is_deleted=1），详情/列表不可见。错误：`20001`、`20005`。

#### GET `/item/mine`（private）我的发布

- 参数同 `/item/list`，但 **page_size 上限 50**：binding 1-100（>100 → `1`）；51-100 不报错但**静默按 10 返回**（>50 的显式拦截是死代码）。status 缺省同 = [0,1]。
- 返回本人全部未删除物品（含已关闭需显式 `status=2`）。

#### POST `/item/:itemID/images`（private，仅本人）覆盖式设置图片

请求体：

```json
{"images": [{"image_url": "/uploads/2026/09/26/x.jpg", "sort_order": 1}]}
```

| 规则 | 值 | 错误码 |
|---|---|---|
| 数量 | ≤3 | 超出 → `20013` |
| sort_order | 1-3，互不重复 | 越界/重复 → `20014` |
| image_url | trim 非空，≤500 字符 | 违反 → `20014` |
| `images: []` | 合法 = 清空全部图片 | — |
| 缺 images 字段 | 绑定失败 | `1` |

- 其他错误：`20001`、`20005`。
- 建议流程：先 `/upload/image` 拿 url，再整体提交（1,2,3 一次给全）。

#### 认领 / 关闭（状态机）

```
            claim(他人)               confirm(发布者, +积分)
 [0 已发布] ────────────→ [1 已认领] ─────────────────→ [2 已关闭]
     ↑  claim/cancel(双方)    │
     └────────────────────────┘
     close(发布者, 0|1 → 2, 不发积分)
```

| 接口 | 调用者 | 约束 | 成功效果 |
|---|---|---|---|
| POST `/item/:itemID/claim` | 非发布者登录用户 | status=0；账号未禁用；`claim_qq_required=true` 时必须已绑 QQ | status→1，记录 claim_user_id/claim_time |
| POST `/item/:itemID/claim/cancel` | 认领者或发布者 | status=1 | status→0，清空认领字段 |
| POST `/item/:itemID/confirm` | 仅发布者 | status=1 | status→2；受益人 +`claim_credit`（当前配置 10）：type=1 拾到帖→发布者，type=0 丢失帖→认领者；站内通知认领者（type=3） |
| POST `/item/:itemID/close` | 仅发布者 | status∈{0,1} | status→2，**不发积分**；若关闭前有进行中的认领，通知认领者（type=3） |

错误映射：

| 场景 | 码 |
|---|---|
| 物品不存在 / 路径 id 非法 | `20001` / `1` |
| 已关闭（claim/cancel/confirm/close 全部） | `20002` |
| 已被认领（claim） | `20003` |
| 认领自己的物品（claim） | `30004` |
| 无认领记录可撤回/确认（cancel/confirm） | `30002` |
| 非认领者也非发布者撤回（cancel） | `30003` |
| 非发布者 confirm/close | `20005` |
| 未绑定 QQ（当前配置强制，claim） | `30006` |
| 账号被禁用（claim） | `10006` |

> **自动关闭**：后台任务**每 5 分钟扫描**（启动时先扫一次），`claim_time` 超过 `claim_auto_close`（当前 24h）的已认领物品自动按 confirm 语义关闭（含发分），并给发布者与认领者双方发站内通知（type=3）。前端不要假设 status 恒定，列表页返回时以响应为准。

### 2.3 upload 模块

#### POST `/upload/image`（private）

- `multipart/form-data`，字段名 **`file`**，单文件。
- 白名单：`image/jpeg`、`image/png`、`image/webp`（**用文件头嗅探，不看扩展名**）；上限 `5MB`（+1KB multipart 开销余量）。
- 成功：`data: {"url": "/uploads/YYYY/MM/DD/uuid.ext"}`（相对 URL）。
- 错误：`9` 过大、`10` 类型不支持、`11` 保存失败、`1` 无文件字段。
- 静态访问：`GET {BASE_URL}/uploads/...`，无需鉴权。
- 用途：物品图（→ item images 接口）、头像（→ `/user/update` avatar）、商城商品图（管理端录入）。

### 2.4 tag 模块

#### GET `/tag/list`（public）

- `data` = Tag[]（sort_order 升序）。前端发帖选标签、筛选器直接用。

#### 管理员（JWT + role≥1）

| 接口 | 请求体 | 错误 |
|---|---|---|
| POST `/tag/create` | `{"name": string(1-50，trim), "color"?: string, "sort_order"?: int}` | `40004` 名长非法、`40002` 重名、`1` 缺 name |
| POST `/tag/update` | `{"id": int64, "name"?/color?/sort_order?}`（增量） | `40001`、`40004`、`40002` 重名（排除自身） |
| DELETE `/tag/:tagID` | — | `40001`、`40005` 被 item_tags 引用、`1` 路径 id 非法 |

- 非 role≥1 → `2`。

### 2.5 location 模块

#### GET `/location/list`（public）

| Query | 说明 |
|---|---|
| parent_id | 传 `0` = 根节点列表（级联选择器第一级）；传具体 id = 其子节点 |
| level | 按层级过滤 |
| 都不传 | 返回全部 |

- `data` = Location[]（sort_order 升序）。级联加载：先 `parent_id=0`，选中后以该 id 查下一级。

#### GET `/item/:itemID/locations`（public）

- `data` = Location[]（根→叶完整链）。无地点 → `[]`；物品不存在 → `20001`。

#### 管理员（JWT + role≥1）

| 接口 | 请求体 | 错误 |
|---|---|---|
| POST `/location/create` | `{"name": string(1-100，trim), "parent_id"?: int64(0=根,省略=0), "address"?: string, "sort_order"?: int}` | `80001` 父级不存在、`80002` 同父重名、`1` 名空/超长 |
| POST `/location/update` | `{"id": int64, "name"?/parent_id?/address?/sort_order?}`（增量） | `80001` 不存在/新父级不存在、`80002` 改名或移动后与目标父级下已有地点重名、`1` 名空/超长/父级指向自身或其后代 |
| DELETE `/location/:locationID` | — | `80005` 有子地点、`80004` 被 items.location_id 引用、`80001` 不存在、`1` 路径 id 非法 |

- update 改 parent_id 时后端在同一事务中重算整棵子树 level。
- 非 role≥1 → `2`。

### 2.6 announcement 模块（公告）

#### 公开端点

| 接口 | Query | 说明 |
|---|---|---|
| GET `/announcement` | `page?`、`page_size?` | 已发布（status=1 且未删除）公告列表，`id DESC`（新→旧）；`data` = AnnouncementListResponse |
| GET `/announcement/:id` | — | 公告详情（**仅 status=1 可见**）；**每次调用 view_count +1（返回值已含本次 +1）** |

| Query | 类型 | 说明 |
|---|---|---|
| page | int | 缺省/0 → 1；负数 → `1` |
| page_size | int | 缺省/0 → 10；binding 1-100：负数或 >100 → `1`（**上限 100 正常生效，无 item 的 51-100 静默坑**） |

- 详情错误：`90001` 不存在/已下架/已删除、`1` 路径 id 非正整数或非数字。
- 列表无结果时 `data.announcements = []`（非 null），`total` 与列表同口径。

#### 管理端点（JWT + role=2，非 role=2 → `2`）

| 接口 | 请求体 / Query | 说明 |
|---|---|---|
| POST `/admin/announcement/create` | `{title, content, type, is_top}` | **创建即发布**：status=1、published_at=当前时间、admin_id 取自 JWT（不信任 body）。成功 `data: {}`（**不返回 id**，跳详情可先拉列表第一条） |
| POST `/admin/announcement/update` | `{id, title?, content?, type?, status?, is_top?}` | 增量更新（传了才改，指针语义）；status 仅 1已发布/2已下架（下架/重新上架）；`published_at` 保持首次发布时间、`view_count`/`created_at` 不可更新 |
| DELETE `/admin/announcement/:id` | — | 软删（is_deleted=1）：公开/管理列表与详情均不可见 |
| GET `/admin/announcement` | `page?`、`page_size?`、`status?` | 管理列表：**含已下架**（id DESC）；`status` 可选筛选 1/2（缺省 = 1+2 全查，0 废弃行不可见）；`status` 非 1/2 → `1` |

create 请求体字段：

| 字段 | 类型 | 必填 | 校验 |
|---|---|---|---|
| title | string | ✅ | 空字符串 → `1`（binding required）；纯空白或 >100 字节 → `90003` |
| content | string | ✅ | 空字符串 → `1`；纯空白 → `90003`。markdown 正文，不限长 |
| type | 0\|1\|2\|3 | ✅（**0 也要显式传**，缺字段/类型错 → `1`） | 0 系统公告 / 1 活动公告 / 2 维护通知 / 3 其他；越界 → `90003` |
| is_top | 0\|1 | ✅（**0 也要显式传**） | 0 否 / 1 是；越界 → `90003` |

update 字段：`id` 必填（缺失或为 0 → `1`）；其余全部可选、传了才改，取值校验同 create（`status` 传 0/3 → `90003`）；仅 id 的空增量 → 成功但无操作；id 不存在/已删除 → `90001`。

#### 错误码

| 码 | 场景 |
|---|---|
| `1` | 绑定失败（缺字段、JSON 类型错、page/page_size 越界、路径 id 非法、管理列表 status 非 1/2） |
| `2` | 未登录 / 非 role=2 调管理端点 |
| `6` | DB 异常 |
| `90001` | 公告不存在/不可见：update、delete、公开详情（含已下架、已删除） |
| `90003` | 业务校验失败：title/content 空白、title 超 100 字节、type/status/is_top 越界 |

#### 注意事项

- 详情接口使 `view_count +1`，勿在轮询/预取中滥用。
- **无草稿态**（创建即发布）；`status=0` 已废弃——历史 status=0 垃圾行在任何列表/详情中不可见（后端提供清理 SQL）。
- `page_size` 上限 100 正常生效，与 item 的静默行为不同，不要照搬 item 的分页容错假设。
- 公告只会出现已发布内容；`status=2` 已下架仅管理列表可见，公开详情返回 `90001`。

### 2.7 notification 模块（站内通知）

#### GET `/notifications`（private）通知列表

| Query | 类型 | 说明 |
|---|---|---|
| limit | int | 每页条数，缺省 10（≤0 归位 10） |
| offset | int | 偏移量，缺省 0（负数归位 0） |
| type | int? | 按类型筛选 |
| is_read | int? | 0 未读 / 1 已读 |
| admin_id | int64? | 按发布者筛选 |

- `data` = NotificationItem[]，`created_at DESC`。错误：`60004`。

#### GET `/notifications/unread-count`（private）未读数

- `data` = int64。Redis 缓存 5 分钟，标记已读/删除后立即失效重建。错误：`60004`。

#### GET `/notifications/:id`（private）详情（自动已读）

- `data` = Notification；未读则自动标记已读（响应中 `is_read=1`）并清未读数缓存。
- 错误：`60001` 不存在/非本人、`60006` 标记已读失败、`1` 路径 id 非法。

#### PUT `/notifications/read`（private）批量已读

请求体：`{"ids": [1,2,3]}`（必填且非空）。

- 只操作自己的未读记录，静默跳过无关 id；不返回条数。错误：`60006`。

#### DELETE `/notifications`（private）批量删除

请求体：`{"ids": [...]}`（必填且非空）。

- 软删除自己的通知；**跳过"自己发给自己的记录"**（user_id=admin_id，管理端群发给自己的那条不可删）。错误：`60007`。

#### POST `/admin/notifications`（private，role≥1）发送通知

请求体：

```json
{"user_ids": [1,2], "send_to_all": false, "type": 0,
 "title": "string ≤100", "content": "string", "related_id": null}
```

- `user_ids` 与 `send_to_all` 二选一：同传 → `1`；`send_to_all=false` 且 `user_ids` 空/全无效 → `1`。
- 异步执行：**接口立即返回成功，实际写入失败仅记后端日志**（无 `60008` 返回路径）。
- 错误：`1`（参数）、`2`（role 不足）。

### 2.8 shop 模块（积分商城）

#### GET `/shop/goods/list`（public）商品列表

| Query | 说明 |
|---|---|
| keyword | 名称模糊匹配 |
| min_price / max_price | 积分闭区间（含边界），0/不传 = 该侧不限；负数 → `1`；min>max → `1` |
| page | 缺省 1 |
| page_size | 缺省 10；**上限 100 正常生效**（与 item 模块不同，无 50 截断） |

- `data` = GoodListResponse；排序 `sort_order ASC, created_at DESC, id DESC`；仅未下架商品。

#### GET `/shop/goods/:goodsID`（public）商品详情

- `data` = Good。错误：`11001` 不存在/已下架、`1` 路径 id 非法。

#### POST `/shop/goods/:goodsID/redeem`（private）积分兑换

- 单事务完成「条件扣库存（防超卖）+ 扣积分（行锁，不足回滚）+ 写订单快照」。
- 前置要求：**用户必须已绑定 QQ**（未绑/格式非法 → `11005`）。
- 成功：`data` = RedeemGoodsResponse（订单号 + 兑换后剩余积分）；随后收到两条站内通知（type=5 积分变动、type=6 商品兑换）与一条 QQ 群 @ 消息（失败均不影响兑换结果）。发货需联系管理员。
- 错误：`10006` 用户问题、`11005` 未绑 QQ、`11001` 商品不存在/下架、`11002` 库存不足（含并发）、`50001` 积分不足、`6`。

#### GET `/shop/orders`（private）我的兑换记录

| Query | 说明 |
|---|---|
| page / page_size | 缺省 1 / 10；上限 100 正常生效 |

- `data` = OrderListResponse（`created_at DESC`）。订单为快照设计，商品改名/下架不影响历史记录。

#### 管理员（JWT + role≥1）

| 接口 | 请求体 | 错误 |
|---|---|---|
| POST `/shop/goods/create` | `{"name": string(1-100，trim), "description"?: string, "image_url"?: string(≤500), "price": int64(1~1000000), "stock"?: int64(≥0，缺省0), "sort_order"?: int}` | `11003` 名空/超长/**重名**、`11004` price 非法、`1` stock 负数/image_url 空/超长 |
| POST `/shop/goods/update` | `{"id": int64, name?/description?/image_url?/price?/stock?/sort_order?}`（增量） | `11001`、`11003`、`11004`、`1` |
| POST `/shop/goods/delete` | `{"id": int64}` | `11001`。软删即下架，历史订单不受影响 |

- 非 role≥1 → `2`。

---

## 3. 前端陷阱清单（生成代码时逐条核对）

1. **成功判定**用 `code === 0`；HTTP 恒 200，axios 的 catch 捕不到业务错误。唯一例外：`/user/jwt-test` 成功返回 `code: -1`。
2. **响应拦截器必须捕获响应头 Authorization 并覆盖 token**（续期机制）。
3. 图片 URL 相对路径：展示处统一 `imgSrc = BASE_URL + url`；`avatar` 同理。
4. `/user/batch` 结果**无序**且未命中时 `data: null` → 需按 id 建映射并兜底空数组。
5. `omitempty` 字段（contact/location_detail/claim_user_id/claim_time/color/address/description/image_url/related_id/read_at/published_at）**可能整个键缺失**，类型定义全部用可选。
6. `/item/list`、`/item/mine`、`/item/search` 不传 status 返回 0+1；"已关闭"标签页必须显式 `status=2`；`/item/search` 的 status 必传且禁 2。
7. item 三接口 `page_size` 51-100 不报错但静默按 10 返回（响应回显你传的值，以 items 长度为准）；`/shop` 两接口上限 100 正常生效。统一传 ≤50 最稳。
8. item `tag_ids` 是指针语义：编辑表单若允许清空标签，必须**显式传 `[]`**，不能省略字段。
9. item update 空对象静默成功；user update 空对象报 `1`。两处行为不同。
10. `logout_all` 必填：登出请求固定发 `{"logout_all": 0|1}`（值非 1 一律按"仅当前会话"处理）。
11. 创建 item 不返回 id → 创建成功后用 `/item/mine` 第一条回填。
12. 详情请求会使 `view_count+1`，不要在轮询/预取中滥用。
13. 长度校验按 **UTF-8 字节**（username/nickname 2-32，password 8-20；item 标题 ≤100、location_detail ≤200、contact ≤100）：`new Blob([s]).size` 或等价方式校验，中文字符算 3。
14. 上传用 `multipart/form-data` 且字段名 `file`；不要手动设置 `Content-Type` 覆盖 boundary；>5MB 前端先拦。
15. 权限不足（role 不够）返回 `2`，与未登录同码——处理上按"需要重新登录或提升角色"提示，勿一律跳转登录页。
16. claim 前若 `claim_qq_required=true`（当前为 true），可先查 `/user/me` 的 `qq` 字段预判，避免用户填完被 `30006` 打回；商城兑换同样要求已绑 QQ（`11005`）。
17. 认领状态可能被自动关闭任务改变（24h 超时，每 5 分钟扫描），UI 不应对 status 做长期缓存。
18. confirm 与 close 的区别：confirm 发积分、close 不发；发布者"我找回来了（没人认领）"用 close，"他认领并且真的还我了"用 confirm。
19. `/admin/add-credit` 的 `operator_id` 由前端显式传入并做身份校验：必须传当前登录管理员的用户 id（可从 `/user/me` 的 `id` 取），传别人或随意值 → `2` 且不执行。
20. `/item/search` 的条件总数是"剔除计"：不存在/非 level3 的 ID 不计入，`min_match` 超过有效条件总数 → `1`。
21. 通知相关：`admin_id=0` 的通知是系统触发；批量删除会跳过"管理端群发给自己的那条"，前端勿把它做成可勾选。
22. QQ 绑定验证码会话 3 分钟内最多 3 次比对机会，但**第 4 次提交仍会比对、第 5 次才报 `10011`**；`get-code` 对同一 QQ 的限制以 QQ 号维度判重（`10007`）。
23. 公告分页与 item 不同：`/announcement`、`/admin/announcement` 的 `page_size` binding 1-100 **上限正常生效**（>100 或负数 → `1`，无 51-100 静默按 10 的坑）；0/缺省 → 10。
24. 公告 create 的 `type`、`is_top` **必填且 0 也要显式传**（缺字段 → `1`）；update 为指针增量，`{"id": n}` 空增量静默成功；`status` 只有 1/2（0 草稿态已废弃）。

---

## 4. 端到端时序（最小可用路径）

```
1  POST /user/create                     注册
2  POST /user/login                      → 响应头取 token
3  GET  /tag/list                        拉标签筛选项
4  GET  /location/list?parent_id=0       级联第一级
5  POST /upload/image                    (multipart) → {url}
6  POST /item/create                     {title, description, type, lost_found_time, location_id, tag_ids:[...]}
7  GET  /item/mine?page=1&page_size=1    → 取新物品 id
8  POST /item/{id}/images                {images:[{image_url, sort_order:1}]}
9  GET  /item/list?status=0&type=0       公开列表
10 GET  /item/{id}                       详情（含地点链/图片/标签）
11 POST /item/{id}/claim                 他账号认领（需已绑 QQ）
12 POST /item/{id}/confirm               发布者确认 → 对方 +10 分
13 GET  /notifications/unread-count      未读通知（认领结果 type=3）
14 GET  /shop/goods/list                 商城列表
15 POST /shop/goods/{id}/redeem          兑换（需已绑 QQ，扣积分+发通知）
16 GET  /shop/orders                     我的兑换记录
17 GET  /announcement?page=1&page_size=10 公告列表（详情会 +1 浏览量）
18 POST /user/logout                     {"logout_all": 0}
```

---

## 5. 附：模块外错误码（仅列出可能见到的）

| 码 | 含义 | | 码 | 含义 |
|---|---|---|---|---|
| 4 | 资源不存在（预留，当前不返回） | | 10013 | QQ 已被注册/绑定 |
| 7 | 未知错误 | | 10014 | 积分不够 |
| 8 | 陈松（机器人）故障 | | 20002 | 物品已关闭 |
| 10004 | 无效签名方式 | | 20006 | 物品类型非法 |
| 10011 | QQ 尝试过多 | | 50001 | 商城兑换积分不足 |
| 11001 | 商品不存在/已下架 | | 11002 | 商品库存不足 |
| 11003 | 商品名非法/重名 | | 11004 | 商品价格非法 |
| 11005 | 兑换需先绑定 QQ | | 60001 | 通知不存在 |
| 90001 | 公告不存在/不可见 | | 90003 | 公告参数或状态错误 |

完整错误码表见 `common_response_code.md`（与 `response/response_code.go` 对齐）。
