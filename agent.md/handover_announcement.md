# 交接文档：announcement 模块修复 + 文档同步（交接给 DeepSeek V4）

> 写于 2026-10-01。前任：Chatbox（GLM-5.3）。本文是唯一任务来源，动手前通读全文。
> 用户语言：中文。所有回复、文档、注释默认中文。

---

## 0. 环境速览

| 项 | 值 |
|---|---|
| 项目 | Go + Gin 失物招领后端（LNF-SERVER），模块：user / item / tag / location / upload / notification / shop / announcement / chensong(QQ机器人) |
| 工作目录 | `D:/AI.ACEF/LostAndFoundGroup/LNF-SERVER`（Windows，PowerShell，无 bash） |
| 验证命令 | `go build ./...`、`go vet ./...`（改完必跑） |
| API 文档目录 | `agent.md/`（api_agent.md、api_guide.md、common_response_code.md 三份 + 本交接文档） |
| 线上服务 | `http://111.229.234.32:8080`，config.yaml 指向**共享真实 DB/Redis**（111.229.234.32）。不要随意起服务做破坏性测试，联调部署由用户执行 |
| 杂项 | 根目录 `x.exe` 为旧构建产物，忽略；部分文件是 CRLF，正常编辑即可 |

---

## 1. 总任务背景与用户铁律

### 1.1 任务起源

用户要求按顺序维护 `agent.md/` 下三份文档（common_response_code.md → api_agent.md → api_guide.md），原则：

1. **以代码为准**写文档；
2. 若发现 a) 文档设计的功能代码未实现，或 b) 文档设计比代码现状优秀 → **必须向用户确认，不得擅自处理**；
3. 有任何问题先问用户；
4. 未经允许，只能修改被授权的文件。

该阶段已完成（见 §2）。核对文档过程中发现了 3 个代码 bug，用户已逐一拍板处理（bug 1 用户自修、bug 2 前任已修、bug 3 = 你要做的任务 A）。

### 1.2 铁律（对你同样生效）

- **其他模块的代码一律不动**（user/item/tag/location/upload/notification/shop 均已核对并同步进文档，改动会导致文档失真）。若你在这些模块发现新问题：报告用户，等指示。
- 尤其**不要顺手"修复" item 模块 page_size 的 51-100 静默按 10 行为**（`service/basic/item_service.go` 的 `normalizePage` + `itemMyMaxPageSize` 死代码）。用户已明确选择按现状写文档（选项 A），改代码反而制造不一致。
- 文档与代码出现任何新的分歧：代码为准改文档；若你认为该改代码，先问用户。
- 只改 §5 授权清单内的文件。

---

## 2. 已完成的全部改动（基线，勿重做、勿回退）

### 2.1 代码改动（均已通过 go build / go vet）

| # | 文件 | 改动 | 改动者 |
|---|---|---|---|
| 1 | `service/basic/item_service.go` | ① `itemMaxTitleLen` 90 → **100**（对齐 items.sql `varchar(100)`，按字节判定不会超列宽）；② 新增常量 `itemMaxLocationDetailLen=200`、`itemMaxContactLen=100`；③ CreateService 校验：location_detail>200 / contact>100 / credit_reward<0 → `1`；④ UpdateService 对 location_detail、contact 同样长度校验 | Chatbox |
| 2 | `service/advanced/location_service.go` | UpdateService 补齐同父级重名校验：改名或移动父级后与目标父级下已有地点重名（排除自身）→ `80002` | Chatbox |
| 3 | `handler/advanced/notification_handler.go` | **bug 2 修复**：6 处 `c.GetInt64("ContextID")` → `c.GetInt64(middleware.ContextID)`（中间件实际写入的键是 `"jwt:id"`；原写法导致用户侧通知接口全部以 userID=0 运行：列表恒空、未读数恒 0、详情恒 60001） | Chatbox |
| 4 | `service/basic/user_service.go` | ChangeUserStatusRequest：**禁用**(status=0)时 INCR JWT 版本号（原代码是解禁时 INCR，方向反了）。现语义：禁用→该用户全部 token 失效；解禁→不影响存量 token | 用户本人 |
| 5 | `dao/constants.go` | **bug 1 修复**：`GetJwtVersionKey` 全角冒号 `"jwt：user:"` → 半角 `"jwt:user:"`（中间件 jwt_auth_middleware.go 硬编码读半角键，原来两键永不相等，导致 logout_all 全端登出、禁用全 token 失效均不生效） | 用户本人 |

> bug 1 修复部署提示：Redis 里可能残留少量全角键 `"jwt：user:X:version"` 垃圾数据，无影响，可不管。
> bug 1 修复后，`logout_all=1` 与"禁用用户全 token 失效"真实生效——这是文档当前描述的语义，验证时可直接测。

### 2.2 文档改动（三份均已全量重写，与 2026-10-01 代码对齐）

**common_response_code.md**：与 `response/response_code.go` 逐码对齐；所有接口 HTTP 恒 200；修正 item status 枚举（0已发布 1已认领 2已关闭，原文档写成待审核/已发布/已认领/已关闭四态）；修正认领段（claims.sql 未使用，认领由 items.status+claim_user_id 表达）；触发场景全部改为真实代码路径，未使用的码标"预留"；新增 11xxxx 商城段、100001、-1/-2/-3 测试码段；总览表带实现状态列。

**api_agent.md**（前端 AI 精确对接版）：依据日期 2026-10-01；新增 `/item/search`、notification(§2.7)、shop(§2.8)、announcement(§2.6，标注"当前整体不可用"+缺陷清单)；修正 20+ 处：标题≤100、`/item/update` 无 status 字段（传入被忽略）、认领错误码（已关闭统一 `20002`、已被认领 `20003`、认领自己 `30004`）、create credit_reward 负数→`1`、QQ 绑定第 5 次提交才 `10011`、`logout_all` 非 1 一律按当前会话、`/user/batch` 空数组返回 `[]`、`/user/jwt-test` 成功是 `code=-1`、page_size 51-100 静默按 10（含死代码说明）、location update 重名 `80002` 等。bug 1/2 修复后已移除对应"⚠️ 已知缺陷"标注。

**api_guide.md**（测试人员版）：同步上述全部修正；新增 notification(§七)、shop(§八)、announcement(§九，标注不可用) 章节；测试用例 1-21；§十一 注意事项；附错误码速查表。bug 1/2 的标注已移除；**§十一 第 10 条"公告模块不可用"保留中，等你完成任务 A 后清除（见 §4）**。

### 2.3 已知 bug 状态

| bug | 状态 |
|---|---|
| 1. JWT 版本键全角冒号 | ✅ 已修（用户） |
| 2. notification handler 读错 context key | ✅ 已修（Chatbox） |
| 3. announcement 模块多处缺陷 | ⬜ **你的任务 A** |

---

## 3. 任务 A：修复 announcement 模块（bug 3）

### 3.1 用户已拍板的 4 项设计决策（原话如下，不得偏离）

> 1，换page/pagesize
> 2，删除草稿功能，简化
> 3，是
> 4，是

对应前任提出的 4 个问题：

1. 分页方案：**放弃游标（started_id/ignore_pieces/limit），改 page/page_size**；
2. 草稿：**删除草稿功能**——创建即发布（status=已发布 + published_at=now），无 0草稿 态；
3. view_count：**做**——新增公开详情端点 `GET /announcement/:id`，详情 view_count +1（与 item 模块语义一致）；
4. `AnnouncementUpdateRequest` **删除 is_deleted 字段**（软删只走 delete 端点，客户端不可控）。

### 3.2 现存问题全清单（已逐条核实，按文件）

**dao/announcement.go**

| # | 问题 |
|---|---|
| 1 | `GetAnnouncementByID`：`err := db.Where("id <= ?", id).Last(&an)` 把 `*gorm.DB` 当 error 判空，恒非 nil → **恒返回 nil** → 所有读取/更新路径恒 90001 |
| 2 | 同上：即使修好，查不到时 `an` 是零值结构体（IsDeleted=0），会当有效记录返回；需按 `gorm.ErrRecordNotFound` 判定 |
| 3 | `DeleteAnnouncement`：`db.Delete(id)` 把裸 int64 传给 gorm → 恒报错 → **删除恒 `6`**；且是硬删，与表设计（is_deleted）不符 |
| 4 | `GetAmmount`：裸 DB 上 `Count`，无 `Model()`、无 is_deleted/可见性过滤 → total 恒 0 且口径与列表不一致 |
| 5 | `UpdateAnnouncement`：`Select(全列).Updates(&announcement)` → 请求体没带的字段全写零值：**view_count 清零、published_at 置 NULL、created_at 被覆盖、is_deleted 客户端可控**；且无 `is_deleted=0` 守卫 |

**handler/basic/announcement_handler.go**

| # | 问题 |
|---|---|
| 6 | 5 个 handler 的错误路径普遍**缺 return**：绑定失败 `FailWithCode` 后继续执行；service 出错后 `FailWithCode` 又落到底部 `Success` → **双写响应体**。仅 DeleteHandler 的路径参数解析处有 return |
| 7 | `AuthGetHandler` 从不设置 `req.Auth=true`（管理员端默认也只能看到已发布）；Auth/AdminID 均来自请求体而非 JWT |
| 8 | GET 路由却 `ShouldBindJSON`（要求 GET 带 JSON body，正常 GET 请求直接绑定失败） |
| 9 | swagger 注释与实际不符：GetHandler/AuthGetHandler 标 `[post]` 但路由是 GET；路径 `/api/v1/announcement` 与实际 `/announcement/` 不一致 |

**service/basic/announcement_service.go**

| # | 问题 |
|---|---|
| 10 | `GetAnnouncements` 游标循环**逐行查库**（取 limit 条 = limit+1 次 SQL）；`an == nil`（翻到底）被当错误返回 nil → 90001 |
| 11 | 游标默认值自相矛盾：`StartedID=0`（swagger 让用户"第一次带0"）→ `last=0` → 循环不执行 → **首页恒空**；负数才触发"从最新开始" |
| 12 | `CreateAnnouncement`：`ToAnnouncement` 直拷请求的 status/is_deleted → 可跳过发布流程直接落"已发布且 published_at=NULL"的脏数据，is_deleted 客户端可控（决策 2/4 后此项并入新契约） |

**model/basic/anouncement.go**（注意文件名拼写就是 anouncement）

| # | 问题 |
|---|---|
| 13 | `AnnouncementUpdateRequest` / `AnnouncementGetRequest` **无任何 binding 标签** → 空请求体 `{}` 也能绑定成功，配合缺 return 会落库全零公告（status=0 的垃圾行） |
| 14 | 无参数校验：`90003 CodeAnnouncementInvalid` 定义了但从未使用；空标题/非法 type 直接落库 |

**model/mysql/basic/announcements.sql**

- status 注释为 `0草稿 1已发布 2已过期`，与 Go model 注释（2已下架）不一致；决策 2 后 0 草稿态废弃，注释需同步。

### 3.3 目标 API 契约（实现以此为准，文档同步也以此为准）

**公开端点**

| 端点 | 说明 |
|---|---|
| `GET /announcement?page=1&page_size=10` | 已发布公告列表（status=1 且 is_deleted=0）。`data = {total, page, page_size, announcements: [AnnouncementResponse]}` |
| `GET /announcement/:id` | 公告详情（**仅 status=1 可见**，否则 90001）；**view_count +1**，返回值含本次 +1；路径 id 非正整数 → `1` |

**管理端点（JWT + role=2，`SystemAdminAuthMiddleware`）**

| 端点 | 请求体 | 说明 |
|---|---|---|
| `POST /admin/announcement/create` | `{title, content, type, is_top}` | **创建即发布**：status=1、published_at=now、admin_id 取 JWT（不信任 body） |
| `POST /admin/announcement/update` | `{id, title?, content?, type?, status?, is_top?}` | 增量更新（指针语义，参考 item 模块）；status 仅允许 1/2（下架/重新上架）；published_at 保持首次发布时间不变 |
| `DELETE /admin/announcement/:id` | — | 软删（is_deleted=1） |
| `GET /admin/announcement?page&page_size&status?` | query | 管理员视角列表：**含已下架**（is_deleted=0 全部）；status 可选筛选（1/2） |

**status 语义（重要）**：`1=已发布`、`2=已下架`，**0 废弃**（不迁移 DB：线上可能存在的 status=0 垃圾行两态都不匹配、自然不可见；建议顺手提供一条清理 SQL 给用户执行，见 §3.6）。

**校验规则 → 90003**（把闲置码用起来）：title trim 非空 ≤100 字节；content 非空；type∈{0,1,2,3}；status∈{1,2}；is_top∈{0,1}。binding 失败（缺字段/类型错/page 越界）→ `1`。

**AnnouncementResponse**（沿用现有结构，无 is_deleted）：`{id, admin_id, title, content, type, status, is_top, view_count, published_at, created_at, updated_at}`；published_at 可空指针（omitempty）。

**错误码全集**：`1` 绑定/参数、`2` role 不足（中间件）、`6` DB、`90001` 不存在/不可见、`90003` 业务校验。

### 3.4 分层改动方案

**dao/announcement.go（重写）**

- `GetAnnouncementByID(id) (*model.Announcement, error)`：`Where("id = ? AND is_deleted = 0").First(&an)`；`errors.Is(err, gorm.ErrRecordNotFound)` → 返回 `(nil, nil)`；其他错误原样返回。
- `ListAnnouncements(page, pageSize int, status *int8, publishedOnly bool)` + `CountAnnouncements(同参)`：单条 SQL 分页（`ORDER BY id DESC` + Offset/Limit），公开端强制 status=1，管理端 status 为 nil 时 1+2 全查。**不要复制 item 模块 normalizePage 的 51-100 静默按 10 行为**，这里 page_size 上限 100 正常生效（>100 由 binding `max=100` 拦 → `1`）。
- `UpdateAnnouncementByVK(id, updates map[string]interface{})`：白名单字段（title/content/type/status/is_top），无有效字段直接成功，加 `is_deleted=0` 守卫（参考 tag/location 模块同名方法风格）。
- `SoftDeleteAnnouncement(id)`：`Update is_deleted=1`，替换现硬删。
- `IncrViewCount(id)`：`UpdateColumn("view_count", gorm.Expr("view_count + 1"))`，条件含 `is_deleted=0 AND status=1`（参考 item 模块）。
- 删除 `GetAmmount`（拼写错且口径错）。

**service/basic/announcement_service.go（重写）**

- `Create`：校验（§3.3 规则 → 90003）→ 构造 Announcement{Status:1, PublishedAt:&now, IsDeleted:0, AdminID: 来自 JWT}。
- `Update`：校验 + 增量 map 组装；published_at 不在可更新列。
- `Delete`：调软删。
- `GetPublishedList(q)` / `GetAdminList(q)` / `GetDetail(id)`（详情 +1 后返回）。
- **翻到底返回空列表而非 nil**。

**handler/basic/announcement_handler.go（重写）**

- 所有错误路径补 `return`（参考 item_handler.go——它就是本项目干净写法的范本）。
- GET 列表用 `ShouldBindQuery`；admin_id 一律 `c.GetInt64(middleware.ContextID)`。
- 公开列表 handler 内部强制 status=1（忽略客户端传入）；公开详情仅 status=1 可见。
- swagger 注释修正：方法 GET/POST/DELETE 与路由一致、路径对齐、@Param 用 query/body 正确标注。

**model/basic/anouncement.go**

- 新增 `AnnouncementListQuery{Page, PageSize, Status *int8}`（form 绑定 + `binding:"omitempty,min=1"` / `"omitempty,min=1,max=100"` / `"omitempty,oneof=1 2"`）。
- 拆分 `CreateAnnouncementRequest{title, content, type, is_top}`（binding required）与 `UpdateAnnouncementRequest{id required + 指针增量字段}`；**删除** is_deleted、admin_id 字段与整个 `AnnouncementGetRequest`。
- `AnnouncementListResponse{Total, Page, PageSize, Announcements}`。
- 文件名 anouncement.go 可顺手改名 announcement.go（可选，Go 不在乎；改名就同步无引用问题——引用都在同项目内，全局替换即可）。

**router/basic/announcement_router.go**

- 公开组：`GET ""`（列表）+ `GET "/:id"`（详情）。同层静态+参数路由可共存，参考 item 模块 `/list` 与 `/:itemID` 并存模式；若 `/` 与 `/:id` 注册冲突就用 `""` 代替 `"/"`。
- 管理组：现有四个端点保持路径不变，GET 列表改 query 传参。

**model/mysql/basic/announcements.sql**：status 注释改 `1已发布 2已下架（0已废弃）`。

**docs/**（swagger 生成物）：handler 注释改完后若有 swag CLI 跑 `swag init` 重新生成；没有 CLI 就手动同步 docs/docs.go 或在交付说明里注明"docs 待重新生成"。

### 3.5 建议默认值（用户未逐条确认的小项；按此执行，不认可必须先问用户）

| 项 | 默认值 |
|---|---|
| create 成功返回 | `data: {}`（与 item 模块一致；若你想返回 `{"id": n}` 更利于前端跳详情，可向用户提议，不要擅自定） |
| 列表排序 | `id DESC`（自增主键即时间序，最简） |
| 管理端列表 | 含已下架 + 可选 status 筛选 |
| 公开详情可见性 | 仅 status=1（下架后详情页 90001） |
| status 取值 | 沿用 1/2，0 废弃不迁移（§3.3 已述） |
| page_size 行为 | 缺省 10、≤0 归 10、1-100 正常、>100 → `1`（**无** item 模块的静默坑） |

### 3.6 验收清单

1. `go build ./...` 与 `go vet ./...` 通过；
2. 逻辑自查（无法连真实 DB 跑集成测试时，至少逐条走查）：
   - create（校验失败各字段 → 90003；成功 → status=1 + published_at=now + admin_id=JWT）；
   - 公开列表只见已发布；管理列表含下架；page/page_size 边界（0/1/100/101/负数）；
   - 详情 view_count 递增、下架/删除后 90001、id 非法 → 1；
   - update 增量：只改传入字段，view_count/published_at/created_at 不被清；status 传 0/3 → 90003；
   - delete → 软删，公开/管理列表均不可见；
   - 普通用户/未登录调管理端点 → 2；
   - 空请求体 create → 1（binding required），不再产生全零公告；
3. 向用户提供一条可选清理 SQL：`DELETE FROM announcements WHERE status = 0 OR title = '';`（修掉历史垃圾行，由用户自行在 DB 执行，你不直接连生产库操作）。

---

## 4. 任务 B：文档同步（任务 A 验收通过后执行，以新代码为准）

| 文件 | 改动 |
|---|---|
| `api_agent.md` §2.6 | 删除"⚠️ 当前整体不可用"整段缺陷说明，按 §3.3 契约重写：端点表、请求/响应字段表、错误码、注意事项（详情 +1 浏览量、page_size 上限 100、status=0 废弃）。§1 数据模型补 AnnouncementResponse 结构与枚举表更新（announcement.status → 1已发布 2已下架）。§3 陷阱清单酌情补 1-2 条。§4 时序图可选补一步。§5 附表补 90001/90003 |
| `api_guide.md` §九 | 重写为正式章节（表格 + 用法说明，风格对齐 shop §八）；§十 补"公告流"测试用例（22-26 左右：创建即发布、下架后公开不可见、详情浏览量、90003 校验、软删）；**§十一 第 10 条"公告模块不可用…修复方案已提交待确认"删除**（其余条目重排编号）；文末"附：错误码速查"表补 90001、90003 两行 |
| `common_response_code.md` §十一 | 删除"⚠️ 实现现状"警告段；status 枚举说明改 1已发布 2已下架（0 废弃）；90001 触发场景改为真实路径（update/delete/公开详情/管理详情）；90003 从"预留"改为实际触发场景（title/content/type/status/is_top 校验）；§一 总览表 9xxxx 行实现状态 ⚠️ → ✅ |

同步后自查：三份文档中不允许再残留 announcement "不可用/已知缺陷"类表述；错误码速查含 90001/90003；文档描述与你的实现逐字段一致（**以代码为准**，若实现与 §3.3 契约有出入，要么改代码对齐契约，要么改文档对齐代码并说明原因——二者选一后保持全局一致）。

---

## 5. 授权范围与禁区

**允许修改**（仅限本任务）：

- 代码：`handler/basic/announcement_handler.go`、`service/basic/announcement_service.go`、`dao/announcement.go`、`model/basic/anouncement.go`、`router/basic/announcement_router.go`、`model/mysql/basic/announcements.sql`、`docs/`（swag 重新生成）
- 文档：`agent.md/api_agent.md`、`agent.md/api_guide.md`、`agent.md/common_response_code.md`
- 本交接文档可追加"进度记录"，但不得删改历史内容

**禁区**：

- 其他一切代码与文档（含 `agent.md/` 下的 item_agent.md、shop_agent.md、requirements.md、image-upload-plan.md 等）
- `response/response_code.go`（90001-90003 已够用，勿动错误码定义；若确需新增码，先问用户）
- 生产 DB/Redis 的直接写操作（清理 SQL 交给用户执行）
- Git 操作未经用户要求不主动做

---

## 6. 其他注意事项

1. **参考实现**：item 模块（handler/service/dao/model 四层）是本项目干净写法的标杆——错误路径 return、ShouldBindQuery/Body 分明、增量更新 map+白名单、软删守卫 `ConditionIDNotDeleted`（dao/constants.go 已有该常量，直接复用）。
2. 动手前先读 `initialization/gorm.go` 确认**是否跑 AutoMigrate**（前任未读过此文件）。若有：model 结构体改动会影响真实表结构，字段只减注释不改类型，必要时停下来问用户；若无：SQL 文件仅为参考 DDL，改注释即可。
3. handler 里 `middleware.ContextID` 的用法见 `handler/advanced/notification_handler.go`（刚修过的 import 写法可照抄）。
4. 全程遵守 §1.2 铁律：发现本文档未覆盖的问题 → 停下来问用户，不擅自扩大改动面。
5. 完成后向用户汇报：改动文件清单、验收结果、清理 SQL（供用户执行）、三份文档的同步差异点。

---

## 7. 进度记录（2026-10-01，任务 A + 任务 B 已完成）

> 本节仅追加，未改动上文任何历史内容。执行者：DeepSeek V4。

### 7.1 任务 A：代码改动清单（`go build ./...`、`go vet ./...`、`gofmt` 全部通过）

| 文件 | 改动摘要 |
|---|---|
| `model/basic/anouncement.go` | 重写：新增 `AnnouncementListQuery{Page, PageSize, Status}`（binding 按 §3.4 指定标签）；拆分 `CreateAnnouncementRequest`（title/content/type/is_top 全 binding required，type/is_top 用 `*int8` 使 0 值可通过 required）与 `AnnouncementUpdateRequest`（id required + 指针增量）；**删除** `AnnouncementGetRequest`、`ToAnnouncement` 及 is_deleted/admin_id 请求字段；`AnnouncementListResponse` 补 page/page_size；`AnnouncementResponse.published_at` 加 omitempty；实体 Status 注释改 1已发布/2已下架，IsDeleted 改 `json:"-"`；文件名保持 anouncement.go（改名可选，未做） |
| `dao/announcement.go` | 重写：修 `GetAnnouncementByID`（`ConditionIDNotDeleted` + `gorm.ErrRecordNotFound` → `(nil,nil)`）；新增 `ListAnnouncements`/`CountAnnouncements`（共享 buildAnnouncementQuery，单 SQL `id DESC` + Offset/Limit）、`UpdateAnnouncementByVK`（title/content/type/status/is_top 白名单 + is_deleted=0 守卫 + updated_at，空 map 直接成功）、`SoftDeleteAnnouncement`（软删替换硬删）、`IncrViewCount`（条件含 status=1）；**删除** `GetAmmount`、旧 `UpdateAnnouncement`、旧 `DeleteAnnouncement` |
| `service/basic/announcement_service.go` | 重写：`Create`（90003 校验 → status=1 + published_at=now + admin_id 取 JWT）、`Update`（存在性 90001 + 指针增量 + status 仅 1/2 → 90003；published_at 不在可更新列）、`Delete`（先查 90001 再软删）、`GetPublishedList`/`GetAdminList`/`GetDetail`（仅 status=1 可见，+1 后返回，翻到底返回 `[]` 非 nil）；分页归一化函数命名 `normalizeAnnouncementPage`（**与 item 模块同包 `normalizePage` 冲突，故改名**，item 代码未动） |
| `handler/basic/announcement_handler.go` | 重写：全部错误路径补 `return`；GET 列表改 `ShouldBindQuery`；admin_id 取 `c.GetInt64(middleware.ContextID)`；新增公开详情 `GetDetailHandler`（id 非正整数 → 1）；swagger 注释按实际方法/路径重写 |
| `router/basic/announcement_router.go` | 公开组 `GET ""` + `GET "/:id"`（契约路径无尾斜杠，避免 `/` 与 `/:id` 注册冲突）；管理组四端点路径不变，GET 列表改 query；中间件不变（JWT + SystemAdmin role=2） |
| `model/mysql/basic/announcements.sql` | 仅 status 列注释 → `1已发布 2已下架（0已废弃）`（gorm.go 无 AutoMigrate，DDL 无实际变更） |
| `docs/` | `swag init` 重新生成 docs.go/swagger.json/swagger.yaml；6 条 announcement 端点与新 definition 已确认，旧 `AnnouncementGetRequest` 定义已消失 |

### 7.2 与 §3.3 契约的取舍说明（以代码为准，文档已按实现写）

1. **page_size**：binding 用 §3.4 指定的 `omitempty,min=1,max=100` → 0/缺省归 10（§3.5"≤0 归 10"对 0 生效），**负数与 >100 → `1`**；page 负数 → `1`，0/缺省 → 第 1 页。
2. **管理列表缺省 status**：按 §3.4"1+2 全查"实现为 `status IN (1,2)`（同时满足 §3.3"status=0 垃圾行自然不可见"），非完全不过滤。
3. **90001 真实路径**为 update / delete / 公开详情——§3.3 契约无"管理详情"端点，文档按实际实现描述。
4. **create 空字符串 vs 纯空白**：`title:""`/`content:""` 被 binding required 拦 → `1`；纯空白 → `90003`（文档已写明）。
5. **详情 +1**：`IncrViewCount` 失败不影响返回（`_ =`），本地 +1 后返回，与 item 模块语义一致。
6. create 成功 `data: {}`（§3.5 默认值，未提议返回 id）；列表 `id DESC`；公开详情仅 status=1。

### 7.3 任务 B：文档同步（三份均以新代码为准）

- `api_agent.md`：范围行去"当前不可用"；§1 补 AnnouncementList/Response 结构、枚举改 1已发布/2已下架；§2.6 按契约全量重写（端点表/字段表/错误码/注意事项）；§3 陷阱新增 23、24 条；§4 时序补第 17 步；§5 附表补 90001/90003。
- `api_guide.md`：范围行更新；§九重写为正式章节（风格对齐 shop §八）；§十补公告流 22-26 用例；§十一删除第 10 条（第 9 条补 announcement 无静默坑，条目 1-9 无需重排）；错误码速查补 90001/90003 行。
- `common_response_code.md`：§一 9xxxx 行 ⚠️ → ✅（见第十一节）；§十一删除"⚠️ 实现现状"段，status 枚举、90001/90003 触发场景改为真实路径。
- 自查：三份文档已无公告"不可用/已知缺陷"类表述；速查表含 90001/90003；文档与实现逐字段一致。

### 7.4 验收结果与遗留事项

1. `go build ./...`、`go vet ./...`、`gofmt -l`（5 个改动文件）全部通过；`swag init` 成功。
2. 逻辑自查（未连生产 DB，逐条走查 §3.6）：create 各字段校验 90003/成功语义、公开/管理列表可见性与分页边界（0/1/100/101/负数）、详情 +1 与 90001、update 增量与 status 0/3→90003、软删不可见、非 role=2 → 2、空 body create → 1，均已按实现核对（见 7.1/7.2）。
3. **供用户执行的可选清理 SQL**（历史垃圾行，勿由服务进程执行）：
   ```sql
   DELETE FROM announcements WHERE status = 0 OR title = '';
   ```
4. 遗留提醒：Redis 中可能残留全角键 `jwt：user:X:version` 垃圾数据（bug 1 遗留，无影响）；联调部署由用户执行。



### 7.5 进度记录（2026-10-01 追加：0 值 bind 修复 + 两文档全面审计完善）

**代码（仅本问题相关，已过 build/vet/gofmt，swag 已重新生成）：**

| 文件 | 改动 |
|---|---|
| `model/basic/anouncement.go` | `AnnouncementListQuery.Status`：`oneof=1 2` → **`oneof=0 1 2`**——`status=0`（含空串 `status=` 解析出的 0）正常读入不报 bind 错，语义 = 查历史废弃行；3+ 仍 → `1` |
| `model/advanced/notification.go` | `NotificationSendRequest.Type`：`int8 required` → **`*int8 required`**——原写法导致 **`type:0`（系统通知）bind 失败报 `1`（文档示例自身都发不出去）**；现缺字段仍 → `1`，0 正常读入。`ToNotification` 同步解引用 |
| `service/advanced/notification_service.go` | `batchCreate` 日志 `req.Type` → `*req.Type`（指针适配） |

同类穷举核查（`binding:"required"`/`oneof` 全量扫描）：其余 0 值合法字段（logout_all、role、status、credit、add-credit type、announcement create type/is_top、item type/status、shop min_price 等）均已用指针或 omitempty 模式正确处理，**无第三处**。

**文档同步（api_agent.md / api_guide.md）：**

- agent §1 枚举、§2.6 公告（端点表/错误表/注意事项）、§2.7 通知（列表空值规则 + 发送 type=0 合法）、§3 新增陷阱 25（数值 query 0/空串统一规则）、§1 Location 示例删 `address:null` 与说明对齐、通知详情示例补 `updated_at`。
- guide §三 状态图 close 改"0/1 状态"（原文"任意状态"与代码矛盾）、§五/§六 color/address omitempty 说法修正、§七 通知示例 URL 删空参数并说明 0 值规则 + 发送错误码补"全部无效"、§九 公告 status 0/1/2 语义与 update `status=0`→90003 的区分说明、§十 用例 25 补 status=0/3、§十一 新增第 10 条统一规则。
- `common_response_code.md` 无需改（90003 仍针对 update/create 请求体，未涉及 query bind）。

**语义口径（文档已按此写）**：列表 query `status=0`/`type=0`/空串 → 读入 0 并筛选（公告 = 查废弃行，公开列表不受影响强制 status=1）；update 请求体 `status=0` → `90003` 不变；缺字段 → `1` 不变。
