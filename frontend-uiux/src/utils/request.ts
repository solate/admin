// Axios instance and request utilities

import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'
import { env } from '@/config/env'
import type { ApiResponse, ApiError } from '@/types/api'

// 创建 axios 实例
export const api: AxiosInstance = axios.create({
  baseURL: env.apiBaseUrl,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

/**
 * 处理 API 错误
 */
const handleApiError = (error: any) => {
  if (!error.response) {
    // 网络错误或请求未发送
    console.error('Network error:', error.message)
    return Promise.reject(error)
  }

  const { status, data } = error.response

  switch (status) {
    case 401:
      // 未授权 - 清除 token 并跳转登录
      localStorage.removeItem('token')
      // 使用路由跳转（需要在实际使用时获取 router 实例）
      if (typeof window !== 'undefined') {
        window.location.href = '/login'
      }
      break

    case 403:
      console.error('Access forbidden:', data)
      break

    case 404:
      console.error('Resource not found:', error.config?.url)
      break

    case 500:
    case 502:
    case 503:
    case 504:
      console.error('Server error:', data)
      break

    default:
      console.error('API error:', status, data)
  }

  return Promise.reject(error)
}

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 从 localStorage 获取 token
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    // 添加租户上下文
    const tenantId = localStorage.getItem('currentTenantId')
    if (tenantId) {
      config.headers['X-Tenant-ID'] = tenantId
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => response,
  handleApiError
)

/**
 * 发送类型安全的请求
 */
export async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const response = await api(config)
  return response.data as T
}

export default api
