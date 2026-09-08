import { getDeviceId } from "../utils/device";
import {
  EVENT_TYPE,
  NON_RETRYABLE_CLOSE_CODES,
  WS_CLOSE_CODE,
  type ClientAckMessage,
  type RealtimeEvent,
  type ServerEventMessage
} from "./protocol";

export type ConnectionStatus = "idle" | "connecting" | "open" | "reconnecting" | "closed";

export interface RealtimeClientOptions {
  /** WebSocket 端点（完整 ws/wss 地址，含 device_id，由 buildRealtimeUrl 生成） */
  url: string;
  /** 收到领域事件（已按 eventId 去重） */
  onEvent: (event: RealtimeEvent) => void;
  /** 网关判定会话失效且不可重连 */
  onSessionInvalid?: (code: number, reason: string) => void;
  /** 连接状态变化 */
  onStatusChange?: (status: ConnectionStatus) => void;
  /** 退避重连基础间隔 */
  baseDelayMs?: number;
  /** 退避重连最大间隔。 */
  maxDelayMs?: number;
  /** 幂等记录条数上限，超出后淘汰最早记录 */
  maxTrackedEventIds?: number;
  logger?: (message: string, ...args: unknown[]) => void;
}

/**
 * WebSocket 客户端：负责连接生命周期、退避重连与事件幂等。
 *
 * 心跳：由服务端每 25 秒发 Ping，浏览器自动回 Pong，客户端不需要额外的心跳消息。
 * 断线检测依赖 close 事件；服务端 45 秒收不到 Pong 会主动关闭，客户端随即重连。
 */
export class RealtimeClient {
  private readonly options: Required<
    Pick<RealtimeClientOptions, "baseDelayMs" | "maxDelayMs" | "maxTrackedEventIds">
  > &
    RealtimeClientOptions;

  private ws: WebSocket | null = null;
  private status: ConnectionStatus = "idle";
  private stopped = true;
  /** 会话已终结：此后不再重连，避免对已失效的凭证发起无限重试 */
  private sessionEnded = false;
  private attempt = 0;
  private reconnectTimer: number | null = null;
  /** 已处理事件 ID：保证同一事件重复投递只执行一次。 */
  private readonly processedEventIds = new Set<string>();
  private readonly processedOrder: string[] = [];

  constructor(options: RealtimeClientOptions) {
    this.options = {
      baseDelayMs: 1000,
      maxDelayMs: 30000,
      maxTrackedEventIds: 200,
      ...options
    };
  }

  /** 建立连接；已在连接中或会话已终结时忽略 */
  start(): void {
    if (this.sessionEnded || !this.stopped) return;
    this.stopped = false;
    this.attempt = 0;
    this.connect();
  }

  /** 主动停止（组件卸载、登出）；不再重连 */
  stop(): void {
    this.stopped = true;
    if (this.reconnectTimer !== null) {
      window.clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      // 移除监听后再关闭，避免触发重连逻辑
      this.ws.onclose = null;
      this.ws.onerror = null;
      this.ws.onmessage = null;
      this.ws.close(WS_CLOSE_CODE.NORMAL, "client stopped");
      this.ws = null;
    }
    this.setStatus("closed");
  }

  getStatus(): ConnectionStatus {
    return this.status;
  }

  // ---------- 内部实现 ----------

  private connect(): void {
    if (this.stopped || this.ws) return;

    const url = this.buildUrl();
    this.setStatus(this.attempt === 0 ? "connecting" : "reconnecting");

    let socket: WebSocket;
    try {
      socket = new WebSocket(url);
    } catch (error) {
      this.log("创建 WebSocket 失败", error);
      this.scheduleReconnect();
      return;
    }
    this.ws = socket;

    socket.onopen = () => {
      this.attempt = 0;
      this.setStatus("open");
      this.log("WebSocket 已连接");
    };

    socket.onmessage = event => {
      if (typeof event.data !== "string") return;
      this.handleMessage(event.data);
    };

    socket.onerror = () => {
      // close 事件一定会跟随 error 事件，重连逻辑统一放在 onclose
      this.log("WebSocket 发生错误");
    };

    socket.onclose = event => {
      this.ws = null;
      this.log("WebSocket 已关闭", event.code, event.reason);
      if (this.stopped) {
        this.setStatus("closed");
        return;
      }
      if (NON_RETRYABLE_CLOSE_CODES.includes(event.code)) {
        this.sessionEnded = true;
        this.stopped = true;
        this.setStatus("closed");
        this.options.onSessionInvalid?.(event.code, event.reason);
        return;
      }
      this.scheduleReconnect();
    };
  }

  /** 带抖动的指数退避重连，最大间隔 30 秒 */
  private scheduleReconnect(): void {
    if (this.stopped) return;
    this.attempt += 1;
    const exponential = Math.min(this.options.maxDelayMs, this.options.baseDelayMs * 2 ** (this.attempt - 1));
    const jitter = Math.random() * 0.3 * exponential;
    const delay = Math.min(this.options.maxDelayMs, exponential + jitter);

    this.setStatus("reconnecting");
    this.log(`${Math.round(delay)}ms 后第 ${this.attempt} 次重连`);
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, delay);
  }

  private handleMessage(raw: string): void {
    let message: ServerEventMessage;
    try {
      message = JSON.parse(raw) as ServerEventMessage;
    } catch {
      this.log("无法解析的消息", raw);
      return;
    }
    if (message.type !== "event" || !message.eventId) return;

    // 先 ACK 再去重：即使重复投递也回复，服务端据此清理待投递记录
    this.sendAck(message.eventId);

    if (this.isProcessed(message.eventId)) {
      this.log("事件重复投递，已幂等忽略", message.eventId);
      return;
    }
    this.rememberEventId(message.eventId);

    const event: RealtimeEvent = {
      eventId: message.eventId,
      eventType: message.eventType ?? "",
      reason: message.reason
    };
    if (event.eventType === EVENT_TYPE.SESSION_REPLACED || event.eventType === EVENT_TYPE.SESSION_TERMINATED) {
      // 会话已死：停止重连，避免用失效凭证反复握手
      this.sessionEnded = true;
      this.stopped = true;
    }
    this.options.onEvent(event);
  }

  private sendAck(eventId: string): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;
    const ack: ClientAckMessage = { type: "ack", eventId };
    this.ws.send(JSON.stringify(ack));
  }

  private isProcessed(eventId: string): boolean {
    return this.processedEventIds.has(eventId);
  }

  private rememberEventId(eventId: string): void {
    this.processedEventIds.add(eventId);
    this.processedOrder.push(eventId);
    const overflow = this.processedOrder.length - this.options.maxTrackedEventIds;
    for (let i = 0; i < overflow; i += 1) {
      const oldest = this.processedOrder.shift();
      if (oldest) this.processedEventIds.delete(oldest);
    }
  }

  private buildUrl(): string {
    // url 由 buildRealtimeUrl 生成（含 device_id），此处直接用，避免两处各拼一半
    return this.options.url;
  }

  private setStatus(status: ConnectionStatus): void {
    if (this.status === status) return;
    this.status = status;
    this.options.onStatusChange?.(status);
  }

  private log(message: string, ...args: unknown[]): void {
    this.options.logger?.(`[realtime] ${message}`, ...args);
  }
}

/** 构造实时连接地址：默认同源 /api/v1/realtime/ws，可用 VITE_WS_BASE_URL 覆盖 */
export function buildRealtimeUrl(deviceId: string = getDeviceId()): string {
  const raw = import.meta.env.VITE_WS_BASE_URL || "/api/v1/realtime/ws";
  const url = new URL(raw, window.location.href);
  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  url.searchParams.set("device_id", deviceId);
  return url.toString();
}
