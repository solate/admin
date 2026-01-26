// Services API

import { api } from '@/utils/request'
import type { ApiResponse, ListResponse, ListParams } from '@/types/api'
import type { Service } from '@/types/models'

/**
 * 服务 API
 */
export const servicesApi = {
  /**
   * 获取服务列表（分页）
   */
  async list(params?: ListParams): Promise<ListResponse<Service>> {
    const res = await api.get<ApiResponse<ListResponse<Service>>>('/services', { params })
    return res.data.data
  },

  /**
   * 根据 ID 获取服务
   */
  async getById(id: string): Promise<Service> {
    const res = await api.get<ApiResponse<Service>>(`/services/${id}`)
    return res.data.data
  },

  /**
   * 创建服务
   */
  async create(data: Partial<Service>): Promise<Service> {
    const res = await api.post<ApiResponse<Service>>('/services', data)
    return res.data.data
  },

  /**
   * 更新服务
   */
  async update(id: string, data: Partial<Service>): Promise<Service> {
    const res = await api.put<ApiResponse<Service>>(`/services/${id}`, data)
    return res.data.data
  },

  /**
   * 删除服务
   */
  async delete(id: string): Promise<void> {
    await api.delete<ApiResponse<void>>(`/services/${id}`)
  },

  /**
   * 切换服务启用状态
   */
  async toggle(id: string, enabled: boolean): Promise<Service> {
    const res = await api.put<ApiResponse<Service>>(`/services/${id}/toggle`, { enabled })
    return res.data.data
  }
}
