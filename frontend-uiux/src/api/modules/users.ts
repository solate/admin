// Users API

import { api } from '@/utils/request'
import type { ApiResponse, ListResponse, ListParams } from '@/types/api'
import type { User } from '@/types/models'

export const usersApi = {
  /**
   * Get users list with pagination
   */
  list: (params?: ListParams) =>
    api.get<ApiResponse<ListResponse<User>>>('/users', { params }),

  /**
   * Get user by ID
   */
  getById: (id: string) => api.get<ApiResponse<User>>(`/users/${id}`),

  /**
   * Create new user
   */
  create: (data: Partial<User>) =>
    api.post<ApiResponse<User>>('/users', data),

  /**
   * Update user
   */
  update: (id: string, data: Partial<User>) =>
    api.put<ApiResponse<User>>(`/users/${id}`, data),

  /**
   * Delete user
   */
  delete: (id: string) => api.delete<ApiResponse<void>>(`/users/${id}`),

  /**
   * Assign roles to user
   */
  assignRoles: (id: string, roles: string[]) =>
    api.put<ApiResponse<User>>(`/users/${id}/roles`, { roles }),

  /**
   * Change password
   */
  changePassword: (data: { oldPassword: string; newPassword: string }) =>
    api.post<ApiResponse<void>>('/user/password/change', data),

  /**
   * Get user profile
   */
  profile: () => api.get<ApiResponse<User>>('/user/profile')
}
