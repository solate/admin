// Tenants API

import { api } from '@/utils/request'
import type { ApiResponse, ListResponse, ListParams } from '@/types/api'
import type { Tenant } from '@/types/models'
import { isMockEnabled, mockTenantHandlers, mockTenants } from '@/mock'

/**
 * 租户 API
 */
export const tenantsApi = {
  /**
   * 获取租户列表（分页）
   */
  async list(params?: ListParams): Promise<ListResponse<Tenant>> {
    if (isMockEnabled()) {
      return mockTenantHandlers.list()
    }
    const res = await api.get<ApiResponse<ListResponse<Tenant>>>('/tenants', { params })
    return res.data.data
  },

  /**
   * 获取所有租户（不分页）
   */
  async getAll(): Promise<Tenant[]> {
    if (isMockEnabled()) {
      return mockTenants
    }
    const res = await api.get<ApiResponse<Tenant[]>>('/tenants/all')
    return res.data.data
  },

  /**
   * 根据 ID 获取租户
   */
  async getById(id: string): Promise<Tenant> {
    if (isMockEnabled()) {
      return mockTenantHandlers.get(id)
    }
    const res = await api.get<ApiResponse<Tenant>>(`/tenants/${id}`)
    return res.data.data
  },

  /**
   * 创建租户
   */
  async create(data: Partial<Tenant>): Promise<Tenant> {
    if (isMockEnabled()) {
      return mockTenantHandlers.create(data)
    }
    const res = await api.post<ApiResponse<Tenant>>('/tenants', data)
    return res.data.data
  },

  /**
   * 更新租户
   */
  async update(id: string, data: Partial<Tenant>): Promise<Tenant> {
    if (isMockEnabled()) {
      const tenant = mockTenants.find(t => t.id === id)
      if (!tenant) throw new Error('Tenant not found')
      return { ...tenant, ...data, updatedAt: new Date().toISOString() }
    }
    const res = await api.put<ApiResponse<Tenant>>(`/tenants/${id}`, data)
    return res.data.data
  },

  /**
   * 删除租户
   */
  async delete(id: string): Promise<void> {
    if (isMockEnabled()) {
      return
    }
    await api.delete<ApiResponse<void>>(`/tenants/${id}`)
  },

  /**
   * 更新租户状态
   */
  async updateStatus(id: string, status: Tenant['status']): Promise<Tenant> {
    if (isMockEnabled()) {
      const tenant = mockTenants.find(t => t.id === id)
      if (!tenant) throw new Error('Tenant not found')
      return { ...tenant, status, updatedAt: new Date().toISOString() }
    }
    const res = await api.put<ApiResponse<Tenant>>('/tenants/status', { id, status })
    return res.data.data
  }
}
