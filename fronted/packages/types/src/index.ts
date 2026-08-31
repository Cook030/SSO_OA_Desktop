import React from "react";

// 通用类型定义
export interface User {
  id: string;
  name: string;
  email: string;
  role: "admin" | "user" | "guest";
  avatar?: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface ApiResponse<T = any> {
  success: boolean;
  data: T;
  message?: string;
  code?: number;
}

export interface PaginationParams {
  page: number;
  pageSize: number;
  total?: number;
}

export interface PaginatedResponse<T> extends ApiResponse<T[]> {
  pagination: {
    page: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
}

// 微前端相关类型
export interface MicroAppConfig {
  name: string;
  entry: string;
  container: string;
  activeRule: string;
}

// DSP 相关类型
export interface Campaign {
  id: string;
  name: string;
  status: "active" | "paused" | "completed";
  budget: number;
  spent: number;
  impressions: number;
  clicks: number;
  ctr: number;
  createdAt: Date;
}

// SSP 相关类型
export interface AdSlot {
  id: string;
  name: string;
  size: string;
  type: "banner" | "video" | "native";
  revenue: number;
  impressions: number;
  fillRate: number;
  createdAt: Date;
}

export interface MenuItem {
  key: string;
  title: string;
  icon?: string | React.ReactNode;
  path?: string;
  children?: MenuItem[];
}
