// utils/ssoRequest.ts
import axios, { type AxiosInstance, type AxiosResponse, type InternalAxiosRequestConfig } from "axios";
import { handleSessionEnded, isSessionReplacedResponse } from "./sessionGuard";
import { redirectToSSOLogin, refreshAccessToken } from "./tokenRefresh";

// 定义 SSO 平台可能返回的统一格式（根据 SSO 接口文档调整）
interface SsoApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
  // reason：机器原因码（如 SESSION_REPLACED），用于区分 401 的具体成因
  reason?: string;
}

// 1. 创建专门用于 SSO 的 axios 实例
const ssoRequest: AxiosInstance = axios.create({
  // baseURL: "https://sso2.maplehaze.cn", // ⬅️ 直接指向 SSO 域名（线上）
  baseURL: "", // 纯本地：走同源 vite proxy（/api/v1/auth/ -> http://127.0.0.1:8081）
  timeout: 10000,
  withCredentials: true, // ⬅️ 关键：跨域请求携带 Cookie，SSO 登录态通常依赖 Cookie
  headers: {
    "Content-Type": "application/json"
  }
});

// 2. 请求拦截器 (可选)
// 如果 SSO 接口也需要特定的 Token 处理，可以在这里添加
// 根据你之前的代码，SSO 可能依赖 Cookie，所以这里可能不需要手动注入 Authorization
ssoRequest.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    return config;
  },
  error => {
    return Promise.reject(error);
  }
);

// 3. 响应拦截器 (可选)
// 处理 SSO 接口特有的响应格式和错误
ssoRequest.interceptors.response.use(
  async (response: AxiosResponse<SsoApiResponse>) => {
    const { code, message, data } = response.data;

    console.log("[ssoRequest] 响应拦截器 - success handler", {
      url: response.config.url,
      httpStatus: response.status,
      body: response.data,
      code,
      message
    });

    if (code === 200) {
      return data;
    }

    // 会话已被顶下线：刷新必然失败，直接走下线流程（dev.md §2 场景三）
    if (isSessionReplacedResponse(response.data)) {
      handleSessionEnded(response.data?.reason);
      return Promise.reject(new Error("账号已在其他设备登录"));
    }

    // 业务层面的 401：尝试刷新 token 后重试，失败则跳转登录
    if (code === 401) {
      if ((response.config as any)._retried) {
        redirectToSSOLogin();
        return Promise.reject(new Error(message || "SSO 登录已过期"));
      }

      const refreshed = await refreshAccessToken();
      if (refreshed) {
        return ssoRequest.request({ ...response.config, _retried: true } as any);
      }
      redirectToSSOLogin();
      return Promise.reject(new Error(message || "SSO 登录已过期"));
    }

    return Promise.reject(new Error(message || "SSO 请求失败"));
  },
  async (error: any) => {
    console.log("[ssoRequest] 响应拦截器 - error handler", {
      url: error.config?.url,
      message: error.message,
      hasResponse: !!error.response,
      status: error.response?.status,
      data: error.response?.data
    });

    if (error.response) {
      const { status, data } = error.response;

      // 会话已被顶下线：刷新必然失败，直接走下线流程
      if (isSessionReplacedResponse(data)) {
        handleSessionEnded(data?.reason);
        return Promise.reject(new Error("账号已在其他设备登录"));
      }

      // HTTP 401：尝试刷新 token 后重试，失败则跳转登录
      if (status === 401) {
        if ((error.config as any)?._retried) {
          redirectToSSOLogin();
          return Promise.reject(new Error("SSO 登录已过期"));
        }

        const refreshed = await refreshAccessToken();
        if (refreshed) {
          return ssoRequest.request({ ...error.config, _retried: true } as any);
        }
        redirectToSSOLogin();
        return Promise.reject(new Error("SSO 登录已过期"));
      }
    }
    return Promise.reject(error);
  }
);

export default ssoRequest;
