// Tenants API

import { api } from '@/utils/request'
import { tenantMock } from '@/mock/handlers'
import { env } from '@/config/env'
import type { ApiResponse, ListResponse, ListParams } from '@/types/api'
import type { Tenant } from '@/types/models'

/**
 * 租户 API
 * 根据 env.useMock 决定使用真实 API 还是 Mock 数据
 */
export const tenantsApi = {
  /**
   * 获取租户列表（分页）
   */
  async list(params?: ListParams): Promise<ListResponse<Tenant>> {
    if (env.useMock) {
      return tenantMock.list(params)
    }
    const res = await api.get<ApiResponse<ListResponse<Tenant>>>('/tenants', { params })
    return res.data.data
  },

  /**
   * 获取所有租户（不分页）
   */
  async getAll(): Promise<Tenant[]> {
    if (env.useMock) {
      return tenantMock.getAll()
    }
    const res = await api.get<ApiResponse<Tenant[]>>('/tenants/all')
    return res.data.data
  },

  /**
   * 根据 ID 获取租户
   */
  async getById(id: string): Promise<Tenant> {
    if (env.useMock) {
      return tenantMock.detail(id)
    }
    const res = await api.get<ApiResponse<Tenant>>(`/tenants/${id}`)
    return res.data.data
  },

  /**
   * 创建租户
   */
  async create(data: Partial<Tenant>): Promise<Tenant> {
    if (env.useMock) {
      return tenantMock.create(data)
    }
    const res = await api.post<ApiResponse<Tenant>>('/tenants', data)
    return res.data.data
  },

  /**
   * 更新租户
   */
  async update(id: string, data: Partial<Tenant>): Promise<Tenant> {
    if (env.useMock) {
      return tenantMock.update(id, data)
    }
    const res = await api.put<ApiResponse<Tenant>>(`/tenants/${id}`, data)
    return res.data.data
  },

  /**
   * 删除租户
   */
  async delete(id: string): Promise<void> {
    if (env.useMock) {
      return tenantMock.delete(id)
    }
    await api.delete<ApiResponse<void>>(`/tenants/${id}`)
  },

  /**
   * 更新租户状态
   */
  async updateStatus(id: string, status: Tenant['status']): Promise<Tenant> {
    if (env.useMock) {
      return tenantMock.updateStatus(id, status)
    }
    const res = await api.put<ApiResponse<Tenant>>('/tenants/status', { id, status })
    return res.data.data
  }
}
