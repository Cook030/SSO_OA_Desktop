import type React from "react";
import { useLayoutEffect } from "react";
import { useLocation } from "react-router";

export interface MhConScrollToTopProps {
  /** 路由 pathname；不传时在 Router 上下文中自动读取 */
  pathname?: string;
}

/** 将页面滚动到顶部 */
export const scrollPageToTop = (): void => {
  window.scrollTo(0, 0);
  document.documentElement.scrollTop = 0;
  document.body.scrollTop = 0;
};

/**
 * 路由切换时滚动到页面顶部，需放在 Router 上下文中使用
 */
export const MhConScrollToTop: React.FC<MhConScrollToTopProps> = ({ pathname: pathnameProp }) => {
  const { pathname: locationPathname } = useLocation();
  const pathname = pathnameProp ?? locationPathname;

  useLayoutEffect(() => {
    scrollPageToTop();
  }, [pathname]);

  return null;
};

export default MhConScrollToTop;
