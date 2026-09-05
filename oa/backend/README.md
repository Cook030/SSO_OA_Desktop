# 员工平台权限管理系统 Backend

企业内部员工平台权限管理系统后端服务，用于管理员工账号、平台资源及员工与平台之间的权限关系。

## 系统角色

| 角色 | 权限 |
|------|------|
| 管理员 (admin) | 管理平台、管理员工、分配员工平台权限 |
| 普通员工 (employee) | 查看个人信息、查看自己的平台权限、修改个人密码 |

## 技术栈

| 组件 | 技术 | 版本 |
|------|------|------|
| 开发语言 | Go | 1.23 |
| Web 框架 | Gin | v1.10.0 |
| ORM | GORM + gorm/gen | v1.25.12 / v0.3.28 |
| 数据库 | MySQL | 8.0 |
| 认证 | SSO Token 校验 (转发 SSO /api/v1/auth/introspect) | - |
| 密码加密 | bcrypt (golang.org/x/crypto) | - |
| 跨域 | gin-contrib/cors | v1.7.0 |
| 配置管理 | Viper | v1.18.2 |
| 日志 | Zap + lumberjack（控制台 + 文件双输出） | v1.27.0 / v2.2.1 |
| API 文档 | Swagger (swaggo) | - |

## 项目结构

```
backend
├── cmd/
│   └── server/
│       └── main.go                          # 程序入口
├── config/
│   ├── config.yaml                          # 实际配置
│   └── config.example.yaml                  # 配置模板（供新成员参考）
├── docs/                                    # Swagger 生成产物（swag init）
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/
│   ├── db_model/                            # 数据库模型（gorm/gen 自动生成）
│   │   ├── query/                           # gen query 类型安全查询
│   │   │   ├── gen.go
│   │   │   ├── sys_platform.gen.go
│   │   │   ├── sys_user.gen.go
│   │   │   └── sys_user_platform.gen.go
│   │   ├── sys_platform.gen.go
│   │   ├── sys_user.gen.go
│   │   └── sys_user_platform.gen.go
│   ├── api_model/                           # API 契约模型（请求体 / 响应 DTO）
│   │   ├── request/                         # 请求体
│   │   │   ├── employee.go
│   │   │   ├── permission.go
│   │   │   └── platform.go
│   │   └── response/                        # 响应 DTO
│   │       ├── employee.go
│   │       ├── permission.go
│   │       └── platform.go
│   ├── client/                              # 外部服务客户端
│   │   ├── sso.go                           # SSO 客户端（introspect 校验 / 撤销用户会话）
│   │   └── sso_test.go
│   ├── repository/                          # 数据访问层（使用 gen query）
│   │   ├── user_repository.go               # 用户数据操作
│   │   ├── platform_repository.go           # 平台数据操作
│   │   ├── user_platform_repository.go      # 权限关系数据操作
│   │   └── user_platform_extra.go           # 权限批量查询
│   ├── service/                             # 业务逻辑层
│   │   ├── shared/                          # 公共工具（常量、校验、权限辅助）
│   │   │   ├── constant.go
│   │   │   ├── format.go
│   │   │   ├── permission_helper.go
│   │   │   └── validator.go
│   │   ├── platform_service.go              # 平台管理业务
│   │   ├── employee_service.go              # 员工管理业务
│   │   └── permission_service.go            # 权限管理业务
│   ├── handler/                             # HTTP 处理器
│   │   ├── platform_handler.go              # 平台管理接口
│   │   ├── employee_handler.go              # 员工管理接口
│   │   └── permission_handler.go            # 权限管理接口
│   ├── middleware/
│   │   ├── request_id.go                    # 全局 request_id 中间件（请求链路追踪 + 请求开始/完成日志）
│   │   ├── auth.go                          # SSO Token 认证 + 管理员权限中间件
│   │   └── auth_test.go
│   ├── router/
│   │   └── router.go                        # 路由注册 + 依赖组装
│   └── utils/                               # 工具层
│       ├── config.go                        # 配置加载(Viper)
│       ├── database.go                      # 数据库连接(GORM)
│       ├── logger.go                        # 日志初始化(Zap)
│       ├── password.go                      # 密码加密与校验(bcrypt)
│       └── response.go                      # 统一响应格式
├── test/                                    # 集成测试（独立数据库，Docker 容器化运行）
│   ├── testutil/                            # 测试工具子包
│   │   ├── assert.go
│   │   ├── cleanup.go
│   │   ├── client.go
│   │   ├── config.go
│   │   └── helpers.go
│   ├── data/
│   │   ├── schema.sql
│   │   └── seed.sql
│   ├── scripts/
│   │   ├── run.sh
│   │   └── wait.sh
│   ├── docker-compose.test.yml
│   ├── Dockerfile.test-runner
│   ├── config.yaml
│   ├── main_test.go
│   ├── auth_test.go
│   ├── employee_test.go
│   ├── flow_test.go
│   ├── permission_test.go
│   └── platform_test.go
├── Dockerfile                               # 多阶段构建（私有 ACR 镜像）
├── Makefile                                 # 构建/测试/Docker 快捷指令
├── gen.yaml                                 # gorm/gen 模型生成配置
├── run.sh                                   # 开发环境启动脚本
├── go.mod
├── go.sum
└── README.md
```

### 分层职责

| 层级 | 职责 | 禁止 |
|------|------|------|
| db_model | 数据库模型定义、表结构映射（gorm/gen 生成） | 业务逻辑 |
| repository | 数据访问（使用 gen query 类型安全查询） | 业务逻辑 |
| api_model | API 契约模型（请求体、响应 DTO） | 业务逻辑 |
| service | 核心业务逻辑、参数校验、事务处理 | 直接操作 HTTP |
| handler | HTTP 请求处理、参数解析、返回响应 | 直接操作数据库 |
| middleware | SSO Token 认证、权限校验 | - |
| shared | Service 层公共工具（常量、校验器、权限辅助函数） | - |

## 快速启动

### 1. 环境要求

- Go 1.23+
- MySQL 8.0+

### 2. 配置数据库

复制模板并填写实际配置：

```bash
cp config/config.example.yaml config/config.yaml
```

编辑 `config/config.yaml`：

```yaml
server:
  port: 8080
  mode: debug  # debug | release（生产环境建议 release）

log:
  path: "logs/app.log"  # 日志文件路径（相对工作目录；容器中为 /app/logs/app.log，目录自动创建）
  max_size: 100         # 单个日志文件最大大小（MB），超过后自动轮转
  max_backups: 10       # 保留的旧日志文件个数
  max_age: 30           # 旧日志文件保留天数
  compress: true        # 是否压缩旧日志文件（gzip）
  local_time: true      # 轮转文件命名是否使用本地时间

mysql:
  host: 127.0.0.1
  port: 3306
  username: root
  password: your_mysql_password
  database: maplehaze_permission
  charset: utf8mb4
  max_idle_conns: 10
  max_open_conns: 100

admin:
  account: admin
  password: your_admin_password
  name: 管理员
  phone: "13800000000"
  email: admin@maplehaze.cn

cors:
  allow_origins:
    - "http://localhost:5173"                 # 本机前端开发地址
    - "https://oa.maplehaze.cn"               # 生产/线上前端域名
    - "http://local-nextui.maplehaze.cn:8004" # 本地/开发前端（sso 接口设计文档示例）
  allow_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allow_headers:
    - "Origin"
    - "Content-Type"
    - "Accept"
  expose_headers: []
  allow_credentials: true        # 允许携带 Cookie 等凭证时，allow_origins 不能为 *，必须显式指定
  max_age: 86400

sso:
  base_url: "http://mh-sso2-svc"                     # SSO 服务地址（K8s 服务名或域名）
  introspect_path: "/api/v1/auth/introspect"         # accessToken 校验接口
  revoke_user_sessions_path: "/api/v1/auth/revoke-user-sessions"  # 撤销用户会话接口（重置密码后调用）
  timeout_second: 5                                  # SSO 请求超时（秒）
```

> `config.yaml` 含实际环境连接信息（数据库地址、密码等），请勿在公开仓库中提交真实敏感配置。

### 3. 启动服务

```bash
cd backend
go mod tidy
go run ./cmd/server
```

服务启动后会自动：
- 连接数据库
- 初始化管理员账号（不存在时创建，账号与初始密码见 `config.yaml` 的 `admin` 段）

### 4. 默认管理员

| 字段 | 值 |
|------|------|
| 账号 | admin |
| 密码 | 见 `config.yaml` 中 `admin.password` |
| 角色 | admin |

### 5. 编译运行（可选）

```bash
cd backend
go build -o server ./cmd/server
./server
```

### 6. Makefile 快捷指令

| 指令 | 说明 |
|------|------|
| `make run` | 本地开发模式启动后端 |
| `make build` | 编译生成可执行文件 `./server` |
| `make build-linux` | 交叉编译 Linux amd64 可执行文件 |
| `make lint` | 格式化并执行 `go vet` 静态检查 |
| `make test` | 运行所有单元测试 |
| `make swag` | 生成 Swagger 接口文档 |
| `make docker-build` | 构建 Docker 镜像（标签: `permission-system:go1.23`） |
| `make docker-run` | 使用默认配置运行 Docker 容器（端口 8080） |
| `make docker-run-with-config` | 挂载本地 `config.yaml` 运行 Docker 容器 |
| `make docker-test` | 容器化运行集成测试（自动创建/清理测试数据库） |
| `make tidy` | 整理 Go 模块依赖 |
| `make clean` | 清理编译产物 |
| `make ci` | 完整 CI 流程：格式化、检查、测试、编译 |

### 7. 运行集成测试

集成测试使用独立的 MySQL 容器，通过 Docker Compose 编排自动管理生命周期：

```bash
cd backend
make docker-test
```

测试过程：容器启动 MySQL → 执行 `schema.sql` 初始化表结构 → 运行 `*_test.go` → 完成后自动清理容器。

测试文件位于 `test/` 目录下，覆盖认证、员工、平台、权限四大模块的业务流程。

## API 接口

共 **12** 个接口。鉴权方式见下文[认证说明](#认证说明)，支持两种方式携带 SSO accessToken。

启动后可访问 Swagger 文档：`http://localhost:8080/swagger/index.html`

### 认证说明

本系统不实现登录接口，前端直接从 SSO 获取 accessToken，后端转发 SSO `/api/v1/auth/introspect` 校验。

accessToken 通过名为 `mh_sso2_access_token` 的 Cookie 携带（与 SSO 接口设计文档一致）。后端仅认 Cookie 通道，不再解析 `Authorization: Bearer` Header。

### 平台管理模块

| Method | URL | 说明 | 鉴权 |
|--------|-----|------|------|
| GET | /api/platforms | 平台列表(分页+permissionCount) | 管理员 |
| POST | /api/platforms | 新增平台 | 管理员 |
| PUT | /api/platforms/:id | 编辑平台 | 管理员 |
| DELETE | /api/platforms/:id | 删除平台(物理删除+级联清理权限) | 管理员 |

### 员工管理模块

| Method | URL | 说明 | 鉴权 |
|--------|-----|------|------|
| GET | /api/employees | 员工列表(搜索/筛选/分页) | 管理员 |
| GET | /api/employees/departments | 部门列表 | 管理员 |
| POST | /api/employees | 新增员工(含权限分配) | 管理员 |
| PUT | /api/employees/:id | 编辑员工(权限全量覆盖) | 管理员 |
| DELETE | /api/employees/:id | 删除员工(物理删除+级联清理权限) | 管理员 |
| PUT | /api/employees/:id/reset-password | 重置员工密码(恢复默认，需再次修改，并调用 SSO 撤销该用户会话强制重新登录) | 管理员 |

**员工列表查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 否 | 姓名/手机号/邮箱模糊搜索 |
| department | string | 否 | 部门筛选（精确匹配） |
| platformId | int | 否 | 平台 ID 筛选 |
| page | int | 否 | 当前页，默认 1 |
| pageSize | int | 否 | 每页条数，默认 20 |

**新增员工请求：**

```json
{
  "name": "张三",
  "phone": "13800000000",
  "emailPrefix": "zhangsan",
  "department": "技术部",
  "platformIds": [1, 2],
  "password": "初始密码"
}
```

> account 取 emailPrefix，email = emailPrefix@maplehaze.cn，初始密码由前端传入（password 字段），后端 bcrypt 加密后存储。

### 权限管理模块

| Method | URL | 说明 | 鉴权 |
|--------|-----|------|------|
| POST | /api/employees/permissions/batch | 批量设置权限(增量新增) | 管理员 |
| DELETE | /api/employees/permissions/batch | 批量删除权限 | 管理员 |

**批量设置权限请求：**

```json
{
  "userIds": [1, 2, 3],
  "platformIds": [1, 2]
}
```

> 增量新增，已存在的权限关系不重复插入（INSERT IGNORE），使用数据库事务保证原子性。

**批量删除权限请求：**

```json
{
  "userIds": [1, 2, 3],
  "platformIds": [1, 2]
}
```

> 删除选中员工的指定平台权限。

## 统一返回格式

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

| code | 说明 |
|------|------|
| 200 | 成功 |
| 400 | 参数错误 |
| 401 | Token 无效或过期 |
| 403 | 无权限(非管理员访问管理接口) |
| 404 | 资源不存在 |
| 500 | 服务器错误 |

## 日志

基于 Zap 实现，同时输出到**控制台（stdout）**与**本地日志文件**，文件写入由 lumberjack 管理自动轮转（按大小/数量/天数切割、可选 gzip 压缩）。

### 日志配置

在 `config.yaml` 的 `log` 段配置：

```yaml
log:
  path: "logs/app.log"  # 日志文件路径（相对工作目录，启动时自动创建目录）
  max_size: 100         # 单个日志文件最大大小（MB），超过后自动轮转
  max_backups: 10       # 保留的旧日志文件个数
  max_age: 30           # 旧日志文件保留天数
  compress: true        # 是否压缩旧日志文件（gzip）
  local_time: true      # 轮转文件命名是否使用本地时间
```

> 容器部署时工作目录为 `/app`，日志落在 `/app/logs/app.log`（K8s 环境建议将该目录挂载为 volume 持久化）。

### 日志级别规范

| 级别 | 用途 | 说明 |
|------|------|------|
| INFO | 正常的重要业务事件 | 启动流程、认证结果、请求开始/完成等；生产环境默认输出 |
| DEBUG | 开发/排查问题 | SSO 请求/原始响应/解析结果等调试细节；仅 `server.mode: debug` 时输出，生产不产生 |
| ERROR | 真正的异常 | SSO 服务异常（网络不通、500 等）；生产环境默认输出 |

**敏感信息保护**：生产 INFO 日志不打印完整 token、密码、Cookie、Authorization、完整 response body；SSO 请求日志仅记录 token 前缀（前 20 位）用于区分请求。

### request_id 请求链路追踪

全局中间件为**每个 HTTP 请求**生成唯一 `request_id`（UUID），从请求开始到认证、业务处理、SSO 调用全程复用同一个 ID。生产环境按 `request_id` 即可串联整条请求链路：

```json
{"level":"info","msg":"HTTP 请求开始","request_id":"f063e85a-...","method":"GET","path":"/api/employees"}
{"level":"info","msg":"SSO introspect 成功","request_id":"f063e85a-...","user_id":1}
{"level":"info","msg":"HTTP 请求完成","request_id":"f063e85a-...","method":"GET","path":"/api/employees","status":200,"latency_ms":18,"user_id":1}
```

排查问题时可回答：谁（`user_id`）请求了什么（`method`/`path`）、用了多久（`latency_ms`）、结果如何（`status`）。

## 认证设计

- **全程 SSO Token**：本系统不签发自己的 JWT，前端所有请求携带 SSO accessToken
- **Token 提取**：从 `mh_sso2_access_token` Cookie 中提取 accessToken（`mh_sso2_access_token` 为 SSO 签发的 Cookie 名，与 SSO 接口设计文档一致），不再解析 `Authorization: Bearer` Header
- **校验流程**：中间件提取 accessToken → 转发 SSO `GET /api/v1/auth/introspect` → 校验返回的 `code == 200` → 用返回的 `data.userId` 查本地用户（userId 与本地 `sys_user.id` 对齐，兼容 number/string 类型）→ 写入上下文(userId/account/role)
- **401 场景**：缺少认证 Cookie、Token 过期/无效、本地用户不存在。Token 过期时透传 SSO 返回的 `msg` 文案；SSO 服务本身异常（网络、500 等）返回"认证服务异常"，避免前端误判为 Token 过期
- **重置密码联动**：重置员工密码成功后调用 SSO `POST /api/v1/auth/revoke-user-sessions` 撤销该用户全部会话与 Refresh Token，强制其重新登录

## 数据库设计

数据库名：`maplehaze_permission`，字符集：`utf8mb4`，存储引擎：`InnoDB`。

共 3 张业务表：

| 表名 | 说明 |
|------|------|
| sys_user | 用户表(管理员+员工，role 字段区分) |
| sys_platform | 平台表 |
| sys_user_platform | 用户平台权限关系表(多对多中间表) |

### 表关系

```
sys_user ──┐
           ├── sys_user_platform (多对多中间表)
sys_platform ─┘
```

- **级联删除**：删除用户/平台时，由代码层负责清理 `sys_user_platform` 中的关联记录（不使用数据库外键约束）
- 所有业务删除均为**物理删除**，不使用软删除标志位

### sys_user 用户表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED | 主键，自增 |
| account | VARCHAR(64) | 登录账号，唯一 |
| password | VARCHAR(255) | 密码(bcrypt 加密) |
| name | VARCHAR(64) | 姓名 |
| phone | VARCHAR(20) | 手机号，唯一 |
| email | VARCHAR(128) | 邮箱，唯一 |
| role | VARCHAR(32) | 角色: admin/employee，默认 employee |
| department | VARCHAR(64) | 所属部门 |
| password_changed | BOOL | 是否已修改初始密码，默认 false；重置密码后置回 false |
| create_time | DATETIME | 创建时间 |
| update_time | DATETIME | 更新时间 |

### sys_platform 平台表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED | 主键，自增 |
| name | VARCHAR(128) | 平台名称，唯一 |
| link | VARCHAR(128) | 平台链接，唯一 |
| create_time | DATETIME | 创建时间 |
| update_time | DATETIME | 更新时间 |

### sys_user_platform 用户平台权限关系表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED | 主键，自增 |
| user_id | BIGINT UNSIGNED | 用户 ID |
| platform_id | BIGINT UNSIGNED | 平台 ID |
| create_time | DATETIME | 创建时间 |

> 联合唯一索引：(user_id, platform_id)

## 跨域配置

项目使用 `github.com/gin-contrib/cors` 中间件处理跨域，前端直接请求后端地址即可，**不再通过 nginx 反向代理**。

在 `config.yaml` 中配置允许的前端来源：

```yaml
cors:
  allow_origins:
    - "http://localhost:5173"
  allow_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allow_headers:
    - "Origin"
    - "Content-Type"
    - "Accept"
  expose_headers: []
  allow_credentials: true
  max_age: 86400
```

**注意事项：**
- `allow_credentials` 设置为 `true` 时，`allow_origins` 不能使用 `*`，必须显式列出前端域名。
- 生产环境请将 `allow_origins` 替换为真实前端部署地址。
- 生产环境需确保后端服务可直接被前端访问（如通过负载均衡或安全组开放端口）。
