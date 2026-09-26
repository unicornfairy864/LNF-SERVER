# 图片上传功能方案（头像 / Item 图片）

> **状态：已定稿并实施（2026-09-26）**。定稿决策：
>
> 1. 融入本项目，不单独建项目；
> 2. 本地磁盘存储（uploads/ 目录 + Gin 静态路由），service 层 Storage 接口抽象，未来可切换对象存储（COS/OSS）；
> 3. **URL 存相对路径**（如 /uploads/2026/09/26/{uuid}.jpg），前端展示时自行拼接服务器源；将来换域名/HTTPS 只改前端配置，数据不用迁移；
> 4. 白名单仅 jpg/png/webp，**不含 gif**；
> 5. 错误码用 0xxxx 通用段：9 文件过大 / 10 类型不支持 / 11 保存失败（common_response_code.md 规划的 8 已被 CodeChenSongError 占用，以代码为准）。
>
> 后续迭代（已排期）：
> - **上传限流**：基于已有 Redis 做每用户每小时 N 次上传的计数限流，中间件形式挂在 /upload 组（V1 明确不做）。

> 结论先行：
> 1. **融入本项目**，不单独建项目；
> 2. **本地磁盘存储起步**（`uploads/` 目录 + Gin 静态路由），但在 service 层抽一个 `Storage` 接口，未来可无痛切换到对象存储（COS/OSS）。
> 3. 数据库结构**不需要改动**：`item_images.image_url`、`users.avatar` 本来就是存 URL 的，缺的只是"URL 从哪来"——即上传接口本身。

---

## 一、现状梳理：与图片上传相关的文件

### 1. 已有的"图片"链路（只存 URL，没有真正的上传能力）

| 文件 | 相关内容 |
|---|---|
| `model/basic/item.go` | `ItemImage` 模型（表 `item_images`：`item_id` / `image_url` / `sort_order` 1-3）、`ItemImageInput`、`SetItemImagesRequest` |
| `dao/item_image_dao.go` | `GetImagesByItemID()`（按 sort_order 升序查询）、`ReplaceImages()`（事务内先删后插的覆盖式写法） |
| `service/basic/item_service.go` | `SetImagesService()`：校验最多 3 张（`itemMaxImages`）、`sort_order` 取值 1-3 且不可重复、URL 长度 ≤500 |
| `handler/basic/item_handler.go` | `SetItemImagesHandler()`：`POST /item/:itemID/images` 的入口 |
| `router/basic/item_router.go` | 路由注册：`private.POST("/:itemID/images", ...)`（JWT 保护） |

### 2. 头像链路（同样是"只存字符串"）

| 文件 | 相关内容 |
|---|---|
| `model/basic/user.go` | `User.Avatar *string`（`varchar(255)`）、`UpdateUserRequest.Avatar` |
| `service/basic/user_service.go` | `UpdateService` 中直接把 `avatar` 当普通字符串写入（`updates["avatar"] = *req.Avatar`） |

### 3. 本次需要动到的基础设施（目前是空白）

| 文件 | 现状 |
|---|---|
| `initialization/router.go` | **没有任何静态文件服务**（无 `r.Static` / `StaticFS`） |
| `config/config.go` + `config.yaml` / `config.yaml.example` | 没有 `storage` 配置段 |
| `go.mod` | 没有任何上传、对象存储 SDK、图像处理相关依赖 |

### 4. 关键结论

全库搜索 `upload` **零结果**。也就是说：前端约定的 `image_url` / `avatar` 字段和数据库表都就位了，但图片 URL 的"来源"是缺失的——**本次要补的就是这个洞**：一个真正接收文件、落盘（或上云）、返回 URL 的上传接口。

---

## 二、决策一：单独项目 or 融入本项目？

**建议：融入本项目。**

| 考量 | 分析 |
|---|---|
| 业务规模 | 失物招领场景：每件物品最多 3 张图、每用户 1 张头像，QPS 很低，图片量完全可预期 |
| 鉴权复用 | JWT 中间件、统一 response 错误码、viper 配置体系都在本仓库；拆出去要么重做一遍，要么引入服务间鉴权 |
| 部署形态 | 目前 MySQL / Redis / ChenSong 全在一台云服务器上，就是单体单机部署，再拆一个上传服务只会增加部署与运维成本 |
| 收益对比 | 拆分的收益（独立扩缩容、多业务共用图床）当前一条都不成立 |

**但要留一个口子**：不要把"写文件"散落在 handler 里，而是收敛到一个 `Storage` 接口（见下文），这样未来切换存储后端不需要动业务代码。

## 三、决策二：文件存哪里？

| 方案 | 优点 | 缺点 | 适配本项目 |
|---|---|---|---|
| **本地磁盘 `uploads/` + Gin 静态路由** | 零依赖、零费用、实现最快；数据在自己手里 | 多实例部署/迁移/备份要自己操心；占服务器磁盘 | ✅ 当前单机部署，最合适 |
| 腾讯云 COS / 阿里云 OSS | 稳定、自带 CDN、不占服务器磁盘、天然支持水平扩展 | 有费用、要管理密钥、多一个外部网络依赖 | 备选（量大或多实例时再切） |
| 自建 MinIO | 免费、S3 兼容、可控 | 要额外维护一个常驻服务 + 磁盘规划，对当前规模属于过度设计 | ❌ 不建议现在上 |
| 数据库 BLOB（MySQL 存文件） | 备份简单（和业务数据一起） | 撑大数据库、拖慢备份、HTTP 读文件要经过应用层，性能差 | ❌ 不考虑 |

**建议：本地磁盘起步 + driver 抽象。**

- 现有部署就是一台服务器，图片总量小（每 item ≤3 张、单张限 5MB 量级），磁盘增长完全可控；
- `config.yaml` 增加 `storage` 段，用 `driver` 字段切换 `local` / 未来的 `cos`；
- 上传 handler 只依赖 `Storage` 接口，切对象存储时新增一个实现即可，业务代码零改动；
- 上线后把 `uploads/` 目录纳入服务器备份脚本即可覆盖数据安全。

## 四、建议的实现要点（草案）

1. **接口**：`POST /api/v1/upload/image`，`multipart/form-data`，字段 `file`，JWT 保护（挂 private 组）。
2. **安全校验**：
   - 大小上限（建议 5MB，写进 config）；
   - 类型白名单：jpg / jpeg / png / webp，且用 `http.DetectContentType` 嗅探文件头，**不信任扩展名和客户端声明的 Content-Type**；
   - **不使用客户端原始文件名**（防路径穿越、防乱码）：服务端用 uuid 重命名；
   - 文件名形如 `uploads/2026/09/26/{uuid}.{ext}`，按日期分目录避免单目录文件爆炸。
3. **返回**：拼接 `storage.base_url` 后返回完整 URL，前端拿到 URL 再走现有的 `POST /item/:itemID/images`（item 图）或 `POST /user/update`（avatar）流程——**现有数据流完全复用，不需要动**。
4. **静态服务**：`initialization/router.go` 中增加 `r.Static("/uploads", cfg.Path)`。注意挂载在根路径而不是 `/api/v1` 前缀下，避免与 API 路由混淆。
5. **（可选，后期）** 头像可以做固定尺寸压缩/裁剪；item 图按 `sort_order=1` 生成缩略图。当前阶段不必做。

## 五、若实施，预计新增 / 改动的文件清单

| 类型 | 文件 | 内容 |
|---|---|---|
| 新增 | `handler/basic/upload_handler.go` | 上传接口 handler + swagger 注释 |
| 新增 | `service/basic/upload_service.go`（或 `service/storage/`） | `Storage` 接口 + `LocalStorage` 实现（保存、校验、URL 拼接） |
| 新增 | `router/basic/upload_router.go` | 注册 `/upload/image`（private 组） |
| 修改 | `config/config.go`、`config.yaml`、`config.yaml.example` | 增加 `storage:` 段（driver / path / max_size / allowed_exts / base_url） |
| 修改 | `initialization/router.go` | 注册静态目录 `/uploads` 与上传路由 |
| 修改 | `response/response_code.go` | 新增上传相关错误码（文件过大、类型不支持、保存失败等） |
| 重新生成 | `docs/docs.go` 等 | `swag init` 更新 swagger 文档 |
