// Tenants API

import { api } from '@/utils/request'
import type { ApiResponse, ListResponse, ListParams } from '@/types/api'
import type { Tenant } from '@/types/models'

export const tenantsApi = {
  /**
   * Get tenants list with pagination
   */
  list: (params?: ListParams) =>
    api.get<ApiResponse<ListResponse<Tenant>>>('/tenants', { params }),

  /**
   * Get all tenants (without pagination)
   */
  getAll: () => api.get<ApiResponse<Tenant[]>>('/tenants/all'),

  /**
   * Get tenant by ID
   */
  getById: (id: string) =>
    api.get<ApiResponse<Tenant>>(`/tenants/${id}`),

  /**
   * Create new tenant
   */
  create: (data: Partial<Tenant>) =>
    api.post<ApiResponse<Tenant>>('/tenants', data),

  /**
   * Update tenant
   */
  update: (id: string, data: Partial<Tenant>) =>
    api.put<ApiResponse<Tenant>>(`/tenants/${id}`, data),

  /**
   * Delete tenant
   */
  delete: (id: string) => api.delete<ApiResponse<void>>(`/tenants/${id}`),

  /**
   * Update tenant status
   */
  updateStatus: (id: string, status: Tenant['status']) =>
    api.put<ApiResponse<Tenant>>('/tenants/status', { id, status })
}
