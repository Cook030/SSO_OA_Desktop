import { reactive } from "vue";

import type { MeResult, UserInfo } from "@/api/auth";
import { fetchMe, login as apiLogin, logout as apiLogout } from "@/api/auth";

interface AuthState {
  /** 当前登录用户；null 表示未登录 */
  user: UserInfo | null;
  /** 用户组/角色等扩展信息 */
  profile: Omit<MeResult, "user"> | null;
  /** 首次会话探测是否完成（路由守卫据此等待） */
  ready: boolean;
}

export const authState = reactive<AuthState>({
  user: null,
  profile: null,
  ready: false,
});

/** 登录并同步用户态 */
export async function signIn(account: string, password: string): Promise<void> {
  const data = await apiLogin(account, password);
  authState.user = data.user;
  authState.profile = null;
}

/** 退出登录（服务端撤销会话并清除 Cookie） */
export async function signOut(): Promise<void> {
  try {
    await apiLogout();
  } finally {
    authState.user = null;
    authState.profile = null;
  }
}

/** 恢复会话：依赖 HttpOnly Cookie 调 /me，返回是否已登录 */
export async function restoreSession(): Promise<boolean> {
  try {
    const me = await fetchMe();
    authState.user = me.user;
    const { user: _user, ...rest } = me;
    authState.profile = rest;
    return true;
  } catch {
    authState.user = null;
    authState.profile = null;
    return false;
  } finally {
    authState.ready = true;
  }
}
