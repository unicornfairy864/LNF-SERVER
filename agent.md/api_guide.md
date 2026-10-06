# LNF-SERVER 前端对接文档（功能介绍 · 测试人员版）

> 读者：前端测试人员。
> 配套文档：`api_agent.md`（给前端 AI 的精确对接文档，含字段表与调用示例）；`common_response_code.md`（完整错误码表）。
> 范围：user、item（含 tag / location / image upload）、notification（站内通知）、shop（积分商城）、announcement（公告）模块。
> 本次新增（2026-10-04）：**Agent 智能助手模块** —— 对话式发帖 / 自然语言找匹配 / 详情页相似推荐，见第十三节；上一批新增 `admin-stats` 见第十二节。

---

## 一、服务器与环境信息

| 项目 | 值 |
|---|---|
| Base URL | `http://111.229.234.32:8080` |
| API 前缀 | `/api/v1`（下文接口路径均已含前缀） |
| Swagger UI | `http://111.229.234.32:8080/api/v1/swagger/index.html` |
| 静态图片目录 | `http://111.229.234.32:8080/uploads/...`（无需鉴权） |

### 统一响应格式（非常重要）

**所有接口 HTTP 状态码恒为 200**（包括参数错误、未登录、权限不足），业务结果看响应体：

```json
{
  "code": 0,
  "message": "操作成功",
  "data": { }
}
```

- `code = 0`：成功；`code != 0`：失败，`message` 为中文错误信息，`data` 为空对象 `{}`。
- 前端判断成功/失败**只看 `code`，不要看 HTTP 状态码**。
- 唯一例外：`GET /user/jwt-test` 成功返回 `code = -1`（调试接口）。

### 鉴权方式

- 登录成功后，token 放在**响应头** `Authorization: Bearer <token>` 中（不在响应体里），前端需要从响应头读取并保存。
- 之后所有需要登录的接口，在**请求头**带 `Authorization: Bearer <token>`。
- **token 会自动续期**：token 签发 6 小时后，任意一次请求的**响应头**里会下发新 token 并作废旧 token。前端必须在每次响应后检查并更新保存的 token，否则会随机出现 `10005 令牌被禁用`。
- token 有效期 24h。

### 角色说明

| role | 含义 | 权限 |
|---|---|---|
| 0 | 普通用户 | 注册默认角色，可发帖/认领/兑换 |
| 1 | 服务管理员 | 额外可管理 tag、location、shop 商品、发通知 |
| 2 | 系统管理员 | 额外可管理用户角色/状态/积分、公告 |

> 注意：权限不足时后端返回的也是 `code=2（未登录或 Token 无效）`，而不是"无权限"语义，测试时不要误判成 token 失效。

---

## 二、user 模块

### 1. 注册 `POST /user/create`（无需登录）

- 请求体：`{"username": "...", "password": "...", "nickname": "..."}`
- 规则：用户名/昵称长度 2-32，密码长度 8-20（**按 UTF-8 字节数算**，1 个汉字 = 3 字节）；用户名唯一。
- 成功不返回用户数据（`data: {}`）。
- 常见错误：`1` 参数不符，`10002` 用户名已被占用，`5` 服务器错误。

### 2. 登录 `POST /user/login`（无需登录）

- 请求体：`{"username": "...", "password": "..."}`
- 成功：`data` 为本人完整用户信息；**token 在响应头 Authorization**。
- 常见错误：`10003` 格式不符，`10001` 用户名或密码错误，`10006` 用户被禁用，`6` 数据库错误。

### 3. 我的信息 `GET /user/me`（需登录）

- 返回本人完整信息（`POST /user/me` 也能用，但响应头会带 Warning 提示建议用 GET）。
- 常见错误：`10006` 用户不存在或被禁用。

### 4. 修改资料 `POST /user/update`（需登录）

- 请求体（增量更新，只传要改的字段）：`{"nickname": "...", "realname": "...", "gender": 0或1, "avatar": "图片URL"}`
- `gender` 为 int8，后端不校验取值，语义由前端定义（如 0/1）。
- `avatar` 填图片上传接口返回的 URL。
- 发送 `{}`（一个字段都不改）会返回 `1 参数错误`。
- 字段超长（如 realname > 50、avatar > 255 字符）会返回 `6 数据库错误`，前端应自行限制长度。

### 5. 登出 `POST /user/logout`（需登录）

- 请求体：`{"logout_all": 0}` 或 `{"logout_all": 1}` —— **该字段必填**。
  - `1`：该账号全部会话登出；**其他任意值（含 0、2 等）一律按"仅当前会话登出"处理**。
- 当前 token 登出后立即失效（`10005`）；`logout_all=1` 时该账号全部会话的 token 失效。

### 6. 批量查用户公开信息 `POST /user/batch`（无需登录）

- 请求体：`{"ids": [1, 2, 3]}`
- 返回公开用户信息数组（id、昵称、性别、头像、角色、时间），**不含 username/qq/credit**。
- 注意：返回顺序与传入顺序**不保证一致**，无效 id 会被过滤；`ids: []` 返回 `data: []`；传入 id 全部未命中时 `data` 为 `null`（前端按空数组处理）。
- 前端用途：物品详情/列表中根据 `user_id` 显示发布者昵称头像。

### 7. QQ 绑定 `POST /user/qq/get-code` → `POST /user/qq/bind`（需登录）

绑定流程（依赖"陈松"QQ 机器人在指定 QQ 群发验证码）：

1. 前端提交 QQ 号 → `POST /user/qq/get-code`，请求体 `{"qq": 12345678}`（范围 10000 ~ 99999999999）。
2. 机器人会在群里 @该用户 发 6 位验证码；**该 QQ 必须已在指定 QQ 群内**，否则报 `10009`。
3. 用户把 6 位验证码填入前端 → `POST /user/qq/bind`，请求体 `{"qq": 12345678, "code": 123456}`。
4. 约束：验证码会话 3 分钟有效；同一会话最多比对 3 次（后端判定 `tries > 3`，即第 5 次提交才报 `10011`，第 4 次仍可尝试）；同一 token 会话或同一 QQ 号存在未失效会话时 `get-code` 返回 `10007`。
5. 绑定成功后会收到一条站内系统通知（通知类型 0，含绑定的 QQ 号）。

> 测试提示：此流程依赖真实 QQ 群和机器人在线，前端测试环境可能无法完整走通；联调时需后端配合确认机器人状态。常见错误码：`10012` 会话不存在或已失效、`10010` QQ 号与申请时不一致、`10008` 验证码错误、`10011` 尝试次数过多、`10013` QQ 已被绑定/账号已有 QQ、`8` 机器人故障、`6`。
> 认领物品和商城兑换都要求先绑定 QQ（见 item / shop 模块）。

### 8. JWT 调试接口 `GET /user/jwt-test`（需登录，仅测试用途）

- 返回 **`code = -1`**（注意不是 0！），`data` 为 `{"jwtID": 1, "jwtRole": 0, "tokenFresh": false}`（tokenFresh 恒为 false，占位字段）。
- **前端正式代码可不接入**；测试时可用它验证 token 是否有效、是否逾期（token 失效时应返回 `2` 或 `10005`）。

### 9. 管理员接口（需 role=2 登录）

| 接口 | 请求体 | 说明 |
|---|---|---|
| `POST /admin/change-role` | `{"id": 用户ID, "role": 0/1/2}` | 修改任意用户角色；role 非法 → `10003` |
| `POST /admin/change-status` | `{"id": 用户ID, "status": 0/1}` | `0` 禁用 / `1` 正常；status 非法 → `10003`。禁用→该用户全部 token 失效；解禁→不影响存量 token |
| `POST /admin/add-credit` | `{"id": 用户ID, "credit": ±N, "type": 0/1/2/3, "description": "备注", "operator_id": 当前登录管理员自己的ID}` | `type`：0 拾金不昧奖励 / 1 认领成功奖励 / 2 违规扣分 / 3 系统调整；积分下限 0，不够扣返回 `10014`；`credit=-99999` 表示清零。**`operator_id` 必填且必须等于当前登录管理员的 id（JWT 身份）**，缺失 → `1`，不符返回 `2` 且不执行 |

---

## 三、item 模块（物品）

### 核心概念

- **type**：`0` 丢失（我丢了东西求线索）、`1` 拾到（我捡到东西等认领）。
- **status**：`0` 已发布（进行中）、`1` 已认领、`2` 已关闭。
- 创建即发布（status=0），无审核环节。

### 1. 公开列表 `GET /item/list`（无需登录）

Query 参数（均可选）：

| 参数 | 说明 |
|---|---|
| `type` | 0 丢失 / 1 拾到 |
| `status` | 0 / 1 / 2；**不传默认返回 0+1（已发布+已认领）**，查已关闭需显式传 `status=2` |
| `location_id` | 按地点筛选 |
| `tag_id` | 按标签筛选（打了该标签的物品） |
| `keyword` | 对标题+描述做模糊匹配 |
| `page` | 默认 1 |
| `page_size` | 默认 10。**>100 报 `1`；传 51-100 不报错但后端按 10 条返回**（响应里的 page_size 字段是回显值，以 items 实际条数为准）。建议固定传 1-50 |

- 返回：`{total, page, page_size, items: [...]}`，按创建时间倒序。
- 列表项已包含：地点链（`locations`）、图片（`images`）、标签（`tags`），可直接渲染卡片。

### 2. 多条件检索 `GET /item/search`（无需登录）

按"最少满足条件数"筛选：`match_count = 命中的标签数 + (地点命中 ? 1 : 0)`，返回 `match_count >= min_match` 的物品。

| 参数 | 必填 | 说明 |
|---|---|---|
| `tag_ids` | — | 逗号分隔，如 `tag_ids=1,2,3`；不存在的标签自动剔除且不计入条件总数 |
| `location_ids` | — | 逗号分隔；**只有第 3 层级（level=3）的地点参与计分**，其他自动忽略 |
| `min_match` | ✅ | ≥1 且 ≤ 有效条件总数；等于条件总数 = 全部条件必须满足；两组都空 → `1` |
| `type` | — | 0 / 1 |
| `status` | ✅ | **必传可多选，只允许 0/1**（如 `status=0,1`；传 2 → `1`） |
| `page` / `page_size` | — | 同 `/item/list` 的分页规则（51-100 静默按 10） |

- 返回结构与 `/item/list` 相同，按创建时间倒序。
- 前端用途："描述了特征帮我找"的场景：勾选若干标签 + 大致地点，至少满足其中 N 个条件。

### 3. 数量统计 `GET /item/count`（无需登录）

- 无参数。返回"正在被寻找的物品数量"：status 为 **0 已发布或 1 已认领**、且**未删除**（is_deleted=0）的物品总数——即与公开列表不传 status 时的筛选范围完全一致。
- 返回：`data` 是**裸数字**（int64），不是对象，如 `{"code": 0, "message": "操作成功", "data": 42}`。
- 前端用途：首页/看板展示"寻找中的物品数"；其值应等于 `GET /item/list`（不传 status）响应的 `total`。
- 错误：`6` 数据库错误。

### 4. 详情 `GET /item/:itemID`（无需登录）

- 返回物品完整信息（同列表项结构），**每次调用浏览量 view_count +1**。
- `20001` 物品不存在或已删除；路径 id 非正整数 → `1`。

### 5. 发布 `POST /item/create`（需登录）

- 请求体：
  ```json
  {
    "title": "必填，非空且≤100字符（按字节）",
    "description": "必填",
    "type": 0,
    "lost_found_time": "2026-09-26T15:04:05+08:00",
    "location_id": 3,
    "location_detail": "教学楼A座201门口（选填，≤200字符）",
    "contact": "微信号/邮箱等（选填，≤100字符）",
    "credit_reward": 0,
    "tag_ids": [1, 2]
  }
  ```
- `lost_found_time` 为 RFC3339 时间字符串（带时区），**必填**。
- `location_id` 必须是 location 列表中存在的 id，否则 `20010`；`tag_ids` 中任一标签不存在则 `40001`（自动去重）；`credit_reward` 负数 → `1`。
- 成功后物品 status=0，但**不返回新物品 id**——前端创建后可通过"我的发布"列表第一条获取。

### 6. 编辑 `POST /item/update`（需登录，仅发布者本人）

- 增量更新：只传要改的字段；**发空对象 `{}` 会静默成功（什么都不改）**，前端应自行拦截。
- **`status` 不能通过此接口修改**：请求体没有该字段，传入会被忽略；状态流转必须走认领/关闭接口。
- `tag_ids`：**只要字段出现就整体替换**标签（传 `[]` 表示清空所有标签）；不传该字段则不动标签。
- `location_id` 传 `0` 表示清除地点；传正数需存在（否则 `20010`）。
- `location_detail` >200 或 `contact` >100 字符 → `1`。
- 错误：`20001` 不存在、`20005` 不是你的物品、`20010` 地点无效、`40001` 标签不存在、`1` 字段非法。

### 7. 删除 `POST /item/delete`（需登录，仅本人）

- 请求体：`{"id": 物品ID}`，逻辑删除，之后列表/详情都看不到。
- 错误：`20001`、`20005`。

### 8. 我的发布 `GET /item/mine`（需登录）

- 参数同公开列表；**status 不传默认也是 0+1**；`page_size` 上限 50（传 51-100 不报错但按 10 条返回，>100 报 `1`）。

### 9. 图片设置 `POST /item/:itemID/images`（需登录，仅本人）

- 请求体：`{"images": [{"image_url": "/uploads/...", "sort_order": 1}, ...]}`
- 最多 **3 张**（超过 `20013`）；`sort_order` 取值 1-3 且不可重复、URL 非空且 ≤500（否则 `20014`）。
- **传 `"images": []` 表示清空图片**；不传 `images` 字段返回 `1`。
- `image_url` 填上传接口返回的相对 URL（见第四节）。
- 覆盖式：每次提交都会整体替换该物品的图片列表。

### 10. 认领与关闭（状态机）

```
             认领 claim                确认 confirm（发积分）
  [0 已发布] ───────────→ [1 已认领] ──────────────→ [2 已关闭]
      ↑  撤回 claim/cancel      │
      └─────────────────────────┘
      0/1 状态：发布者 close 关闭（不发积分）
      删除 delete：任意状态下仅本人可删（列表/详情不可见）
```

| 接口 | 谁可以调 | 前置条件 | 说明 |
|---|---|---|---|
| `POST /item/:itemID/claim` | 任何登录用户（非发布者） | status=0 | 认领后 status=1；**当前配置要求先绑定 QQ**，未绑定返回 `30006` |
| `POST /item/:itemID/claim/cancel` | 认领者**或**发布者 | status=1 | 撤回后恢复 status=0，可再次被认领 |
| `POST /item/:itemID/confirm` | 仅发布者 | status=1 | 确认找回并关闭（status=2），给对方加 `claim_credit` 分（当前配置 **10 分**）；通知认领者 |
| `POST /item/:itemID/close` | 仅发布者 | status=0 或 1 | 自己找回了直接关闭，**不发积分**；若有进行中的认领会通知认领者 |

- 积分归属：拾到帖（type=1）确认后发给帖主；丢失帖（type=0）确认后发给认领者。
- **自动关闭**：物品被认领超过 **24 小时** 未确认，系统自动按 confirm 语义关闭并发分（后台每 5 分钟扫描一次，服务启动时也会先扫一次），并给双方发站内通知。测试时注意状态可能"自己"变化，需重新拉取列表。

常见错误：`20001` 不存在、`20002` 已关闭（claim/cancel/confirm/close 统一返回此码）、`20003` 已被认领、`30004` 不能认领自己的帖子、`30002` 无认领记录、`30003` 无权撤回、`20005` 非发布者、`30006` 未绑 QQ、`10006` 账号被禁用。

---

## 四、图片上传 upload 模块

`POST /upload/image`（需登录）

- 请求格式：`multipart/form-data`，文件字段名固定为 **`file`**。
- 仅支持 **jpg / png / webp**（不支持 gif，按文件头识别不看扩展名）；单文件最大 **5MB**。
- 返回：`{"code": 0, "data": {"url": "/uploads/2026/09/26/xxx.jpg"}}` —— **相对 URL**。
- 展示时需自行拼接服务器源：`http://111.229.234.32:8080` + `url`。
- 错误：`9` 文件过大、`10` 类型不支持、`11` 保存失败、`1` 未带文件字段。

**典型用法**：
- 物品图片：上传 → 拿到 url → `POST /item/:itemID/images` 提交。
- 头像：上传 → 拿到 url → `POST /user/update` 的 `avatar` 字段。

---

## 五、tag 模块（标签）

### 普通用户 / 游客

- `GET /tag/list`（无需登录）：返回全部标签，按 sort_order 升序，用于发帖选标签和列表筛选器。
- 标签字段：`{id, name, color, sort_order, created_at, updated_at}`（color 为 null 时整个键缺失）。

### 管理员（role≥1）

| 接口 | 请求体 | 错误 |
|---|---|---|
| `POST /tag/create` | `{"name": "必填1-50字", "color": "#FF0000可选", "sort_order": 0}` | `40004` 名称非法、`40002` 重名 |
| `POST /tag/update` | `{"id": 1, ...要改的字段}`（增量） | `40001` 不存在、`40004` 名称非法、`40002` 重名 |
| `DELETE /tag/:tagID` | — | `40001` 不存在、`40005` 仍被物品引用 |

---

## 六、location 模块（地点，树形）

### 普通用户 / 游客

- `GET /location/list`（无需登录）：`parent_id=0` 返回根节点（级联选择器第一级）；传 `parent_id=某id` 返回其子地点；都不传返回全部。按 sort_order 升序。
- 地点字段：`{id, name, parent_id, level, address, sort_order, created_at, updated_at}`（address 为 null 时整个键缺失）。
- `GET /item/:itemID/locations`（无需登录）：返回该物品地点的完整链路（根→叶），用于详情页"校区→楼栋→房间"展示；物品无地点时返回 `[]`；物品不存在 → `20001`。
- `/item/search` 只有 **level=3** 的地点参与计分，检索入口的地点选择器应只展示第 3 层。

### 管理员（role≥1）

| 接口 | 请求体 | 错误 |
|---|---|---|
| `POST /location/create` | `{"name": "必填1-100字", "parent_id": 0, "address": "可选", "sort_order": 0}` | `80001` 父级不存在、`80002` 同父级重名、`1` 名空/超长 |
| `POST /location/update` | `{"id": 1, ...要改的字段}`（增量；改 parent_id 会重算子树层级） | `80001` 不存在/新父级不存在、`80002` 改名或移动后与目标父级下已有地点重名、`1` 名空/超长/父级指向自身或后代 |
| `DELETE /location/:locationID` | — | `80005` 有子地点、`80004` 被物品引用、`80001` 不存在 |

---

## 七、notification 模块（站内通知）

- 通知类型 `type`：`0` 系统通知 / `1` 物品匹配 / `2` 认领申请 / `3` 认领结果 / `4` 评论回复 / `5` 积分变动 / `6` 商品兑换。
- `admin_id=0` 表示系统自动触发（QQ 绑定、认领关闭、兑换等）。
- 详情响应字段：`read_at` **未读时整个键缺失**（不是 `null`，前端按可选处理，判断已读用 `is_read`）；内部字段 `is_deleted` **不返回**（列表/详情均无该键）。

### 用户侧（需登录）

| 接口 | 说明 |
|---|---|
| `GET /notifications?limit=10&offset=0` | 我的通知列表（`created_at DESC, id DESC` 稳定排序；limit 缺省 10、**上限 100**）。`type`/`is_read`/`admin_id` 不筛就**省略**——传 0 或空串（如 `type=`）都会按 0 参与筛选且不报错（`type=0` 只看系统通知）；`type` 超出 0-6、`is_read` 超出 0-1 → `1` |
| `GET /notifications/unread-count` | 未读数量（int64；有 5 分钟 Redis 缓存，已读/删除后立刻刷新） |
| `GET /notifications/:id` | 通知详情，**未读会自动标记已读**；不存在/非本人 → `60001` |
| `PUT /notifications/read` | 批量已读，请求体 `{"ids": [1,2]}`（必填非空，单次 **1-200 条**，超出 → `1`）；只操作自己的未读记录 |
| `DELETE /notifications` | 批量删除（软删），请求体 `{"ids": [...]}`（单次 **1-200 条**）；**"管理端群发给自己的那条"会被跳过不可删** |

### 管理侧（role≥1）

- `POST /admin/notifications`：请求体 `{"user_ids": [1,2], "send_to_all": false, "type": 0, "title": "≤100字", "content": "..."}`（`user_ids` 单次 **≤1000**，超出 → `1`）。
- `user_ids` 与 `send_to_all` 二选一（同传、全空或全部为无效 id → `1`）；**`type` 必填，`0`（系统通知）是合法值**；**异步发送，接口立即返回成功**（写入失败只记后端日志，不报错给前端）。

---

## 八、shop 模块（积分商城）

### 普通用户 / 游客

- `GET /shop/goods/list`（无需登录）：`keyword`（名称模糊）、`min_price`/`max_price`（积分闭区间，0/不传=不限，min>max → `1`）、`page`/`page_size`（缺省 1/10，**上限 100 正常生效**）。返回 `{total, page, page_size, items}`，按 sort_order 升序+创建时间倒序。
- `GET /shop/goods/:goodsID`（无需登录）：商品详情；不存在/已下架 → `11001`。
- 商品字段：`{id, name, description, image_url, price, stock, sort_order, ...}`；`image_url` 相对路径展示需拼接；`stock=0` 不可兑换。

### 兑换（需登录 + 已绑定 QQ）

- `POST /shop/goods/:goodsID/redeem`：单事务完成扣库存（防超卖）+ 扣积分 + 写订单。
- 前置：**必须已绑定 QQ**（未绑/格式非法 → `11005`）。
- 成功返回：`{"order_no": "订单号", "goods_id": 1, "goods_name": "...", "price": 20, "credit": 剩余积分, "created_at": "..."}`。
- 成功后自动收到两条站内通知（type=5 积分变动、type=6 商品兑换）和一条 QQ 群 @ 消息；**领奖需联系管理员**（线下发货）。
- 错误：`10006` 账号问题、`11005` 未绑 QQ、`11001` 商品不存在/下架、`11002` 库存不足、`50001` 积分不足。
- `GET /shop/orders`（需登录）：我的兑换记录，`page`/`page_size` 分页，按时间倒序；订单为快照设计，商品改名/下架不影响历史记录。

### 管理员（role≥1）

| 接口 | 请求体 | 错误 |
|---|---|---|
| `POST /shop/goods/create` | `{"name": "必填1-100字", "description": "可选", "image_url": "可选≤500", "price": 1~1000000, "stock": 0, "sort_order": 0}` | `11003` 名空/超长/**重名**、`11004` 价格非法、`1` stock 负数等 |
| `POST /shop/goods/update` | `{"id": 1, ...要改的字段}`（增量） | `11001`、`11003`、`11004` |
| `POST /shop/goods/delete` | `{"id": 1}` | `11001`。软删=下架，历史订单不受影响 |

> 注意：商品重名没有独立错误码，统一返回 `11003`（商品名称无效）。

---

## 九、announcement 模块（公告）

### 普通用户 / 游客

- `GET /announcement?page=1&page_size=10`（无需登录）：已发布公告列表，按 id 倒序（新的在前）。返回 `{total, page, page_size, announcements}`；`page` 缺省 1，`page_size` 缺省 10、**上限 100 正常生效**（>100 或负数 → `1`，与 item 的 51-100 静默按 10 不同）。无结果时 `announcements` 为 `[]`。
- `GET /announcement/:id`（无需登录）：公告详情，**仅已发布（status=1）可见**；**每次调用浏览量 +1（返回值已含本次 +1）**，勿轮询滥用。不存在/已下架/已删除 → `90001`；id 非正整数 → `1`。
- 公告字段：`{id, admin_id, title, content(markdown 正文), type, status, is_top, view_count, published_at, created_at, updated_at}`；`published_at` 可能为 null（键缺失）。
- 状态 `status`：`1` 已发布 / `2` 已下架；**没有草稿态（0 已废弃）**，创建即发布。
- 类型 `type`：`0` 系统公告 / `1` 活动公告 / `2` 维护通知 / `3` 其他。

### 管理员（role=2）

| 接口 | 请求体 / Query | 说明与错误 |
|---|---|---|
| `POST /admin/announcement/create` | `{"title": "必填", "content": "必填markdown", "type": 0, "is_top": 0}` | **创建即发布**（status=1、发布时间=当前时间、发布人=登录管理员 JWT）；成功 `data: {}`（**不返回 id**，跳详情可先拉列表第一条）。`type`/`is_top` 必填且 **0 也要显式传**（缺字段 → `1`）；标题空白或超 100 字节、内容空白、type 非 0-3、is_top 非 0/1 → `90003` |
| `POST /admin/announcement/update` | `{"id": 1, ...要改的字段}`（增量） | 可改 title/content/type/status/is_top；`status` 仅 `1` 重新上架 / `2` 下架（传 0/3 → `90003`）；发布时间保持首次不变，view_count/created_at 不可改；不存在/已删除 → `90001`；只传 id → 成功但无操作 |
| `DELETE /admin/announcement/:id` | — | 软删（is_deleted=1）：公开/管理列表与详情均不可见；不存在 → `90001`、id 非法 → `1` |
| `GET /admin/announcement` | `page` / `page_size` / `status` | 管理列表：**含已下架**，id 倒序；`status` 可选 0/1/2——缺省 = 1+2 全部，`1`/`2` 按状态筛，**`0` = 查历史废弃行**（清理后恒为空），其他值 → `1` |

> 注意：以上四个接口非 role=2 调用返回 `2`；历史 `status=0` 垃圾数据在公开列表/详情任何情况不可见、缺省管理列表也不返回，仅管理列表显式 `status=0` 可查（后端提供清理 SQL，由运维在数据库执行）。注意区分：update 请求体的 `status=0` 仍是业务错误（`90003`），只有列表 query 的 `status=0` 合法。

---

## 十、建议测试用例（冒烟 + 边界）

**用户流**
1. 注册（正常 / 重名 / 短密码 / 11 个汉字昵称——按字节算应失败）。
2. 登录 → 从响应头取 token → `/user/me` 校验身份一致；`/user/jwt-test` 应返回 `code=-1`。
3. `/user/update` 改昵称、头像；`{}` 空更新应返回 `1`。
4. `logout_all=0` 后旧 token 不可用（`10005`）；另一端登录后 `logout_all=1`，全部端 token 失效。
5. `/user/batch` 传有效+无效 id 混合，确认乱序返回与过滤；全部无效时 `data: null`。
6. QQ 绑定（需真实群环境，可作选测）：验证码错 3 次仍可比对、第 5 次提交报 `10011`、会话过期 `10012`。

**物品流**
7. 发布丢失帖/拾到帖各一条 → 公开列表可见（status=0）。
8. 列表筛选：type / status / location_id / tag_id / keyword / 分页逐项验证；不传 status 应包含已认领的物品；page_size=60 验证实际返回 10 条；`GET /item/count`（无需登录）返回的 `data`（裸数字）应等于不传 status 的 `/item/list` 的 `total`（status 0+1 且未删除）。
9. `/item/search`：两组条件选 2 标签+1 地点，min_match=2 应命中"满足任意两条"的物品；传 status=2 应报 `1`。
10. 详情页 view_count 递增；地点链、标签、图片正确展示。
11. 另一账号认领该物品（未绑 QQ → `30006`，绑定后成功）→ 发布者 confirm → 双方积分变化（认领者或帖主 +10）+ 双方收到 type=3 通知。
12. 认领后 claim/cancel 撤回（用认领者、发布者各测一次）→ 物品恢复可认领；第三人撤回 → `30003`。
13. 发布者 close 直接关闭（不发积分）→ 再认领应报 `20002`。
14. 非本人 update/delete/images → `20005`；update 传 status 字段确认被忽略。
15. 图片：超 3 张（`20013`）、重复 sort_order（`20014`）、`images: []` 清空。
16. 上传：>5MB（`9`）、gif（`10`）、jpg/png/webp 正常路径拼接展示。
17. 标签/地点管理接口用 role≥1 账号测；普通账号应得到 `2`。地点改名撞已有名 → `80002`。

**商城流**
18. 未绑 QQ 兑换 → `11005`；积分不足 → `50001`；库存不足 → `11002`。
19. 正常兑换：积分扣减、返回订单号与剩余积分、`/shop/orders` 可见记录、收到 type=5/6 两条通知。
20. 并发兑换库存为 1 的商品，仅一人成功。

**通知流**
21. 列表分页/筛选（type/is_read）；未读数与已读联动；详情自动已读；批量已读/删除；"群发给自己的那条"删除被跳过。

**公告流**
22. role=2 创建公告 → 公开列表立即可见（创建即发布：status=1、published_at 非空、admin_id=登录管理员）；缺 type/is_top → `1`；标题全空白或 type=9 → `90003`。
23. update `status=2` 下架 → 公开列表与详情不可见（详情 `90001`），管理列表仍可见 → `status=1` 重新上架恢复；传 `status=0/3` → `90003`。
24. 公开详情浏览量：连续两次 `GET /announcement/:id`，view_count 每次 +1 且返回值含本次 +1；id 非法（如 `abc`、`0`）→ `1`。
25. 列表分页边界：page 缺省/0 → 第 1 页；page_size=100 正常返回；page_size=101 或负数 → `1`；无数据时 `announcements: []` 而非 null；管理列表 `status=0` 合法（返回历史废弃行，清理后为空）、`status=3` → `1`。
26. `DELETE /admin/announcement/:id` 软删 → 公开/管理列表均不可见、详情 `90001`；删除不存在的 id → `90001`；普通用户（role=0）调任一管理端点 → `2`；空请求体 create → `1`。

---

## 十一、已知注意事项与风险（重要）

1. **HTTP 状态码恒为 200**：成功失败都看 `code`（唯一例外 `/user/jwt-test` 成功是 `-1`）。
2. **token 从响应头拿、响应头续期**：前端拦截器必须每次响应都更新 Authorization。
3. **无 CORS 配置**：后端未开启跨域支持。前端本地开发（不同端口/域名）浏览器会直接拦截请求，**需要前端 dev server 配代理**，或前后端同域部署。若联调遇到"请求根本发不出去/预检失败"，先查这条，并与后端确认是否补 CORS（含 `Access-Control-Expose-Headers: Authorization`）。
4. 权限不足返回 `code=2`（"未登录或 Token 无效"），容易与 token 失效混淆，测试时用 `/user/jwt-test` 区分。
5. 图片 URL 是相对路径，展示必须拼服务器源。
6. 列表/批量查询等场景 `data` 可能为 `null`（如 batch 全部未命中），前端需做兜底。
7. 用户名/昵称/密码长度限制按**字节**计算（1 汉字=3 字节），前端校验规则需与后端一致；item 标题 ≤100、location_detail ≤200、contact ≤100 同理。
8. 认领超 24h 自动关闭并发分（每 5 分钟扫描），状态可能无操作自行变化。
9. item 系接口 page_size 传 51-100 会静默按 10 返回（不报错）；shop、announcement 接口无此问题（上限 100 正常生效）。统一传 ≤50 最稳。
10. 数值筛选 query 参数（notification `type/is_read/admin_id`、item `type/status`、公告管理列表 `status`）传 `0` 或空串一律按 0 参与筛选且**不报错**（公告 `status=0` = 查历史废弃行，清理后为空）；不筛请直接省略参数。**例外**：notification 的 `type` 只接受 0-6、`is_read` 只接受 0-1，**越界 → `1`**（2026-10-07 起）。发通知 `type=0`（系统通知）是合法值，缺字段才报 `1`；注意 update 请求体的 `status=0` 另当别论（`90003`）。

---

## 十二、管理员数据分析模块（admin-stats，role≥1）

管理端数据看板共 **9 个只读 GET** 接口，全部挂在 `/api/v1/admin/stats/*`。**role=1（服务管理员）与 role=2（系统管理员）都可访问**；role=0 或未登录返回 `code=2`。除 CSV 导出成功响应外，均为统一 JSON 信封（HTTP 200，成功 `code=0`）。

### 通用参数与口径（先读）

| 参数 | 说明 |
|---|---|
| `start_date` / `end_date` | `YYYY-MM-DD`，含首尾自然日；必须**成对**出现；`end_date` 不能是未来日期（今天合法） |
| `days` | 1-366，含今天的最近 N 个自然日；**与日期对互斥**（缺省时各接口有默认值） |

- 时间一律按**东八区**自然日计算；`start_date/end_date` 与 `trend.date` 都是 `YYYY-MM-DD`（不是 RFC3339）。
- 参数问题（日期非法、start>end、只传单边日期、日期与 days 同传、跨度>366 天、枚举值非法）→ `code=1`；数据库异常 → `6`；未登录/权限不足 → `2`。
- 百分比保留 2 位小数；分母为 0 返回 `0`。环比上期为 0 且本期 >0 时 `change_percent` 为 `null` 且 `comparable=false`。
- 状态口径：`0` 已发布（pending）/ `1` 已认领（claimed，要求 `claim_user_id` 非空）/ `2` 已关闭。"成功归还" = `status=2` 且 `claim_user_id`、`claim_time` 均非空；`status=2` 但无认领人 = 发帖者自行关闭（未归还）。
- ⚠️ `claim_time` 是**认领时间**，不是归还完成时间；归还时长用 `updated_at - created_at` 近似。撤回认领会清空认领字段，所以历史认领无法还原。

### 接口清单

| 接口 | Query 参数 | 说明 |
|---|---|---|
| `GET /admin/stats/overview` | 日期对 或 `days` | 概览：缺省当前自然月（本月 1 日~今天），对比上一个完整自然月；显式范围时对比紧邻等长周期 |
| `GET /admin/stats/trend` | 同上（默认 `days=30`） | 按日趋势，连续补零、日期升序，长度 = 区间天数 |
| `GET /admin/stats/locations` | 同上 + `limit`（默认 4，0-100） | 按物品**直接地点**聚合，NULL 单列 `unknown`，不向父地点归并 |
| `GET /admin/stats/time-heatmap` | 同上（默认 `days=30`） | 发布时段热力图：7 行（周一~周日）× 24 列（0-23 点），空桶为 0 |
| `GET /admin/stats/items/stagnant` | `days`（**滞留天数**，默认 7，1-3650）、`page`（默认 1）、`page_size`（默认 10，1-100） | 滞留待处理清单：`status=0` 且创建已超 `days` 天；按创建时间升序 |
| `GET /admin/stats/items/high-view` | 同上 + `min_views`（默认 50，1-2147483647） | 高浏览低认领：`status=0` 且 `view_count>=min_views`；按浏览降序 |
| `GET /admin/stats/return-duration` | 同上（默认 `days=30`）+ `group_by`（`none`\|`location`\|`type`，默认 `none`） | 近似归还时长（平均/中位数，单位秒） |
| `GET /admin/stats/distribution` | 同上（默认 `days=30`）+ `dimension`（**必填** `type`\|`tag`） | 按类型或标签的发布/归还分布 |
| `GET /admin/stats/funnel` | 同上（默认 `days=30`） | 注册/发布/认领/归还四层活跃人数与相邻比率（非嵌套漏斗） |

> 注意：`items/*` 两个清单接口的 `days` 含义是"滞留天数阈值"，**不接受** `start_date/end_date`；只返回物品摘要（`id/title/type/status/location_id/location_name/view_count/created_at/stagnant_days`），**不含联系方式与描述**。

### 各接口响应要点与示例

**overview**（发布/归还/待处理的环比 + 归还率 + 超 24h 未处理数）

```json
{
  "period": {"start_date": "2026-10-01", "end_date": "2026-10-03"},
  "published": {"value": 10, "previous": 8, "change_percent": 25.00, "comparable": true},
  "returned": {"value": 4, "previous": 2, "change_percent": 100.00, "comparable": true},
  "pending": {"value": 5, "previous": 4, "change_percent": 25.00, "comparable": true},
  "return_rate": {"value": 40.00, "previous": 25.00, "change_percent": 60.00, "comparable": true},
  "pending_over_24h": 2
}
```

- 发布队列按 `created_at` 归入区间，状态取**当前快照**；`pending_over_24h` = 队列内 `status=0` 且创建时间早于当前时刻 24 小时。
- 归还率 = `returned / (returned + pending + claimed + closed_without_return)`。

**trend**

```json
{"start_date": "2026-09-04", "end_date": "2026-10-03",
 "points": [{"date": "2026-09-04", "published": 3, "returned": 1}]}
```

- `points` 覆盖区间每一天（无数据补 0），日期升序；`published` 与 `returned` 都按物品**创建日**归属。

**locations**

```json
{"total": 10,
 "locations": [{"location_id": 3, "name": "教学楼A", "count": 4, "percent": 40.00}],
 "unknown": {"name": "unknown", "count": 1, "percent": 10.00}}
```

- `percent` 分母 = 区间内全部未删除发布量（`total`）；`limit>0` 只返回前 `limit` 个地点，`unknown` 永远单列且不占名额；`limit=0` 返回全部地点桶，但已知地点桶超过 1000 个 → `1`。
- 排序 `count` 降序、`location_id` 升序；`unknown` 无 `location_id` 键。

**time-heatmap**

```json
{"start_date": "2026-09-04", "end_date": "2026-10-03",
 "matrix": [[/* 7 行 × 24 列整数 */]]}
```

- `matrix[7][24]` 整数矩阵，行 0=周一…行 6=周日，列=小时；按 `created_at` 东八区归属，空桶为 0。

**items/stagnant 与 items/high-view**

```json
{"total": 5, "page": 1, "page_size": 10,
 "items": [{"id": 10, "title": "丢失黑色钱包", "type": 0, "status": 0,
            "location_id": 3, "location_name": "教学楼A", "view_count": 3,
            "created_at": "2026-09-25T18:00:00+08:00", "stagnant_days": 8}]}
```

- `stagnant_days` 为整数天数；`location_id` 可为 `null`（这时 `location_name` 为空串）。

**return-duration**

```json
{"start_date": "2026-09-04", "end_date": "2026-10-03", "group_by": "none", "approximate": true,
 "overall": {"group_id": null, "count": 7, "average_seconds": 3600, "median_seconds": 1800},
 "groups": [{"group_id": 3, "count": 2, "average_seconds": 1200, "median_seconds": 1000}]}
```

- 样本 = `status=2` 且认领字段非空、`updated_at` 落在区间且 `updated_at >= created_at`（排除负时长）；时长 = `updated_at-created_at` 秒。
- `approximate` 恒为 `true`（`updated_at` 可能被其他更新污染）；`group_by=none` 时只看 `overall`，`groups` 为 `[]`；空样本各项为 0。

**distribution**

```json
{"start_date": "2026-09-04", "end_date": "2026-10-03", "dimension": "tag", "total": 3,
 "buckets": [{"bucket_id": 1, "bucket_name": "证件", "published": 2, "returned": 1, "return_rate": 50.00, "percent": 66.67}]}
```

- `dimension=type`：`bucket_id` 0/1，名称 `lost`/`found`，每个物品只进一个桶。
- `dimension=tag`：一个物品可进多个桶，`published` 之和可大于 `total`；未打标签单列 `bucket_id=null`、名称 `untagged`；已删除的标签关联仍保留桶（名称为空串）。
- 已知标签桶超过 1000 个 → `1`；排序 `published` 降序、`bucket_id` 升序。

**funnel**

```json
{"start_date": "2026-09-04", "end_date": "2026-10-03",
 "stages": [{"stage": "registered", "users": 12}, {"stage": "published", "users": 8},
            {"stage": "claimed", "users": 4}, {"stage": "returned", "users": 2}],
 "adjacent_ratios": [66.67, 50.00, 50.00]}
```

- `stages` 固定顺序 `registered/published/claimed/returned`；各层口径互相独立（注册按 `users.created_at`、发布按 `items.created_at`、认领按 `claim_time`、归还按 `updated_at`），**不是嵌套队列，人数不保证递减**。
- `adjacent_ratios` 长度 3：`published/registered`、`claimed/published`、`returned/claimed`，百分比可超过 100，分母为 0 返回 0。

### CSV 导出（`export=csv`）

支持导出的 6 个报表：`overview`、`trend`、`locations`、`distribution`、`return-duration`、`time-heatmap`。

- 请求方式：在对应接口上加 `export=csv`，其余参数不变（`locations` 仍支持 `limit`）。`export` 只接受空值（走 JSON）或 `csv`，传其他值 → `1`。
- **成功响应不是 JSON**：HTTP 200，`Content-Type: text/csv; charset=utf-8`，含 UTF-8 BOM，`Content-Disposition` 带中文文件名 `管理员统计-{report}-{start_date}-{end_date}.csv`（另有 ASCII 回退名 `stats.csv`）。测试时用浏览器/Postman 保存文件后用 Excel 打开，中文不应乱码。
- 限制：最多 10000 数据行（`return-duration` 含 `overall` 行，每组一行），整体 30 秒超时，每 200 行 flush；超行数 → `1`。
- 失败行为：**开始写流之前**失败仍是 HTTP 200 JSON（参数/超限 `1`、超时 `5`、数据库 `6`）；**已开始写流之后**失败会直接中断，文件不完整且流尾没有 JSON，前端按下载失败处理。
- `funnel`、`items/stagnant`、`items/high-view` **不支持导出**：传 `export=csv` 会被静默忽略并返回正常 JSON 成功响应。

### 注意事项与建议测试点

1. 权限：用 role=1、role=2 各测一遍应可用；role=0 或无 token → `2`。本模块没有 role=2 专属接口。
2. 日期边界：只传 `start_date` → `1`；日期与 `days` 同传 → `1`；`end_date` 传明天 → `1`；跨度 367 天 → `1`；`days=0` 或 `days=367` → `1`；`days=366` 正常。
3. `overview` 缺省应返回本月 1 日~今天，并对比上一个完整月；`change_percent` 为 2 位小数。
4. 上期为 0 且本期 >0 时，对应 `change_percent` 应为 `null` 且 `comparable=false`；上期与本期都为 0 时为 `0`。
5. `trend` 无数据的日期应出现且计数为 0；日期连续不跳日。
6. `locations`：构造一条无地点的物品，应出现在 `unknown` 且不计入 `locations`；对比 `limit=0` 与 `limit=3` 的返回条数；`percent` 之和（不含 unknown）不应超过 100。
7. `distribution`：`dimension` 缺失或传其他值 → `1`；tag 模式下一个物品挂两个标签时 `published` 之和应大于 `total`，且存在 `untagged` 桶。
8. `return-duration`：`group_by` 传非法值 → `1`；无样本时 `count/average_seconds/median_seconds` 均为 0；响应 `approximate=true`。
9. `funnel`：`adjacent_ratios` 应有 3 个元素且允许 >100；各层人数不要求递减。
10. CSV：目标报表加 `export=csv` 应下载到以 BOM 开头、中文文件名正确的文件；参数非法时仍是 JSON（`1`）；`export=json` 等其他值 → `1`；`funnel` 传 `export=csv` 仍是 JSON 成功响应。
11. 24 小时自动关闭任务会改变 `status`，统计结果随时间变化，前端不要长期缓存看板数据。

---

## 十三、Agent 智能助手模块（对话式发帖 · 自然语言找匹配 · 相似推荐）

> 本次新增（2026-10-04）。API 面向普通用户（role=0，`/agent/*` 需登录）；同一套能力已接入 **QQ 群机器人**（群里 @机器人 即可，前端无需开发）。
> 后端落点：`agent/orchestrator/*`（编排）、`handler/advanced/agent_handler.go`、路由 `/api/v1/agent/*`；数据契约 `model/basic/agent.go`。

### 1. 应用场景（先读这节，决定前端怎么接入）

一句话概括：**把“填表单发帖”和“翻列表找东西”变成“说一句话”。**

| # | 场景 | 用户说什么 | 后端做什么 | 前端建议承载位置 |
|---|---|---|---|---|
| **A** | **懒人发帖**（失物/招领） | 「我昨天下午在图书馆三楼丢了个黑色保温杯，带吸管的」 | 判类 → 抽取标题/描述/标签/地点/时间 → 返回草稿（`stage=need_confirm`）+ 缺失项 → 用户回一句话（补充信息 或「确认」）→ **建帖** | 发帖页的「🤖 一句话发帖」入口 / 对话式弹窗 |
| **B** | **找匹配**（我丢的东西有没有人捡到） | 「有人捡到黑色水杯吗」 | 判类（疑问句→`match`）→ 抽取特征 → 召回（标签 + 地点全链 + 时间 + 全文）→ LLM 精排 → 返回候选（`score`/`reasons`） | 首页/搜索页的「描述物品，让助手帮你找」输入框 |
| **C** | **详情页相似推荐** | （无输入，打开某条帖子） | 按该帖的类型/标签/地点计分，返回相似的其他帖子 | 物品详情页底部「相似帖子」 |
| **D** | **QQ 群机器人**（同一套能力） | 群里 `@机器人 我在图书馆捡到一个黑色保温杯` | 完全相同的链路；群内先回执「我正在思考」→ 回复草稿/结果 → 再次确认后建帖（联系方式自动补发送者 QQ） | 前端无需开发 |

**A 与 B 不需让用户选**：后端按句式自动判类 —— **陈述句/要求登记 → 发帖（A）；疑问句（「有没有人…吗」「帮我找找」）→ 找匹配（B）**。
场景 A/B/D 共用主入口 `POST /agent/chat`；场景 C 是独立只读接口 `GET /item/:itemID/similar`。

### 2. 接口清单（统一响应信封，HTTP 恒 200）

| 方法 | 路径 | 登录 | 用途 | 消耗 LLM |
|---|---|---|---|---|
| POST | `/agent/chat` | 需登录 | **主入口**：发帖（草稿+确认）/ 找匹配 / 闲聊兜底 | 是（1~3 次） |
| POST | `/agent/match` | 需登录 | 只读匹配：不做会话、不建帖，直接返回候选 | 是（1~2 次） |
| POST | `/agent/extract` | 需登录 | 只读抽取：给表单“智能填充”用 | 是（1 次） |
| POST | `/agent/session/close` | 需登录 | 关闭当前会话（幂等） | 否 |
| GET | `/item/:itemID/similar` | **无需登录** | 详情页相似帖子推荐（纯 SQL） | 否 |

### 3. `POST /agent/chat`（主入口）

请求：

```json
{
  "session_id": "上一轮返回的会话ID（第一轮不传）",
  "text": "我昨天下午在图书馆三楼丢了个黑色保温杯，带吸管的",
  "image_urls": ["/uploads/2026/10/xxx.jpg"],
  "action": "auto"
}
```

| 字段 | 说明 |
|---|---|
| `text` | 必填；≤ 500 字符（超出 `120003`） |
| `image_urls` | 可选，最多 3 张；`/` 开头视为本站相对路径（后端自动拼公网前缀），也可直接传公网 URL（多模态） |
| `action` | `auto`（默认，按文本自动处理）/ `confirm`（确认发布当前草稿）/ `cancel`（放弃当前草稿） |

响应 `data` 示例（发帖场景）：

```json
{
  "session_id": "as_24337219-ae87-46a9-b39b-675eb7246184",
  "stage": "need_confirm",
  "reply": "我帮你初步写好了招领信息：「拾到黑色保温杯」…还缺时间、联系方式。回复补充信息我会合并后发布；回复「确认」也会发布；回复「取消」则放弃；其他内容我不会发布。",
  "intent": "create_found",
  "questions": ["还缺：时间、联系方式", "大概什么时候捡到的？"],
  "matches": [],
  "similar": [],
  "draft": {
    "type": 1,
    "title": "拾到黑色保温杯",
    "description": "在图书馆三楼捡到一个黑色保温杯，带吸管。",
    "tag_ids": [33, 53],
    "tag_names": ["水杯", "黑色"],
    "location_id": 30,
    "location_name": "屏峰校区/图书馆",
    "location_detail": "三楼",
    "missing_fields": ["time", "contact"],
    "followup_question": "大概什么时候捡到的？"
  },
  "created_item_id": null
}
```

**`stage` 取值与前端应对**：

| stage | 含义 | 前端建议 |
|---|---|---|
| `need_confirm` | 已生成草稿，等用户补充/确认 | 展示 `draft` 预览 + “确认发布”按钮（点了就带 `session_id` 再发 `{"text":"确认"}` 或 `{"action":"confirm"}`） |
| `created` | **已建帖** | 用 `created_item_id` 跳详情页，提示可在“我的发布”修改 |
| `matched` | 找到候选 | 渲染 `matches`（`item` 即标准 `ItemResponse`，可复用列表卡片）；`similar` 为同/异类型补充推荐 |
| `no_match` | 没找到候选 | 展示 `reply`，引导用户走发帖（场景 A） |
| `chitchat` | 闲聊/无关 | 展示 `reply`（固定话术），**不是错误** |
| `cancelled` | 用户放弃 / 会话结束 | 清掉本地 `session_id` |

**`intent` 取值**：`create_lost`（失物）/ `create_found`（招领）/ `match`（找匹配）/ `chitchat` / `other`。

### 4. 建帖是“两步确认”（务必按此交互）

1. **第一步**：用户描述 → 返回 `need_confirm` + `draft`（`tag_ids` 可直接复用给 `/item/create`）
2. **第二步（唯一一轮）**：用户回复 →
   - 是**补充信息**（「门牌号是202」）→ 合并进草稿 → **直接建帖**（`stage=created`）
   - 是**明确确认**（「确认」「可以」）→ **直接建帖**
   - 含**取消/拒绝**语义（「算了」「不发了」）→ `stage=cancelled`，**不建帖**
   - 是**其他内容** → `stage=cancelled`（不建帖）；LLM 故障 → `120002`（不建帖，草稿保留可重试）

> 要点：**除“补充信息”和“明确确认”外，一律不会建帖**；前端不要假设“随便回一句就会发布”。

### 5. 会话规则（严格模式）

- 一个用户同时只有 **1 个会话**；`session_id` 由后端返回，前端保存并在**下一轮原样回传**。
- **不带 `session_id`** → 新开会话并**覆盖**旧会话（旧 `session_id` 立即失效）。
- **带错的/过期的 `session_id`** → `120001`，**不会**新建（前端可提示“会话已过期”并让用户重新描述）。
- 会话 30 分钟有效（每次交互续期）；单会话最多 3 轮用户消息，超出自动开新会话。

### 6. 限流与错误码（12xxxx 段）

| 限制 | 阈值 | 返回 |
|---|---|---|
| 每用户 | 10 次/分钟（三个 `/agent/*` 接口共享计数） | `120004` |
| 全系统 | 30 次/分钟 | `120004` |

| code | 含义 | 前端建议 |
|---|---|---|
| `120001` | 会话不存在或已过期 | 清掉本地 `session_id`，提示重新描述 |
| `120002` | 智能服务暂时不可用（LLM 超时/输出非法） | 提示稍后再试；草稿保留可重试 |
| `120003` | 输入不合法（空 / 超 500 字 / 图片超 3 张） | 提示用户 |
| `120004` | 操作过于频繁 | 提示稍后再试 |
| `120005` | 当前会话状态不允许该操作（如无草稿却 `action=confirm`） | 本地重置会话 |
| `120006` | Agent 功能未开启（后端 `openai.agent_enabled=false`） | 隐藏入口 |
| `120007` | 未匹配到相关帖子 | **预留**（当前不返回；未匹配走 `stage=no_match`） |

### 7. 只读辅助接口

- **`POST /agent/match`**：请求 `{"text":"...","image_urls":[],"top_n":3}` → `{intent, entities, matches, similar, summary, verdict}`；`verdict` = `strong_match` / `ambiguous` / `no_match`；`top_n` 默认 3、最大 10。**不建会话、不建帖**。
- **`POST /agent/extract`**：请求 `{"text":"...","image_urls":[]}` → `{intent, is_lnf_context, draft, missing_fields, questions}`，用于**表单智能填充**（把 `draft` 字段灌进发帖表单，最后仍走 `/item/create`）。
- **`POST /agent/session/close`**：请求 `{"session_id":"..."}`（可空）→ 关闭该用户当前会话，幂等。

### 8. `GET /item/:itemID/similar`（无需登录）

- Query：`limit`（默认 5，最大 10）
- 返回：`data` = `ItemResponse[]`（与列表项同构，可直接复用卡片组件）
- 规则：同类型优先，不足用相反类型补齐；按“标签命中 + 地点全链（忽略楼号）+ 全文相关度 + 时间”排序；**不调用 LLM、无额外成本**
- 物品不存在/已删除 → `20001`

### 9. 建议测试用例（新增）

1. **发帖两步**：`/agent/chat` 发「我在图书馆三楼捡到一个黑色保温杯」→ 应 `stage=need_confirm`、`intent=create_found`、`draft.type=1`、`location_detail="三楼"`；再带 `session_id` 发「确认」→ `stage=created` 且 `created_item_id` 有值，`/item/:id` 可查到。
2. **补充信息即发布**：同上第一轮后回「我手机13800000000」→ 直接 `created`，详情里 `contact` 为该手机号。
3. **不发布会话**：第一轮后回「随便说点什么」→ `stage=cancelled`，且“我的发布”无新增。
4. **取消**：第一轮后回「算了不发了」→ `stage=cancelled`，无新增。
5. **询问式判类**：「有人捡到黑色水杯吗」→ `intent=match`（**不应**出现 `draft`、不应建帖）。
6. **会话严格性**：不带 `session_id` 连发两次 → 第二次的 `session_id` 与第一次不同；带乱写的 `session_id` → `120001`。
7. **限流**：同一账号 1 分钟内连续请求 11 次 → 第 11 次 `120004`。
8. **只读抽取**：`/agent/extract` 传「我丢了个黑色保温杯」→ `draft.tag_names` 含「水杯」「黑色」，`missing_fields` 含 `location`/`time`。
9. **相似推荐**：任意存在的 id 调 `/item/:id/similar?limit=5` → 不超 5 条且**不含自己**；不存在的 id → `20001`。
10. **图片**：带 1~3 张 `image_urls` 正常返回；超过 3 张会被截断为前 3 张；`QQ 图床`等公网 URL 同样可用。

### 10. 注意事项

- **agent 建帖会写站内通知**：成功后发帖人收到 `type=0` 系统通知（「智能助手已为你发布信息」）；QQ 侧匹配成功额外写 `type=1 物品匹配`（API 侧匹配不写，避免与响应重复）。
- **联系方式**：仅当用户**明确给出**（手机/微信/邮箱/QQ）才写入 `contact`；**QQ 侧自动用发送者 QQ 兜底**；不会编造或推断。
- **地点**：必须命中地点词表（否则 `null`，后端建帖时补 `140「其他地点」`）；`location_detail` 只放链路表达不了的细节（楼层/门牌），前端展示 = **地点全链 + 详情**，不要重复拼接。
- **标签**：`tag_ids` 一定来自 `tags` 表，可直接提交 `/item/create`；`tag_names` 仅供展示。
- **LLM 延迟**：单次 `/agent/chat` 通常 2~8 秒（判类 + 召回 + 精排），前端务必加 loading；建议请求超时 ≥ 30 秒。

---

## 附：错误码速查（本次对接范围）

| 码 | 含义 | | 码 | 含义 |
|---|---|---|---|---|
| 0 | 成功 | | 20010 | 地点无效 |
| 1 | 请求参数错误 | | 20013 | 图片最多 3 张 |
| 2 | 未登录/Token 无效/权限不足 | | 20014 | 图片参数错误 |
| 5 | 服务器内部错误 | | 30002 | 认领记录不存在 |
| 6 | 数据库错误 | | 30003 | 无权撤回认领 |
| 8 | 机器人（陈松）故障 | | 30004 | 不能认领自己的物品 |
| 9 | 上传文件过大 | | 30006 | 认领前需绑定 QQ |
| 10 | 上传类型不支持 | | 40001 | 标签不存在 |
| 11 | 上传保存失败 | | 40002 | 标签名重复 |
| 10001 | 用户名或密码错误 | | 40004 | 标签名非法 |
| 10002 | 用户名已被占用 | | 40005 | 标签使用中 |
| 10003 | 注册/登录格式不符 / role、status 非法 | | 50001 | 商城兑换积分不足 |
| 10005 | 令牌被禁用（登出/续期） | | 60001 | 通知不存在 |
| 10006 | 用户不存在或被禁用 | | 60004 | 通知查询失败 |
| 10007 | QQ 验证会话已存在 | | 60006 | 通知更新失败 |
| 10008 | 验证码错误 | | 60007 | 通知删除失败 |
| 10009 | 用户不在 QQ 群内 | | 80001 | 地点不存在 |
| 10010 | QQ 号不一致 | | 80002 | 地点名重复 |
| 10011 | 尝试次数过多 | | 80004 | 地点使用中 |
| 10012 | 会话不存在/失效 | | 80005 | 有子地点 |
| 10013 | QQ 已被绑定 | | 11001 | 商品不存在/已下架 |
| 10014 | 积分不够扣 | | 11002 | 商品库存不足 |
| 20001 | 物品不存在 | | 11003 | 商品名非法/重名 |
| 20002 | 物品已关闭 | | 11004 | 商品价格非法 |
| 20003 | 物品已被认领 | | 11005 | 兑换需先绑定 QQ |
| 20005 | 无权操作该物品 | | 20006 | 物品类型非法 |
| 90001 | 公告不存在/不可见 | | 90003 | 公告参数或状态错误 |

完整错误码以后端 `response/response_code.go` 与 `common_response_code.md` 为准。
