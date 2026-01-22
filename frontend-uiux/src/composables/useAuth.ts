// Authentication composable

import { computed } from 'vue'
import { useAuthStore } from '@/stores/modules/auth'
import { authStorage } from '@/utils/storage'
import type { User, LoginCredentials, RegisterData } from '@/types/models'

export function useAuth() {
  const authStore = useAuthStore()

  const user = computed(() => authStore.user)
  const token = computed(() => authStore.token)
  const isAuthenticated = computed(() => authStore.isAuthenticated)
  const userRole = computed(() => authStore.userRole)
  const tenantId = computed(() => authStore.tenantId)

  const login = async (credentials: LoginCredentials): Promise<User> => {
    const result = await authStore.login(credentials)
    return result as User
  }

  const logout = async (): Promise<void> => {
    await authStore.logout()
  }

  const register = async (data: RegisterData): Promise<User> => {
    const result = await authStore.register(data)
    return result as User
  }

  const refreshToken = async (): Promise<void> => {
    // Implementation for token refresh
    const currentToken = authStorage.getToken()
    if (currentToken) {
      // Call API to refresh token
      // await apiService.auth.refreshToken(currentToken)
    }
  }

  const hasRole = (roles: string | string[]): boolean => {
    const userRoles = [userRole.value]
    if (typeof roles === 'string') {
      return userRoles.includes(roles)
    }
    return roles.some((role) => userRoles.includes(role))
  }

  const hasPermission = (permission: string): boolean => {
    // Check if user has specific permission
    // Implementation depends on permission system
    return true
  }

  return {
    user,
    token,
    isAuthenticated,
    userRole,
    tenantId,
    login,
    logout,
    register,
    refreshToken,
    hasRole,
    hasPermission
  }
}
