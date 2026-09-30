# item / tags / location / item_image API 实施文档

> 本模块已实施完毕并经用户逐条验收，**以本文与代码为准**；`go build ./...` 编译通过，swagger 文档（docs/）已由 `swag init` 生成。

## 一、已确认的关键决定（Q1-Q7）

| #  | 决定                                                                                                                                                     |
|----|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| Q1 | `items.status` 三态：`0已发布（默认，创建即免审核） 1已认领 2已关闭`；同步修正 `items.sql` 过时索引注释                                                  |
| Q2 | Item↔Tag 关联：创建/更新 item 时传 `tag_ids` 数组（方案 A）                                                                                              |
| Q3 | 图片：独立接口 `POST /item/:itemID/images` 覆盖式设置（方案 A），与 item 提交分离                                                                        |
| Q4 | 不设 admin 的 item 状态接口；`ChangeItemStatusRequest` 已由用户删除                                                                                      |
| Q5 | `ServiceAdminAuthMiddleware` 读取 `jwt:role` 的 bug 已由用户修复（role≥1 放行）                                                                          |
| Q6 | `CreateLocationRequest` 放宽 binding：允许 `parent_id=0`（根节点）、`sort_order=0`；`level` 由 service 计算                                              |
| Q7 | 删除策略：表无 `is_deleted` 则硬删除；location 删除检查 **items.location_id 引用 + 子元素**；tag 删除检查 **item_tags 引用**；`agent.md/item.txt` 已删除 |

### 已核实的 gin v1.12 路由约束

1. 只支持 `:param`，不支持 `{param}`（既有 `location_router.go` 的 `{itemID}` 是 bug，已改为 `:itemID`）。
2. 同一位置的参数名必须完全一致（`/item/:itemID` 各处统一）。
3. 静态与参数路径可并存（`/cmd/vet` + `/cmd/:tool` 测试通过）→ `GET /item/list` 与 `GET /item/:itemID` 无冲突。
4. 现有 `location_router.go` 的 `admin := userGroup.Group("")` 会把 `/create` 注册到根路径，已改为
   `Group("/location")`。

## 二、接口总表

### Item（`router/basic/item_router.go`）

| 分级    | 方法 | 路径                      | Handler                | 说明                                                                                                  |
|---------|------|---------------------------|------------------------|-------------------------------------------------------------------------------------------------------|
| Public  | GET  | `/item/list`              | `ListItemHandler`      | 筛选+分页+关键词；query: `type,status,location_id,tag_id,keyword,page,page_size`；**未指定 status 时默认返回 0/1（已发布+已认领）** |
| Public  | GET  | `/item/:itemID`           | `GetItemHandler`       | 详情：地点链+标签+图片；`view_count+1`                                                                |
| Private | POST | `/item/create`            | `CreateItemHandler`    | 带 `tag_ids`；`user_id` 取 JWT；`status` 默认 0                                                       |
| Private | POST | `/item/update`            | `UpdateItemHandler`    | 增量更新，仅本人；`tag_ids` 可选（传了就整体替换关联）                                                |
| Private | POST | `/item/delete`            | `DeleteItemHandler`    | 软删 `is_deleted=1`（items 表有 is_deleted），仅本人                                                  |
| Private | GET  | `/item/mine`              | `ListMyItemHandler`    | 我的发布，分页默认10、上限50；未指定 status 时默认返回 0/1                                            |
| Private | POST | `/item/:itemID/images`    | `SetItemImagesHandler` | 覆盖式设置图片；≤3 张；sort_order 1-3 互不重复；仅本人                                                |

### Item 认领 / 关闭（同在 `item_router.go`，Private）

| 方法 | 路径                        | Handler               | 说明                                                                                             |
|------|-----------------------------|-----------------------|--------------------------------------------------------------------------------------------------|
| POST | `/item/:itemID/claim`        | `ClaimItemHandler`    | 认领：仅 status=0 可认领；不能认领自己的物品（30004）；`server.claim_qq_required` 开启时未绑 QQ 返回 30006；并发条件更新（仅 status=0 生效） |
| POST | `/item/:itemID/claim/cancel` | `WithdrawClaimHandler`| 撤回认领：**认领者或发帖者双方均可**；仅 status=1 可撤，撤回后恢复 status=0；已关闭 30005 / 无认领 30002 / 无权 30003 |
| POST | `/item/:itemID/confirm`      | `ConfirmClaimHandler` | 发帖者确认由他人找回：status=1→2，同事务给受益人加 `server.claim_credit` 积分；受益人：拾物帖(type=1)归发帖者，失物帖(type=0)归认领者 |
| POST | `/item/:itemID/close`        | `CloseSelfHandler`    | 发帖者关闭自己的帖子（自己已找回，不发积分）：status=0/1→2，清空认领字段                          |

### Location（`router/advanced/location_router.go`）

| 分级   | 方法   | 路径                      | Handler                   | 说明                                                                                                |
|--------|--------|---------------------------|---------------------------|-----------------------------------------------------------------------------------------------------|
| Public | GET    | `/location/list`          | `ListLocationHandler`     | query: `parent_id,level` 均可选；缺省返回全部（按 sort_order）                                      |
| Public | GET    | `/item/:itemID/locations` | `GetItemLocationsHandler` | 物品 location 的祖先链（根→叶）                                                                     |
| Admin  | POST   | `/location/create`        | `CreateLocationHandler`   | level 由 service 计算（parent=0 → 1，否则 parent.level+1）；parent 不存在报 80001；同父重名报 80002 |
| Admin  | POST   | `/location/update`        | `UpdateLocationHandler`   | ID + 可选字段增量更新；改 parent 时校验存在/非自身/非后代，并在事务中重算子树 level                 |
| Admin  | DELETE | `/location/:locationID`   | `DeleteLocationHandler`   | 硬删除；先查子元素(80005)再查 items 引用(80004)                                                     |

### Tag（`router/advanced/tag_router.go`）

| 分级   | 方法   | 路径          | Handler            | 说明                                          |
|--------|--------|---------------|--------------------|-----------------------------------------------|
| Public | GET    | `/tag/list`   | `ListTagHandler`   | 按 sort_order                                 |
| Admin  | POST   | `/tag/create` | `CreateTagHandler` | 查重 uk_name（40002）；name 空或 >50 报 40004 |
| Admin  | POST   | `/tag/update` | `UpdateTagHandler` | 增量；改名查重（排除自身）                    |
| Admin  | DELETE | `/tag/:tagID` | `DeleteTagHandler` | 硬删除；被 item_tags 引用报 40005             |

## 三、响应码（response/response_code.go）

### 本模块新增/使用

| 码    | 常量                    | 消息                         |
|-------|-------------------------|------------------------------|
| 20001 | CodeItemNotFound        | 物品不存在                   |
| 20005 | CodeItemNoPermission    | 无权操作该物品               |
| 20006 | CodeItemTypeInvalid     | 物品类型非法                 |
| 20010 | CodeItemLocationInvalid | 地点无效                     |
| 20013 | CodeItemImageTooMany    | 物品图片最多3张              |
| 20014 | CodeItemImageInvalid    | 物品图片参数错误             |
| 40001 | CodeTagNotFound         | 标签不存在                   |
| 40002 | CodeTagDuplicate        | 标签名称已存在               |
| 40004 | CodeTagNameInvalid      | 标签名称无效                 |
| 40005 | CodeTagInUse            | 标签正在被使用，无法删除     |
| 80001 | CodeLocationNotFound    | 地点不存在                   |
| 80002 | CodeLocationDuplicate   | 地点名称重复                 |
| 80004 | CodeLocationInUse       | 地点正在被物品引用，无法删除 |
| 80005 | CodeLocationHasChildren | 存在子地点，无法删除         |

### 认领段 3xxxx（与 2xxxx 部分语义重复，编码不同、常量名加 Claim 前缀）

| 码    | 常量                        | 消息                     | 使用情况                     |
|-------|-----------------------------|--------------------------|------------------------------|
| 30001 | CodeClaimItemAlreadyClaimed | 物品已被认领             | claim 并发重查               |
| 30002 | CodeClaimNotFound           | 认领记录不存在           | withdraw / confirm          |
| 30003 | CodeClaimNoPermission       | 无权审核该认领申请       | withdraw 非双方             |
| 30004 | CodeClaimSelfItem           | 不能认领自己发布的物品   | claim                       |
| 30005 | CodeClaimItemClosed         | 物品已关闭               | claim/withdraw/confirm/close |
| 30006 | CodeClaimQQRequired         | 认领前请先绑定QQ         | claim（受配置开关控制）     |
| 30007 | CodeClaimDescriptionTooLong | 认领描述过长             | 已定义，预留未使用          |
| 30008 | CodeClaimDuplicate          | 重复提交认领申请         | 已定义，预留未使用          |

## 四、文件清单与分层职责（实际实现）

1. **response/response_code.go**：上表全部码+消息（含 2xxxx/3xxxx 认领段）。
2. **model/mysql/basic/items.sql**：已修正过时注释（首页信息流、状态筛选、地点筛选语义）。
3. **model**
    - `model/basic/item.go`：`ItemImage`（TableName=`item_images`）；`CreateItemRequest.TagIDs`；
      `UpdateItemRequest{ID required, TagIDs *[]int64}`；`ListItemQuery`（form 标签）、
      `SetItemImagesRequest`/`ItemImageInput`、`ItemListResponse`。
    - `model/advanced/tag.go`：`Tag`、`ItemTag`（TableName=`item_tags`，uk_item_tag + idx_tag_item）、
      `CreateTagRequest`/`UpdateTagRequest`。
    - `model/advanced/location.go`：放宽 `CreateLocationRequest` binding（parent_id/sort_order 可省略）；
      `UpdateLocationRequest`、`ListLocationRequest`（form 标签）。
4. **dao**（复用 `ConditionIDNotDeleted`；item 软删，tag/location 硬删）
    - `item_dao.go`：`buildItemQuery`（每次调用返回全新 DB，避免 Count 污染链；tag_id 走 item_tags
      子查询、keyword 走 LIKE）、`GetItemPage`（分页）、`GetItemByID`、`CreateItemWithTags`（事务）、
      `UpdateItemWithTags`（事务：增量 updates + tagIDs 非 nil 时先删后插、按 seen map 去重）、
      `SoftDeleteItem`、`IncrViewCount`、`CountItemsByLocationID`；
      认领部分：`ClaimItem`（条件更新 status 0→1）、`WithdrawClaim`（1→0 清空认领字段）、
      `CloseItem`（0/1→2）、`CloseItemWithCredit`（1→2 同事务发积分，affected=0 表示并发已变化）、
      `ListExpiredClaimedItems`（认领超时扫描）。
    - `item_image_dao.go`：`GetImagesByItemID`（sort_order 升序）、`ReplaceImages`（事务先删后插）。
    - `tag_dao.go`：列表/按 ID/按名/按 ID 集合、创建、VK 更新、硬删。
    - `item_tag_dao.go`：`GetTagsByItemID`（联查 tags）、`CountByTagID`（删除前引用检查）。
    - `location_dao.go`：按 ID/parent+name/parent/level/全部查询、`GetLocationsByIDs`、创建、VK 更新、
      `CountChildren`、`GetLocationChain`（沿 parent 上溯，限深 50 防环，返回根→叶）、
      `MoveLocation`（事务改 parent 并 BFS 重算子树 level）、`DeleteLocationByID`。
    - `dao/enter.go` 注册 `ItemDao/ItemImageDao/ItemTagDao/TagDao/LocationDao`。
5. **service**
    - `service/basic/item_service.go`：
      列表/详情/CRUD：`ListPublicService`、`ListMyService`（均默认 status IN (0,1)；mine 的 page_size
      上限 50）、`GetDetailService`（组装+浏览量）、`CreateService`（type 0/1、location 存在、tags 全
      存在、status=0）、`UpdateService`（本人、location 指针语义 0=清除）、`DeleteService`（本人软删）、
      `SetImagesService`（归属+≤3+sort 1-3 不重复+URL ≤500）。
      认领/关闭：`ClaimService`（含 QQ 绑定开关校验、并发重查）、`WithdrawClaimService`（双方可撤）、
      `ConfirmClaimService`（`claimBeneficiary`：type=1 归发帖者，否则归认领者）、`CloseSelfService`、
      `AutoCloseExpiredClaimsService`（扫描认领超时物品按确认语义自动关闭并发分，单批上限 200；
      由 `initialization.StartClaimAutoCloseScheduler` 每 5 分钟调用，`main.go` 启动）。
      关闭通知（2026-09-30 接入，经 `notifyClaimClosed` 走 notification 模块 `Create`，type=3 认领结果、
      adminID=0、relatedID=物品ID，失败仅记日志不影响主流程）：超时自动关闭通知发帖者+认领者双方；
      手动确认关闭仅通知认领者（确认者=发帖者本人不自我通知）；文案按受益人身份区分积分发放说明；
      `notificationService` 为包内共享实例（user 模块 QQ 绑定也使用），直接依赖 service/advanced
      避免 basic↔service 循环引用。
      `CloseSelfService`（发帖者自行找回关闭，清认领不发分）：关闭前若存在进行中认领（status=1），
      经 `notifySelfCloseToClaimer` 通知认领者（type=3，adminID=0，relatedID=物品ID，失败仅记日志）；
      status=0 直接关闭无认领，不通知。
    - `service/advanced/tag_service.go`、`location_service.go`（含 level 计算、链查询、删除双重引用
      检查、改父防环 `isDescendant`）。
    - `service/enter.go` 注册 `ItemService/TagService/LocationService`。
6. **handler**：`basic/item_handler.go`（含认领/关闭 4 个 handler）、`advanced/location_handler.go`、
   `advanced/tag_handler.go`；**每个 handler 上方完整 swagger 注释**（@Summary/@Tags/@Accept/@Produce/
   @Param/@Success/@Router，路径含 `/api/v1` 前缀）；`handler/enter.go` 注册 `TagHandler`。
7. **router**
    - `basic/item_router.go`：按总表注册（public/private；admin 组保留空壳待 audit 模块）。
    - `advanced/location_router.go`：分组与 `:itemID` 已修正。
    - `advanced/tag_router.go`。
    - `router/enter.go`、`initialization/router.go` 注册 `ItemRouter/TagRouter`。
8. **配置**（`config/server_config.go`，mapstructure 标签）
    - `server.claim_qq_required`：认领是否要求已绑定 QQ（user.qq 非空检查开关）。
    - `server.claim_credit`：认领成功每次加的积分数量。
    - `server.claim_auto_close`：认领后自动关闭时长（time.Duration），<=0 表示不自动关闭。
9. **验证**：`go build ./...` 编译通过；swagger 文档已生成（`docs/docs.go|swagger.json|swagger.yaml`）。

## 五、业务校验与错误映射速查

- 参数绑定失败 / page 归一化失败 / page_size 超上限（mine>50）→ `CodeParamError(1)`
- type 不在 {0,1} → `20006`；location_id 指向不存在地点 → `20010`
- item 不存在或已删 → `20001`；非发布者操作 → `20005`
- tag 不存在 → `40001`；tag 重名 → `40002`；tag 被引用 → `40005`
- location 不存在 → `80001`；同父重名 → `80002`；被 items 引用 → `80004`；有子元素 → `80005`
- 图片 >3 张 → `20013`；sort_order 越界/重复/URL 越长 → `20014`
- 认领：非 status=0 → `30001/30005`；认领自己的 → `30004`；未绑 QQ（开关开启）→ `30006`
- 撤回/确认：无认领记录 → `30002`；非双方/非本人 → `30003`/`20005`；已关闭 → `30005`
- 其余 DB 异常 → `CodeDatabaseError(6)`

## 六、多条件最小匹配检索（已实施）

> 供 agent 复用的公开接口：`GET /api/v1/item/search`（Public 无需登录）。

### 接口语义

- 计分：`match_count = |item.tags ∩ tag_ids| + (item.location_id ∈ 有效location_ids ? 1 : 0)`，
  返回 `match_count >= min_match` 的物品。
- location 命中规则：**只对 level=3 的地点计分**；传入 location_ids 中非 level3（或不存在）的 ID
  自动忽略，item 命中记 1 分，否则 0 分，无其他数值。
- tag_ids 去重后不存在的 ID 同理剔除（两者均按"剔除计"口径参与校验）。

### 请求参数（query，数组支持逗号分隔与重复参数两种写法）

| 参数          | 必填 | 说明                                                                 |
|---------------|------|----------------------------------------------------------------------|
| `tag_ids`     | 否   | 标签ID列表，如 `tag_ids=1,2,3`                                        |
| `location_ids`| 否   | 地点ID列表（仅 level=3 生效），如 `location_ids=4,5`                  |
| `min_match`   | 是   | 最少满足条件数，≥1（binding: required,min=1）                          |
| `type`        | 否   | 0丢失 1拾到                                                           |
| `status`      | 是   | 多选，仅允许 0已发布/1已认领，传 2 或为空报参数错误                     |
| `page` / `page_size` | 否 | 默认 1 / 10，page_size 上限 100                                  |

### 校验与边界（用户已确认）

- `min_match` 强制必传且 ≥1。
- `tag_ids` 与 `location_ids` 可某一项为空（另一维度计 0 分），但两项都为空 → 参数错误
  （由 min_match ≤ 条件总数=0 拦截）。
- 条件总数 = 去重后真实存在的 tag 数 + level3 地点数；**min_match ≤ 条件总数**，违反报参数错误
  （用户确认取 ">="：相等 = 全部条件必须满足，允许）。
- 结果恒为空的合法请求（理论上不会出现，因 min_match > 总数已被拦截）返回空列表。
- 结果过滤：`is_deleted=0` + 指定 status（多选）+ 可选 type；排序 `created_at DESC, id DESC`；
  返回结构沿用 `ItemListResponse` 分页（total/page/page_size/items）。

### 实现落点

- `model/basic/item.go`：新增 `ItemMatchQuery`（`collection_format:"csv"`；status 用
  `required,min=1,max=2,dive,oneof=0 1` 拦截空/含2）。
- `dao/item_dao.go`：`buildMinMatchExpr`（标签命中走 item_tags 子查询 + 地点命中 CASE WHEN，
  某维度为空时计分表达式回退为常量 0）、`GetItemsByMinMatch`（Count/Find 各自重建条件链）。
- `service/basic/item_service.go`：`SearchService`（tag 去重过滤存在性、location 过滤 level=3、
  条件总数校验、复用 `buildPagedList`）。
- `handler/basic/item_handler.go`：`SearchItemHandler`（含 swagger 注释）。
- `router/basic/item_router.go`：public 组注册 `GET /item/search`（静态路径与 `/:itemID` 并存无冲突）。
- 验证：`go build ./...`、`go vet ./...` 通过；swagger 文档需重新 `swag init`（待用户批准）。
