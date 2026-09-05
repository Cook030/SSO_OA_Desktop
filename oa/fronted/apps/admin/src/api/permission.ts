// src/api/permission.ts
import request from "@/utils/request";

// 修改密码（响应直接返回 null）
export const changePassword = (data: { oldPassword: string; newPassword: string }) =>
  request.put<void>("/auth/password", data);

// 平台列表（响应直接返回 { total, list }）
export const getPlatforms = (params: { page?: number; pageSize?: number }) =>
  request.get<{ total: number; list: any[] }>("/platforms", { params });

// 新增平台（响应直接返回 { id, name, link }）
export const createPlatform = (data: { name: string; link: string }) =>
  request.post<{ id: number; name: string; link: string }>("/platforms", data);

// 员工列表
export const getEmployees = (params: {
  keyword?: string;
  department?: string;
  platformId?: number;
  page?: number;
  pageSize?: number;
}) => request.get<{ total: number; list: any[] }>("/employees", { params });

// 批量设置权限
export const batchSetPermissions = (data: { userIds: number[]; platformIds: number[] }) =>
  request.post<{ affectedCount: number }>("/employees/permissions/batch", data);
