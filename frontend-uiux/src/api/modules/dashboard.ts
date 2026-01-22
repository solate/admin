// Dashboard API

import { api } from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { DashboardStats, ChartData, AuditLog } from '@/types/models'

export const dashboardApi = {
  /**
   * Get dashboard statistics
   */
  stats: () => api.get<ApiResponse<DashboardStats>>('/dashboard/stats'),

  /**
   * Get dashboard chart data
   */
  charts: () => api.get<ApiResponse<ChartData>>('/dashboard/charts'),

  /**
   * Get recent activity
   */
  activity: (params?: { limit?: number }) =>
    api.get<ApiResponse<AuditLog[]>>('/dashboard/activity', { params })
}

// Additional API modules that can be added later
export const rolesApi = {
  list: (params?: any) => api.get('/roles', { params }),
  getById: (id: string) => api.get(`/roles/${id}`),
  create: (data: any) => api.post('/roles', data),
  update: (id: string, data: any) => api.put(`/roles/${id}`, data),
  delete: (id: string) => api.delete(`/roles/${id}`)
}

export const permissionsApi = {
  list: () => api.get('/permissions'),
  getByRole: (roleId: string) => api.get(`/permissions/role/${roleId}`)
}

export const auditLogsApi = {
  list: (params?: any) => api.get('/audit-logs', { params }),
  getById: (id: string) => api.get(`/audit-logs/${id}`)
}

export const notificationsApi = {
  list: (params?: any) => api.get('/notifications', { params }),
  markAsRead: (id: string) => api.put(`/notifications/${id}/read`),
  markAllAsRead: () => api.put('/notifications/read-all'),
  unreadCount: () => api.get('/notifications/unread-count')
}

export const settingsApi = {
  get: (key: string) => api.get(`/settings/${key}`),
  set: (key: string, value: any) => api.put(`/settings/${key}`, { value }),
  getAll: () => api.get('/settings')
}
