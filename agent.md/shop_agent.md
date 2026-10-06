# shop / goods / orders（积分兑换商城）实施文档

> 状态：**Step 1~5 ✅（2026-09-29，swag init 已由用户执行）｜notification 兑换触发点已接入（2026-09-30：type=5/6，见 §六）**。
> 需求已由用户逐条确认（见下），实施过程以本文为准。

## 〇、用户确认记录（2026-09-29）

| #  | 结论                                                                                                                                     |
|----|------------------------------------------------------------------------------------------------------------------------------------------|
| Q1 | shop = **积分兑换商城**，消耗 `users.credit`（拾金不昧奖励积分）                                                                          |
| Q2 | 范围 **b**：goods + orders + 完整兑换流程（事务扣库存+扣积分）                                                                            |
| Q3 | 图片**单张** `image_url` 字段，不建图片表                                                                                                |
| Q4 | 代码目录 `advanced/`；API 前缀 `/shop/goods`                                                                                             |
| Q5 | 响应码：积分不足与 goods 相关码**合并放同一新段**（见 §三）                                                                              |
| Q6 | `credit_logs.type` 新增 **4积分兑换**，同步修改 credit_logs.sql 注释                                                                      |
| Q7 | 兑换成功后通过 chensong 模块 client 发送**群消息到 activated_group**：at QQ + 昵称 + 商品名 + 引导联系管理员；**不建发货状态字段**（简化） |
| 附加1 | 兑换记录需写入 notification（**type=5 积分变动为第一项**，占位见 §六）                                                               |
| 附加2 | **QQ 绑定门槛**：只有绑定过 QQ 的用户才能使用商品兑换功能（复用用户模块 5~11 位判定语义，见 §五）                                      |
| 附加3 | goods 表**不设 status 字段**（用户明确"表中status删除"），仅靠 `is_deleted` 控制下架                                                 |
| 附加4 | model 结构体取**单数 `Good`**（用户决定：gorm 默认命名策略即映射 goods 表），文件名 `good.go`；`TableName()` 仍显式兑底                   |

## 一、表结构（交付 SQL 文件，由用户手动执行；项目无 AutoMigrate）

### goods（`model/mysql/advanced/goods.sql`）

```sql
CREATE TABLE `goods` (
    `id`          BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name`        VARCHAR(100) NOT NULL                COMMENT '商品名称',
    `description` TEXT         DEFAULT NULL            COMMENT '商品描述',
    `image_url`   VARCHAR(500) DEFAULT NULL            COMMENT '商品图片URL（单张）',
    `price`       INT          NOT NULL                COMMENT '兑换所需积分',
    `stock`       INT          NOT NULL DEFAULT 0      COMMENT '库存（0不可兑换）',
    `sort_order`  INT          NOT NULL DEFAULT 0      COMMENT '排序序号',
    `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_deleted`  TINYINT      NOT NULL DEFAULT 0      COMMENT '逻辑删除(下架): 0否 1是',
    PRIMARY KEY (`id`),
    KEY `idx_is_deleted_sort` (`is_deleted`, `sort_order`),
    KEY `idx_name` (`name`, `is_deleted`, `sort_order`),
    KEY `idx_price` (`is_deleted`, `price`, `sort_order`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '积分商城商品表';
```

- 无 `status`：公开列表即"未删除商品"；下架 = `is_deleted=1`（删除接口语义）。
- `stock=0` 表示暂时不可兑换（如需补货场景），与删除区分。
- 索引说明（用户 2026-09-29 确认，修正为复合索引模式；同日二次复核修正两处过誉声明，方案 A：
  保持三条索引不动，商品量级小，残留 filesort 可忽略；MySQL 8.0 降序索引严格免排序方案已评估，暂不采用）：
    - `idx_is_deleted_sort(is_deleted, sort_order)`：默认列表（无价格筛选）的过滤列。
      实际 ORDER BY 为 `sort_order ASC, created_at DESC, id DESC`（混合方向且含索引外列），
      仍需一次 filesort——几十~几百行量级开销可忽略。
    - `idx_price(is_deleted, price, sort_order)`：积分区间列表——等值 is_deleted=0 在前、
      price 范围居中、sort_order 收尾，单索引完成过滤+缩小排序集（范围后仍有一次小额排序，
      行集已被范围大幅缩小）。
    - `idx_name(name, is_deleted, sort_order)`：查重 `WHERE name=? AND is_deleted=0` 两列等值
      全走索引。keyword 搜索实际为 `LIKE '%kw%'` 双侧通配，不走本索引定位；
      尾列 sort_order 为兼容排序的冗余列（无害）。
      不用 UNIQUE——软删行会占用唯一键阻碍下架商品名复用，唯一性由 service 对未删除集合查重。
- **已建表库需手动补索引**（goods.sql 已同步，但表已创建；若先前已建旧索引先 DROP）：

```sql
ALTER TABLE goods DROP INDEX idx_name, ADD KEY idx_name (name, is_deleted, sort_order);
ALTER TABLE goods ADD KEY idx_price (is_deleted, price, sort_order);
```

### orders（`model/mysql/advanced/orders.sql`）

```sql
CREATE TABLE `orders` (
    `id`         BIGINT   NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `order_no`   VARCHAR(32) NOT NULL              COMMENT '业务单号（时间戳+随机，幂等展示用）',
    `user_id`    BIGINT   NOT NULL                 COMMENT '兑换用户ID',
    `goods_id`   BIGINT   NOT NULL                 COMMENT '商品ID',
    `goods_name` VARCHAR(100) NOT NULL             COMMENT '商品名称快照（防商品后续改名/删除）',
    `price`      INT      NOT NULL                 COMMENT '成交积分快照',
    `qq`         VARCHAR(50) NOT NULL              COMMENT '用户QQ快照（发群消息at用）',
    `nickname`   VARCHAR(50) NOT NULL              COMMENT '用户昵称快照（发群消息用）',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '兑换时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_no` (`order_no`),
    KEY `idx_user_created` (`user_id`, `created_at`),
    KEY `idx_goods_created` (`goods_id`, `created_at`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '积分兑换订单表';
```

- 简化设计：无状态字段、无软删；快照字段保证商品变更/删除后历史记录可读、群消息可发。
- `order_no` 生成规则：`时间戳(14位) + 6位随机数`，service 层生成。

### credit_logs.sql 同步修改

`type` 注释追加：`类型: 0拾金不昧奖励 1认领成功奖励 2违规扣分 3系统调整 4积分兑换`

## 二、接口总表（`router/advanced/shop_router.go`，新建）

| 分级   | 方法   | 路径                        | Handler                | 说明                                                                 |
|--------|--------|-----------------------------|------------------------|----------------------------------------------------------------------|
| Public | GET    | `/shop/goods/list`          | `ListGoodsHandler`     | query: `keyword,min_price,max_price,page,page_size`；`is_deleted=0`，sort_order 升序；**积分区间筛选用闭区间（含边界），只传其一为单边筛选，min>max 或负数报参数错误**（用户 2026-09-29 确认新增；已实现） |
| Public | GET    | `/shop/goods/:goodsID`      | `GetGoodsHandler`      | 详情；不存在/已删 → 11001                                            |
| Admin  | POST   | `/shop/goods/create`        | `CreateGoodsHandler`   | `ServiceAdminAuthMiddleware`（role≥1）；name 必填 ≤100，price ≥1     |
| Admin  | POST   | `/shop/goods/update`        | `UpdateGoodsHandler`   | 增量更新（指针语义）；库存直接 set；软删商品返回 11001               |
| Admin  | POST   | `/shop/goods/delete`        | `DeleteGoodsHandler`   | 软删 `is_deleted=1`（下架）；已有订单不受影响                        |
| Private| POST   | `/shop/goods/:goodsID/redeem` | `RedeemGoodsHandler` | 兑换：见 §四流程；未绑 QQ → 11005                                    |
| Private| GET    | `/shop/orders`                | `ListMyOrdersHandler`  | 我的兑换记录；query: `page,page_size`；created_at 降序               |

- 无独立上下架接口（status 字段已删），下架即 delete（软删）；"重新上架"由管理员重新创建或后续需求再议。
- admin 组写接口用 POST（与项目既有 tag/location 的 create/update/delete POST 风格一致）。

## 三、响应码（response/response_code.go，新段 11xxxx 商城）

| 码    | 常量                  | 消息               | 触发场景                                   |
|-------|-----------------------|--------------------|--------------------------------------------|
| 11001 | CodeGoodsNotFound     | 商品不存在         | 详情/更新/删除/兑换时商品缺失或已删        |
| 11002 | CodeGoodsStockNotEnough | 商品库存不足     | 兑换时 stock≤0（含并发扣减失败）           |
| 11003 | CodeGoodsNameInvalid  | 商品名称无效       | name 空 / >100 / 重复（uk 校验，service 层查重） |
| 11004 | CodeGoodsPriceInvalid | 商品价格无效       | price <1 或 >1_000_000                     |
| 11005 | CodeShopQQRequired    | 使用商城功能需先绑定QQ | 兑换前校验 user.qq 非空              |

- **积分不足不新设码**：复用现成的 `50001 CodeCreditInsufficient`（积分余额不足）。Q5"合并"按此落地：
  11xxx 段承载 goods 自身错误，积分类错误归 5xxxx 既有段，两段在同一次兑换响应中都可能出现。
- Msg map 已同步补充上述消息文案（Step 2 完成）。

## 四、兑换核心流程（service 层，单个 DB 事务）

```
RedeemGoodsService(userID, goodsID):
  1. user := GetUserByID(userID)；user.ID==0 或 status=0 → 10006
     user.QQ == nil/空 → 11005（QQ 绑定门槛，附加2）
     绑定合法性按用户模块同款规则：5~11 位纯数字（见 §五）
  2. goods := GetGoodsByID(goodsID)；ID==0 或 is_deleted=1 → 11001
  3. goods.stock <= 0 → 11002（快速失败，真实扣减在事务内）
  4. user.Credit < goods.price → 50001（快速失败）
  5. 事务 {
       a. 条件更新扣库存：UPDATE goods SET stock=stock-1 WHERE id=? AND is_deleted=0 AND stock>0
          affected==0 → 11002（并发下另一人先扣完）
       b. 扣积分：UserDao.AddUserCreditTx(tx, userID, -price, 4, "积分兑换:"+goods.Name, 0)
          行锁读余额，不足时返回 error → 整体回滚（余额扣减失败 → 50001）
       c. 写订单：INSERT orders（order_no/goods_name/price/qq/nickname 快照）
       // 通知不进事务，提交成功后统一发送（见第 6 步）
     }
  6. 事务提交成功后：
     - 站内通知（2026-09-30 已接入，同步发送，失败仅记日志，详见 §六）：
       先 type=5 积分变动（变动金额+事务后余额），再 type=6 商品兑换（订单号+引导联系管理员）
     - 发送群消息（异步，失败不影响兑换结果，仅记日志）：
       chensong.Client.SendGroupMessage(
         "[CQ:at,qq="+qq+"] "+nickname+" 成功用 "+price+" 积分兑换了「"+goodsName+"」，"+
         "领取奖励请联系管理员～", global.LNF_CONFIG.ChenSong.ActivatedGroup)
       res.Status != "ok" 视为失败（与 QQGetCode 现有判定一致）
```

- 事务内 b 失败自动回滚 a 的扣库存；c 失败同理；a affected=0 时不执行 b/c 直接返回。
- 加分/扣分流水由 `AddUserCreditTx` 统一写入 `credit_logs`（type=4）。

## 五、QQ 判定规则（与用户模块对齐）

- 用户模块现有判定位于 `handler/basic/user_handler.go`：`req.QQ >= 10000 && req.QQ <= 99999999999`
  （即 **5~11 位**，覆盖最小 QQ 号 10000）。
- shop 侧按同一数值区间实现（`strconv.ParseInt` 校验 5~11 位数值范围），
  集中在 `service/advanced/shop_service.go` 一个小函数内，兑换与发消息共用。
- `users.qq` 列为 varchar，快照字段 `orders.qq` 保持字符串透传，不做二次格式化。

## 六、notification 兑换触发点（2026-09-30 已实施）

原三处 `TODO(shop)` 占位已全部落地为实际调用。`NotificationService.Create(adminID=0, userID, ...)`
现有调用方：item 认领关闭/确认/自行关闭（type=3）、user QQ 绑定成功（type=0）、shop 兑换（本节）。

兑换事务提交成功后的发送顺序（`shop_service.go` 的 `notifyRedeemNotifications`）：

1. **type=5 积分变动**（第一项）：title `积分变动提醒`，
   content `你在积分商城兑换商品「{商品名}」，积分 -{价格}，当前余额 {余额} 分。`
   （余额与响应 credit 同源：事务后重读 users）
2. **type=6 商品兑换**：title `商品兑换成功`，
   content `你已用 {价格} 积分兑换「{商品名}」（订单号 {订单号}），领取奖励请联系管理员～`
   （引导语与 chensong 群消息口径一致）

- 两条均 adminID=0；事务提交后同步发送，失败仅记日志不影响兑换结果与响应。
- 通知不进兑换事务：`Create` 走全局 DB 连接，且通知失败不应回滚兑换。
- 群内 at 提醒仍由 chensong 异步发送（`notifyRedeem`）。
- `model/advanced/notification.go` 与 `notifications.sql` 的 type 注释已补 `6商品兑换`（纯注释，无 DDL）；
  notification_handler.go 的 List @Param type 注释同步。

## 七、文件变更清单（Step 2-5 逐项执行）

1. **Step 2 数据层 ✅（2026-09-29 完成，两张表已由用户执行 SQL 建好）**
   - 新建 `model/mysql/advanced/goods.sql`、`model/mysql/advanced/orders.sql`（§一）
   - 修改 `model/mysql/advanced/credit_logs.sql`：type 注释追加 `4积分兑换`
   - `response/response_code.go`：新增 11xxxx 段 5 个码 + Msg 文案（§三）
   - 新建 `model/advanced/good.go`（附加4 单数命名）：`Good`（TableName=goods）、
     `CreateGoodRequest`、`UpdateGoodRequest{ID required + 指针字段}`、`DeleteGoodRequest`、
     `ListGoodQuery`（form）、`GoodListResponse`；`model/advanced/order.go`：`Order`
     （TableName=orders）、`OrderListQuery`、`OrderListResponse`、
     `RedeemGoodsResponse{order_no,goods_id,goods_name,price,credit,created_at}`
2. **Step 3 dao + service ✅（2026-09-29 完成）**
   - 新建 `dao/good_dao.go`：`GetGoodByID`、`GetGoodByName`（查重仅看未删除，下架商品名可复用）、
     `GetGoodPage`（keyword LIKE name、**min_price/max_price 积分区间筛选（闭区间，已实现）**、
     分页、`buildGoodQuery` 模式仿 item_dao、sort_order 升序→created_at 降序）、`CreateGood`、
     `UpdateGoodByVK`、`SoftDeleteGood`、`RedeemStockTx`（事务内条件更新扣库存：
     `stock>0` 才扣防超卖，affected=0 表示库存不足）
   - 新建 `dao/order_dao.go`：`CreateOrderTx`、`GetOrdersByUserID`（分页，created_at 降序）
   - `dao/enter.go` 注册 `GoodDao/OrderDao`
   - 新建 `service/advanced/shop_service.go`：`ListService`、`GetDetailService`、`CreateService`、
     `UpdateService`、`DeleteService`、`ListMyOrdersService`、`RedeemGoodsService`
     （§四全流程；事务错误经库存哨兵/积分消息映射区分 11002 与 50001）、`validateQQFormat`（§五）、
     `generateOrderNo`（14 位时间戳+6 位随机）、`notifyRedeem`（chensong 群消息，异步，
     失败不影响兑换结果）
   - `service/enter.go` 注册 `ShopService`
3. **Step 4 handler + router ✅（2026-09-29 完成）**
   - 新建 `handler/advanced/shop_handler.go`：7 个 handler，每个含完整 swagger 注释
     （@Summary/@Tags/@Accept/@Produce/@Param/@Success/@Router，路径含 `/api/v1` 前缀）
   - 新建 `router/advanced/shop_router.go`：public/admin/private 三组（§二）
   - `handler/enter.go` 注册 `ShopHandler`；`router/enter.go` 注册 `ShopRouter`；
     `initialization/router.go` Advanced 段注册 `router.ShopRouter.CreateRouter(api)`
   - 完成交付：`shop_handler.go` 7 个 handler + `parseGoodsID` 辅助；`shop_router.go`
     三组路由（`:goodsID` 参数名各组一致，list 静态路径与 :goodsID 并存，同 item 模式）；
     积分区间筛选一并实现（`ListGoodQuery` 增 `min_price/max_price` 字段、
     `buildGoodQuery` 区间条件、`ListService` min>max 校验、list 接口 swagger 补 @Param）
4. **Step 5 验证与文档回写 ✅（2026-09-29）**
   - `go build ./...` + `go vet ./...` 通过（含 Step 4 与积分区间筛选，2026-09-29 复跑）
   - 本文 §〇 状态行更新为"已完成"；接口总表核对实际路由
   - `swag init` 待用户批准（沿用 item 模块约定）
5. **顺带修正**：`model/mysql/advanced/localtions.sql` 文件名拼写错误（localtions→locations）——
   本次不做，仅记录（改名涉及 git 历史，由用户决定）。

## 八、错误映射速查（shop）

- 参数绑定失败 → `CodeParamError(1)`
- 商品不存在/已删 → `11001`；名称无效/重名 → `11003`；price <1 或越界 → `11004`
- 兑换：库存不足/并发扣减失败 → `11002`；积分不足 → `50001`；未绑 QQ → `11005`
- 用户不存在/禁用 → `10006`；群消息发送失败 → 不影响兑换结果，仅记日志
- 其余 DB 异常 → `CodeDatabaseError(6)`
