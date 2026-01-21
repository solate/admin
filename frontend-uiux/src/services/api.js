import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

// Create axios instance
export const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor - Add auth token
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

// Response interceptor - Handle errors
api.interceptors.response.use(
  (response) => response,
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

// API Service Methods
export const apiService = {
  // Auth
  auth: {
    login: (credentials) => api.post('/auth/login', credentials),
    logout: () => api.post('/auth/logout'),
    register: (data) => api.post('/auth/register', data),
    refreshToken: (refreshToken) => api.post('/auth/refresh', { refreshToken }),
    me: () => api.get('/auth/me')
  },

  // Tenants
  tenants: {
    list: (params) => api.get('/tenants', { params }),
    getAll: () => api.get('/tenants/all'),
    getById: (id) => api.get(`/tenants/${id}`),
    create: (data) => api.post('/tenants', data),
    update: (id, data) => api.put(`/tenants/${id}`, data),
    delete: (id) => api.delete(`/tenants/${id}`),
    updateStatus: (id, status) => api.put('/tenants/status', { id, status })
  },

  // Users
  users: {
    list: (params) => api.get('/users', { params }),
    getById: (id) => api.get(`/users/${id}`),
    create: (data) => api.post('/users', data),
    update: (id, data) => api.put(`/users/${id}`, data),
    delete: (id) => api.delete(`/users/${id}`),
    assignRoles: (id, roles) => api.put(`/users/${id}/roles`, { roles }),
    changePassword: (data) => api.post('/user/password/change', data),
    profile: () => api.get('/user/profile')
  },

  // Roles
  roles: {
    list: (params) => api.get('/roles', { params }),
    getById: (id) => api.get(`/roles/${id}`),
    create: (data) => api.post('/roles', data),
    update: (id, data) => api.put(`/roles/${id}`, data),
    delete: (id) => api.delete(`/roles/${id}`)
  },

  // Permissions
  permissions: {
    list: () => api.get('/permissions'),
    getByRole: (roleId) => api.get(`/permissions/role/${roleId}`)
  },

  // Audit Logs
  auditLogs: {
    list: (params) => api.get('/audit-logs', { params }),
    getById: (id) => api.get(`/audit-logs/${id}`)
  },

  // Services
  services: {
    list: (params) => api.get('/services', { params }),
    getById: (id) => api.get(`/services/${id}`),
    create: (data) => api.post('/services', data),
    update: (id, data) => api.put(`/services/${id}`, data),
    delete: (id) => api.delete(`/services/${id}`),
    toggle: (id, enabled) => api.put(`/services/${id}/toggle`, { enabled })
  },

  // Notifications
  notifications: {
    list: (params) => api.get('/notifications', { params }),
    markAsRead: (id) => api.put(`/notifications/${id}/read`),
    markAllAsRead: () => api.put('/notifications/read-all'),
    unreadCount: () => api.get('/notifications/unread-count')
  },

  // Dashboard
  dashboard: {
    stats: () => api.get('/dashboard/stats'),
    charts: () => api.get('/dashboard/charts'),
    activity: () => api.get('/dashboard/activity')
  },

  // Settings
  settings: {
    get: (key) => api.get(`/settings/${key}`),
    set: (key, value) => api.put(`/settings/${key}`, { value }),
    getAll: () => api.get('/settings')
  }
}

export default api
