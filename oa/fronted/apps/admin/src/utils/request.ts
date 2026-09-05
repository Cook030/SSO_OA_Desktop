import type { AxiosInstance, AxiosResponse, InternalAxiosRequestConfig } from "axios";
import axios from "axios";
import { redirectToSSOLogin, refreshAccessToken } from "./tokenRefresh";

// 定义后端统一返回格式
interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
}

// 创建 axios 实例
const request: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 10000,
  withCredentials: true,
  headers: {
    "Content-Type": "application/json"
  }
});

// ----- 请求拦截器：自动注入 Token -----
// request.interceptors.request.use(
//   (config: InternalAxiosRequestConfig) => {
//     //  test_token
//     const token = "wo-shi-szq";

//     if (token) {
//       // 按照文档要求：Authorization: Bearer <access_token>
//       config.headers.Authorization = `Bearer ${token}`;
//     }

//     return config;
//   },
//   error => {
//     return Promise.reject(error);
//   }
// );

request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // Token 已通过 HttpOnly Cookie 自动附加，前端无需任何操作。
    return config;
  },
  error => {
    return Promise.reject(error);
  }
);

// ----- 响应拦截器：统一解包 & 错误处理 -----
request.interceptors.response.use(
  async (response: AxiosResponse<ApiResponse>) => {
    const { code, message, data } = response.data;

    console.log("[request] 响应拦截器 - success handler", {
      url: response.config.url,
      httpStatus: response.status,
      body: response.data,
      code,
      message
    });

    // 业务成功
    if (code === 200) {
      return data;
    }

    // 业务层面的 401：尝试刷新 token 后重试，失败则跳转登录
    if (code === 401) {
      // 已重试过的请求仍返回 401，直接跳转登录
      if ((response.config as any)._retried) {
        redirectToSSOLogin();
        return Promise.reject(new Error(message || "登录已过期"));
      }

      const refreshed = await refreshAccessToken();
      if (refreshed) {
        // 刷新成功，重试原请求（新 cookie 已由浏览器自动更新）
        return request.request({ ...response.config, _retried: true } as any);
      }
      // 刷新失败，跳转 SSO 登录页
      redirectToSSOLogin();
      return Promise.reject(new Error(message || "登录已过期"));
    }

    // 其他业务错误
    return Promise.reject(new Error(message || "请求失败"));
  },
  async (error: any) => {
    console.log("[request] 响应拦截器 - error handler", {
      url: error.config?.url,
      message: error.message,
      hasResponse: !!error.response,
      status: error.response?.status,
      data: error.response?.data
    });

    if (error.response) {
      const { status, data } = error.response;

      // HTTP 401：尝试刷新 token 后重试，失败则跳转登录
      if (status === 401) {
        if ((error.config as any)?._retried) {
          redirectToSSOLogin();
          return Promise.reject(new Error("登录已过期，请重新登录"));
        }

        const refreshed = await refreshAccessToken();
        if (refreshed) {
          return request.request({ ...error.config, _retried: true } as any);
        }
        redirectToSSOLogin();
        return Promise.reject(new Error("登录已过期，请重新登录"));
      }

      if (status === 403) {
        return Promise.reject(new Error(data?.message || "没有操作权限，请联系管理员"));
      }

      if (status === 500) {
        return Promise.reject(new Error("服务器内部错误，请稍后重试"));
      }

      return Promise.reject(new Error(data?.message || `请求异常 (${status})`));
    }

    if (error.code === "ECONNABORTED" || error.message.includes("timeout")) {
      return Promise.reject(new Error("请求超时，请检查网络"));
    }

    return Promise.reject(error);
  }
);

export default request;
