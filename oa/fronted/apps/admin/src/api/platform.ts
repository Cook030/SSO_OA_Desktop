// src/api/platform.ts
import request from "@/utils/request";

export interface Platform {
  id: number;
  name: string;
  link: string;
}

export const getPlatforms = (params?: { page?: number; pageSize?: number }) =>
  request.get<{ total: number; list: Platform[] }>("/platforms", { params });
