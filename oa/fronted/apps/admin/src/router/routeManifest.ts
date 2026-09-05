import type { FlatRouteMenuDefinition } from "@mh-repo/utils";

export const dspRouteManifest: FlatRouteMenuDefinition[] = [
  {
    key: "/",
    title: "首页",
    path: "/",
    index: true,
    hideInMenu: true
  },
  {
    key: "/permission",
    title: "权限管理",
    icon: "FolderOpenOutlined",
    group: true
  },
  {
    // 访问 /permission 时重定向到默认子页 /permission/platforms，不在菜单中展示
    key: "/permission::redirect",
    path: "/permission",
    routePath: "permission",
    hideInMenu: true
  },
  {
    key: "/permission/platforms",
    title: "平台管理",
    path: "/permission/platforms",
    routePath: "/permission/platforms",
    parentKey: "/permission"
  },
  // {
  //   key: "/permission/platforms/AddPlatform",
  //   title: "平台新增",
  //   path: "AddPlatform", // 作为子路由时，path 通常不需要带前缀斜杠
  //   routePath: "/permission/platforms/AddPlatform",
  //   parentKey: "/permission/platforms", //  指向父级路由的 key
  //   hideInMenu: true //防止它作为子菜单展开显示
  // },
  {
    key: "/permission/employees",
    title: "员工管理",
    path: "/permission/employees",
    routePath: "/permission/employees",
    parentKey: "/permission"
  },
  {
    key: "/permission/employees/AddEmployee",
    title: "员工新增",
    path: "AddEmployee", // 作为子路由时，path 通常不需要带前缀斜杠
    routePath: "/permission/employees/AddEmployee",
    parentKey: "/permission/employees", //  指向父级路由的 key
    hideInMenu: true //防止它作为子菜单展开显示
  }
  // {
  //   key: "/permission/password",
  //   title: "登录密码修改",
  //   path: "/permission/password",
  //   routePath: "/permission/password",
  //   parentKey: "/permission"
  // }
];
