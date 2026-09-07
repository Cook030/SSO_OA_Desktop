# 企业 OA Agent：第一迭代

本服务是当前工作台的真实执行后端：Python + FastAPI + LangGraph，支持真实模型工具调用、SSO 登录与刷新、持久化会话、流式任务事件、角色分配预览、确认后执行和 OA 回查。前端为 `workbench/frontend` 的原生 TypeScript；无需 React。

## 本轮范围

- 查询员工（姓名/账号，分页）、部门、平台、角色、权限点；在多轮对话里继续指代已查询人员。
- 角色归属差异比较。该工具不会把角色差异直接声称为有效权限差异；可以继续查角色权限点和权限目录。
- 单员工增量分配 1–20 个角色，确认前不写入；始终保留当前已有角色。
- 计划绑定实际操作者、任务和原始参数；客户端只提交同意/拒绝，不能修改执行参数。
- 确认前后复核身份、权限、目标员工、已有角色及新增角色的权限点。变化则拒绝执行并要求重新生成计划。
- 真实 LangGraph `interrupt()` / `Command(resume=...)`；等待确认的任务能通过持久化 checkpoint 恢复。
- Bearer 是工作台随机会话标识，数据库只保存其 SHA-256；SSO 凭据使用 Fernet 加密。密码仅在登录时转发 SSO。
- 工作台 token 只在前端内存保存，刷新页面/重开客户端需要重新登录；会话历史和待确认任务保留在服务端。
- SSE 事件可按 ID 重放，用户隔离覆盖会话、消息、计划、任务、事件和审计。

## 本地运行

要求 Python 3.12+、uv。以下命令均在 `agent-server/` 执行。

```powershell
uv sync --frozen
uv run python -c "from cryptography.fernet import Fernet; print(Fernet.generate_key().decode())"
```

参考 `.env.example` 新建 `.env`，填写生成的 `AGENT_ENCRYPTION_KEY`、实际 OA/SSO 地址、模型名称、模型网关地址与密钥。使用长期稳定的加密密钥，重启后不要更换，否则现有会话凭据无法解密。

```powershell
uv run uvicorn app.main:create_app --factory --host 127.0.0.1 --port 8090 --workers 1
```

业务表在启动时按 ORM 模型自动创建（`Base.metadata.create_all`，已存在则跳过），不需要单独的迁移步骤。

本地默认用 SQLite，业务数据和 LangGraph checkpoints 分别保存在 `data/agent.db` 和 `data/checkpoints.db`。数据库不保存 OA 密码或明文 SSO Token。模型未配置时允许登录和查历史，发送任务返回明确的 503。

另开终端运行前端：

```powershell
cd workbench/frontend
pnpm install --frozen-lockfile
pnpm dev
```

浏览器打开 `http://127.0.0.1:5173`，服务地址填 `http://127.0.0.1:8090`，使用现有企业 SSO 账号登录。远程部署需要 HTTPS，并显式配置 `AGENT_CORS_ORIGINS`。

桌面打包在 `workbench/` 执行 `F:\GOPATH\bin\wails.exe build`。新的 Wails 入口仅绑定 `Desktop`（只暴露 `AppName()`），Go 侧不参与业务，仅作为嵌入 `frontend/dist` 的窗口容器；前端源码不引用任何 `wailsjs` 绑定，因此浏览器模式与桌面模式共用同一套前端代码。旧的 `App` 绑定、正则 Agent、本地账号、Mock Skill、React 页面、旧 YAML 配置和旧 `wailsjs` 绑定均已删除，需要对照时查 Git 历史。用户已有的本地历史数据库不在仓库内，本轮不涉及。

## PostgreSQL / Docker

`compose.yaml` 使用 PostgreSQL 保存业务表及 LangGraph checkpoints。设置 `.env` 中的 `AGENT_DB_PASSWORD`（建议使用字母数字随机串，其他字符需 URL 编码），并将 OA/SSO 地址设置为容器可访问地址。在 Docker Desktop 上，宿主机服务可用 `http://host.docker.internal:实际端口`。

```powershell
docker compose up --build -d
```

只支持 **一个 Agent worker / 一个实例**。当前运行器使用进程内任务调度；数据库任务表持久化状态，但尚无分布式队列。关闭桌面不会停止已提交任务；Agent 服务重启会将执行中任务标记失败、写入中的计划标记 `unknown`，绝不自动重放可能已提交的写操作。等待确认任务保持可恢复。

启动时使用数据库 advisory lock（PostgreSQL）或文件锁（SQLite）阻止第二个实例误执行任务恢复。

首次启动时自动建表（幂等），LangGraph 官方 saver 管理自身 checkpoint schema。后续新增字段/索引需要手工执行 DDL。加密密钥和数据库需一起备份。内网部署建议经反向代理终止 TLS，当前 Compose 端口仅绑定本机。

## 已核实的 OA/SSO 契约

| 操作 | 当前仓库接口 |
|---|---|
| 企业登录 / 刷新 / 退出 | `POST /api/v1/auth/login`, `/refresh`, `/logout` |
| 身份、管理员标记、权限码 | `GET /api/me/permissions` |
| 员工搜索 | `GET /api/employees?keyword=...&page=...&pageSize=20` |
| 部门 | `GET /api/employees/departments` |
| 平台 | `GET /api/platforms` |
| 角色目录 | `GET /api/roles`（data 是数组） |
| 员工角色 | `GET /api/users/{id}/roles`（data 是数组） |
| 角色权限点 | `GET /api/roles/{id}/permissions`（data 是 ID 数组） |
| 权限目录 | `GET /api/permissions` |
| 增量分配 | `POST /api/roles/users`，`userIds` + `roleIds` |

OA 仅从 `mh_sso2_access_token` Cookie 读取 Token。OA 业务错误可能仍返回 HTTP 200，必须同时检查 JSON `code`。SSO 消息字段是 `msg`，OA 是 `message`；服务对外使用受控错误文案，不回显上游响应中的敏感字段。

当前 SSO 没有授权码流程。本轮代理其已有账号登录/刷新接口，未虚构浏览器 OAuth 回调或 Token Exchange。

## 与 workbench 的接口契约

前端（`workbench/frontend/src`）只通过 HTTP + SSE 访问本服务，没有 Go ↔ Python 的进程内调用。

| 前端 | 后端 | 说明 |
|---|---|---|
| `main.ts` 登录 | `POST /api/auth/login` | 返回 `{token, actor, expires_at}` |
| `main.ts` | `POST /api/auth/logout` | 可能返回 `{warning}`（远端撤销未确认） |
| `main.ts` | `GET /api/me` | `model_configured`，前端据此禁用输入 |
| `main.ts` | `GET/POST /api/conversations`、`GET /api/conversations/{id}` | 详情含 messages 与 tasks |
| `main.ts` | `POST /api/conversations/{id}/messages` | 202 + Task；未配模型返回 503 |
| `main.ts` | `GET /api/tasks/{id}` | 含 plans 与 audit |
| `main.ts` | `POST /api/operation-plans/{id}/decision` | 仅 `{approved}`，`extra="forbid"` |
| `main.ts` | `POST /api/tasks/{id}/cancel` | |
| `api.ts` | `GET /api/tasks/{id}/events?after=` | SSE，Bearer 放 header |

SSE 事件（`app/runtime.py`、`app/graph.py` 发出）：`task.running`、`message.delta`、`message.completed`、`tool.started`、`tool.completed`、`operation.executing`、`operation.confirmation_required`、`task.completed`、`task.failed`。前端渲染前六种；`operation.confirmation_required` 与两个终态事件前端不直接消费，改为在流结束后重新拉取任务详情。流按 60 秒服务端断流 + 前端 `after` 游标续接，最多重试 5 次。

确认卡片的计划载荷由后端构造、前端只渲染：`action`、`employee`、`before_roles`、`added_roles`（含 `permission_ids`，供执行前复核）、`after_roles`。前端只提交同意或拒绝，不回传员工与角色参数。

契约目前没有共享定义，前端类型在 `workbench/frontend/src/main.ts` 手写维护，也没有契约级自动化测试。**修改事件名、计划载荷或字段含义时，必须同步后端、前端类型与本节表格。**

## 可靠性边界与后续迭代

1. **本轮不是完整企业上线版本。** 删除/替换角色、创建员工、重置密码、批量员工操作、审批流和多 Agent 暂未开放。
2. OA 仍是最终权限裁决层。当前 OA 的部门级数据范围能力没有在本轮新增，不能宣称已实现部门隔离。
3. 当前 OA 增量接口没有原子的版本条件或业务幂等键。本轮通过单实例确认消费、执行前复核和增量写入防止重复确认及覆盖已有角色；复核与写入之间仍存在并发窗口。角色权限变化的严格防护需要 OA 增加事务内条件写入，再扩展全量覆盖和高风险操作。
4. 请求超时不会重试写入，标记为结果待核查。增量接口会跳过已存在的分配，但这不等于整个工作流具有 exactly-once 保证。
5. 重名候选通过模型追问、部门/账号信息及确认卡片明确对象；本轮未实现独立人员选择组件。真实模型理解率需用实际企业指令集评估。
6. 事件按节点和回答分片持久化；目前纯文本渲染，避免引入 Markdown/HTML 安全依赖。长会话达到输入大小上限时提示新建任务。
7. 尚未完成原生安全凭据存储、浏览器 SSO 跳转、自动更新、生产级监控告警、数据保留清理和不可篡改审计。审计在应用层只追加，数据库管理员仍能改动。
8. 模型数据范围由管理员配置网关控制。只读员工工具不向模型提供手机号和邮箱；不要向未经企业授权的模型服务配置真实数据。

## 验证

```powershell
uv run pytest -q
uv run ruff check app tests
uv run ruff format --check app tests
```

测试使用真实 LangGraph 和 SQLite saver，OA 与模型通过边界替身隔离，不修改真实人员权限。覆盖确认/拒绝、重复确认、参数篡改、越权、权限撤回、计划过期、前置数据冲突、OA 超时/拒绝、SSO 刷新、事件重放、重启恢复等场景。PostgreSQL 部署与真实模型/SSO/OA 联调需在有相应服务和凭据的环境执行。

UI 隔离验证入口（只用于本地测试，绝不可部署）：

```powershell
uv run uvicorn tests.demo_server:create_demo --factory --host 127.0.0.1 --port 8091
```

该入口登录填任意测试字符串，前端服务地址指向 8091；第一次任务输入“给张三增加运营只读角色”。它使用内存中的测试 OA 和脚本模型，与真实服务完全隔离。测试模块未打入 Docker 镜像。

框架行为依据：[LangGraph interrupts](https://docs.langchain.com/oss/python/langgraph/interrupts)、[Python persistence](https://docs.langchain.com/oss/python/langgraph/persistence)。特别注意恢复时中断节点从头运行，因此确认节点在 `interrupt()` 之前不执行写入。
