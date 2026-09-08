import { useEffect } from "react";

import { startRealtime, stopRealtime } from "../realtime";

/**
 * 在应用主布局挂载期间维持一条实时连接。
 *
 * 断线重连、心跳、ACK 与幂等都在 realtime 模块内部处理，
 * 本 Hook 只负责生命周期的开关。
 */
export function useRealtimeSession(): void {
  useEffect(() => {
    startRealtime();
    return () => stopRealtime();
  }, []);
}
