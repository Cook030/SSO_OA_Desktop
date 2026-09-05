import { buildConfiguredRouteTree, type ConfiguredRouteTreeNode } from "@mh-repo/utils";
import type React from "react";
import { lazy, Suspense } from "react";
import { createBrowserRouter, Navigate, type RouteObject } from "react-router";
import { dspRouteManifest } from "./routeManifest";

const App = lazy(() => import("../App"));
const Exception403 = lazy(() => import("../pages/Exception/403"));
const Exception404 = lazy(() => import("../pages/Exception/404"));
const Exception500 = lazy(() => import("../pages/Exception/500"));
const PermissionPlatform = lazy(() => import("../pages/permission/platforms"));
const PermissionEmployees = lazy(() => import("../pages/permission/employees"));
// const PermissionPassword = lazy(() => import("../pages/permission/password"));
const AddEmployee = lazy(() => import("../pages/permission/employees/AddEmployee"));
// const AddPlatform = lazy(() => import("../pages/permission/platforms/AddPlatform"));

const withSuspense = (element: React.ReactElement) => {
  return <Suspense fallback={<div className="p-6 text-sm text-gray-500">页面加载中...</div>}>{element}</Suspense>;
};

const routeElementMap: Record<string, React.ReactElement> = {
  "/": <Navigate to="/permission/platforms" replace />,
  "/permission::redirect": <Navigate to="/permission/platforms" replace />,
  "/exception/403": withSuspense(<Exception403 />),
  "/exception/404": withSuspense(<Exception404 />),
  "/exception/500": withSuspense(<Exception500 />),
  "/permission/platforms": withSuspense(<PermissionPlatform />),
  "/permission/employees": withSuspense(<PermissionEmployees />),
  // "/permission/password": withSuspense(<PermissionPassword />),
  "/permission/employees/AddEmployee": withSuspense(<AddEmployee />)
  // "/permission/platforms/AddPlatform": withSuspense(<AddPlatform />)
};

const buildRouteObjects = (routeTree: ConfiguredRouteTreeNode[]): RouteObject[] => {
  return routeTree.reduce<RouteObject[]>((acc, node) => {
    const children = node.children ? buildRouteObjects(node.children) : undefined;
    const element = routeElementMap[node.key];

    if (node.key.endsWith("::__group")) {
      if (children?.length) {
        acc.push({ children });
      }
      return acc;
    }

    if (node.index) {
      if (element) {
        acc.push({
          index: true,
          element
        });
      }
      return acc;
    }

    if (node.path || element || children?.length) {
      acc.push({
        ...(node.path ? { path: node.path } : {}),
        ...(element ? { element } : {}),
        ...(node.hideInMenu ? { hideInMenu: true } : {}),
        ...(children?.length ? { children } : {})
      });
    }

    return acc;
  }, []);
};

// 路由配置
export const createRouter = (basename: string = "/admin"): ReturnType<typeof createBrowserRouter> => {
  const routeDefinitions = buildRouteObjects(buildConfiguredRouteTree("dsp", dspRouteManifest));

  return createBrowserRouter(
    [
      {
        path: "/",
        element: withSuspense(<App />),
        children: [
          ...routeDefinitions,
          {
            path: "/permission/employees/AddEmployee",
            element: withSuspense(<AddEmployee />)
          },
          // {
          //   path: "/permission/platforms/AddPlatform",
          //   element: withSuspense(<AddPlatform />)
          // },
          { path: "*", element: withSuspense(<Exception404 />) }
        ]
      }
    ],
    {
      basename
    }
  );
};

export default createRouter;
