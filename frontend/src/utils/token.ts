/**
 * Token 管理模块
 * 负责 access_token、refresh_token 和租户信息的存储、获取、刷新
 */

import { authApi, type TenantInfo, type RoleInfo } from '../api'

const TOKEN_KEY = 'access_token'
const REFRESH_TOKEN_KEY = 'refresh_token'
const USER_ID_KEY = 'user_id'
const USERNAME_KEY = 'username'
const EMAIL_KEY = 'email'
const PHONE_KEY = 'phone'
const TENANT_INFO_KEY = 'tenant_info' // 完整的租户信息
const ROLES_INFO_KEY = 'roles_info' // 角色列表

// Token刷新状态管理
let isRefreshing = false
let refreshSubscribers: Array<(token: string) => void> = []

/**
 * 保存token信息（登录成功后）
 */
export function saveTokens(data: {
  access_token: string
  refresh_token: string
  user_id?: string // 改为可选，支持先保存 token 再获取用户信息
  username?: string
  email?: string
  phone?: string
  tenant?: TenantInfo
  roles?: RoleInfo[]
}) {
  localStorage.setItem(TOKEN_KEY, data.access_token)
  localStorage.setItem(REFRESH_TOKEN_KEY, data.refresh_token)
  if (data.user_id !== undefined) {
    localStorage.setItem(USER_ID_KEY, data.user_id)
  }
  if (data.username !== undefined) {
    localStorage.setItem(USERNAME_KEY, data.username || '')
  }
  if (data.email !== undefined) {
    localStorage.setItem(EMAIL_KEY, data.email || '')
  }
  if (data.phone !== undefined) {
    localStorage.setItem(PHONE_KEY, data.phone || '')
  }
  if (data.tenant) {
    localStorage.setItem(TENANT_INFO_KEY, JSON.stringify(data.tenant))
  }
  if (data.roles) {
    localStorage.setItem(ROLES_INFO_KEY, JSON.stringify(data.roles))
  }
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
 * 获取用户ID
 */
export function getUserId(): string | null {
  return localStorage.getItem(USER_ID_KEY)
}

/**
 * 获取租户ID（上次选择的）
 */
export function getLastTenantId(): string | null {
  const tenantInfo = getTenantInfo()
  return tenantInfo?.tenant_id || null
}

/**
 * 获取租户信息
 */
export function getTenantInfo(): TenantInfo | null {
  const info = localStorage.getItem(TENANT_INFO_KEY)
  if (info) {
    try {
      return JSON.parse(info)
    } catch {
      return null
    }
  }
  return null
}

/**
 * 获取角色信息列表
 */
export function getRolesInfo(): RoleInfo[] {
  const info = localStorage.getItem(ROLES_INFO_KEY)
  if (info) {
    try {
      return JSON.parse(info)
    } catch {
      return []
    }
  }
  return []
}

/**
 * 获取用户信息
 */
export function getUserInfo() {
  const tenantInfo = getTenantInfo()
  const rolesInfo = getRolesInfo()
  return {
    user_id: localStorage.getItem(USER_ID_KEY),
    user_name: localStorage.getItem(USERNAME_KEY),
    email: localStorage.getItem(EMAIL_KEY),
    phone: localStorage.getItem(PHONE_KEY),
    tenant_id: tenantInfo?.tenant_id || null,
    tenant: tenantInfo,
    roles: rolesInfo
  }
}

/**
 * 清除所有token
 */
export function clearTokens() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
  localStorage.removeItem(USER_ID_KEY)
  localStorage.removeItem(USERNAME_KEY)
  localStorage.removeItem(EMAIL_KEY)
  localStorage.removeItem(PHONE_KEY)
  localStorage.removeItem(TENANT_INFO_KEY)
  localStorage.removeItem(ROLES_INFO_KEY)
}

/**
 * 解析JWT token获取过期时间
 * @throws 如果 token 格式无效则抛出错误
 */
function parseJwt(token: string): { exp?: number } {
  try {
    // 去除 Bearer 前缀
    const cleanToken = token.replace(/^Bearer\s+/i, '').trim()

    const parts = cleanToken.split('.')
    if (parts.length !== 3) {
      throw new Error('Invalid token format: token must have 3 parts separated by dots')
    }
    const base64Url = parts[1]
    if (!base64Url) {
      throw new Error('Invalid token format: missing payload')
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
    throw new Error(`Failed to parse JWT token: ${e instanceof Error ? e.message : 'Unknown error'}`)
  }
}

/**
 * 检查token是否即将过期（5分钟内）
 */
export function isTokenExpiringSoon(): boolean {
  const token = getAccessToken()
  if (!token) return true

  try {
    const payload = parseJwt(token)
    if (!payload.exp) return true

    const now = Math.floor(Date.now() / 1000)
    const timeLeft = payload.exp - now

    // 如果剩余时间少于5分钟，认为即将过期
    return timeLeft < 300
  } catch {
    // token 格式无效，视为需要刷新
    return true
  }
}

/**
 * 检查token是否已过期
 */
export function isTokenExpired(): boolean {
  const token = getAccessToken()
  if (!token) return true

  try {
    const payload = parseJwt(token)
    if (!payload.exp) return true

    const now = Math.floor(Date.now() / 1000)
    return payload.exp <= now
  } catch {
    // token 格式无效，视为已过期
    return true
  }
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
    console.warn('刷新token失败: refresh_token 不存在')
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

    // 更新租户信息
    if (response.tenant) {
      localStorage.setItem(TENANT_INFO_KEY, JSON.stringify(response.tenant))
    }
    // 更新角色信息
    if (response.roles) {
      localStorage.setItem(ROLES_INFO_KEY, JSON.stringify(response.roles))
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
  getUserId,
  getLastTenantId,
  getTenantInfo,
  getRolesInfo,
  getUserInfo,
  clearTokens,
  isTokenExpired,
  isTokenExpiringSoon,
  refreshAccessToken,
  ensureValidToken
}
