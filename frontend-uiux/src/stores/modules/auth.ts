// Authentication store

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authStorage } from '@/utils/storage'
import { authApi } from '@/api/modules/auth'
import type { User } from '@/types/models'
import type { LoginRequest, RegisterRequest } from '@/types/api'

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null)
  const token = ref<string | null>(authStorage.getToken())
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Computed
  const isAuthenticated = computed(() => !!token.value)
  // 确保始终返回有效的角色值，避免 null
  const userRole = computed(() => user.value?.role || 'user')
  const tenantId = computed(() => user.value?.tenantId || null)

  // Actions
  async function login(credentials: LoginRequest) {
    isLoading.value = true
    error.value = null

    try {
      const response = await authApi.login(credentials)
      user.value = response.user
      token.value = response.token
      authStorage.setToken(response.token)
      return response.user
    } catch (err) {
      error.value = err instanceof Error ? err.message : '登录失败'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function logout() {
    isLoading.value = true

    try {
      await authApi.logout()
      user.value = null
      token.value = null
      authStorage.removeToken()
    } catch (err) {
      error.value = err instanceof Error ? err.message : '登出失败'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function register(data: RegisterRequest) {
    isLoading.value = true
    error.value = null

    try {
      const response = await authApi.register(data)
      user.value = response.user
      token.value = response.token
      authStorage.setToken(response.token)
      return response.user
    } catch (err) {
      error.value = err instanceof Error ? err.message : '注册失败'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function fetchUser() {
    if (!token.value) return null

    isLoading.value = true
    error.value = null

    try {
      user.value = await authApi.me()
      return user.value
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取用户信息失败'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  return {
    // State
    user,
    token,
    isLoading,
    error,

    // Computed
    isAuthenticated,
    userRole,
    tenantId,

    // Actions
    login,
    logout,
    register,
    fetchUser
  }
})
