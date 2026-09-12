











# 🤖 LNF-SERVER Agent Protocol

## 🌌 系统概述

**LNF-SERVER** 是一个基于 Go 1.26.4 构建的失物招领智能服务系统，融合了传统数据库存储与 AI Agent 能力。本系统旨在通过智能化的方式连接失主与拾获者，让每一件丢失的物品都能找到回家的路。

## 🧠 核心能力矩阵

### 📊 数据持久化层
- **MySQL 集成**: 基于 GORM 的数据库抽象层
- **配置管理**: Viper 驱动的动态配置系统
- **事务支持**: 完整的数据库事务处理能力

### 🤖 AI Agent 模块
- **LLM 客户端**: 支持多模型接入（默认 GLM-4.7）
- **智能匹配**: 基于语义理解的物品匹配算法
- **可扩展架构**: 预留 Agent 扩展接口

### 🛡️ API 网关层
- **RESTful API**: 标准化的 API 设计
- **统一响应格式**: 结构化的错误处理机制
- **中间件支持**: 认证、日志、限流等能力

## 🚀 快速启动协议

### 环境初始化
```bash
# 克隆代码仓库
git clone https://github.com/unicornfairy864/LNF-SERVER.git
cd LNF-SERVER

# 安装依赖
go mod download
```

### 配置系统
复制示例配置文件并根据你的环境进行修改：
```bash
cp config.yaml.example config.yaml
```

编辑 `config.yaml`：
```yaml
# 数据库连接配置
mysql:
  database_host: "localhost"
  database_port: "3306"
  database_dbname: "your_database_name"
  database_user: "your_username"
  database_password: "your_password"

# AI Agent 配置
api:
  openai_key: "your_api_key"
  openai_base_url: "https://open.bigmodel.cn/api/paas/v4"
  default_model: "glm-4.7"
```

### 启动服务
```bash
go run main.go
```

## 📁 项目架构

```
LNF-SERVER/
├── agent/              # AI Agent 核心模块
│   ├── enter.go       # Agent 入口
│   └── llm_client.go  # LLM 客户端实现
├── api/               # API 路由层
│   └── v1/           # API 版本 1
│       └── system/   # 系统相关 API
├── config/            # 配置管理
├── core/             # 核心功能模块
├── dao/              # 数据访问对象
├── global/           # 全局变量与常量
├── middleware/       # 中间件
├── model/            # 数据模型
├── response/         # 响应结构体
├── service/          # 业务逻辑层
├── test/             # 测试文件
├── docs/             # 文档
├── main.go           # 程序入口
├── config.yaml       # 配置文件
└── go.mod            # Go 模块定义
```

## 🔌 API 响应协议

所有 API 遵循统一的响应格式：

```json
{
    "code": 0,
    "message": "Success",
    "data": {}
}
```

### 状态码定义

| 代码 | 消息 | 说明 |
|------|------|------|
| 0 | Success | 请求成功 |
| 500 | Internal Server Error | 服务器内部错误 |
| 1001 | Incorrect username or password | 用户名或密码错误 |
| 1002 | User already exists | 用户已存在 |
| 1003 | Invalid Param | 参数无效 |
| 1004 | Invalid token | Token 无效 |

## 🧪 测试协议

系统内置测试模块，位于 `test/` 目录下。运行测试：

```bash
go test ./...
```

## 🤝 贡献指南

我们欢迎所有形式的贡献！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 详见 LICENSE 文件

## 🌟 致谢

感谢所有为 LNF-SERVER 项目做出贡献的开发者！

---

**注意**: 本系统仍在积极开发中，部分功能可能尚未完全实现。欢迎提出 Issue 和 Pull Request！

*让科技温暖每一个失而复得的瞬间* 🌈
