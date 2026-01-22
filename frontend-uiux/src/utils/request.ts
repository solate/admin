// Axios instance and request utilities

import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse } from 'axios'
import type { ApiResponse, ApiError } from '@/types/api'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

// Create axios instance
export const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    // Add tenant context from localStorage
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

// Response interceptor
api.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error) => {
    if (error.response) {
      // Handle 401 - Unauthorized
      if (error.response.status === 401) {
        localStorage.removeItem('token')
        window.location.href = '/login'
      }

      // Handle 403 - Forbidden
      if (error.response.status === 403) {
        console.error('Access forbidden:', error.response.data)
      }

      // Handle 404 - Not Found
      if (error.response.status === 404) {
        console.error('Resource not found:', error.config.url)
      }

      // Handle 500 - Server Error
      if (error.response.status >= 500) {
        console.error('Server error:', error.response.data)
      }
    }

    return Promise.reject(error)
  }
)

// Helper function to make typed requests
export async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const response = await api(config)
  return response.data as T
}

export default api
