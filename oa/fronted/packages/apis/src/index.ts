import type { ApiResponse, PaginatedResponse, User, Campaign, AdSlot } from "@mh-repo/types";

// API 基础配置
const API_BASE_URL = process.env.VITE_API_BASE_URL || "https://api.example.com";

class ApiClient {
  private baseURL: string;
  private token: string | null = null;

  constructor(baseURL: string = API_BASE_URL) {
    this.baseURL = baseURL;
  }

  setToken(token: string) {
    this.token = token;
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<ApiResponse<T>> {
    const url = `${this.baseURL}${endpoint}`;
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(options.headers as Record<string, string>)
    };

    if (this.token) {
      headers.Authorization = `Bearer ${this.token}`;
    }

    try {
      const response = await fetch(url, {
        ...options,
        headers
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error("API request failed:", error);
      throw error;
    }
  }

  // 用户相关 API
  async getUsers(): Promise<ApiResponse<User[]>> {
    return this.request<User[]>("/users");
  }

  async getUserById(id: string): Promise<ApiResponse<User>> {
    return this.request<User>(`/users/${id}`);
  }

  async createUser(user: Omit<User, "id" | "createdAt" | "updatedAt">): Promise<ApiResponse<User>> {
    return this.request<User>("/users", {
      method: "POST",
      body: JSON.stringify(user)
    });
  }

  async updateUser(id: string, user: Partial<User>): Promise<ApiResponse<User>> {
    return this.request<User>(`/users/${id}`, {
      method: "PUT",
      body: JSON.stringify(user)
    });
  }

  async deleteUser(id: string): Promise<ApiResponse<void>> {
    return this.request<void>(`/users/${id}`, {
      method: "DELETE"
    });
  }

  // DSP 相关 API
  async getCampaigns(): Promise<ApiResponse<Campaign[]>> {
    return this.request<Campaign[]>("/dsp/campaigns");
  }

  async getCampaignById(id: string): Promise<ApiResponse<Campaign>> {
    return this.request<Campaign>(`/dsp/campaigns/${id}`);
  }

  async createCampaign(campaign: Omit<Campaign, "id" | "createdAt">): Promise<ApiResponse<Campaign>> {
    return this.request<Campaign>("/dsp/campaigns", {
      method: "POST",
      body: JSON.stringify(campaign)
    });
  }

  // SSP 相关 API
  async getAdSlots(): Promise<ApiResponse<AdSlot[]>> {
    return this.request<AdSlot[]>("/ssp/adslots");
  }

  async getAdSlotById(id: string): Promise<ApiResponse<AdSlot>> {
    return this.request<AdSlot>(`/ssp/adslots/${id}`);
  }

  async createAdSlot(adSlot: Omit<AdSlot, "id" | "createdAt">): Promise<ApiResponse<AdSlot>> {
    return this.request<AdSlot>("/ssp/adslots", {
      method: "POST",
      body: JSON.stringify(adSlot)
    });
  }
}

// 导出单例实例
export const apiClient = new ApiClient();

// 导出类型和工具函数
export { ApiClient };
export type { ApiResponse, PaginatedResponse };
