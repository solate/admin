// Services API

import { api } from '@/utils/request'
import type { ApiResponse, ListResponse, ListParams } from '@/types/api'
import type { Service } from '@/types/models'

export const servicesApi = {
  /**
   * Get services list with pagination
   */
  list: (params?: ListParams) =>
    api.get<ApiResponse<ListResponse<Service>>>('/services', { params }),

  /**
   * Get service by ID
   */
  getById: (id: string) => api.get<ApiResponse<Service>>(`/services/${id}`),

  /**
   * Create new service
   */
  create: (data: Partial<Service>) =>
    api.post<ApiResponse<Service>>('/services', data),

  /**
   * Update service
   */
  update: (id: string, data: Partial<Service>) =>
    api.put<ApiResponse<Service>>(`/services/${id}`, data),

  /**
   * Delete service
   */
  delete: (id: string) => api.delete<ApiResponse<void>>(`/services/${id}`),

  /**
   * Toggle service enabled status
   */
  toggle: (id: string, enabled: boolean) =>
    api.put<ApiResponse<Service>>(`/services/${id}/toggle`, { enabled })
}
