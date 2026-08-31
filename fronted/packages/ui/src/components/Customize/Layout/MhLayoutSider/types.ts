import type React from "react";

export interface MhLayoutSiderMenuItem {
  key: string;
  title: React.ReactNode;
  icon?: React.ReactNode;
  path?: string;
  children?: MhLayoutSiderMenuItem[];
  /** 为 true 时点击在新标签页打开（需配合 appBasePath） */
  openInNewTab?: boolean;
}

export interface MhLayoutSiderProps {
  menuItems: MhLayoutSiderMenuItem[];
  /** 当前路由 pathname，用于高亮与展开父级菜单 */
  pathname: string;
  /** 菜单项点击导航 */
  onNavigate: (path: string) => void;
  /** 应用 base path，如 /dsp，用于新标签页打开完整路径 */
  appBasePath?: string;
  collapsed?: boolean;
  defaultCollapsed?: boolean;
  onCollapse?: (collapsed: boolean) => void;
  width?: number;
  collapsedWidth?: number;
  inlineIndent?: number;
  className?: string;
  siderClassName?: string;
  menuClassName?: string;
  footer?: React.ReactNode;
}

export type MatchedMenuState = {
  selectedKey: string | null;
  ancestorKeys: string[];
};
