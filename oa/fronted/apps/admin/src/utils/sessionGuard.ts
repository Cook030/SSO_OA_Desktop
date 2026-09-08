import { MhModal } from "@mh-repo/ui";

import { isSessionEnded, markSessionEnded, redirectToSSOLogin } from "./tokenRefresh";

/**
 * 会话终结的统一入口。
 *
 * 触发来源有两条，最终都汇聚到这里，保证用户只看到一次提示、只跳转一次：
 *  1. WebSocket 推送 session_replaced（在线设备，体验更好）；
 *  2. HTTP 401 且 reason=SESSION_REPLACED（离线/断网/网关不可达时的兜底）。
 */

/** 服务端返回的机器原因码 */
export const REASON_SESSION_REPLACED = "SESSION_REPLACED";

interface MaybeReasonPayload {
  code?: unknown;
  reason?: unknown;
}

/** 判断响应体是否表示"会话已被置换" */
export function isSessionReplacedResponse(payload?: MaybeReasonPayload | null): boolean {
  return payload?.reason === REASON_SESSION_REPLACED;
}

function messageOf(reason?: string): string {
  switch (reason) {
    case "password_changed":
      return "密码已被修改，请重新登录。";
    case "revoked_by_admin":
      return "登录状态已被管理员终止，请重新登录。";
    case "replaced_by_new_login":
      return "您的账号已在其他设备登录，当前会话已退出。";
    default:
      return "当前登录状态已失效，请重新登录。";
  }
}

/**
 * 处理"会话终结"：提示 → 阻止后续刷新与请求 → 跳转登录页。
 * 全局只执行一次（防重入）。
 */
export function handleSessionEnded(reason?: string): void {
  if (isSessionEnded()) return;
  markSessionEnded();

  MhModal.warning({
    title: "登录状态已失效",
    content: messageOf(reason),
    okText: "重新登录",
    closable: false,
    maskClosable: false,
    keyboard: false,
    onOk: () => redirectToSSOLogin()
  });
}
