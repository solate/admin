// Authentication store

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authStorage } from '@/utils/storage'
import { authApi } from '@/api/modules/auth'
import type { User } from '@/types/models'

export interface LoginCredentials {
  email: string
  password: string
}

export interface RegisterData {
  name: string
  email: string
  password: string
}

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null)
  const token = ref<string | null>(authStorage.getToken())
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Computed
  const isAuthenticated = computed(() => !!token.value)
  const userRole = computed(() => user.value?.role || null)
  const tenantId = computed(() => user.value?.tenantId || null)

  // Actions
  async function login(credentials: LoginCredentials) {
    isLoading.value = true
    error.value = null

    try {
      // Mock implementation - replace with actual API call
      // const response = await authApi.login(credentials)
      // const { user: userData, token: accessToken } = response.data

      // Mock login
      await new Promise((resolve) => setTimeout(resolve, 500))
      const userData: User = {
        id: '1',
        name: credentials.email.split('@')[0],
        email: credentials.email,
        role: 'admin',
        tenantId: 'tenant-1'
      }
      const accessToken = 'mock-jwt-token-' + Date.now()

      user.value = userData
      token.value = accessToken
      authStorage.setToken(accessToken)

      return userData
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Login failed'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function logout() {
    isLoading.value = true

    try {
      // Mock implementation - replace with actual API call
      // await authApi.logout()

      user.value = null
      token.value = null
      authStorage.removeToken()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Logout failed'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function register(data: RegisterData) {
    isLoading.value = true
    error.value = null

    try {
      // Mock implementation - replace with actual API call
      // const response = await authApi.register(data)
      // const { user: userData, token: accessToken } = response.data

      // Mock registration
      await new Promise((resolve) => setTimeout(resolve, 500))
      const userData: User = {
        id: Date.now().toString(),
        name: data.name,
        email: data.email,
        role: 'user',
        tenantId: null
      }
      const accessToken = 'mock-jwt-token-' + Date.now()

      user.value = userData
      token.value = accessToken
      authStorage.setToken(accessToken)

      return userData
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Registration failed'
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
      // Mock implementation - replace with actual API call
      // const response = await authApi.me()
      // user.value = response.data

      return user.value
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch user'
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
