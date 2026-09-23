# item / tags / location / item_image API 实施计划

> 本计划已经用户逐条验收，实施过程以本文为准。

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

1. 只支持 `:param`，不支持 `{param}`（既有 `location_router.go` 的 `{itemID}` 是 bug，须改为 `:itemID`）。
2. 同一位置的参数名必须完全一致（`/item/:itemID` 各处统一）。
3. 静态与参数路径可并存（`/cmd/vet` + `/cmd/:tool` 测试通过）→ `GET /item/list` 与 `GET /item/:itemID` 无冲突。
4. 现有 `location_router.go` 的 `admin := userGroup.Group("")` 会把 `/create` 注册到根路径，必须改为
   `Group("/location")`。

## 二、接口总表

### Item（`router/basic/item_router.go`）

| 分级    | 方法 | 路径                   | Handler                | 说明                                                                                                  |
|---------|------|------------------------|------------------------|-------------------------------------------------------------------------------------------------------|
| Public  | GET  | `/item/list`           | `ListItemHandler`      | 筛选+分页+关键词；query: `type,status,location_id,tag_id,keyword,page,page_size`；公开默认 `status=0` |
| Public  | GET  | `/item/:itemID`        | `GetItemHandler`       | 详情：地点链+标签+图片；`view_count+1`                                                                |
| Private | POST | `/item/create`         | `CreateItemHandler`    | 带 `tag_ids`；`user_id` 取 JWT；`status` 默认 0                                                       |
| Private | POST | `/item/update`         | `UpdateItemHandler`    | 增量更新，仅本人；`tag_ids` 可选（传了就整体替换关联）                                                |
| Private | POST | `/item/delete`         | `DeleteItemHandler`    | 软删 `is_deleted=1`（items 表有 is_deleted），仅本人                                                  |
| Private | GET  | `/item/mine`           | `ListMyItemHandler`    | 我的发布，分页默认10、上限50；不过滤状态（除非显式传）                                                |
| Private | POST | `/item/:itemID/images` | `SetItemImagesHandler` | 覆盖式设置图片；≤3 张；sort_order 1-3 互不重复；仅本人                                                |

### Location（`router/advanced/location_router.go`）

| 分级   | 方法   | 路径                      | Handler                   | 说明                                                                                                |
|--------|--------|---------------------------|---------------------------|-----------------------------------------------------------------------------------------------------|
| Public | GET    | `/location/list`          | `ListLocationHandler`     | query: `parent_id,level` 均可选；缺省返回全部（按 sort_order）                                      |
| Public | GET    | `/item/:itemID/locations` | `GetItemLocationsHandler` | 物品 location 的祖先链（根→叶）                                                                     |
| Admin  | POST   | `/location/create`        | `CreateLocationHandler`   | level 由 service 计算（parent=0 → 1，否则 parent.level+1）；parent 不存在报 80001；同父重名报 80002 |
| Admin  | POST   | `/location/update`        | `UpdateLocationHandler`   | ID + 可选字段增量更新；改 parent 时校验存在/非自身/非后代，并在事务中重算子树 level                 |
| Admin  | DELETE | `/location/:locationID`   | `DeleteLocationHandler`   | 硬删除；先查子元素(80005)再查 items 引用(80004)                                                     |

### Tag（新建 `router/advanced/tag_router.go`）

| 分级   | 方法   | 路径          | Handler            | 说明                                          |
|--------|--------|---------------|--------------------|-----------------------------------------------|
| Public | GET    | `/tag/list`   | `ListTagHandler`   | 按 sort_order                                 |
| Admin  | POST   | `/tag/create` | `CreateTagHandler` | 查重 uk_name（40002）；name 空或 >50 报 40004 |
| Admin  | POST   | `/tag/update` | `UpdateTagHandler` | 增量；改名查重（排除自身）                    |
| Admin  | DELETE | `/tag/:tagID` | `DeleteTagHandler` | 硬删除；被 item_tags 引用报 40005             |

## 三、新增响应码（response/response_code.go）

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

## 四、文件变更清单与分层职责

1. **response/response_code.go**：新增上表码+消息。
2. **model/mysql/basic/items.sql**：修正 3 处过时注释（首页信息流 `status=0`、"审核队列"改为状态筛选语义、地点筛选
   `status=0`）。
3. **model**
    - `model/basic/item.go`：补全 `ItemImage`（按 item_images.sql，TableName=`item_images`）；`CreateItemRequest` 加
      `TagIDs`；`UpdateItemRequest` 加 `ID`(required)+`TagIDs *[]int64`；新增 `ListItemQuery`（form 标签）、
      `SetItemImagesRequest`/`ItemImageInput`、`ItemListResponse`。
    - `model/advanced/tag.go`：新增 `CreateTagRequest`/`UpdateTagRequest`/`ItemTag`（TableName=`item_tags`）。
    - `model/advanced/location.go`：放宽 `CreateLocationRequest` binding；新增 `UpdateLocationRequest`、
      `ListLocationRequest`（form 标签）。
4. **dao**（复用 `ConditionIDNotDeleted`；item 软删，tag/location 硬删）
    - 新建 `item_dao.go`：分页（tag_id 走 item_tags 子查询、keyword 走 LIKE）、按 ID 查、`CreateItemWithTags`（事务）、
      `UpdateItemWithTags`（事务）、软删、`IncrViewCount`、`CountItemsByLocationID`。
    - 新建 `item_image_dao.go`：按 item 查（sort_order 升序）、`ReplaceImages`（事务：先删后插）。
    - 新建 `tag_dao.go`：列表/按 ID/按名/按 ID 集合、创建、VK 更新、硬删。
    - 新建 `item_tag_dao.go`：按 item 联查 tags、按 tag 统计引用、`ReplaceInTx`。
    - 补全 `location_dao.go`：按 ID/parent/level/全部查询、`GetLocationChain`（沿 parent 上溯，限深防环）、创建、VK 更新、硬删、
      `CountChildren`、`MoveLocation`（事务改 parent 并 BFS 重算子树 level）。
    - `dao/enter.go` 注册 `ItemDao/ItemImageDao/ItemTagDao/TagDao/LocationDao`。
5. **service**
    - 新建 `service/basic/item_service.go`：`ListPublicService`（缺省 status=0）、`MyListService`、`GetDetailService`
      （组装+浏览量）、`CreateService`（type 0/1、location 存在、tags 全存在、status=0）、`UpdateService`（本人、location 指针语义
      0=清除）、`DeleteService`（本人软删）、`SetImagesService`（归属+≤3+sort 1-3 不重复）。
    - 新建 `service/advanced/tag_service.go`、`location_service.go`（含 level 计算、链查询、删除双重引用检查、改父防环）。
    - `service/enter.go` 注册 `ItemService/TagService/LocationService`。
6. **handler**：补 `basic/item_handler.go`、`advanced/location_handler.go`，新建 `advanced/tag_handler.go`； **每个 handler
   上方完整 swagger 注释**（@Summary/@Tags/@Accept/@Produce/@Param/@Success/@Router，路径含 `/api/v1` 前缀）；
   `handler/enter.go` 注册 `TagHandler`。
7. **router**
    - `basic/item_router.go`：按总表注册（public/private；admin 组保留空壳待 audit 模块）。
    - `advanced/location_router.go`：修正分组与 `:itemID`，补全 public/admin。
    - 新建 `advanced/tag_router.go`。
    - `router/enter.go`、`initialization/router.go` 注册 `ItemRouter/TagRouter`。
8. **验证**：`go build ./...` 编译通过 + lint 检查；不运行测试；`swag init` 待用户另行批准。

## 五、业务校验与错误映射速查

- 参数绑定失败 / page 归一化失败 → `CodeParamError(1)`
- type 不在 {0,1} → `20006`；location_id 指向不存在地点 → `20010`
- item 不存在或已删 → `20001`；非发布者操作 → `20005`
- tag 不存在 → `40001`；tag 重名 → `40002`；tag 被引用 → `40005`
- location 不存在 → `80001`；同父重名 → `80002`；被 items 引用 → `80004`；有子元素 → `80005`
- 图片 >3 张 → `20013`；sort_order 越界/重复/URL 越长 → `20014`
- 其余 DB 异常 → `CodeDatabaseError(6)`
