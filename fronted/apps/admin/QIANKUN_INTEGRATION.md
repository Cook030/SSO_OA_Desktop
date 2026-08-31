# Qiankun 微前端集成说明

## 概述

DSP 应用支持两种运行模式：

1. **独立运行模式**：Header 占据整行，Sidebar 在 Header 下方
2. **微前端模式**：作为 qiankun 子应用运行，隐藏 Header 和 Logo，保留 Sidebar 菜单

## 实现原理

### 环境检测

通过 `qiankunWindow.__POWERED_BY_QIANKUN__` 判断当前运行环境：

```typescript
import { qiankunWindow } from "vite-plugin-qiankun/dist/helper";

const isInQiankun = qiankunWindow.__POWERED_BY_QIANKUN__;
```

### 布局适配

在 `AppLayout.tsx` 中根据运行环境调整布局：

```typescript
<MhLayout>
  {/* Header 仅在独立模式下显示，占据整行 */}
  {!isInQiankun && (
    <Header style={{ width: '100%', zIndex: 10 }}>
      DSP 需求方平台
    </Header>
  )}

  {/* Sidebar + Content 布局 */}
  <MhLayout>
    <Sider
      style={{
        // 独立模式：固定定位在 Header 下方（top: 64px）
        // 微前端模式：相对定位
        position: isInQiankun ? 'relative' : 'fixed',
        top: isInQiankun ? 0 : 64,
        height: isInQiankun ? '100vh' : 'calc(100vh - 64px)',
      }}
    >
      {/* Logo 仅在独立模式下显示 */}
      {!isInQiankun && <div>DSP 广告管理</div>}
      <Menu />
    </Sider>

    <Content
      style={{
        // 独立模式：添加左边距
        // 微前端模式：无左边距
        marginLeft: isInQiankun ? 0 : (collapsed ? 80 : 240),
      }}
    >
      <Outlet />
    </Content>
  </MhLayout>
</MhLayout>
```

## 运行模式对比

### 独立运行模式

- **访问地址**：http://local-nextui.maplehaze.cn:8002/dsp/
- **布局**：Header 占据整行 + Sidebar + Content
- **导航**：使用自己的侧边栏菜单
- **用户信息**：显示在自己的 Header 中
- **定位方式**：Header 粘性定位在顶部，Sidebar 固定定位在 Header 下方

```
┌─────────────────────────────────────┐
│  DSP 需求方平台          欢迎回来    │ ← Header（占据整行）
├─────────┬───────────────────────────┤
│  Logo   │                           │
│  DSP    │                           │
├─────────┤                           │
│ 📊 概览 │      页面内容区域          │
│ 📢 活动 │                           │
│         │                           │
└─────────┴───────────────────────────┘
   ↑ Sidebar（在 Header 下方）
```

### 微前端模式

- **访问地址**：由主应用决定
- **布局**：Sidebar + Content（无 Header，无 Logo）
- **导航**：使用 DSP 自己的侧边栏菜单
- **用户信息**：使用主应用的 Header
- **定位方式**：Sidebar 使用相对定位，不会挡住主应用 Header

```
┌─────────────────────────────────────┐
│      主应用的 Header                 │ ← 主应用控制（不被遮挡）
├─────────┬───────────────────────────┤
│ 📊 概览 │                           │ ← 无 Logo
│ 📢 活动 │      DSP 页面内容区域      │
│         │                           │
│         │                           │
└─────────┴───────────────────────────┘
   ↑ Sidebar（相对定位）
```

## 设计优势

### 为什么 Header 占据整行？

1. **视觉统一**：Header 横跨整个页面，提供统一的顶部导航体验
2. **空间利用**：充分利用顶部空间展示应用信息和用户操作
3. **层级清晰**：Header 在最上层，Sidebar 和 Content 在下层，层级关系明确
4. **符合习惯**：符合常见的后台管理系统布局习惯

### 为什么保留 Sidebar？

1. **独立导航**：DSP 子应用有自己的页面导航逻辑
2. **用户体验**：用户可以在 DSP 内部快速切换页面
3. **功能完整性**：保持子应用的功能独立性
4. **视觉一致性**：Sidebar 提供清晰的功能区分

### 为什么隐藏 Header 和 Logo？

1. **避免重复**：主应用已有 Header，不需要重复显示
2. **空间利用**：节省垂直空间，提供更多内容展示区域
3. **统一体验**：用户信息、退出等操作统一在主应用 Header 中
4. **不遮挡主应用**：使用相对定位，Logo 隐藏，确保不挡住主应用 Header

### 定位方式的差异

| 模式       | Header 显示 | Sidebar 定位         | Sidebar Top | 主内容区左边距      | Logo 显示 |
| ---------- | ----------- | -------------------- | ----------- | ------------------- | --------- |
| 独立模式   | ✅ 整行显示 | `position: fixed`    | `64px`      | `marginLeft: 240px` | ✅ 显示   |
| 微前端模式 | ❌ 隐藏     | `position: relative` | `0`         | `marginLeft: 0`     | ❌ 隐藏   |

**为什么这样设计？**

- **独立模式**：
  - Header 占据整行，粘性定位在顶部
  - Sidebar 固定定位在 Header 下方（top: 64px）
  - 主内容区需要添加左边距避免被 Sidebar 遮挡
- **微前端模式**：
  - 无 Header，使用主应用的 Header
  - Sidebar 相对定位，自然占位，不会覆盖主应用内容
  - 主内容区无需额外边距

## 生命周期

在 `main.tsx` 中配置了 qiankun 生命周期钩子：

```typescript
const initQianKun = () => {
  renderWithQiankun({
    mount(props) {
      render(props.container);
      // 监听主应用状态变化
      props.onGlobalStateChange?.(res => {
        console.log("主应用状态变化:", res);
      });
    },
    update() {},
    bootstrap() {},
    unmount() {}
  });
};
```

## 路由配置

路由 basename 设置为 `/dsp`，确保在两种模式下都能正确工作：

```typescript
// router/index.tsx
export const createRouter = (basename: string = "/dsp") => {
  return createBrowserRouter([...], { basename });
};
```

## 样式隔离

- 使用 Ant Design 的 ConfigProvider 确保样式一致性
- Sidebar 使用浅色主题（theme="light"）
- 内容区域保持灰色背景（#f0f2f5）与主应用协调

## 测试建议

### 独立模式测试

```bash
cd apps/dsp
pnpm dev
# 访问 http://local-nextui.maplehaze.cn:8002/dsp/
# 应该看到：
# - Header 占据整行，粘性定位在顶部
# - Sidebar 在 Header 下方（带 Logo）
# - Content 在 Sidebar 右侧
```

### 微前端模式测试

需要在主应用中配置并加载 DSP 子应用。

- 应该看到：
  - 无 Header（使用主应用 Header）
  - 相对定位的 Sidebar（无 Logo）
  - Content 在 Sidebar 右侧
- 主应用 Header 不应被遮挡

## 注意事项

1. **路由同步**：确保主应用和子应用的路由能够正确同步
2. **状态共享**：通过 qiankun 的 `onGlobalStateChange` 监听主应用状态
3. **样式冲突**：避免使用全局样式，使用 CSS Modules 或 Tailwind
4. **资源加载**：确保静态资源路径正确配置
5. **Sidebar 宽度**：固定宽度 240px（折叠时 80px），确保与主应用布局协调
6. **定位方式**：微前端模式下使用相对定位，避免遮挡主应用内容
7. **Logo 隐藏**：微前端模式下隐藏 Logo 区域，节省空间并避免视觉冲突
8. **Header 高度**：Header 高度为 64px，Sidebar 的 top 值需要对应调整
9. **内容区高度**：独立模式下内容区高度为 `calc(100vh - 64px)`，扣除 Header 高度
