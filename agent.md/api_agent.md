# LNF-SERVER 前端 AI 对接文档（api_agent）

> 读者：前端 AI / 代码生成代理。本文提供精确的接口契约、字段表、行为语义与陷阱清单，供直接生成对接代码。
> 读者：人类可读版见 `api_guide.md`。Swagger：`http://111.229.234.32:8080/api/v1/swagger/index.html`
> 范围：user、item（含 tag、location、image upload）。公告模块不在范围。
> 依据：后端代码现状（2026-09-28），与 handler swagger 注释冲突处以本文实现语义为准。

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
| 角色数值 | `0` 普通 / `1` 服务管理员 / `2` 系统管理员 |
| 权限不足 | 返回 `code=2`（与未登录同码），不是 403 语义 |
| ⚠️ CORS | 后端**未配置 CORS**。浏览器跨源调用会被预检拦截；前端 dev server 必须配代理或同域部署。且因 token 走自定义响应头，跨域还需服务端 `Access-Control-Expose-Headers: Authorization`（当前缺失） |

### 通用失败码（任意接口可能返回）

`1` 参数错误、`5` 服务器错误、`6` 数据库错误、`2` 未登录/Token 无效/Token 被禁用（区分：`10004` 签名方式无效、`10005` token 已禁用）、`7` 未知错误。

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

### ItemListResponse（`/item/list`、`/item/mine`）

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

- **相对 URL**。展示/入库展示时拼 `${BASE_URL}${url}`；提交给 item images / avatar 时可直接存相对路径（展示时再拼）。

### 枚举

| 字段 | 取值 |
|---|---|
| `item.type` | 0 丢失 / 1 拾到 |
| `item.status` | 0 已发布 / 1 已认领 / 2 已关闭 |
| `user.role` | 0 普通 / 1 服务管理员 / 2 系统管理员 |
| `user.status` | 0 禁用 / 1 正常 |
| `user.gender` | int8，后端不校验取值，语义由前端定义 |
| `add-credit.type` | 0 拾金不昧 / 1 认领成功 / 2 违规扣分 / 3 系统调整 |

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
- 错误：`10003` 格式不符、`10001` 凭证错误、`10006` 被禁用。

#### GET `/user/me`（private）本人信息

- 成功：`data` = UserResponse。错误：`10006`。
- `POST /user/me` 等价可用，响应头带 `Warning` 提示改用 GET。

#### POST `/user/logout`（private）登出

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| logout_all | int | ✅（必须显式传） | 0 仅当前会话 / 1 全部会话 |

- 不传该字段 → `1`（pointer+required，`{}` 也不行）。
- 成功后本地 token 应丢弃。

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

请求体：`{"ids": [1,2,3]}`（必填字段；`[]` 合法）。

- 成功：`data` = PublicUserResponse[]。
- 语义：**顺序不保证**、无效 id 被过滤；全部无效/未命中时 `data` 可能为 **null**（**前端必须兜底为 `[]`**）。

#### POST `/user/qq/get-code`（private）申请 QQ 验证码

请求体：`{"qq": 10000~99999999999}`（number，越界 → `1`）。

- 语义：机器人在指定 QQ 群 @该 QQ 发送 6 位验证码；会话 3 分钟（bind_timeout）。
- 错误：`10007` 已有未失效会话、`10009` 不在群内、`8` 机器人故障、`6`。

#### POST `/user/qq/bind`（private）提交验证码

请求体：`{"qq": number, "code": 100000~999999}`（均 number，越界 → `1`）。

- 语义：与会话中记录比对；尝试计数 ≤3 次（bind_max_tries）。
- 错误：`10008` 码错、`10010` qq 与申请时不一致、`10011` 次数超限、`10012` 会话不存在/过期、`10013` 该 QQ 已绑定或本账号已有 QQ、`10006`。
- 成功后 `/user/me` 的 `qq` 字段可查。

#### GET `/user/jwt-test`（private，**仅测试用途**）

- 返回 `data: {"jwtID": int, "jwtRole": int, "tokenFresh": false}`。`tokenFresh` 恒 false（占位）。
- **前端业务代码可不接入**；可用于联调期验证 token 有效性/逾期（失效 → `2`/`10005`）。

#### 管理员（JWT + role=2）

| 接口 | 请求体 | 成功 | 专属错误 |
|---|---|---|---|
| POST `/admin/change-role` | `{"id": int64, "role": 0\|1\|2}` | `{}` | `10003` role 非法 |
| POST `/admin/change-status` | `{"id": int64, "status": 0\|1}` | `{}` | `10003` status 非法。禁用→该用户全部 token 失效；解禁→版本号+1 |
| POST `/admin/add-credit` | `{"id": int64, "credit": int64, "type": 0\|1\|2\|3, "description"?: string, "operator_id": int64}` | `{}` | `10014` 积分不够扣。`credit=-99999` 清零；credit 可为负但结果下限 0。**`operator_id` 必填且必须等于当前登录管理员的 JWT 身份**：缺失/为 0 → `1`，与登录身份不符 → `2`（不执行） |

- 非 role=2 调用 → `2`。

### 2.2 item 模块

#### GET `/item/list`（public）公开列表

Query（全部可选）：

| 参数 | 类型 | 说明 |
|---|---|---|
| type | 0\|1 | 精确匹配 |
| status | 0\|1\|2 | **缺省 = [0,1]（已发布+已认领）**；查已关闭必须显式 `status=2` |
| location_id | int64 | 精确匹配 |
| tag_id | int64 | 物品包含该标签 |
| keyword | string | title/description LIKE `%kw%` |
| page | int≥1 | 缺省 1 |
| page_size | int 1-100 | 缺省 10 |

- 成功：`data` = ItemListResponse；排序 `created_at DESC, id DESC`。
- 绑定失败（如 page=0）→ `1`。

#### GET `/item/:itemID`（public）详情

- 成功：`data` = ItemResponse；**每次调用 view_count +1**（返回值已含本次 +1）。
- 错误：`20001` 不存在/已删除。

#### POST `/item/create`（private）发布

请求体：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| title | string | ✅ | 非空（trim）且 ≤100 字符，否则 `1` |
| description | string | ✅ | text，不限长 |
| type | 0\|1 | ✅ | 其他值 → `20006` |
| lost_found_time | RFC3339 | ✅ | 缺失/格式错 → `1` |
| location_id | int64? | — | 必须存在 → 否则 `20010` |
| location_detail | string? | — | ≤200 字符 |
| contact | string? | — | ≤100 字符 |
| credit_reward | int? | — | ≥0 |
| tag_ids | int64[]? | — | 任一不存在 → `40001`；自动去重 |

- 成功：`data: {}`，物品 status=0。**不返回新 id**——前端创建后应调 `GET /item/mine?page=1&page_size=1` 取最新一条。

#### POST `/item/update`（private，仅本人）增量编辑

请求体：

| 字段 | 类型 | 语义 |
|---|---|---|
| id | int64 | ✅ 必填 |
| title | string? | trim 非空 ≤100 |
| description | string? | |
| status | 0\|1\|2? | 可直接改状态（但优先走 claim/confirm/close 流程） |
| location_id | int64? | **传 0 = 清除地点**；传正数需存在 |
| location_detail / contact | string? | |
| lost_found_time | RFC3339? | |
| credit_reward | int? | ≥0 |
| tag_ids | int64[]? | **指针语义：字段出现即整体替换**（`[]` = 清空）；不出现 = 不动 |

- **`{}`（无任何有效字段）→ 成功但无操作**（前端应自行拦截空提交）。
- 错误：`20001`、`20005` 非本人、`1`、`20010`、`40001`、`6`。

#### POST `/item/delete`（private，仅本人）

请求体：`{"id": int64}`。软删除（is_deleted=1），详情/列表不可见。错误：`20001`、`20005`。

#### GET `/item/mine`（private）我的发布

- 参数同 `/item/list`，但：**page_size 上限 50**（51-100 → `1`）；status 缺省同 = [0,1]。
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
| POST `/item/:itemID/confirm` | 仅发布者 | status=1 | status→2；受益人 +`claim_credit`（当前配置 10）：type=1 拾到帖→发布者，type=0 丢失帖→认领者 |
| POST `/item/:itemID/close` | 仅发布者 | status∈{0,1} | status→2，**不发积分** |

错误映射：

| 场景 | 码 |
|---|---|
| 物品不存在 | `20001` |
| 已关闭 | claim→`20002`；cancel/confirm→`30005`；close→`20002` |
| 已被认领 | claim→`20003`（或 `30001`） |
| 认领自己的物品 | `30004` |
| 无认领记录可撤回/确认 | `30002` |
| 非认领者也非发布者撤回 | `30003` |
| 非发布者 confirm/close | `20005` |
| 未绑定 QQ（当前配置强制） | `30006` |
| 账号被禁用 | `10006` |

> **自动关闭**：后端定时任务每批扫描，`claim_time` 超过 `claim_auto_close`（当前 24h）的已认领物品自动 confirm（含发分）。前端不要假设 status 恒定，列表页返回时以响应为准。

### 2.3 upload 模块

#### POST `/upload/image`（private）

- `multipart/form-data`，字段名 **`file`**，单文件。
- 白名单：`image/jpeg`、`image/png`、`image/webp`（**用文件头嗅探，不看扩展名**）；上限 `5MB`（+1KB multipart 开销余量）。
- 成功：`data: {"url": "/uploads/YYYY/MM/DD/uuid.ext"}`（相对 URL）。
- 错误：`9` 过大、`10` 类型不支持、`11` 保存失败、`1` 无文件字段。
- 静态访问：`GET {BASE_URL}/uploads/...`，无需鉴权。
- 用途：物品图（→ item images 接口）、头像（→ `/user/update` avatar）。

### 2.4 tag 模块

#### GET `/tag/list`（public）

- `data` = Tag[]（sort_order 升序）。前端发帖选标签、筛选器直接用。

#### 管理员（JWT + role≥1）

| 接口 | 请求体 | 错误 |
|---|---|---|
| POST `/tag/create` | `{"name": string(1-50), "color"?: string(十六进制), "sort_order"?: int}` | `40002` 重名、`40004` 名长非法、`1` |
| POST `/tag/update` | `{"id": int64, "name"?/color?/sort_order?}`（增量） | `40001`、`40002` |
| DELETE `/tag/:tagID` | — | `40001`、`40005` 被 item_tags 引用 |

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
| POST `/location/create` | `{"name": string, "parent_id"?: int64(0=根,省略=0), "address"?: string, "sort_order"?: int}` | `80001` 父级不存在、`80002` 同父重名 |
| POST `/location/update` | `{"id": int64, "name"?/parent_id?/address?/sort_order?}`（增量；改 parent 校验非自身非后代并重算子树 level） | `80001`、`80002`、`1` |
| DELETE `/location/:locationID` | — | `80005` 有子地点、`80004` 被 items.location_id 引用、`80001` |

---

## 3. 前端陷阱清单（生成代码时逐条核对）

1. **成功判定**用 `code === 0`；HTTP 恒 200，axios 的 catch 捕不到业务错误。
2. **响应拦截器必须捕获响应头 Authorization 并覆盖 token**（续期机制）。
3. 图片 URL 相对路径：展示处统一 `imgSrc = BASE_URL + url`；`avatar` 同理。
4. `/user/batch` 结果**无序**且可能 `data: null` → 需按 id 建映射并兜底空数组。
5. `omitempty` 字段（contact/location_detail/claim_user_id/claim_time/color/address）**可能整个键缺失**，类型定义全部用可选。
6. `/item/list` 与 `/item/mine` 不传 status 返回 0+1；"已关闭"标签页必须显式 `status=2`。
7. `/item/mine` 的 page_size >50 → `1`；`/item/list` 上限 100。
8. item `tag_ids` 是指针语义：编辑表单若允许清空标签，必须**显式传 `[]`**，不能省略字段。
9. item update 空对象静默成功；user update 空对象报 `1`。两处行为不同。
10. `logout_all` 必填：登出请求固定发 `{"logout_all": 0|1}`。
11. 创建 item 不返回 id → 创建成功后用 `/item/mine` 第一条回填。
12. 详情请求会使 `view_count+1`，不要在轮询/预取中滥用。
13. 长度校验按 **UTF-8 字节**（username/nickname 2-32，password 8-20）：`new Blob([s]).size` 或等价方式校验，中文字符算 3。
14. 上传用 `multipart/form-data` 且字段名 `file`；不要设置 `Content-Type` 手动覆盖 boundary；>5MB 前端先拦。
15. 权限不足（role 不够）返回 `2`，与未登录同码——处理上按"需要重新登录或提升角色"提示，勿一律跳转登录页。
16. claim 前若 `claim_qq_required=true`（当前为 true），可先查 `/user/me` 的 `qq` 字段预判，避免用户填完被 `30006` 打回。
17. 认领状态可能被自动关闭任务改变（24h 超时），UI 不应对 status 做长期缓存。
18. confirm 与 close 的区别：confirm 发积分、close 不发；发布者"我找回来了（没人认领）"用 close，"他认领并且真的还我了"用 confirm。
19. `/admin/add-credit` 的 `operator_id` 由前端显式传入并做身份校验：必须传当前登录管理员的用户 id（可从 `/user/me` 的 `id` 取），传别人或随意值 → `2` 且不执行。

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
13 POST /user/logout                     {"logout_all": 0}
```

---

## 5. 附：模块外错误码（仅列出可能见到的）

| 码 | 含义 | | 码 | 含义 |
|---|---|---|---|---|
| 4 | 资源不存在 | | 10013 | QQ 已被注册/绑定 |
| 7 | 未知错误 | | 10014 | 积分不够 |
| 8 | 陈松（机器人）故障 | | 20002 | 物品已关闭 |
| 10004 | 无效签名方式 | | 20007 | 积分奖励不能为负（保留） |
| 10011 | QQ 尝试过多 | | 20009 | 标题空（保留，实际走 `1`） |
