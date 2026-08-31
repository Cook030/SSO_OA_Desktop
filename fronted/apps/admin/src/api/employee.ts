// src/api/employee.ts
import request from "@/utils/request";

// ===== 员工列表 =====
export interface Employee {
  id: number;
  account: string;
  name: string;
  phone: string;
  email: string;
  department: string;
  platformPermissions: { id: number; name: string }[];
}

export interface EmployeeListResponse {
  total: number;
  list: Employee[];
}

export const getEmployees = (params: {
  keyword?: string;
  department?: string;
  platformId?: number;
  page?: number;
  pageSize?: number;
}) => request.get<EmployeeListResponse>("/employees", { params });

// ===== 新增员工 =====
export interface CreateEmployeeRequest {
  name: string;
  phone: string;
  emailPrefix: string;
  department: string;
  platformIds: number[];
}

export const createEmployee = (data: CreateEmployeeRequest) =>
  request.post<{ id: number; name: string; account: string; email: string }>("/employees", data);

// ===== 编辑员工 =====
export interface UpdateEmployeeRequest {
  name: string;
  phone: string;
  emailPrefix: string;
  department: string;
  platformIds: number[];
}

export const updateEmployee = (id: number, data: UpdateEmployeeRequest) =>
  request.put<{ id: number; name: string; phone: string; email: string; department: string }>(`/employees/${id}`, data);

// ===== 删除员工 =====
export const deleteEmployee = (id: number) => request.delete(`/employees/${id}`);

// ===== 部门列表 =====
export const getDepartments = () => request.get<string[]>("/employees/departments");

// ===== 批量设置权限 =====
export const batchSetPermissions = (data: { userIds: number[]; platformIds: number[] }) =>
  request.post<{ affectedCount: number }>("/employees/permissions/batch", data);

// ===== 批量清除权限 =====
export const batchClearPermissions = (data: { userIds: number[] }) =>
  request.delete<{ affectedCount: number }>("/employees/permissions/batch", { data });
