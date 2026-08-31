/** 统一响应结构（接口文档 §2） */
export interface ApiResponse<T> {
  code: number;
  msg: string;
  data: T | null;
}

/** 业务错误：携带服务端业务码与提示语 */
export class ApiError extends Error {
  constructor(
    message: string,
    readonly code: number,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

/** 内存中的 access token（Bearer 备用；主认证走 HttpOnly Cookie） */
let memoryAccessToken: string | null = null;

export function setAccessToken(token: string | null): void {
  memoryAccessToken = token;
}

const BASE_URL = "/api/v1/auth";

/** 防并发：多个请求同时 401 时只发起一次 refresh */
let refreshing: Promise<boolean> | null = null;

async function tryRefresh(): Promise<boolean> {
  refreshing ??= rawRequest<{ accessToken: string }>("/refresh", { method: "POST", body: "{}" })
    .then((data) => {
      setAccessToken(data.accessToken);
      return true;
    })
    .catch(() => false)
    .finally(() => {
      refreshing = null;
    });
  return refreshing;
}

async function rawRequest<T>(path: string, init: RequestInit): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    Accept: "application/json",
    ...(init.headers as Record<string, string> | undefined),
  };
  if (memoryAccessToken) {
    headers.Authorization = `Bearer ${memoryAccessToken}`;
  }

  let res: Response;
  try {
    res = await fetch(BASE_URL + path, { ...init, headers, credentials: "include" });
  } catch {
    throw new ApiError("网络异常，请稍后重试", -1);
  }

  const body = (await res.json().catch(() => null)) as ApiResponse<T> | null;
  if (body?.code === 200 && body.data !== undefined) {
    return body.data as T;
  }
  throw new ApiError(body?.msg ?? `请求失败（${res.status}）`, body?.code ?? res.status);
}

/**
 * 认证接口请求封装：
 * - credentials: include 携带 HttpOnly Cookie
 * - 401 时自动用 refresh token 换新并重试一次
 */
export async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  try {
    return await rawRequest<T>(path, init);
  } catch (err) {
    const isAuthEndpoint = path === "/login" || path === "/register" || path === "/refresh";
    if (err instanceof ApiError && err.code === 401 && !isAuthEndpoint && (await tryRefresh())) {
      return rawRequest<T>(path, init);
    }
    throw err;
  }
}
