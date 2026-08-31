# mh_next_ui 架构说明

## 1. 项目定位

`mh_next_ui` 是基于 `pnpm workspace` 的前端 monorepo，目标是统一 ADX 多应用体系的工程和 UI 标准。

- 主应用：`container`
- 子应用：`dsp`、`ssp`、`rbac`
- 认证相关：`sso2`
- 公共包：`@mh-repo/ui`、`@mh-repo/utils`、`@mh-repo/apis`、`@mh-repo/types`

## 2. 目录结构

- `apps/*`：业务应用（每个 app 独立 `vite.config.ts` 和 `package.json`）
- `packages/ui`：`Mh*` 组件封装层（对 Ant Design 的统一包装）
- `packages/utils`：菜单树、格式化、基础工具函数
- `packages/apis`：接口能力封装
- `packages/types`：跨应用共享类型
- `docs`：架构、编码和 AI 协作规范

## 3. 技术栈

- React + TypeScript + Vite
- Ant Design（通过 `@mh-repo/ui` 封装）
- qiankun（微前端）
- TanStack Query（请求缓存）
- GraphQL codegen（类型与 SDK 生成）
- Less + Tailwind（并存，按现有文件约定使用）

## 4. 微前端运行模型

### container（主应用）

- 负责头部导航和子应用挂载点；
- 管理平台菜单配置（`platforms` + `allMenus`）；
- 通过路由切换子应用路径（如 `/dsp/*`、`/rbac/*`）。

### 子应用（dsp/ssp/rbac）

- 可独立运行，也可在 qiankun 中被挂载；
- 侧边栏菜单来源为“本地路由清单 + 服务端菜单配置合并”；
- 菜单、路由、权限（rbac）保持同源约束。

## 5. 菜单数据流（关键）

1. container 调用 `mhSso2_getMenuConfig` 获取 `platforms/allMenus`。
2. 结果缓存至 localStorage（`menu-config-cache`、`menu-config-{platform}`）。
3. 子应用读取平台菜单并叠加本地默认清单，保证可配置和可回退。
4. 头部平台弹层（PlatformMegaMenu）必须和侧边栏同源（同 `allMenus`）。

## 6. 依赖边界

- `apps/*` 可以依赖 `packages/*`。
- `packages/*` 不应依赖 `apps/*`（禁止反向耦合）。
- UI 通用能力放 `packages/ui`，业务逻辑留在 `apps/*`。

## 7. 常用命令

- 全量开发：`pnpm dev`
- 单应用开发：`pnpm --filter <app> dev`
- 全量构建：`pnpm build`
- 预览构建：`pnpm preview:build && pnpm preview:preview`
