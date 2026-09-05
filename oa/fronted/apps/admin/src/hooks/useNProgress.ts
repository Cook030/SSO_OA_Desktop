import NProgress from "nprogress";
import { useEffect } from "react";
import { useLocation } from "react-router";

/**
 * useNProgress Hook
 *
 * 监听路由变化并自动控制 NProgress 进度条的显示和隐藏
 *
 * 功能：
 * - 监听 React Router 的 location 变化
 * - 在路由开始切换时调用 NProgress.start()
 * - 在路由切换完成后调用 NProgress.done()
 * - 处理快速连续路由切换（通过清理函数取消前一个进度条）
 * - 处理组件卸载时的资源清理
 * - 提供 NProgress 未加载时的降级处理
 *
 * 验证需求：
 * - 需求 1.1: 主应用路由切换时立即显示进度条
 * - 需求 1.2: 主应用路由加载完成后进度条消失
 * - 需求 5.1: 快速连续路由切换时取消前一个进度条
 * - 需求 5.3: 组件卸载时清理所有定时器
 * - 需求 6.1: 清理所有定时器和事件监听器
 * - 需求 6.3: 返回清理函数
 * - 需求 8.1: NProgress 未加载时静默失败
 * - 需求 8.2: 开发环境输出警告信息
 *
 * @example
 * ```typescript
 * function App() {
 *   // 在应用根组件中使用
 *   useNProgress();
 *
 *   return <div>...</div>;
 * }
 * ```
 *
 * @remarks
 * - 必须在 React Router 的上下文中使用（useLocation 依赖）
 * - 如果 NProgress 库未加载，会在开发环境输出警告并静默失败
 * - 使用 setTimeout 确保在 DOM 更新后调用 NProgress.done()
 * - 清理函数会在路由变化和组件卸载时执行，防止内存泄漏
 */
export function useNProgress(): void {
  const location = useLocation();

  useEffect(() => {
    // 降级处理：检查 NProgress 是否已加载
    if (typeof NProgress === "undefined" || !NProgress) {
      if (process.env.NODE_ENV === "development") {
        console.warn("[useNProgress] NProgress is not loaded. Progress bar will not be displayed.");
      }
      return;
    }

    // 路由开始变化：启动进度条
    NProgress.start();

    // 使用 setTimeout 确保在 DOM 更新后完成进度条
    // 这样可以让进度条在新页面渲染完成后再消失
    const timer = setTimeout(() => {
      NProgress.done();
    }, 0);

    // 清理函数：处理快速连续路由切换和组件卸载
    return () => {
      // 清理定时器，防止内存泄漏
      clearTimeout(timer);

      // 立即完成进度条
      // 这对于快速连续路由切换很重要：
      // - 如果用户在前一个路由加载完成前触发新的路由切换
      // - 清理函数会取消前一个进度条
      // - 新的 effect 会启动新的进度条
      NProgress.done();
    };
  }, [location.pathname, location.search]);
  // 注意：我们监听 pathname 和 search，但不监听 hash
  // 因为 hash 变化通常不需要显示进度条（页面内锚点跳转）
}
