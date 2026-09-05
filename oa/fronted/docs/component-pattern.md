# mh_next_ui 组件模式

## 1. 分层模型

1. `packages/ui`：通用 UI 封装层（`Mh*` 组件）。
2. `apps/*/components`：业务组件层（面向某个 app）。
3. `apps/*/pages`：路由页面层（组合业务组件）。

原则：能复用的 UI 能力优先下沉到 `packages/ui` 或 `packages/utils`。

## 2. `Mh*` 封装模式

`packages/ui` 组件遵循统一模式：

- `MhXxxProps extends AntdXxxProps`
- 透传原有 props
- 保持 `Mh` 前缀导出

收益：

- 统一主题和行为入口；
- 避免各 app 直接依赖细节实现。

## 3. 布局组件模式

- 使用 `MhLayout` + `MhLayoutHeader` + `MhLayoutSider` 组合页面骨架。
- Header 负责平台级导航与用户信息；
- Sider 负责菜单展示与路由联动；
- 业务内容通过 `Content` 或页面组件承载。

## 4. 菜单组件模式

### 侧边栏

- 输入使用树形菜单节点；
- icon 支持字符串映射到组件；
- 跳转统一由 `onNavigate` 驱动。

### 头部平台弹层

- 数据源必须与侧边栏同源（`allMenus`）；
- 平台无菜单时不展示弹层；
- 仅负责展示与导航，不承载业务状态。

## 5. 文件组织约定

- 组件目录：`ComponentName/index.tsx` + `index.module.less`
- 样式与组件同目录，便于维护
- 复杂转换逻辑提取为纯函数，便于复用和测试

## 6. 图标与菜单一致性

- 各应用统一使用 `apps/<app>/src/components/SvgComp`（逻辑由 `packages/ui/src/components/General/SvgComp` 的 `createSvgComp` 提供，经 `@mh-repo/ui` 导出）：
  - Ant Design 图标：`name="SearchOutlined"` 等 Outlined 名称；
  - MapleHaze 自有图标：`name="mh-home"` 等 `mh-` 前缀（资源 SVG 或内联图形）。
- 已接入：`dsp`、`container`、`rbac`、`ssp`、`sso2`；各 app 在 `iconRegistry.ts` 维护本应用图标表。
- 菜单配置中的字符串 `icon` 通过 `renderSvgIcon(name)` 解析，与侧边栏同一映射。
- 分组图标需复用侧边栏同一映射逻辑；
- 不允许头部菜单与侧边栏出现不同图标语义。
