/**
 * 设备唯一标识。
 *
 * 生成策略（前端负责）：
 * - 首次访问时生成 UUID v4 并持久化到 localStorage，同一浏览器后续复用；
 * - 无 crypto.randomUUID 时退化为 crypto.getRandomValues，再退化为时间戳 + 随机数。
 *
 * 用途（后端负责校验与绑定）：
 * - 登录时通过 X-MH-Device-Id 上报，写入会话记录；
 * - WebSocket 握手时通过 query 携带，网关与会话绑定的设备标识比对。
 *
 * 注意：它是"设备标签"而不是凭证。单设备登录的权威依据是 SSO 的 current_session，
 * 即使这里拿不到设备标识（隐私模式 / 禁用 localStorage），登录与踢线依然正常工作。
 */

const DEVICE_ID_STORAGE_KEY = "mh_device_id";

/** 设备类型：与后端 utils.NormalizeDeviceType 的取值保持一致 */
export type DeviceType = "desktop" | "mobile" | "tablet" | "unknown";

function generateUUID(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  if (typeof crypto !== "undefined" && typeof crypto.getRandomValues === "function") {
    const bytes = crypto.getRandomValues(new Uint8Array(16));
    // 按 RFC 4122 置位 version(4) 与 variant
    bytes[6] = (bytes[6] & 0x0f) | 0x40;
    bytes[8] = (bytes[8] & 0x3f) | 0x80;
    const hex = Array.from(bytes, b => b.toString(16).padStart(2, "0")).join("");
    return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
  }
  return `${Date.now().toString(16)}-${Math.random().toString(16).slice(2, 10)}`;
}

function readStoredDeviceId(): string | null {
  try {
    const value = window.localStorage.getItem(DEVICE_ID_STORAGE_KEY);
    return value && value.trim() ? value : null;
  } catch {
    // 隐私模式或 localStorage 被禁用：降级为内存态标识
    return null;
  }
}

let memoryDeviceId: string | null = null;

/** 获取（必要时生成并持久化）本设备唯一标识 */
export function getDeviceId(): string {
  if (memoryDeviceId) return memoryDeviceId;

  const stored = readStoredDeviceId();
  if (stored) {
    memoryDeviceId = stored;
    return stored;
  }

  const created = generateUUID();
  memoryDeviceId = created;
  try {
    window.localStorage.setItem(DEVICE_ID_STORAGE_KEY, created);
  } catch {
    // 持久化失败不影响本次会话使用
  }
  return created;
}

/** 依据 UA 粗略判定设备类型，仅用于展示与审计 */
export function getDeviceType(): DeviceType {
  if (typeof navigator === "undefined") return "unknown";
  const ua = navigator.userAgent;
  if (/iPad|Tablet/i.test(ua)) return "tablet";
  if (/Mobile|Android|iPhone|iPod/i.test(ua)) return "mobile";
  return "desktop";
}

/** 登录等 HTTP 请求携带的设备标识头 */
export function deviceHeaders(): Record<string, string> {
  return {
    "X-MH-Device-Id": getDeviceId(),
    "X-MH-Device-Type": getDeviceType()
  };
}
