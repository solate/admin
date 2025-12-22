/**
 * Token 管理模块
 * 负责 access_token 和 refresh_token 的存储、获取、刷新
 */

import { authApi } from '../api'

const TOKEN_KEY = 'access_token'
const REFRESH_TOKEN_KEY = 'refresh_token'
const USER_ID_KEY = 'user_id'
const USER_NAME_KEY = 'user_name'

// Token刷新状态管理
let isRefreshing = false
let refreshSubscribers: Array<(token: string) => void> = []

/**
 * 保存token信息
 */
export function saveTokens(data: {
  access_token: string
  refresh_token: string
  user_id: string
  user_name: string
}) {
  localStorage.setItem(TOKEN_KEY, data.access_token)
  localStorage.setItem(REFRESH_TOKEN_KEY, data.refresh_token)
  localStorage.setItem(USER_ID_KEY, data.user_id)
  localStorage.setItem(USER_NAME_KEY, data.user_name)
}

/**
 * 获取 access_token
 */
export function getAccessToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

/**
 * 获取 refresh_token
 */
export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

/**
 * 获取用户信息
 */
export function getUserInfo() {
  return {
    user_id: localStorage.getItem(USER_ID_KEY),
    user_name: localStorage.getItem(USER_NAME_KEY)
  }
}

/**
 * 清除所有token
 */
export function clearTokens() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
  localStorage.removeItem(USER_ID_KEY)
  localStorage.removeItem(USER_NAME_KEY)
}

/**
 * 解析JWT token获取过期时间
 */
function parseJwt(token: string): { exp?: number } {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) {
      return {}
    }
    const base64Url = parts[1]
    if (!base64Url) {
      return {}
    }
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    )
    return JSON.parse(jsonPayload)
  } catch (e) {
    return {}
  }
}

/**
 * 检查token是否即将过期（5分钟内）
 */
export function isTokenExpiringSoon(): boolean {
  const token = getAccessToken()
  if (!token) return true

  const payload = parseJwt(token)
  if (!payload.exp) return true

  const now = Math.floor(Date.now() / 1000)
  const timeLeft = payload.exp - now
  
  // 如果剩余时间少于5分钟，认为即将过期
  return timeLeft < 300
}

/**
 * 检查token是否已过期
 */
export function isTokenExpired(): boolean {
  const token = getAccessToken()
  if (!token) return true

  const payload = parseJwt(token)
  if (!payload.exp) return true

  const now = Math.floor(Date.now() / 1000)
  return payload.exp <= now
}

/**
 * 订阅token刷新完成事件
 */
function subscribeTokenRefresh(callback: (token: string) => void) {
  refreshSubscribers.push(callback)
}

/**
 * 通知所有订阅者token已刷新
 */
function onTokenRefreshed(token: string) {
  refreshSubscribers.forEach((callback) => callback(token))
  refreshSubscribers = []
}

/**
 * 刷新token
 * 返回新的 access_token，如果刷新失败返回 null
 */
export async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) {
    return null
  }

  // 如果正在刷新，返回一个Promise，等待刷新完成
  if (isRefreshing) {
    return new Promise((resolve) => {
      subscribeTokenRefresh((token: string) => {
        resolve(token)
      })
    })
  }

  isRefreshing = true

  try {
    const response = await authApi.refreshToken({ refresh_token: refreshToken })
    
    // 更新本地token
    const newAccessToken = response.access_token
    localStorage.setItem(TOKEN_KEY, newAccessToken)
    
    // 如果返回了新的refresh_token，也更新它
    if (response.refresh_token) {
      localStorage.setItem(REFRESH_TOKEN_KEY, response.refresh_token)
    }

    // 通知所有等待的请求
    onTokenRefreshed(newAccessToken)
    
    return newAccessToken
  } catch (error) {
    console.error('刷新token失败:', error)
    // 刷新失败，清除所有token
    clearTokens()
    return null
  } finally {
    isRefreshing = false
  }
}

/**
 * 检查并在需要时刷新token
 * 对业务代码透明
 */
export async function ensureValidToken(): Promise<boolean> {
  // 如果token不存在，返回false
  if (!getAccessToken()) {
    return false
  }

  // 如果token已过期或即将过期，尝试刷新
  if (isTokenExpired() || isTokenExpiringSoon()) {
    const newToken = await refreshAccessToken()
    return newToken !== null
  }

  return true
}

export default {
  saveTokens,
  getAccessToken,
  getRefreshToken,
  getUserInfo,
  clearTokens,
  isTokenExpired,
  isTokenExpiringSoon,
  refreshAccessToken,
  ensureValidToken
}

