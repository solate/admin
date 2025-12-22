import axios from 'axios'
import type { AxiosInstance, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import { getAccessToken, refreshAccessToken, clearTokens } from '../utils/token'

// 开发环境使用代理，生产环境使用环境变量
const baseURL = import.meta.env.MODE === 'development' ? '/api' : (import.meta.env.VITE_API_BASE_URL || '/api')

const http: AxiosInstance = axios.create({
  baseURL,
  timeout: 15000
})

// 是否正在刷新token的标志
let isRefreshing = false
// 重试队列，用于存储因token过期而失败的请求
let retryQueue: Array<{
  config: InternalAxiosRequestConfig
  resolve: (value: any) => void
  reject: (error: any) => void
}> = []

/**
 * 请求拦截器：自动添加token
 */
http.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = getAccessToken()
    if (token) {
      config.headers = config.headers || {}
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

/**
 * 响应拦截器：处理业务错误和token过期
 */
http.interceptors.response.use(
  (response: AxiosResponse) => {
    const data = response.data
    // 后端返回 code: 200 或 code: 0 都表示成功
    if (typeof data === 'object' && data && 'code' in data && data.code !== 0 && data.code !== 200) {
      ElMessage.error(data.msg || data.message || '请求失败')
      return Promise.reject(data)
    }
    return data?.data ?? data
  },
  async (error) => {
    const originalRequest = error.config
    const status = error?.response?.status

    // 处理401未授权错误（token过期或无效）
    if (status === 401 && !originalRequest._retry) {
      // 如果是刷新token的请求失败了，直接跳转登录
      if (originalRequest.url?.includes('/auth/refresh-token')) {
        clearTokens()
        ElMessage.error('登录已过期，请重新登录')
        setTimeout(() => {
          location.href = '/login'
        }, 1000)
        return Promise.reject(error)
      }

      // 标记该请求已重试过，避免无限循环
      originalRequest._retry = true

      if (!isRefreshing) {
        isRefreshing = true

        try {
          // 尝试刷新token
          const newToken = await refreshAccessToken()

          if (newToken) {
            // 刷新成功，更新原请求的token并重试
            originalRequest.headers.Authorization = `Bearer ${newToken}`

            // 处理队列中的请求
            retryQueue.forEach(({ config, resolve, reject }) => {
              config.headers.Authorization = `Bearer ${newToken}`
              http
                .request(config)
                .then((response) => resolve(response))
                .catch((err) => reject(err))
            })
            retryQueue = []

            // 重试原请求
            return http(originalRequest)
          } else {
            // 刷新失败，清除token并跳转登录
            clearTokens()
            ElMessage.error('登录已过期，请重新登录')
            retryQueue = []
            setTimeout(() => {
              location.href = '/login'
            }, 1000)
            return Promise.reject(error)
          }
        } catch (refreshError) {
          // 刷新失败
          clearTokens()
          retryQueue = []
          ElMessage.error('登录已过期，请重新登录')
          setTimeout(() => {
            location.href = '/login'
          }, 1000)
          return Promise.reject(refreshError)
        } finally {
          isRefreshing = false
        }
      } else {
        // 如果正在刷新token，将请求加入队列
        return new Promise((resolve, reject) => {
          retryQueue.push({ config: originalRequest, resolve, reject })
        })
      }
    }

    // 处理其他错误
    if (status === 403) {
      ElMessage.error('没有权限访问')
    } else if (status === 404) {
      ElMessage.error('请求的资源不存在')
    } else if (status === 500) {
      ElMessage.error('服务器错误')
    } else {
      ElMessage.error(error?.response?.data?.msg || error?.response?.data?.message || error.message || '网络错误')
    }

    return Promise.reject(error)
  }
)

export default http


