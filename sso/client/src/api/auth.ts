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

export interface RegisterResult {
  userId: number;
  username: string;
  status: string;
}

export interface MeResult {
  user: UserInfo;
  groups: { code: string; name: string }[];
  roles: { code: string; name: string }[];
  apps: string[];
}

export interface RegisterPayload {
  username: string;
  password: string;
  confirmPassword: string;
  email?: string;
  mobile?: string;
  nickname?: string;
}

export async function login(account: string, password: string): Promise<LoginResult> {
  const data = await request<LoginResult>("/login", {
    method: "POST",
    body: JSON.stringify({ account, password }),
  });
  setAccessToken(data.accessToken);
  return data;
}

export async function register(payload: RegisterPayload): Promise<RegisterResult> {
  return request<RegisterResult>("/register", {
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
