# mh_next_ui 项目分析与技能包

## 0. 项目身份

**全栈偏前端架构师（20年经验）**

他是一位有 20 年实战经验的前端技术专家，长期负责大型 Web 系统的架构设计与工程落地，精通 React/Vue/TypeScript、工程化体系、性能优化与可维护性建设。  
同时具备扎实的 Go 后端能力，熟悉 Go 微服务架构、数据库设计、缓存与消息队列、API 设计与服务治理，能从前后端协同视角推动系统稳定、高效交付。  
技术风格务实，强调“业务可落地 + 架构可演进 + 代码可维护”，擅长复杂项目拆解、技术选型、规范建设与团队提效。

## 1. 项目分析（简版）

### 1.1 项目定位

- 这是一个广告中台前端 `monorepo`，核心形态是：`Container 主应用 + DSP/SSP 微前端子应用 + 共享基础包`。
- 主应用负责平台切换、菜单权限管理、微前端挂载；子应用负责各自业务页面与路由。

### 1.2 技术栈

- 工程：`pnpm workspace` + `turbo`
- 前端：`React` + `TypeScript` + `Vite` + `React Router`
- 微前端：`qiankun` + `vite-plugin-qiankun`
- 数据：`GraphQL` + `graphql-request` + `graphql-codegen` + `@tanstack/react-query`
- UI：`@mh-repo/ui`（对 antd 的封装组件库）
- 规范：`eslint` + `prettier` + `husky` + `lint-staged` + `changesets`

### 1.3 目录职责

- `apps/container`：主应用，负责微前端注册、平台菜单、菜单配置弹窗。
- `apps/dsp`：DSP 子应用，支持独立运行和被容器挂载。
- `apps/ssp`：SSP 子应用，支持独立运行和被容器挂载。
- `packages/ui`：组件库与测试（vitest/jest/playwright）。
- `packages/types`：跨应用共享类型。
- `packages/apis`：通用 API client（当前更偏模板层能力）。

### 1.4 当前核心能力

- 微前端装载与路由切换（`/dsp`、`/ssp`）
- 菜单配置中心（读取/编辑/保存）
- 菜单配置多层同步（GraphQL、localStorage、自定义事件、跨标签 storage 事件）
- 子应用布局双模式适配（独立模式 / qiankun 模式）

---

## 2. 项目技能包（可直接复用）

以下“技能包”可作为 AI 或团队成员的标准执行单元，覆盖从本地开发、联调、发布到排障的完整链路。

## 技能包 A：本地开发环境与多应用联调

### 适用场景

- 新同学接手项目
- 本地同时联调 `container/dsp/ssp`

### 标准步骤

1. 安装依赖并校验 Node/Pnpm 版本（Node `>=20`，pnpm `>=10`）。
2. 单应用开发优先使用 `pnpm dev:container`、`pnpm dev:dsp`、`pnpm dev:ssp`。
3. 跨应用联调使用 `pnpm dev:qiankun`，观察路由切换与挂载行为。
4. 开发完成后至少执行一次 `pnpm turbo:build` 验证 monorepo 构建完整性。

### 常用命令

```bash
pnpm install --frozen-lockfile
pnpm dev:container
pnpm dev:dsp
pnpm dev:ssp
pnpm dev:qiankun
pnpm turbo:build
```

### 验收标准

- 三个应用可独立启动。
- 容器模式下子应用可正确挂载和卸载。
- 全仓构建通过，无类型和打包错误。

---

## 技能包 B：微前端接入与路由治理

### 适用场景

- 新增子应用
- 调整子应用入口、激活路由或部署地址

### 标准步骤

1. 在 `apps/container/src/qiankun/index.js` 注册应用（`name/entry/container/activeRule`）。
2. 在 `apps/container/src/main.tsx` 确保 `start()` 只启动一次。
3. 子应用 `main.tsx` 保留 `renderWithQiankun` 生命周期与独立运行分支。
4. 校验子应用路由 `basename` 与容器前缀一致（如 `/dsp`、`/ssp`）。
5. 回归浏览器前进/后退、直达路由、刷新等场景。

### 验收标准

- 子应用独立访问正常。
- 容器跳转无白屏、无重复挂载、无路由错位。
- 多次切换应用后内存与事件监听无明显泄漏。

---

## 技能包 C：菜单权限配置中心

### 适用场景

- 菜单结构调整、平台增删改
- 菜单权限配置实时生效与跨页面同步

### 标准步骤

1. 使用 `mhSso2_getMenuConfig` 拉取菜单配置。
2. 在 `MenuConfig` 完成平台与菜单编辑。
3. 使用 `mhSso2_saveMenuConfig` 持久化。
4. 同步更新 localStorage（`menu-config-{platform}`）。
5. 触发统一事件 `menu-config-updated`，驱动 container/dsp/ssp 同步刷新。
6. 子应用读取策略采用“API 优先，缓存兜底”。

### 验收标准

- 保存后 container 顶部平台和子应用侧边栏即时更新。
- 同页更新依赖自定义事件，新标签页依赖 `storage` 事件，均可生效。
- API 异常时仍可使用缓存降级展示。

### 排查重点

- localStorage key 是否正确写入。
- 事件名与监听方是否一致。
- 平台过滤逻辑是否正确（`dsp` / `ssp`）。

---

## 技能包 D：GraphQL 文档与 SDK 生成闭环

### 适用场景

- Schema 变更
- 新增查询/变更字段后同步类型和请求 SDK

### 标准步骤

1. 在目标应用执行 GraphQL 文档生成。
2. 运行 codegen 生成 `generated` 代码与 TS 类型。
3. 修复调用处编译错误与字段映射问题。
4. 回归关键页面与菜单配置链路。

### 常用命令

```bash
pnpm --filter container graphql:rebuild
pnpm --filter dsp graphql:rebuild
pnpm --filter ssp graphql:rebuild
```

### 验收标准

- `generated` 可编译且无 lint 错误。
- 请求参数、返回类型与页面消费一致。
- 不手改 `generated` 目录文件。

---

## 技能包 E：子应用布局双模式适配

### 适用场景

- 统一 DSP/SSP 在独立模式与容器模式下的 UI 行为
- 修复 header、sider、内容区偏移错位

### 标准步骤

1. 基于 `qiankunWindow.__POWERED_BY_QIANKUN__` 分支处理布局。
2. 独立模式显示完整框架；容器模式隐藏重复 header/logo。
3. 按模式调整 `Sider` 固定策略与内容区 margin。
4. 验证折叠、滚动、路由切页时的视觉一致性。

### 验收标准

- 容器模式不遮挡主应用 header。
- 独立模式布局完整，折叠后偏移正确。
- NProgress 与页面切换一致，不闪烁不残留。

---

## 技能包 F：共享包开发（UI/Types/APIs/Utils）

### 适用场景

- 抽象跨应用能力
- 更新 `@mh-repo/ui` 组件或共享类型

### 标准步骤

1. 在对应 `packages/*/src` 完成改造，保证导出入口完整。
2. `@mh-repo/ui` 组件补充 vitest 测试，必要时增加 e2e。
3. 执行构建与消费端回归，确保 workspace 引用可用。
4. 检查是否引入 breaking change（props/type/样式行为）。

### 常用命令

```bash
pnpm --filter @mh-repo/ui test
pnpm --filter @mh-repo/ui e2e:vitest
pnpm --filter @mh-repo/ui build
```

### 验收标准

- 构建产物与类型声明可用。
- 业务应用无编译错误、无运行时回归。
- 组件在 container/dsp/ssp 表现一致。

---

## 技能包 G：质量门禁与发布准备

### 适用场景

- 提交前自检
- 发布前全链路检查

### 标准步骤

1. 执行格式化与静态检查（`prettier/eslint/cspell`）。
2. 运行关键构建与测试。
3. 检查 Docker/nginx 配置是否与应用路由一致。
4. 使用 changesets 或既有发布流程生成版本与发布记录。

### 常用命令

```bash
pnpm lint:prettier
pnpm lint:eslint
pnpm lint:spellcheck
pnpm turbo:build
pnpm release
```

### 验收标准

- 代码规范检查通过。
- 构建与主要测试通过。
- 发布产物可部署且路由、静态资源路径正确。

---

## 技能包 H：线上问题排查与应急修复

### 适用场景

- 微前端白屏、菜单不同步、GraphQL 请求异常
- 环境配置变更引发的回归

### 标准步骤

1. 先判断问题层级：容器、子应用、共享包、API/Schema。
2. 复现并抓取关键信号：路由、网络请求、localStorage、事件触发、控制台报错。
3. 优先做“最小修复”并验证主链路恢复。
4. 补充回归点和文档，避免同类问题重复出现。

### 排查速查表

- 挂载失败：检查 `entry/activeRule/container` 与子应用生命周期。
- 菜单异常：检查 `mhSso2_getMenuConfig`、`menu-config-updated`、缓存 key。
- 类型错误：执行 `graphql:rebuild`，确认是否 schema 变化未同步。
- 样式错位：校验独立/嵌入双模式的布局分支。

---

## 3. 团队协作建议（落地版）

- 新需求先映射到技能包，再开始开发，减少“边改边猜”。
- 菜单与权限相关改动必须同时验证 API、缓存、事件三层同步。
- GraphQL 变更必须走 `graphql:rebuild`，禁止直接手改 generated。
- PR 合并前至少完成一次 `dev:qiankun` 联调和一次 `turbo:build`。
- 共享包改动默认视为“跨应用变更”，至少回归 container + 一个子应用。
