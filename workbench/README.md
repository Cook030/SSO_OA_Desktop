# 智工工作台（OA Agent 桌面 MVP）

类似 Codex / WorkBuddy 的会话式桌面工作台，但业务系统只连接当前 OA。

## 当前能力

在对话框输入：

```text
给张三分配 A 平台权限
```

Agent 会按顺序调用当前 OA 的接口：

1. `GET /api/employees?keyword=张三` 查询员工；
2. `GET /api/platforms` 查询平台；
3. `POST /api/employees/permissions/batch` 执行增量授权；
4. 在会话中保留任务结果和本地事件记录。

员工或平台存在同名项时，Agent 会停止执行并提示用户提供更明确的信息。

## 安全开关

首次启动配置默认在 `%APPDATA%/MapleHaze/AIWorkbench/config/workbench.yaml` 生成，且默认：

```yaml
oa:
  mode: dry-run
```

此模式会真实查询 OA，但不会授予权限。完成联调后改为：

```yaml
oa:
  mode: execute
```

工作台会从 `OA_ACCESS_TOKEN` 环境变量读取现有 OA 管理员的 SSO Access Token，并作为 `mh_sso2_access_token` Cookie 调用当前 OA；Token 不写入 YAML 或数据库。

## 运行与打包

```powershell
cd workbench
F:\GOPATH\bin\wails.exe dev
F:\GOPATH\bin\wails.exe build
```

产物位于 `build/bin/workbench.exe`。MVP 初始本地入口账号为 `admin / admin123`；后续可替换为 OA 身份适配器。
