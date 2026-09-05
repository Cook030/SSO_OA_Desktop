import React from "react";
import { createRoot } from "react-dom/client";
import {
  DashboardOutlined,
  FileTextOutlined,
  MhButton,
  type MhConBreadcrumbItem,
  MhConException,
  type MhConExceptionStatus,
  MhConfigProvider,
  MhConModal,
  MhConModalCloseButton,
  MhConPageHeader,
  MhLayout,
  MhLayoutHeader,
  MhLayoutSider,
  type MhLayoutSiderMenuItem,
  SettingOutlined
} from "../src";

const { Content } = MhLayout;

const sectionStyle: React.CSSProperties = {
  marginBottom: 48
};

const sectionTitleStyle: React.CSSProperties = {
  fontSize: 18,
  fontWeight: 600,
  marginBottom: 16,
  paddingBottom: 8,
  borderBottom: "1px solid #eee"
};

const demoMenuItems: MhLayoutSiderMenuItem[] = [
  {
    key: "dashboard",
    title: "工作台",
    icon: <DashboardOutlined />,
    path: "/demo/dashboard"
  },
  {
    key: "settings",
    title: "系统设置",
    icon: <SettingOutlined />,
    children: [
      {
        key: "settings-user",
        title: "用户管理",
        path: "/demo/settings/user"
      },
      {
        key: "settings-role",
        title: "角色管理",
        path: "/demo/settings/role"
      }
    ]
  },
  {
    key: "table",
    title: "数据表格",
    icon: <FileTextOutlined />,
    path: "/demo/table"
  }
];

const demoBreadcrumbItems: MhConBreadcrumbItem[] = [
  { key: "home", title: "首页", path: "/", clickable: true },
  { key: "demo", title: "Customize 组件", path: "/demo", clickable: true },
  { key: "current", title: "开发预览" }
];

const exceptionStatuses: MhConExceptionStatus[] = ["403", "404", "500"];

const App: React.FC = () => {
  const [pathname, setPathname] = React.useState("/demo/dashboard");
  const [modalOpen, setModalOpen] = React.useState(false);
  const [exceptionStatus, setExceptionStatus] = React.useState<MhConExceptionStatus>("404");

  const handleNavigate = (path: string) => {
    setPathname(path);
  };

  const handleBreadcrumbClick = (item: MhConBreadcrumbItem) => {
    if (item.path) {
      setPathname(item.path);
    }
  };

  return (
    <MhConfigProvider>
      <div style={{ padding: "24px 0" }}>
        <h1 style={{ margin: "0 0 32px" }}>枫岚 UI — Customize 组件开发预览</h1>

        <section style={sectionStyle}>
          <h2 style={sectionTitleStyle}>Layout · MhLayoutHeader / MhLayoutSider</h2>
          <div style={{ border: "1px solid #eee", borderRadius: 8, overflow: "hidden", background: "#fff" }}>
            <MhLayout style={{ minHeight: 400 }}>
              <MhLayoutHeader brandTitle="Maplehaze" platformLabel="UI Dev" userGreeting="您好" userRole="开发者" />
              <MhLayout>
                <MhLayoutSider
                  menuItems={demoMenuItems}
                  pathname={pathname}
                  onNavigate={handleNavigate}
                  appBasePath="/demo"
                />
                <Content style={{ padding: 24, background: "#f5f7fa" }}>
                  <p>
                    当前路径：<code>{pathname}</code>
                  </p>
                  <p>点击左侧菜单切换高亮与展开状态。</p>
                </Content>
              </MhLayout>
            </MhLayout>
          </div>
        </section>

        <section style={sectionStyle}>
          <h2 style={sectionTitleStyle}>Content · MhConPageHeader</h2>
          <div style={{ background: "#fff", borderRadius: 8, overflow: "hidden" }}>
            <MhConPageHeader
              title="页面标题示例"
              breadcrumbItems={demoBreadcrumbItems}
              onBreadcrumbClick={handleBreadcrumbClick}
              currentPath="/demo/preview"
            />
          </div>
        </section>

        <section style={sectionStyle}>
          <h2 style={sectionTitleStyle}>Content · MhConModal</h2>
          <MhButton type="primary" onClick={() => setModalOpen(true)}>
            打开定制弹窗
          </MhButton>
          <MhConModal
            open={modalOpen}
            title="定制弹窗标题"
            headerExtra={<MhConModalCloseButton onClick={() => setModalOpen(false)} />}
            footer={
              <>
                <MhButton onClick={() => setModalOpen(false)}>取消</MhButton>
                <MhButton type="primary" onClick={() => setModalOpen(false)}>
                  确定
                </MhButton>
              </>
            }
            onCancel={() => setModalOpen(false)}
          >
            <p>这是 MhConModal 的内容区域，支持自定义 header、footer 和关闭按钮。</p>
          </MhConModal>
        </section>

        <section style={sectionStyle}>
          <h2 style={sectionTitleStyle}>Content · MhConException</h2>
          <div style={{ marginBottom: 16, display: "flex", gap: 8 }}>
            {exceptionStatuses.map(status => (
              <MhButton
                key={status}
                type={exceptionStatus === status ? "primary" : "default"}
                onClick={() => setExceptionStatus(status)}
              >
                {status}
              </MhButton>
            ))}
          </div>
          <div style={{ background: "#fff", borderRadius: 8, overflow: "hidden" }}>
            <MhConException
              status={exceptionStatus}
              breadcrumbItems={demoBreadcrumbItems}
              onBreadcrumbClick={handleBreadcrumbClick}
              onBackHome={() => setPathname("/demo/dashboard")}
            />
          </div>
        </section>
      </div>
    </MhConfigProvider>
  );
};

const container = document.getElementById("root");
if (container) {
  const root = createRoot(container);
  root.render(<App />);
}
