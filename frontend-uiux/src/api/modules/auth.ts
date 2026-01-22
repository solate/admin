// Authentication API

import { api } from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { User, LoginCredentials, RegisterData } from '@/types/models'

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  name: string
  email: string
  password: string
}

export interface AuthResponse {
  user: User
  token: string
  refreshToken: string
}

export const authApi = {
  /**
   * User login
   */
  login: (credentials: LoginRequest) =>
    api.post<ApiResponse<AuthResponse>>('/auth/login', credentials),

  /**
   * User logout
   */
  logout: () => api.post<ApiResponse<void>>('/auth/logout'),

  /**
   * User registration
   */
  register: (data: RegisterRequest) =>
    api.post<ApiResponse<AuthResponse>>('/auth/register', data),

  /**
   * Refresh access token
   */
  refreshToken: (refreshToken: string) =>
    api.post<ApiResponse<{ token: string }>>('/auth/refresh', { refreshToken }),

  /**
   * Get current user info
   */
  me: () => api.get<ApiResponse<User>>('/auth/me')
}
