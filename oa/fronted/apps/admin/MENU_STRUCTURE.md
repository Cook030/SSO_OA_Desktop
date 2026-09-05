# DSP 菜单结构说明

## 菜单层级

DSP 应用采用多级菜单结构，提供清晰的功能导航。

### 一级菜单

1. **📊 数据概览** - 首页仪表盘
2. **📢 广告活动** - 广告活动管理（含子菜单）
3. **📑 报表中心** - 数据报表（含子菜单）
4. **⚙️ 系统设置** - 系统配置（含子菜单）

### 完整菜单树

```
📊 数据概览 (/)
📢 广告活动
  ├─ 📋 活动列表 (/campaigns)
  ├─ ➕ 创建活动 (/campaigns/create)
  └─ 📈 数据分析
      ├─ 📊 分析概览 (/campaigns/analytics)
      └─ 📄 详细分析 (/campaigns/analytics/detail)
📑 报表中心
  ├─ 📊 概览报表 (/reports/overview)
  └─ 📄 详细报表 (/reports/detail)
⚙️ 系统设置
  ├─ 👤 账户设置 (/settings/account)
  └─ 🔔 通知设置 (/settings/notification)
```

## 路由配置

| 路径                          | 组件                    | 说明                 |
| ----------------------------- | ----------------------- | -------------------- |
| `/`                           | Dashboard               | 数据概览首页         |
| `/campaigns`                  | CampaignList            | 广告活动列表         |
| `/campaigns/create`           | CampaignCreate          | 创建广告活动         |
| `/campaigns/analytics`        | CampaignAnalytics       | 数据分析概览         |
| `/campaigns/analytics/detail` | CampaignAnalyticsDetail | 详细分析             |
| `/campaigns/:id`              | CampaignDetail          | 活动详情（动态路由） |
| `/reports/overview`           | ReportOverview          | 概览报表             |
| `/reports/detail`             | ReportDetail            | 详细报表             |
| `/settings/account`           | SettingsAccount         | 账户设置             |
| `/settings/notification`      | SettingsNotification    | 通知设置             |

## 菜单配置

菜单配置位于 `apps/dsp/src/components/Layout/AppLayout.tsx` 中：

```typescript
const menuItems: MenuItem[] = [
  {
    key: "/",
    label: "数据概览",
    icon: "📊",
    path: "/"
  },
  {
    key: "/campaigns",
    label: "广告活动",
    icon: "📢",
    path: "/campaigns",
    children: [
      {
        key: "/campaigns/list",
        label: "活动列表",
        icon: "📋",
        path: "/campaigns"
      },
      {
        key: "/campaigns/create",
        label: "创建活动",
        icon: "➕",
        path: "/campaigns/create"
      },
      {
        key: "/campaigns/analytics",
        label: "数据分析",
        icon: "📈",
        path: "/campaigns/analytics"
      }
    ]
  }
  // ... 其他菜单项
];
```

## 菜单特性

### 1. 多级展开

- 点击带有子菜单的项目会展开/收起子菜单
- 子菜单项缩进显示，层级清晰

### 2. 路由导航

- 点击菜单项自动跳转到对应路由
- 当前路由对应的菜单项高亮显示

### 3. 图标支持

- 每个菜单项都有对应的 emoji 图标
- 图标在菜单折叠时仍然可见

### 4. 折叠功能

- Sidebar 支持折叠/展开
- 折叠时只显示图标，节省空间
- 展开时显示完整的菜单文本

## 添加新菜单

### 步骤 1：创建页面组件

```bash
# 在 apps/dsp/src/pages/ 下创建新页面
mkdir apps/dsp/src/pages/NewPage
touch apps/dsp/src/pages/NewPage/index.tsx
```

### 步骤 2：添加路由

在 `apps/dsp/src/router/index.tsx` 中添加路由配置：

```typescript
import NewPage from "../pages/NewPage";

// 在 children 数组中添加
{
  path: "new-page",
  element: <NewPage />
}
```

### 步骤 3：添加菜单项

在 `apps/dsp/src/components/Layout/AppLayout.tsx` 的 `menuItems` 中添加：

```typescript
{
  key: "/new-page",
  label: "新页面",
  icon: "🆕",
  path: "/new-page"
}
```

或者作为子菜单添加：

```typescript
{
  key: "/parent",
  label: "父菜单",
  icon: "📁",
  children: [
    {
      key: "/parent/new-page",
      label: "新页面",
      icon: "🆕",
      path: "/parent/new-page"
    }
  ]
}
```

## 菜单项类型定义

```typescript
interface MenuItem {
  key: string; // 唯一标识
  label: string; // 显示文本
  icon?: string; // 图标（emoji 或图标组件）
  path?: string; // 路由路径
  children?: MenuItem[]; // 子菜单
}
```

## 注意事项

1. **key 唯一性**：每个菜单项的 key 必须唯一
2. **路径一致性**：菜单项的 path 应与路由配置的 path 一致
3. **子菜单限制**：建议最多使用 2 级菜单，避免层级过深
4. **图标选择**：使用清晰易懂的图标，保持视觉一致性
5. **命名规范**：使用简洁明了的菜单名称，避免过长

## 未来扩展

可以考虑的功能扩展：

- 菜单权限控制（根据用户角色显示不同菜单）
- 菜单搜索功能
- 菜单收藏/常用功能
- 动态菜单配置（从后端获取）
- 菜单徽章（显示未读消息数等）
