import { getDeviceId } from "../utils/device";
import { handleSessionEnded } from "../utils/sessionGuard";
import { EVENT_TYPE, WS_CLOSE_CODE } from "./protocol";
import { RealtimeClient, buildRealtimeUrl } from "./wsClient";

/**
 * 实时连接的应用层入口：单例客户端 + 事件到业务动作的映射。
 *
 * 职责划分：
 * - wsClient 只管连接（建连、心跳、退避重连、ACK、幂等）；
 * - 本模块只管语义（收到什么事件 → 做什么业务动作）；
 * - 会话是否真的失效由 SSO 判定，前端不做"自我判断"，只展示与跳转。
 */

let client: RealtimeClient | null = null;

function resolveReason(eventReason?: string): string | undefined {
  return eventReason || undefined;
}

function createClient(): RealtimeClient {
  return new RealtimeClient({
    url: buildRealtimeUrl(getDeviceId()),
    onEvent: event => {
      if (event.eventType === EVENT_TYPE.SESSION_REPLACED || event.eventType === EVENT_TYPE.SESSION_TERMINATED) {
        handleSessionEnded(resolveReason(event.reason));
      }
    },
    onSessionInvalid: code => {
      // 网关主动关闭且不可重连：等价于会话终结
      handleSessionEnded(code === WS_CLOSE_CODE.DEVICE_MISMATCH ? "device_mismatch" : undefined);
    },
    logger: (message, ...args) => {
      if (import.meta.env.DEV) {
        console.log(message, ...args);
      }
    }
  });
}

/** 启动实时连接（幂等） */
export function startRealtime(): void {
  client ??= createClient();
  client.start();
}

/** 停止实时连接（登出、卸载） */
export function stopRealtime(): void {
  client?.stop();
}
