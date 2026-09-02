import { request, setAccessToken } from "./http";

/** 当前用户信息（接口文档 §4/§9 user 对象） */
export interface UserInfo {
  id: number;
  account: string;
  name: string;
  phone: string;
  email: string;
  role: string;
  department: string;
  passwordChanged: number;
  createTime: string;
  updateTime: string;
}

export interface LoginResult {
  accessToken: string;
  refreshToken: string;
  tokenType: string;
  expiresIn: number;
  refreshExpiresIn: number;
  user: UserInfo;
}

export interface MeResult {
  user: UserInfo;
  groups: { code: string; name: string }[];
  roles: { code: string; name: string }[];
  apps: string[];
}

export async function login(account: string, password: string): Promise<LoginResult> {
  const data = await request<LoginResult>("/login", {
    method: "POST",
    body: JSON.stringify({ account, password }),
  });
  setAccessToken(data.accessToken);
  return data;
}

export interface ChangePasswordPayload {
  password: string;
  confirmPassword: string;
}

/** 修改密码（成功后服务端会撤销全部会话并清除 Cookie，需重新登录） */
export async function changePassword(payload: ChangePasswordPayload): Promise<void> {
  await request<null>("/change-password", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function logout(): Promise<void> {
  try {
    await request<null>("/logout", { method: "POST", body: "{}" });
  } finally {
    setAccessToken(null);
  }
}

export async function fetchMe(): Promise<MeResult> {
  return request<MeResult>("/me", { method: "GET" });
}

export interface UpdateProfilePayload {
  nickname: string;
  email?: string;
  mobile?: string;
}

/** 更新个人资料（姓名/邮箱/手机号），返回更新后的用户信息 */
export async function updateProfile(payload: UpdateProfilePayload): Promise<UserInfo> {
  return request<UserInfo>("/me", { method: "PUT", body: JSON.stringify(payload) });
}
