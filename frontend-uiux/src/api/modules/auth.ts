// Authentication API

import { api } from '@/utils/request'
import type { ApiResponse, LoginRequest, RegisterRequest, AuthResponse } from '@/types/api'
import type { User } from '@/types/models'
import { isMockEnabled, mockAuthHandlers } from '@/mock'

/**
 * 认证 API
 */
export const authApi = {
  /**
   * 用户登录
   */
  async login(credentials: LoginRequest): Promise<AuthResponse> {
    if (isMockEnabled()) {
      return mockAuthHandlers.login(credentials)
    }
    const res = await api.post<ApiResponse<AuthResponse>>('/auth/login', credentials)
    return res.data.data
  },

  /**
   * 用户登出
   */
  async logout(): Promise<void> {
    if (isMockEnabled()) {
      return mockAuthHandlers.logout()
    }
    await api.post<ApiResponse<void>>('/auth/logout')
  },

  /**
   * 用户注册
   */
  async register(data: RegisterRequest): Promise<AuthResponse> {
    if (isMockEnabled()) {
      return mockAuthHandlers.register(data)
    }
    const res = await api.post<ApiResponse<AuthResponse>>('/auth/register', data)
    return res.data.data
  },

  /**
   * 刷新访问令牌
   */
  async refreshToken(refreshToken: string): Promise<{ token: string }> {
    if (isMockEnabled()) {
      return { token: 'mock-refreshed-token-' + Date.now() }
    }
    const res = await api.post<ApiResponse<{ token: string }>>('/auth/refresh', { refreshToken })
    return res.data.data
  },

  /**
   * 获取当前用户信息
   */
  async me(): Promise<User> {
    if (isMockEnabled()) {
      const token = localStorage.getItem('token') || ''
      return mockAuthHandlers.me(token)
    }
    const res = await api.get<ApiResponse<User>>('/auth/me')
    return res.data.data
  }
}
