// Users API

import { api } from '@/utils/request'
import type { ApiResponse, ListResponse, ListParams } from '@/types/api'
import type { User } from '@/types/models'

/**
 * 用户 API
 */
export const usersApi = {
  /**
   * 获取用户列表（分页）
   */
  async list(params?: ListParams): Promise<ListResponse<User>> {
    const res = await api.get<ApiResponse<ListResponse<User>>>('/users', { params })
    return res.data.data
  },

  /**
   * 根据 ID 获取用户
   */
  async getById(id: string): Promise<User> {
    const res = await api.get<ApiResponse<User>>(`/users/${id}`)
    return res.data.data
  },

  /**
   * 创建用户
   */
  async create(data: Partial<User>): Promise<User> {
    const res = await api.post<ApiResponse<User>>('/users', data)
    return res.data.data
  },

  /**
   * 更新用户
   */
  async update(id: string, data: Partial<User>): Promise<User> {
    const res = await api.put<ApiResponse<User>>(`/users/${id}`, data)
    return res.data.data
  },

  /**
   * 删除用户
   */
  async delete(id: string): Promise<void> {
    await api.delete<ApiResponse<void>>(`/users/${id}`)
  },

  /**
   * 为用户分配角色
   */
  async assignRoles(id: string, roles: string[]): Promise<User> {
    const res = await api.put<ApiResponse<User>>(`/users/${id}/roles`, { roles })
    return res.data.data
  },

  /**
   * 修改密码
   */
  async changePassword(data: { oldPassword: string; newPassword: string }): Promise<void> {
    await api.post<ApiResponse<void>>('/user/password/change', data)
  },

  /**
   * 获取用户个人资料
   */
  async profile(): Promise<User> {
    const res = await api.get<ApiResponse<User>>('/user/profile')
    return res.data.data
  }
}
