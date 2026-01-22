// Users API

import { api } from '@/utils/request'
import { userMock } from '@/mock/handlers'
import { env } from '@/config/env'
import type { ApiResponse, ListResponse, ListParams } from '@/types/api'
import type { User } from '@/types/models'

/**
 * 用户 API
 * 根据 env.useMock 决定使用真实 API 还是 Mock 数据
 */
export const usersApi = {
  /**
   * 获取用户列表（分页）
   */
  async list(params?: ListParams): Promise<ListResponse<User>> {
    if (env.useMock) {
      return userMock.list(params)
    }
    const res = await api.get<ApiResponse<ListResponse<User>>>('/users', { params })
    return res.data.data
  },

  /**
   * 根据 ID 获取用户
   */
  async getById(id: string): Promise<User> {
    if (env.useMock) {
      return userMock.detail(id)
    }
    const res = await api.get<ApiResponse<User>>(`/users/${id}`)
    return res.data.data
  },

  /**
   * 创建用户
   */
  async create(data: Partial<User>): Promise<User> {
    if (env.useMock) {
      return userMock.create(data)
    }
    const res = await api.post<ApiResponse<User>>('/users', data)
    return res.data.data
  },

  /**
   * 更新用户
   */
  async update(id: string, data: Partial<User>): Promise<User> {
    if (env.useMock) {
      return userMock.update(id, data)
    }
    const res = await api.put<ApiResponse<User>>(`/users/${id}`, data)
    return res.data.data
  },

  /**
   * 删除用户
   */
  async delete(id: string): Promise<void> {
    if (env.useMock) {
      return userMock.delete(id)
    }
    await api.delete<ApiResponse<void>>(`/users/${id}`)
  },

  /**
   * 为用户分配角色
   */
  async assignRoles(id: string, roles: string[]): Promise<User> {
    if (env.useMock) {
      return userMock.assignRoles(id, roles)
    }
    const res = await api.put<ApiResponse<User>>(`/users/${id}/roles`, { roles })
    return res.data.data
  },

  /**
   * 修改密码
   */
  async changePassword(data: { oldPassword: string; newPassword: string }): Promise<void> {
    if (env.useMock) {
      await userMock.changePassword(data)
      return
    }
    await api.post<ApiResponse<void>>('/user/password/change', data)
  },

  /**
   * 获取用户个人资料
   */
  async profile(): Promise<User> {
    if (env.useMock) {
      return userMock.profile()
    }
    const res = await api.get<ApiResponse<User>>('/user/profile')
    return res.data.data
  }
}
