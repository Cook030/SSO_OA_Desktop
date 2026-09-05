import { MhApp, MhConfigProvider, MhTheme } from "@mh-repo/ui";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import React from "react";
import type { Root } from "react-dom/client";
import ReactDOM from "react-dom/client";
import { RouterProvider } from "react-router";
import { qiankunWindow, renderWithQiankun } from "vite-plugin-qiankun/dist/helper";
import { createRouter } from "./router";
import "./index.css";
import "nprogress/nprogress.css";
import NProgress from "nprogress";

// import "./utils/request";

// 创建 QueryClient 实例
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30000, // 30 秒
      refetchOnWindowFocus: false
    }
  }
});

const BODY_CLASS_NAME = "mh-dsp-body";
let reactRoot: Root | null = null;

const syncBodyBackground = (enabled: boolean) => {
  document.body.classList.toggle(BODY_CLASS_NAME, enabled);
};

// qiankun 生命周期
const initQianKun = () => {
  renderWithQiankun({
    mount(props) {
      syncBodyBackground(false);
      render(props.container);
      // 监听主应用传值
      props.onGlobalStateChange?.((res: { count: string }) => {
        console.log("主应用状态变化:", res.count);
      });
    },
    update() {},
    bootstrap() {},
    unmount(props) {
      syncBodyBackground(false);
      reactRoot?.unmount();
      reactRoot = null;
      props.container && (props.container.innerHTML = "");
      // 清理 NProgress 状态
      if (typeof NProgress !== "undefined") {
        NProgress.done();
        NProgress.remove();
      }
    }
  });
};

const RouterShell: React.FC<{ basename: string }> = ({ basename }) => {
  const [router, setRouter] = React.useState(() => createRouter(basename));

  React.useEffect(() => {
    const rebuildRouter = () => {
      setRouter(createRouter(basename));
    };

    window.addEventListener("menu-config-updated", rebuildRouter as EventListener);
    window.addEventListener("storage", rebuildRouter);

    return () => {
      window.removeEventListener("menu-config-updated", rebuildRouter as EventListener);
      window.removeEventListener("storage", rebuildRouter);
    };
  }, [basename]);

  return <RouterProvider router={router} />;
};

// 渲染函数
const getMountElement = (container?: HTMLElement) => {
  if (!container) {
    const root = document.getElementById("root");
    if (!root) {
      throw new Error("Root element #root not found");
    }
    return root;
  }

  return (container.querySelector("#root") as HTMLElement | null) ?? container;
};

const render = (container?: HTMLElement) => {
  const appDom = getMountElement(container);
  reactRoot?.unmount();
  reactRoot = ReactDOM.createRoot(appDom);
  reactRoot.render(
    <React.StrictMode>
      <QueryClientProvider client={queryClient}>
        <MhConfigProvider theme={{ algorithm: MhTheme.defaultAlgorithm }}>
          <MhApp>
            <RouterShell basename="/admin" />
          </MhApp>
        </MhConfigProvider>
        {process.env.NODE_ENV === "development" && <ReactQueryDevtools initialIsOpen={false} />}
      </QueryClientProvider>
    </React.StrictMode>
  );
};

// 判断当前应用是否在主应用中
if (qiankunWindow.__POWERED_BY_QIANKUN__) {
  initQianKun();
} else {
  syncBodyBackground(true);
  // 确保 DOM 就绪后再挂载：构建产物中 import() 位于 <head>，
  // 模块缓存命中时会在 <body> 解析前执行，导致 #root 不存在
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", () => render());
  } else {
    render();
  }
}
