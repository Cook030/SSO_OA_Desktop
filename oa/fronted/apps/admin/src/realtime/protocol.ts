/**
 * 前端与 WebSocket Gateway 的最小协议（dev.md §5）。
 * 常量与 ws_gateway/internal/protocol 保持一致，修改需同步。
 */

/** 服务端 → 客户端消息 */
export interface ServerEventMessage {
  type: "event";
  eventId: string;
  eventType?: string;
  reason?: string;
}

/** 客户端 → 服务端消息（目前只有确认） */
export interface ClientAckMessage {
  type: "ack";
  eventId: string;
}

/** 事件类型 */
export const EVENT_TYPE = {
  SESSION_REPLACED: "session_replaced",
  SESSION_TERMINATED: "session_terminated"
} as const;

/** WebSocket 应用级关闭码 */
export const WS_CLOSE_CODE = {
  NORMAL: 1000,
  /** 会话无效 / 已被顶下线：不得用旧凭证重连 */
  SESSION_INVALID: 4001,
  /** 同一会话建立新连接，旧连接被取代 */
  DUPLICATE_SESSION: 4002,
  /** 设备标识与会话绑定不一致 */
  DEVICE_MISMATCH: 4003,
  /** 网关优雅关闭：可以立即重连 */
  SERVER_SHUTDOWN: 4004,
  /** 写队列积压：可以重连 */
  SLOW_CONSUMER: 4005,
  /** 心跳超时：可以重连 */
  HEARTBEAT_TIMEOUT: 4006
} as const;

export type WsCloseCode = (typeof WS_CLOSE_CODE)[keyof typeof WS_CLOSE_CODE];

/**
 * 收到该关闭码后不应再尝试重连：
 * 会话已死，重连必然被网关拒绝；应转向 HTTP 401 的下线流程。
 */
export const NON_RETRYABLE_CLOSE_CODES: readonly number[] = [
  WS_CLOSE_CODE.SESSION_INVALID,
  WS_CLOSE_CODE.DUPLICATE_SESSION,
  WS_CLOSE_CODE.DEVICE_MISMATCH
];

/** 解析后的领域事件 */
export interface RealtimeEvent {
  eventId: string;
  eventType: string;
  reason?: string;
}
