import type React from "react";
import { useEffect, useMemo, useState } from "react";
import { LeftOutlined, RightOutlined } from "../../../General/Icon";
import { MhLayout } from "../../../Layout/Layout";
import { MhMenu } from "../../../Navigation/Menu";
import styles from "./index.module.less";
import { convertToAntdMenuItems, findMatchedMenuState } from "./menuUtils";
import type { MhLayoutSiderProps } from "./types";

const { Sider } = MhLayout;

export { findMatchedMenuState, matchPath, normalizePath } from "./menuUtils";
export type { MatchedMenuState, MhLayoutSiderMenuItem, MhLayoutSiderProps } from "./types";

export const MhLayoutSider: React.FC<MhLayoutSiderProps> = ({
  menuItems,
  pathname,
  onNavigate,
  appBasePath,
  collapsed: controlledCollapsed,
  defaultCollapsed = false,
  onCollapse,
  width = 256,
  collapsedWidth = 80,
  inlineIndent = 22,
  className,
  siderClassName,
  menuClassName,
  footer
}) => {
  const [uncontrolledCollapsed, setUncontrolledCollapsed] = useState(defaultCollapsed);
  const collapsed = controlledCollapsed ?? uncontrolledCollapsed;

  const setCollapsed = (next: boolean) => {
    if (controlledCollapsed === undefined) {
      setUncontrolledCollapsed(next);
    }
    onCollapse?.(next);
  };

  const [openKeys, setOpenKeys] = useState<string[]>([]);

  const matchedMenuState = useMemo(() => findMatchedMenuState(menuItems, pathname), [menuItems, pathname]);
  const selectedKeys = matchedMenuState.selectedKey ? [matchedMenuState.selectedKey] : [];

  const antdMenuItems = useMemo(
    () =>
      convertToAntdMenuItems(menuItems, {
        onNavigate,
        appBasePath,
        styles: {
          menuItemLabel: styles["mh-layout-sider-menu-item-label"],
          menuItemText: styles["mh-layout-sider-menu-item-text"],
          menuItemLink: styles["mh-layout-sider-menu-item-link"]
        }
      }),
    [menuItems, onNavigate, appBasePath]
  );

  useEffect(() => {
    if (collapsed) {
      setOpenKeys([]);
      return;
    }
    setOpenKeys(matchedMenuState.ancestorKeys);
  }, [collapsed, matchedMenuState.ancestorKeys]);

  useEffect(() => {
    window.dispatchEvent(
      new CustomEvent("mh-layout-sider-collapse", {
        detail: {
          collapsed,
          width: collapsed ? collapsedWidth : width
        }
      })
    );
  }, [collapsed, width, collapsedWidth]);

  const collapseBtnLeft = collapsed ? collapsedWidth - 1 : width - 1;

  return (
    <div className={[styles["mh-layout-sider-wrap"], className].filter(Boolean).join(" ")}>
      <Sider
        collapsible
        trigger={null}
        collapsed={collapsed}
        onCollapse={setCollapsed}
        theme="light"
        width={width}
        collapsedWidth={collapsedWidth}
        className={[styles["mh-layout-sider"], siderClassName].filter(Boolean).join(" ")}
      >
        <MhMenu
          mode="inline"
          selectedKeys={selectedKeys}
          openKeys={openKeys}
          onOpenChange={keys => setOpenKeys(keys as string[])}
          inlineIndent={inlineIndent}
          items={antdMenuItems}
          className={[styles["mh-layout-sider-menu"], menuClassName].filter(Boolean).join(" ")}
        />
        {footer ? <div className={styles["mh-layout-sider-footer"]}>{footer}</div> : null}
      </Sider>
      <button
        type="button"
        className={styles["mh-layout-sider-collapse-btn"]}
        style={{ left: collapseBtnLeft }}
        onClick={() => setCollapsed(!collapsed)}
        aria-label={collapsed ? "展开侧边栏" : "收起侧边栏"}
      >
        {collapsed ? <RightOutlined /> : <LeftOutlined />}
      </button>
    </div>
  );
};

export default MhLayoutSider;
