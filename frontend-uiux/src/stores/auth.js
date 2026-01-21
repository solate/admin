import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)
  const token = ref(localStorage.getItem('token') || null)

  const isAuthenticated = computed(() => !!token.value)
  const userRole = computed(() => user.value?.role || null)
  const tenantId = computed(() => user.value?.tenantId || null)

  function login(credentials) {
    // Mock login - replace with actual API call
    return new Promise((resolve) => {
      setTimeout(() => {
        user.value = {
          id: '1',
          name: credentials.email.split('@')[0],
          email: credentials.email,
          role: 'admin',
          tenantId: 'tenant-1'
        }
        token.value = 'mock-jwt-token-' + Date.now()
        localStorage.setItem('token', token.value)
        resolve(user.value)
      }, 500)
    })
  }

  function logout() {
    user.value = null
    token.value = null
    localStorage.removeItem('token')
  }

  function register(data) {
    // Mock registration - replace with actual API call
    return new Promise((resolve) => {
      setTimeout(() => {
        user.value = {
          id: Date.now().toString(),
          name: data.name,
          email: data.email,
          role: 'user',
          tenantId: null
        }
        token.value = 'mock-jwt-token-' + Date.now()
        localStorage.setItem('token', token.value)
        resolve(user.value)
      }, 500)
    })
  }

  return {
    user,
    token,
    isAuthenticated,
    userRole,
    tenantId,
    login,
    logout,
    register
  }
})
