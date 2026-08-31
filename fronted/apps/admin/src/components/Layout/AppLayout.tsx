import type { MenuItem } from "@mh-repo/types";
import {
  MhDropdown,
  MhLayout,
  MhLayoutHeader,
  MhLayoutSider,
  type MhLayoutSiderMenuItem,
  MhMessage
} from "@mh-repo/ui";
import { PoweroffOutlined } from "@mh-repo/ui/components/General/Icon";
import { buildMenuNodesFromFlatRoutes, mergeMenuNodesWithServerOverlay } from "@mh-repo/utils";
import type React from "react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Outlet, useLocation, useNavigate } from "react-router";
import { qiankunWindow } from "vite-plugin-qiankun/dist/helper";
import logo from "../../assets/logo.svg";
import { useNProgress } from "../../hooks/useNProgress";
import { dspRouteManifest } from "../../router/routeManifest";
import { fetchMenuConfig, loadMenuConfigFromCache } from "../../services/menuConfigService";
import ssoRequest from "../../utils/ssoRequest";
import { renderSvgIcon } from "../SvgComp";
import styles from "./index.module.less";

const { Content } = MhLayout;

const applyIconComponents = (items: MenuItem[]): MenuItem[] => {
  return items.map(item => ({
    ...item,
    icon: typeof item.icon === "string" ? renderSvgIcon(item.icon) : item.icon,
    children: item.children ? applyIconComponents(item.children) : undefined
  }));
};

const defaultMenuItems: MenuItem[] = applyIconComponents(buildMenuNodesFromFlatRoutes(dspRouteManifest) as MenuItem[]);
const routeManifestMap = new Map(dspRouteManifest.map(item => [item.key, item]));

const toSiderMenuItems = (items: MenuItem[]): MhLayoutSiderMenuItem[] =>
  items.map(item => ({
    key: item.key,
    title: item.title,
    icon: item.icon,
    path: item.path,
    openInNewTab: routeManifestMap.get(item.key)?.openInNewTab,
    children: item.children ? toSiderMenuItems(item.children) : undefined
  }));

const isSameMenu = (a: MenuItem[], b: MenuItem[]): boolean => {
  if (a.length !== b.length) return false;

  for (let i = 0; i < a.length; i += 1) {
    const itemA = a[i];
    const itemB = b[i];

    if (itemA.key !== itemB.key || itemA.title !== itemB.title || itemA.path !== itemB.path) {
      return false;
    }

    const childrenA = itemA.children || [];
    const childrenB = itemB.children || [];

    if (!isSameMenu(childrenA, childrenB)) {
      return false;
    }
  }

  return true;
};

const AppLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();

  useNProgress();

  const isInQiankun = qiankunWindow.__POWERED_BY_QIANKUN__;

  const [menuItems, setMenuItems] = useState<MenuItem[]>(() => {
    const cachedMenu = loadMenuConfigFromCache("dsp");
    return cachedMenu && cachedMenu.length > 0
      ? mergeMenuNodesWithServerOverlay(defaultMenuItems, cachedMenu)
      : defaultMenuItems;
  });
  const [user, setUser] = useState<any>(null);

  useEffect(() => {
    const loadMenuConfig = async () => {
      const apiMenu = await fetchMenuConfig("dsp");

      if (apiMenu && apiMenu.length > 0) {
        const mergedMenu = mergeMenuNodesWithServerOverlay(defaultMenuItems, apiMenu);
        setMenuItems(prev => (isSameMenu(prev, mergedMenu) ? prev : mergedMenu));
        return;
      }

      const cachedMenu = loadMenuConfigFromCache("dsp");
      if (cachedMenu && cachedMenu.length > 0) {
        const mergedMenu = mergeMenuNodesWithServerOverlay(defaultMenuItems, cachedMenu);
        setMenuItems(prev => (isSameMenu(prev, mergedMenu) ? prev : mergedMenu));
        return;
      }

      setMenuItems(prev => (isSameMenu(prev, defaultMenuItems) ? prev : defaultMenuItems));
    };

    loadMenuConfig();

    const handleConfigUpdate = (event: Event) => {
      const detail = (event as CustomEvent).detail;
      if (detail?.platform === "dsp" || !detail?.platform) {
        loadMenuConfig();
      }
    };

    window.addEventListener("menu-config-updated", handleConfigUpdate as EventListener);
    window.addEventListener("storage", loadMenuConfig);

    return () => {
      window.removeEventListener("menu-config-updated", handleConfigUpdate as EventListener);
      window.removeEventListener("storage", loadMenuConfig);
    };
  }, []);

  const siderMenuItems = useMemo(() => toSiderMenuItems(menuItems), [menuItems]);
  const handleNavigate = useCallback((path: string) => navigate(path), [navigate]);
  const items = [
    {
      key: "logout",
      icon: <PoweroffOutlined />,
      label: "退出系统"
    }
  ];

  // ==== 定义退出登录的处理函数 ====
  const handleLogout = async () => {
    try {
      // 调用后端的 SSO 登出接口
      // 后端需要负责：清除本地 Session，并请求 SSO 认证中心销毁全局会话
      await ssoRequest.post("/api/v1/auth/logout", {});

      MhMessage.success("已安全退出系统");
    } catch (error: any) {
      // 即使接口报错，也建议强制清除本地状态并跳转，防止用户卡死
      console.error("SSO 登出异常:", error);
    } finally {
      // 清除本地残留的 Token（防御性编程）
      localStorage.removeItem("token");

      // 清除 Cookie 中的 SSO 访问令牌
      // 注意：如果 token 设置了 HttpOnly 属性，前端 JavaScript 将无法通过 document.cookie 访问或删除它。
      // 在这种情况下，必须由后端在 /v1/auth/logout 接口中通过设置 Cookie 的 Max-Age 为 0 来清除。
      document.cookie = "mh_sso2_access_token=; Max-Age=0; path=/;";

      // 跳转到 SSO 登录页
      // 使用从网页解析中获取的完整登录 URL
      window.location.href = "https://sso2.maplehaze.cn";
    }
  };

  // ==== 获取用户信息 ====
  const userInfo = async () => {
    try {
      // ssoRequest 响应拦截器已返回 response.data.data，res 本身就是 data 对象
      // API 结构: { code: 200, data: { user: { name: "..." } }, msg: "success" }
      const res: any = await ssoRequest.get("/api/v1/auth/me");
      // MhMessage.success("获取用户信息成功");
      setUser(res?.user?.name ?? null);
    } catch {
      MhMessage.error("获取用户信息失败");
    }
  };

  useEffect(() => {
    userInfo();
  }, []);

  // 绑定到菜单点击事件
  const handleMenuClick = (e: { key: string }) => {
    if (e.key === "logout") {
      handleLogout();
    }
  };

  return (
    <MhLayout className={`${styles["dsp-app-layout"]} ${isInQiankun ? styles["dsp-app-layout--qiankun"] : ""}`}>
      {!isInQiankun && (
        <MhLayoutHeader
          logoUrl={logo}
          brandTitle="Maplehaze OA"
          userRole={<span style={{ cursor: "pointer" }}>管理员</span>}
          userGreeting={
            <MhDropdown menu={{ items, onClick: handleMenuClick }} placement="bottomLeft" trigger={["hover"]}>
              <span style={{ cursor: "pointer", display: "inline-flex", alignItems: "center" }}>
                {`${user}，您好~`}
              </span>
            </MhDropdown>
          }
        />
      )}

      <MhLayout className={styles["dsp-layout-shell"]}>
        <MhLayoutSider
          menuItems={siderMenuItems}
          pathname={location.pathname}
          onNavigate={handleNavigate}
          appBasePath="/admin"
        />

        <Content className={`${styles["dsp-content"]} main-app-content`}>
          <Outlet />
        </Content>
      </MhLayout>
    </MhLayout>
  );
};

export default AppLayout;
