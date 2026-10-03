# 任务：为 LNF-SERVER 增加管理员数据分析 API（仪表盘 + 扩展统计）

## 一、需求范围与实施批次

为管理员端新增数据分析能力，共 11 个接口，分三批实施。路由名为建议值，阶段 1 细化后经用户批准。
 
### 批次 1 —— 仪表盘核心
| 接口 | 功能 | 要点 |
|---|---|---|
| GET /admin/stats/overview | 总览 KPI | 累计发布、成功归还、待处理、归还率，各含较上月环比；附超 24h 未处理数 |
| GET /admin/stats/trend | 近 30 日趋势 | 按天聚合发布量/归还量双序列；缺数日期补 0；归还按 claim_time 归日 |
| GET /admin/stats/locations | 地点分布 | `limit` 缺省 4（仪表盘高频地点）；`limit=0` 返回全量地点分布；含 count 与 percent |

### 批次 2 —— 归还效率与滞留分析
| 接口 | 功能 | 要点 |
|---|---|---|
| GET /admin/stats/return-duration | 平均归还时长 | `claim_time - created_at` 的均值与中位数；支持按地点/类型分组；支持时间范围 |
| GET /admin/stats/items/stagnant | 滞留待处理清单 | status=0 且发布超 N 天（N 可传，缺省值阶段 1 确认）；分页、按时长排序；返回物品摘要字段 |
| GET /admin/stats/items/high-view | 高浏览低认领清单 | view_count ≥ 阈值 且未认领超 N 天；阈值参数化；分页 |
| GET /admin/stats/time-heatmap | 发布时段热力 | 星期几 × 小时的发布量矩阵；支持时间范围 |

### 批次 3 —— 维度拆分、漏斗与导出
| 接口 | 功能 | 要点 |
|---|---|---|
| GET /admin/stats/distribution | 类型/标签维度分析 | `dimension=type\|tag`；各分桶的发布量、归还率；type 指失物/拾物 |
| GET /admin/stats/funnel | 用户参与漏斗 | 注册→发布→认领→归还四级转化；口径见 §三-3，与用户确认后再实现 |
| （横切）CSV 导出 | 统一报表导出 | 聚合接口支持 `export=csv`；见 §三-4；批次 3 最后实现 |

批次排序原因：批次 1/2 共享同一套时间与状态口径，连续实现成本最低；漏斗依赖口径确认、导出依赖接口面稳定，故放批次 3。

## 二、必读材料（动手前按顺序阅读）
1. `agent.md/api_guide.md` —— 接口规范（HTTP 恒 200、code/data 响应结构、鉴权、admin 路由约定）
2. `agent.md/api_agent.md` —— 前后端对接现状，避免与已有接口/字段/错误码冲突
3. `agent.md/common_response_code.md` + `response/response_code.go` —— 错误码申请与登记规则
4. `handler/` `service/` `dao/` `model/` `router/` 中 admin 相关代码（如 announcement、user 管理），摸清现有管理员能力与分层写法
5. `agent.md/item_agent.md`、`model/mysql/basic/items.sql` —— items 状态枚举、索引、软删约定

## 三、统一指标与口径约定（所有接口必须遵守；与代码冲突处以核实结果为准）

### 1. 通用规则
- 所有统计排除 `is_deleted=1`
- 时间统一东八区，按自然日/自然月聚合
- 所有 GET /admin/stats/* 支持 `start_date`/`end_date`（含当日）或 `days` 参数；各接口缺省值在文档中写明
- 分页遵循项目现有 page/page_size 约定
- 响应遵循项目规范：HTTP 恒 200，`code=0` 成功

### 2. 状态口径
- 发布量：`is_deleted=0`，按 `created_at` 落在周期内
- 待处理：`status=0`
- 成功归还：`status IN (1,2)` 且 `claim_user_id IS NOT NULL`
- ⚠️ 阶段 1 必须核实：发布者 close 无认领的物品是否也会置 status=2。若是，此类记录单独计入「已关闭/撤回」口径，归还率分母 = 归还 + 待处理 + 已关闭，并在文档中写清恒等式
- 归还时间一律取 `claim_time`

### 3. 专属口径（阶段 1 列入待确认清单）
- 漏斗：推荐「周期内活跃口径」——期内新增注册 / 期内发过物品 / 期内认领过（distinct claim_user_id）/ 期内完成归还的用户数；cohort 按注册月队列作为备选。实现前与用户确认
- 归还时长中位数：先确认 MySQL 版本；≥8.0 可用窗口函数，否则用近似方案，admin_agent.md 中写明所选方法
- 高浏览低认领阈值：view_count 下限与滞留天数缺省值（如浏览 ≥50、发布 ≥7 天）由用户确认
- 清单类接口（stagnant/high-view）返回字段白名单在阶段 1 列出待批（建议：id、title、type、status、location、view_count、created_at、滞留天数，不含联系方式）

### 4. CSV 导出
- 复用各接口同一查询逻辑，通过 `export=csv` 触发，不另起一套统计代码
- CSV 加 UTF-8 BOM（保证 Excel 打开中文不乱码）；文件名按 RFC 5987 编码支持中文
- 流式写出（分批查询 + flush），防大数据量撑爆内存；文件不落盘
- 首批覆盖：overview、trend、locations、distribution、return-duration、time-heatmap；清单类接口后续再议

## 四、工作流程（严格分阶段，未经用户确认不得跨阶段）

**阶段 1 —— 现状调研与规划（只读）**
- 阅读 admin 相关代码，新建 `agent.md/admin_agent.md`，记录：现有 admin 路由与能力清单、涉及的表/模型/索引、本次需求差距分析、三批接口设计（路由、请求/响应 JSON 结构、错误码、SQL 与索引方案）、实施步骤拆分
- 核实 §三-2 的 close 语义与 MySQL 版本，输出**待确认问题清单**（口径、阈值、字段白名单、路由命名等），一次性向用户提出
- 将完整计划呈现给用户，**获得明确批准后才能开始编码**

**阶段 2 —— 编码（仅限用户批准的范围）**
- 按现有分层（handler → service → dao → model）实现，按批次交付；每批完成后向用户简报，批准后进入下一批
- 路由挂 admin 分组，鉴权方式与现有 admin 接口一致
- 新增错误码按既有码段规则登记到 `response/response_code.go`，并同步 `agent.md/common_response_code.md`
- 如需加索引/建表：SQL 放入 `model/mysql/` 对应目录，在 admin_agent.md 说明执行方式，**实际执行交由用户**
- 每批完成后执行 `go build ./...` 确保可编译

**阶段 3 —— 文档同步（编码完成后）**
- 更新 `agent.md/api_agent.md` 与 `agent.md/api_guide.md`：接口表、请求/响应示例、错误码、前端对接注意事项，与代码逐一对应
- 补齐 handler 的 swagger 注释；`swag init` 由用户执行，完成后提醒用户即可，不要自行改动 `docs/` 生成物
- 回写 `agent.md/admin_agent.md` 的实施状态（已完成批次、遗留问题）

## 五、硬性约束
- **范围外一律只读**。可写文件仅限：`agent.md/` 下的文档、用户明确批准的新增/修改代码与 SQL
- 禁止：git commit/push 等任何 git 写操作、删除文件、修改 `config.yaml`、直接对数据库执行变更
- 遵循现有响应规范：HTTP 恒 200，`code=0` 成功；新错误码不得与现有码段冲突
- 指标口径、接口命名、返回字段不得自行发明，与本文档或用户确认结果保持一致

## 六、沟通规则
- 任何需求歧义、设计取舍、范围变更，先询问用户，不得自行决定
- 阶段 1 的问题一次性批量提出，不要逐条追问
- 每阶段/每批次结束输出简报（做了什么 / 改了哪些文件 / 下一步），经用户确认后继续
- 若发现文档与代码不一致，以代码为准，并在对应文档中标注修正
