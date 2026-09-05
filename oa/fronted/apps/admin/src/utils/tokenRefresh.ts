import axios from "axios";

const SSO_LOGIN_URL = "https://sso2.maplehaze.cn/login?redirect=https%3A%2F%2Foa.maplehaze.cn%2F";
const REFRESH_ENDPOINT = "https://sso2.maplehaze.cn/api/v1/auth/refresh";

let refreshPromise: Promise<boolean> | null = null;
let hasRedirected = false;

/** 跳转到 SSO 登录页（防重入，避免多次跳转） */
export function redirectToSSOLogin(): void {
  if (hasRedirected) return;
  hasRedirected = true;
  window.location.href = SSO_LOGIN_URL;
}

/**
 * 调用 SSO 刷新 token 接口。
 *
 * refresh_token 存储在 HttpOnly cookie (mh_sso2_refresh_token, Domain=.maplehaze.cn) 中，
 * 浏览器会自动携带，无需手动传入。
 * 成功后 SSO 通过 Set-Cookie 更新 mh_sso2_access_token 和 mh_sso2_refresh_token，
 * 后续请求会自动携带新的 access_token。
 *
 * @returns true = 刷新成功，可重试原请求；false = 刷新失败，需跳转登录页
 */
export function refreshAccessToken(): Promise<boolean> {
  // 并发去重：多个 401 同时到达时只发一次刷新请求
  if (refreshPromise) {
    return refreshPromise;
  }

  refreshPromise = (async () => {
    try {
      const response = await axios.post(
        REFRESH_ENDPOINT,
        {},
        {
          withCredentials: true,
          headers: { "Content-Type": "application/json" },
          timeout: 10000
        }
      );

      if (response.data?.code === 200) {
        console.log("[tokenRefresh] token 刷新成功");
        return true;
      }
      console.warn("[tokenRefresh] token 刷新失败:", response.data);
      return false;
    } catch (error) {
      console.warn("[tokenRefresh] token 刷新异常:", error);
      return false;
    } finally {
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}
