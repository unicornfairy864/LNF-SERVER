# 管理员数据分析模块实施文档

> 状态：阶段1已批准；批次1、2完成；批次3接口及CSV已实现，自动验证通过，真实库验收待用户执行。
> 最新进展：9个GET接口已接入，6个聚合接口支持CSV；共享API文档、Swagger同步及数据库实测由用户执行。
> 范围：`agent.md/admin.md` 的管理员统计需求；以 `.clinerules/admin-stats.md` 为最高约束。
> 调研日期：2026-10-03。

## 1. 现状结论

### 1.1 技术与分层

- Go module：`github.com/unicornfairy864/LNF-SERVER`；`go.mod` 声明 Go 1.26.4，本机 Go 1.27.0。
- Web/ORM：Gin 1.12、GORM 1.31、MySQL driver 1.6。
- 项目分层：handler -> service -> dao -> model，路由独立位于 router。
- 所有 JSON 响应 HTTP 200，信封为 `{code,message,data}`；成功 `code=0`。
- `initialization/gorm.go` 不运行 AutoMigrate；DDL 文件不会自动执行。
- DSN 使用 `parseTime=True&loc=Local`，代码未固定 `time.Local`，数据库 session time_zone 也未设置。

### 1.2 管理员能力

- role=1：失物招领/服务管理员；`ServiceAdminAuthMiddleware` 接受所有 role != 0。
- role=2：系统管理员；`SystemAdminAuthMiddleware` 只接受 role=2。
- item router 已有 `/admin` 空路由组并使用 role>=1，但统计模块应建立独立 `/admin/stats` 路由。
- 当前没有统计 handler/service/dao/model/router，也没有管理员统计专用错误码。

### 1.3 数据模型与历史语义

- `items.status`：0 已发布、1 已认领、2 已关闭；逻辑删除字段 `is_deleted`。
- 认领时写入 `claim_user_id`、`claim_time`；其中 `claim_time` 是认领发生时间。
- 撤回认领会恢复 status=0，并清空 `claim_user_id`、`claim_time`。
- 发帖者自行关闭允许 status 0/1 -> 2，并清空 `claim_user_id`、`claim_time`。
- 确认认领或自动关闭执行 status 1 -> 2，保留 `claim_user_id`、`claim_time`，但没有独立完成时间；`updated_at` 是当前唯一近似完成时间。
- 因此当前快照不能还原已撤回认领，也不能还原“从已认领状态自行关闭”的历史。
- 地点是层级树，item 仅保存一个直接 `location_id`；标签通过 `item_tags` 多对多关联。

### 1.4 现有索引

- items：`(is_deleted,type,status,lost_found_time)`、`(user_id,is_deleted,created_at)`、`(is_deleted,status,created_at)`、`(location_id,is_deleted,type,status,lost_found_time)`、`(claim_user_id,is_deleted,claim_time)`。
- users：`created_at` 单列索引。
- item_tags：唯一 `(item_id,tag_id)`，以及 `(tag_id,item_id)`。
- 第一批按 created_at/status/claim_time 聚合可部分使用现有索引；组合时间查询仍需在真实数据量和 `EXPLAIN` 后决定是否加索引。

## 2. 需求差距与风险

1. 原需求声称 11 个接口，正文只有 9 个 GET 接口和 CSV 横切能力。
2. 原需求把 `claim_time` 称为归还时间，但代码语义是认领时间。
3. 没有精确的归还完成时间；以 `updated_at` 近似会被其他更新污染，但 status=2 后现有业务通常不再更新 item。
4. items 快照无法提供严格的认领事件历史；漏斗与历史认领数只能做当前快照近似。
5. 数据库实际 MySQL 版本未知；中位数 SQL 方案不能确定。
6. 应用与数据库时区未知，不能假设 `loc=Local` 就是东八区。
7. CSV 文件流与统一 JSON 信封互斥，需要定义为成功响应例外。

## 3. 建议统一契约

以下为建议值，用户确认后才成为实现契约。

### 3.1 权限与时间

- `/api/v1/admin/stats/*` 使用 JWT + `ServiceAdminAuthMiddleware`，即 role>=1。
- 日期格式固定 `YYYY-MM-DD`，业务时区固定 `Asia/Shanghai`。
- `start_date` 与 `end_date` 必须同时出现；均为含首尾自然日，内部转换为 `[start 00:00, end+1 day 00:00)`。
- `days` 与日期对互斥；`days` 范围 1-366，表示包含今天在内的最近 N 个自然日。
- 同时传两种范围、只传单边日期、非法日期、start>end 或结束日期晚于今天均返回 code=1。
- 普通统计默认 `days=30`；overview 默认当前自然月，同时返回上一个完整自然月对比；清单默认不限制创建上界，仅按阈值筛选。
- 百分比保留两位小数，四舍五入；分母为 0 时返回 0，不返回 null/NaN。
- 环比 `(current-previous)/previous*100`；previous=0 且 current=0 返回 0，previous=0 且 current>0 返回 null，并附 `comparable=false`。

### 3.2 暂定归还口径

- `pending`：`status=0`。
- `claimed`：`status=1 AND claim_user_id IS NOT NULL`，表示认领中，不等于成功归还。
- `returned`：`status=2 AND claim_user_id IS NOT NULL AND claim_time IS NOT NULL`，表示确认或自动关闭的成功归还快照。
- `closed_without_return`：`status=2 AND claim_user_id IS NULL`。
- 用户确认同批次口径：先按 `created_at` 将物品归入发布周期，再以当前快照判断该队列的 returned/pending/claimed/closed；趋势中 returned 也归入该物品创建日。
- overview/trend/locations 不把 `updated_at` 当作归还事件时间；没有独立完成时间，不声称可计算准确归还完成趋势/时长。
- 批次 2 按用户批准以 `updated_at-created_at` 计算近似归还时长，明确 updated_at 可被其他更新污染；`claim_time` 是认领时间，不能作为完成时间。
- 归还率：`returned / (returned + pending + claimed + closed_without_return)`，即所有未删除发布快照。
- 当前结构无法统计撤回认领事件，漏斗认领层只能按当前仍保留的 distinct claim_user_id 近似。

## 4. 接口设计

所有 JSON 数据均置于统一信封的 `data` 中；所有 items 查询显式过滤 `is_deleted=0`。

### 批次 1

#### GET `/admin/stats/overview`

参数：可选 `start_date/end_date` 或 `days`。缺省当前自然月；对比期为紧邻且天数相同的上一周期，若缺省当前月则对比上一个完整自然月。

```json
{
  "period":{"start_date":"2026-10-01","end_date":"2026-10-03"},
  "published":{"value":10,"previous":8,"change_percent":25.00,"comparable":true},
  "returned":{"value":4,"previous":2,"change_percent":100.00,"comparable":true},
  "pending":{"value":5,"previous":4,"change_percent":25.00,"comparable":true},
  "return_rate":{"value":40.00,"previous":25.00,"change_percent":60.00,"comparable":true},
  "pending_over_24h":2
}
```

发布、returned、pending 均按 created_at 纳入同一发布队列，状态按当前快照判断；不是历史时点状态快照。`pending_over_24h` 按当前时刻前 24 小时的 created_at 阈值筛选。

#### GET `/admin/stats/trend`

参数：可选日期范围，默认 days=30，最大 366。按自然日返回连续数组并在 service 补零。

```json
{"start_date":"2026-09-04","end_date":"2026-10-03","points":[{"date":"2026-09-04","published":3,"returned":1}]}
```

发布与 returned 均按 created_at 自然日归属；returned 额外要求当前 status=2 且 claim 字段非空。

#### GET `/admin/stats/locations`

参数：日期范围默认 days=30；`limit` 默认 4，范围 0-100，0 返回全部直接地点桶。

```json
{"total":10,"locations":[{"location_id":3,"name":"教学楼A","count":4,"percent":40.00}],"unknown":{"count":1,"percent":10.00}}
```

按 item 直接 location_id 聚合，不向父地点归并；分母为期内全部未删除发布量；NULL 单列 unknown；排序 count DESC, location_id ASC。

### 批次 2

#### GET `/admin/stats/return-duration`

参数：日期范围默认 days=30；`group_by=none|location|type`，默认 none。样本是期内近似完成的 returned（按 updated_at 纳入范围），时长为 `updated_at-created_at` 秒；明确为近似归还时长，不能还原精确归还完成事件。

返回 overall 和 groups，每组包含 `count/average_seconds/median_seconds`。实际 MySQL >=8 时使用窗口函数精确中位数；否则由 DAO 仅拉取有界的排序时长列并在 Go 中计算，最大范围 366 天。

#### GET `/admin/stats/items/stagnant`

参数：`days` 默认 7，范围 1-3650；`page` 默认 1；`page_size` 默认 10、最大 100。筛选当前 status=0 且 created_at <= now-days，排序 created_at ASC, id ASC。

#### GET `/admin/stats/items/high-view`

参数：`min_views` 默认 50，范围 1-2147483647；`days` 默认 7；分页同上。筛选 status=0、view_count>=min_views、created_at<=阈值，排序 view_count DESC, created_at ASC, id ASC。

两类清单只返回：`id,title,type,status,location_id,location_name,view_count,created_at,stagnant_days`，不返回 description/contact/user 信息。

#### GET `/admin/stats/time-heatmap`

参数：日期范围默认 days=30、最大 366。返回 `matrix[7][24]`；星期顺序 Monday-Sunday，小时 0-23，按 created_at 东八区聚合。

### 批次 3

#### GET `/admin/stats/distribution`

参数：日期范围默认 days=30；必填 `dimension=type|tag`。

- type 每个 item 只属于一个桶。
- tag 中一个 item 可进入多个桶，因此桶 count 总和可大于发布总数；percent 分母仍为期内 distinct item 总数。
- 未打标签 item 单列 `untagged`；排序 count DESC, bucket id ASC。
- 每桶返回 `published/returned/return_rate`；returned 使用当前快照近似。

#### GET `/admin/stats/funnel`

采用周期内活跃近似口径：期内新增注册用户；期内 distinct 发布用户；期内当前仍保留 claim_time 的 distinct 认领用户；期内近似完成 returned 涉及的 distinct claim_user_id。各层不是严格嵌套 cohort，因此仅返回相邻比率，不宣称严格漏斗；若需严格漏斗必须新增事件表并另立需求。

#### CSV 横切能力

- 覆盖 overview、trend、locations、distribution、return-duration、time-heatmap。
- `export=csv` 时成功响应不使用 JSON 信封，返回 `text/csv; charset=utf-8`、UTF-8 BOM、RFC 5987 文件名；参数/查询失败仍返回现有 HTTP 200 JSON 错误信封。
- 复用同一 service/DAO 结果，不另写统计口径；最大日期范围 366 天；流式写出并 flush，不落盘。

## 5. 实现与查询方案

- model：每个接口使用专属 query/response DTO，日期字符串在 service 解析；不得复用含 contact 的 ItemResponse。
- handler：`ShouldBindQuery`，失败 code=1；调用 service；Swagger 注释写完整 `/api/v1/admin/stats/...`。
- service：统一日期范围解析、东八区边界、百分比/环比、日期补零和中位数；不在内存加载完整 item 行。
- dao：固定 SQL 片段和白名单 group_by；聚合返回精简 row DTO；禁止用户输入拼接 SQL。
- 时间查询使用 `>= start AND < end`。用户确认部署应用与数据库均使用东八区。
- 索引先不新增。批次完成后由用户在真实库执行 EXPLAIN；必要时建议新增 `(is_deleted,created_at)`、`(is_deleted,status,updated_at)`、`(is_deleted,status,view_count,created_at)`，实际执行由用户决定。

## 6. 文件计划

### AI 可新建并修改

- `agent.md/admin_agent.md`
- `model/advanced/admin_stats.go`
- `dao/admin_stats_dao.go`
- `service/advanced/admin_stats_service.go`
- `service/advanced/admin_stats_service_test.go`
- `handler/advanced/admin_stats_handler.go`
- `router/advanced/admin_stats_router.go`
- 后续确有必要时，新建同模块测试文件和 `model/mysql/advanced/admin_stats_indexes.sql`，需再次获批。

### 需要用户手工修改

编码完成后 AI 提供最小代码片段，由用户修改：

- `dao/enter.go`：注册 `AdminStatsDao`。
- `service/enter.go`：注册 `AdminStatsService`。
- `handler/enter.go`：注册 `AdminStatsHandler`。
- `router/enter.go`：注册 `AdminStatsRouter`。
- `initialization/router.go`：调用 `router.AdminStatsRouter.CreateRouter(api)`。
- 批次 3 完成后，用户手工更新共享 API 文档并执行 `swag init`。

## 7. 验证方案

- service 表驱动测试覆盖日期默认值/互斥/边界、东八区跨日、环比除零、补零、中位数奇偶样本。
- DAO 查询通过 GORM DryRun 或 sqlmock 不可行（项目无 sqlmock 依赖）；不新增依赖，提供固定测试数据 SQL 与真实测试库验收步骤。
- 接口验收覆盖无 token、role=0、role=1/2、非法参数、空数据和正常结果。
- 每批执行：仅对模块新文件 gofmt，随后 `go test ./...`、`go build ./...`、`git diff --check`。

## 8. 待用户确认

用户于 2026-10-03 选择方案 A：批准本节第 1-8 项；实际环境为 MySQL 8.x，应用与数据库均使用东八区；授权按三个批次连续开发并在每批完成后报告。另确认 overview/trend/locations 使用同批次（发布队列）口径：先按 `created_at` 将物品归入周期，再按当前快照判断 returned/pending/claimed/closed；归还归属以物品 `created_at` 为准，不使用 `updated_at` 作为归还事件时间。

1. 接口总量按正文的 9 个 GET + CSV 横切能力执行，不补不存在的第 10/11 个 GET。
2. 权限使用 role>=1。
3. 接受第 3.2 节快照近似：returned 用保留 claim 字段的 status=2；完成时间用 updated_at；时长改为 updated_at-created_at。
4. 接受第 3.1 节日期、默认值、最大 366 天、百分比和环比规则。
5. 地点按直接 location_id，不向父级归并；NULL 单列。
6. stagnant 默认 7 天、high-view 默认 50 浏览且 7 天、page_size 最大 100，字段使用白名单。
7. distribution 的 tag 桶允许重复计数；漏斗接受非严格的当前快照近似。
8. CSV 成功响应作为 JSON 信封例外，失败仍返回 HTTP 200 JSON。
9. 请提供实际 MySQL 大版本（5.7 或 8.x）以及部署服务器/数据库是否均使用东八区；若未知，批次 1 可先实现与版本无关部分，但 return-duration 暂停。

## 9. 实施状态

- [x] 阶段 1：代码与文档调研
- [x] 阶段 1：接口、查询、文件与测试计划
- [x] 用户确认第 8 节（方案 A：MySQL 8.x，应用/数据库东八区）及同批次归还口径
- [x] 批次 1：overview / trend / locations（代码、入口和自动测试完成；真实库验收由用户执行）
- [x] 批次 2：return-duration / stagnant / high-view / time-heatmap（代码和自动测试完成；真实库验收由用户执行）
- [x] 批次 3：distribution / funnel / CSV（代码及自动测试完成；真实库验收待用户执行）
- [x] 用户完成必要公共入口注册（dao、router及initialization；handler/service使用统计专属包入口，不需要额外公共变量）
- [ ] 用户完成共享文档及Swagger生成

### 2026-10-03 批次 1 验证记录

- overview/trend/locations 已实现；批次 1 仍待公共入口集成、全仓编译和测试库验收，不标记为完整验收通过。
- overview 默认当月截至今天，对比上一个完整自然月；显式范围对比紧邻等长周期。显式日期跨度补齐最大 366 天校验。
- return_rate 改用独立数值 DTO，40% 输出数值 40 而非缩放整数 4000；JSON 数字不保证尾随零，客户端展示两位小数。当前各百分比使用 math.Round 显式四舍五入至两位小数，各桶独立舍入，总和可能为 99.99/100.01，不进行尾差分摊。
- pending_over_24h 限定选定创建队列，阈值为查询当前时刻减 24 小时，不使用统计周期结束时间作为当前时刻。
- 测试通过：`go test service/advanced/admin_stats_service.go service/advanced/admin_stats_service_test.go -count=1 -v`，覆盖日期互斥/非法日期/最大跨度、默认月份/上月/闰月/跨时区、等长基期、趋势补零、比例 JSON、地点 limit/NULL 桶/空数据。
- `go test ./dao ./model/advanced` 通过（包编译，无 DAO 集成测试）；`git diff --check` 通过。统计 Go 文件已格式化。
- `go test ./...` 与 `go build ./...` 均失败于既有 `service/advanced/comment_service.go:46-48` 语法错误。该文件保持不变，需要用户手工修复；handler/router 编译及 HTTP 联调尚未验证。
- 本次纠正文档中未经批准的“发布至认领时长”改名，批次 2 恢复第 8 节已批准的 `updated_at-created_at` 近似口径；overview/trend 同创建队列口径保持不变。
- 路由文件直接使用专属 handler/service/DAO 类型及变量，当前不依赖公共 enter 文件注册。最小集成方案：用户在 `initialization/router.go` import 增加 `routerAdvanced "github.com/unicornfairy864/LNF-SERVER/router/advanced"`，在 Advanced 路由注册块增加 `(&routerAdvanced.AdminStatsRouter{}).CreateRouter(api)`。需要用户手工修改；无须同时修改各层 enter 文件。
- 剩余风险：locations limit=0 返回全部桶的最大行数保护尚未确定；DAO 聚合未连接测试库验证，EXPLAIN 尚未执行；Swagger 注释待补齐和只读核验共享生成物。

### 2026-10-03 公共入口修改后复核

- 用户在 dao/enter.go 注册 AdminStatsDao；AI 移除专属 dao/admin_stats_dao.go 中的重复变量声明，保留并复用公共入口变量。
- 用户在 initialization/router.go 调用 router.AdminStatsRouter.CreateRouter(api)，但 router/enter.go 尚无 AdminStatsRouter 声明。需要用户手工在该文件现有 var 块增加 `AdminStatsRouter advanced.AdminStatsRouter`；其现有 advanced import 已满足要求。
- 三个批次 1 handler 已补齐 Swagger 注释，未修改 docs 生成物。
- 最新统计单文件测试、dao/model 包检查与 service/advanced 包测试通过；全仓测试中 handler、router 包编译通过，评论模块不再出现语法错误。
- 最新 go test ./... 与 go build ./... 仅报告 initialization/router.go:27 的 undefined: router.AdminStatsRouter；git diff --check 通过。待用户补齐路由声明后重新验证。
- 上一节“无需修改各层 enter”仅适用于直接实例化路由方案；用户已选择公共 router 入口调用，实际应补齐上述 router/enter.go 声明。无需额外注册 handler/service 变量。

### 2026-10-03 路由声明补齐后验证

- 只读确认 router/enter.go 已声明 `AdminStatsRouter advanced.AdminStatsRouter`，公共入口接入完成。
- `go test ./...`、`go build ./...`、`git diff --check` 全部通过。上述结果是编译与现有测试结果，不代表已连接数据库或启动 HTTP 服务联调。
- 批次 1 暂不勾选完整验收：limit=0 最大桶数未获确认，真实 MySQL 聚合及 EXPLAIN 尚未人工执行。

### 批次 1 固定样例 SQL（用户在 MySQL 8 测试环境只读执行）

以下 CTE 不创建表、不插入数据，验证空边界、删除排除、撤回/自行关闭/确认关闭/自动关闭快照。该样例不替代 GORM 的真实查询联调。

```sql
WITH fixture AS (
  SELECT 1 AS id, CAST('2026-10-01 00:00:00' AS DATETIME) AS created_at,
         0 AS status, NULL AS claim_user_id, CAST(NULL AS DATETIME) AS claim_time, 0 AS is_deleted
  UNION ALL SELECT 2, '2026-10-01 08:00:00', 0, NULL, NULL, 0 -- 撤回认领
  UNION ALL SELECT 3, '2026-10-01 09:00:00', 2, NULL, NULL, 0 -- 自行关闭
  UNION ALL SELECT 4, '2026-10-01 10:00:00', 2, 101, '2026-10-01 10:30:00', 0 -- 确认关闭
  UNION ALL SELECT 5, '2026-10-01 11:00:00', 2, 102, '2026-10-01 11:30:00', 0 -- 自动关闭
  UNION ALL SELECT 6, '2026-10-01 12:00:00', 1, 103, '2026-10-01 12:30:00', 0
  UNION ALL SELECT 7, '2026-10-01 13:00:00', 2, 104, '2026-10-01 13:30:00', 1 -- 已删除
  UNION ALL SELECT 8, '2026-10-02 00:00:00', 0, NULL, NULL, 0 -- 上界排除
  UNION ALL SELECT 9, '2026-09-30 23:59:59', 0, NULL, NULL, 0 -- 下界之前排除
)
SELECT COUNT(*) AS published,
       COALESCE(SUM(status=0), 0) AS pending,
       COALESCE(SUM(status=1 AND claim_user_id IS NOT NULL), 0) AS claimed,
       COALESCE(SUM(status=2 AND claim_user_id IS NOT NULL AND claim_time IS NOT NULL), 0) AS returned,
       COALESCE(SUM(status=2 AND claim_user_id IS NULL), 0) AS closed
FROM fixture
WHERE is_deleted=0
  AND created_at >= '2026-10-01 00:00:00'
  AND created_at < '2026-10-02 00:00:00';
```

- 预期 published=6、pending=2、claimed=1、returned=2、closed=1，return_rate=33.33。
- 将日期范围换成 `[2026-10-03,2026-10-04)`，预期所有聚合为0；service 趋势应返回对应自然日的零值点。
- 地点固定样本建议：直接地点10三件、直接地点20两件、NULL一件，预期 total=6、percent=50/33.33/16.67；limit=1 只保留地点10，unknown仍为一件，分母仍为6。父地点不合并。地点硬删除后保留该ID桶，名称为空字符串；NULL桶单列unknown（现有SQL行为，未扩展隐私字段）。
- 用户在测试库对 dao/admin_stats_dao.go 中三个固定聚合 SELECT 添加 EXPLAIN，日期替换为真实业务范围，记录 key、rows、filtered 与 Extra。现有 idx_items_audit 在只过滤 is_deleted 和 created_at 时，中间 status 未限定，不能声称命中完整时间索引范围；地点 GROUP BY 与日期趋势预计可能使用临时表/排序，需以实际计划为准。不新增索引，不直接连接数据库。
- 待确认方案：limit=0 最多1000个非NULL地点桶（unknown单列），超过上限返回code=1，不静默截断；正limit沿用1-100。用户确认前不实现该上限，也不记录为已批准。

### 2026-10-03 地点上限决定与批次切换

- 原问题：limit=0 全量桶数保护未定义。用户明确批准最多1000个非NULL地点桶，unknown单列；超限code=1，不静默截断。影响locations及未来对应CSV。
- 用户批准完成批次1后继续批次2。新增白名单文件 `dao/admin_stats_dao_test.go` 属于原批准的管理员统计DAO测试文件范围。
- DAO全量模式最多读取1002个桶用于检测超限；正limit模式在数据库限制非NULL桶数量，额外一次聚合返回unknown。所有模式总量分母不受limit影响。
- 固定SQL样例和EXPLAIN步骤见上一节，由用户在测试库执行；AI未连接数据库。

### 2026-10-03 批次1完成简报与批次2进展

- 批次1新增1000桶边界测试通过，go test ./dao ./service/advanced、go test ./...、go build ./...、git diff --check全部通过。
- 批次2已开始time-heatmap：WEEKDAY(created_at)周一为0、HOUR(created_at)为0-23，按批准的东八区DATETIME语义与半开范围聚合；model固定[7][24]int64矩阵，service补零，handler沿用统一信封，路由沿用现有鉴权。
- 热力图固定SQL检查建议：在测试环境 `SELECT WEEKDAY('2026-10-05 00:00:00'), HOUR('2026-10-05 00:00:00'), WEEKDAY('2026-10-11 23:59:59'), HOUR('2026-10-11 23:59:59');` 预期0、0、6、23。真实查询需验证已删除记录和范围上界不参与聚合。
- 待确认清单page最大值（page_size已批准最大100）；建议page最大10000，page超限code=1。
- 待确认近似时长出现updated_at<created_at的异常数据行为；建议排除负时长样本，空集count=0、average_seconds=0、median_seconds=0。确认前不实现受影响接口。

### 2026-10-03 批次2边界确认与实施

- 用户批准page最大10000（超限code=1）、排除updated_at<created_at异常样本、空样本count/average_seconds/median_seconds全部返回0。影响stagnant/high-view/return-duration。
- 清单响应data为 `{total,page,page_size,items}`；items各项按第4节字段白名单，location_id允许null，缺失地点名称为空字符串，created_at为JSON时间；stagnant_days使用TIMESTAMPDIFF(DAY,created_at,当前时刻)，代表完整24小时天数。
- 时长响应data为 `{start_date,end_date,group_by,approximate:true,overall,groups}`；overall及groups项为 `{group_id,count,average_seconds,median_seconds}`。none模式groups为空数组；location模式NULL地点单独group_id=null；type模式group_id为类型整数。groups按group_id ASC，NULL在前。无样本overall为零指标，group_id=null、groups为空数组。平均和中位数单位秒，数据库ROUND保留两位小数，返回JSON数字。
- MySQL8窗口函数精确计算中位数：奇数n取floor((n+1)/2)，偶数n取floor((n+1)/2)与floor((n+2)/2)的均值。overall和分组使用相同过滤快照，分组查询至多两次SQL；无全量样本读取，无N+1。
- 清单count和分页两次查询，固定ORDER BY和最终id稳定键，最多100行、偏移最大999900；高浏览查询可能额外排序，需人工EXPLAIN真实数据成本。
- 时长查询按updated_at范围及成功归还快照过滤，现有索引不覆盖updated_at，预计需扫描候选关闭记录并使用窗口排序临时结果；不擅自新增索引，用户测试库EXPLAIN后评估。
- 奇偶中位数固定只读SQL样例（由用户执行，AI未连接数据库）：

```sql
WITH samples AS (
 SELECT 1 AS group_id, 10 AS duration UNION ALL SELECT 1,20 UNION ALL SELECT 1,30
 UNION ALL SELECT 2,10 UNION ALL SELECT 2,20 UNION ALL SELECT 2,30 UNION ALL SELECT 2,40
), ranked AS (
 SELECT group_id,duration,ROW_NUMBER() OVER(PARTITION BY group_id ORDER BY duration) AS rn,
 COUNT(*) OVER(PARTITION BY group_id) AS n FROM samples
)
SELECT group_id,COUNT(*) AS count,ROUND(AVG(duration),2) AS average_seconds,
 ROUND(AVG(CASE WHEN rn IN(FLOOR((n+1)/2),FLOOR((n+2)/2)) THEN duration END),2) AS median_seconds
FROM ranked GROUP BY group_id ORDER BY group_id;
```

- 预期组1 count=3、average=20、median=20；组2 count=4、average=25、median=25。真实样本另验证撤回与自行关闭不纳入、确认与自动关闭纳入、负时长排除、零时长纳入、NULL地点和删除记录。

### 2026-10-03 批次2简报

- 完成4个GET接口及Swagger注释，权限沿用已接入的JWT与ServiceAdminAuthMiddleware，无新增公共入口改动。
- 修改文件：model/advanced/admin_stats.go、dao/admin_stats_dao.go及其测试、service/advanced/admin_stats_service.go及其测试、handler/advanced/admin_stats_handler.go、router/advanced/admin_stats_router.go、本实施文档。未修改公共文件或Swagger生成物。
- 模块测试 `go test ./dao ./service/advanced -count=1`、`go test ./...`、`go build ./...`、`git diff --check` 全部通过。测试覆盖分页与浏览/天数阈值边界、非法分组、热力补零，以及固定SQL白名单/过滤条件检查。MySQL奇偶中位数运行结果尚未验证，仅提供可复现SQL验收样例，不虚报数据库集成测试。
- 未决风险：真实库EXPLAIN和端到端鉴权/数据库响应需用户验收；时长近似值可能被updated_at其他更新污染；overall/分组与清单count/list分开查询，并发变更时不保证同一事务快照；深分页可有较高扫描成本。共享API文档及Swagger产物由用户手工同步。
- 下一批distribution/funnel/CSV在用户批准后继续；CSV最大行数与超时策略仍需在实施前确认，不能仅凭366天日期上限推定安全。

### 2026-10-03 批次3批准

- 用户明确批准完整方案：distribution/funnel默认30天、最大366天，沿用快照近似；相邻比率可超过100%，分母为0返回0。标签桶最多1000个，untagged单列，超限code=1。
- CSV覆盖overview、trend、locations、distribution、return-duration、time-heatmap；最多10000条数据行，整体超时30秒，每批读取/flush最多200行。写出前失败保持HTTP200 JSON：参数/超限code=1，超时code=5，数据库错误code=6；流开始后失败终止并记录日志，不追加JSON。
- 文件名：管理员统计-{报表}-{开始日期}-{结束日期}.csv；Content-Type为text/csv; charset=utf-8，加UTF-8 BOM及RFC5987文件名。
- 批准新建测试文件：handler/advanced/admin_stats_handler_test.go。其余仅修改既有统计专属文件，不修改公共入口、共享文档、依赖或Swagger生成物。
- distribution响应：{start_date,end_date,dimension,total,buckets:[{bucket_id,bucket_name,published,returned,return_rate,percent}]}，NULL标签桶名称untagged；删除的标签关联仍按tag_id保留桶，名称为空，不丢失物品计数。type名称lost/found。
- funnel响应：{start_date,end_date,stages:[{stage,users}],adjacent_ratios:[百分比]}，stage依次registered/published/claimed/returned，不是嵌套cohort；注册按users.created_at，发布按items.created_at，认领按claim_time，归还按updated_at。注册排除软删除用户，其余层按未删除items当前快照。

### 批次3验收与交付

- 新增distribution/funnel路由，复用JWT和ServiceAdminAuthMiddleware。新增handler统计测试文件，其余修改仅限统计白名单文件。
- CSV复用原service与DAO口径，查询传递30秒context；潜在较多的distribution/duration分组采用数据库Rows游标逐行读取，每次读取1行（不超过200），返回有界聚合DTO后写出，每200行flush。不会读入全表items，不落盘。注意不是数据库与HTTP同时流式输出：为保证查询失败时仍可返回JSON，会先缓冲有界聚合DTO和CSV行。duration导出最多取10000个分组，加overall后若超过10000数据行即拒绝，不静默截断。
- 自动验证：模块测试、go test ./...、go build ./...、git diff --check通过。新增测试覆盖CSV BOM/表头/中文文件名/转义/flush/写入失败/写出前取消与超限、HTTP200错误信封、6种报表列数、整数百分比舍入、标签重叠计数及1000桶边界、非嵌套漏斗比率、SQL白名单及过滤条件。未连接真实MySQL，不声称数据库集成验收完成。
- 测试库人工SQL验证：复制dao/admin_stats_dao.go中statsDistributionSQL("tag"/"type")及statsFunnelSQL，将每对?替换为同一东八区自然日范围，例如'2026-10-01 00:00:00'和'2026-10-04 00:00:00'。分别执行原SELECT与EXPLAIN，检查key/rows/filtered/Extra；tag关联应使用uk_item_tag或idx_tag_item，聚合可能使用临时表。funnel的OR时间过滤可能扫描较大范围，需真实计划评估，不新增索引。
- 固定验收数据预期：范围内item A为returned并带tag 1、2，item B为pending且带tag 1，item C无标签且自行关闭；总发布3，tag 1 published=2/returned=1/percent=66.67/return_rate=50，tag 2 published=1/returned=1/percent=33.33/return_rate=100，untagged published=1/returned=0/percent=33.33。占比和为133.33是允许的；移除A认领字段后returned降低，软删除A后所有桶排除A。上界created_at恰等于end的记录不计入。
- 漏斗验收：注册/发布/认领/归还各按对应时间字段distinct去重；同一用户多物品只计1，撤回后清空认领字段不计认领，自行关闭不计归还，确认和自动关闭计归还；注册层排除软删users，各items层排除软删items。各层非嵌套，不要求人数递减。
- 需要用户手工修改共享文档：agent.md/api_agent.md与agent.md/api_guide.md的管理员统计接口表加入9个GET完整/api/v1路径，参数/响应以本文件第4节与批次2/3追加契约为准；6个聚合接口加入export=csv及CSV成功信封例外、BOM、文件名、行数、超时和流中失败行为。response_code无需新增。随后由用户手工执行swag init；AI不生成或修改docs目录。
- 剩余风险：真实MySQL8查询/EXPLAIN、端到端权限验证待用户执行；updated_at归还近似可能受更新污染；多SQL不保证同事务快照；底层ResponseWriter不支持SetWriteDeadline时，仅context取消和逐行检查提供超时保护，阻塞网络写入需服务器写超时配置保证；CSV中的地点/标签文本未作Excel公式中和，导出给Excel使用时需注意不可信名称。