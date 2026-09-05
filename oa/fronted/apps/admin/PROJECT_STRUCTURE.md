# DSP 项目结构说明

## 目录结构

```
apps/dsp/src/
├── components/              # 业务组件
│   └── Layout/
│       └── AppLayout.tsx   # 主布局组件（包含 Header 和 Sidebar）
├── pages/                   # 页面组件
│   ├── Dashboard/          # 数据概览页
│   │   └── index.tsx
│   ├── CampaignList/       # 广告活动列表页
│   │   └── index.tsx
│   └── CampaignDetail/     # 广告活动详情页
│       └── index.tsx
├── App.tsx                  # 主应用组件
├── main.tsx                 # 应用入口
├── index.css                # 全局样式
└── App.css                  # 应用样式
```

## 路由配置

- `/` - 数据概览（Dashboard）
- `/campaigns` - 广告活动列表（CampaignList）
- `/campaigns/:id` - 广告活动详情（CampaignDetail）

## 布局说明

### AppLayout 组件

- **侧边栏（Sider）**：
  - 可折叠
  - 包含 Logo 和导航菜单
  - 固定定位，始终可见
- **顶部导航栏（Header）**：
  - 显示应用标题
  - 用户信息和退出按钮
  - 粘性定位，滚动时固定在顶部
- **内容区域（Content）**：
  - 使用 React Router 的 Outlet 渲染子路由
  - 灰色背景（#f0f2f5）

## 技术栈

- **React 18** + TypeScript
- **React Router v6** - 路由管理
- **TanStack Query** - 数据状态管理
- **Ant Design** - UI 组件库（通过 @mh-repo/ui）
- **Tailwind CSS** - 样式工具
- **qiankun** - 微前端框架

## 开发指南

### 启动开发服务器

```bash
cd apps/dsp
pnpm dev
```

### 访问地址

- 独立运行：http://local-nextui.maplehaze.cn:8002/dsp/
- 微前端模式：由主应用加载

## 下一步开发

1. 完善 Dashboard 页面（数据统计卡片、图表）
2. 实现 CampaignList 页面（列表、搜索、筛选、分页）
3. 实现 CampaignDetail 页面（详情信息、数据趋势）
4. 添加 API 服务层
5. 实现数据管理 Hooks
