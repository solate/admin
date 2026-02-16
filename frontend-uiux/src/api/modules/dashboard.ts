// Dashboard API

import { api } from '@/utils/request'
import type { ApiResponse, ListParams } from '@/types/api'
import type { DashboardStats, ChartData, AuditLog } from '@/types/models'
import { isMockEnabled, mockDashboardHandlers } from '@/mock'

// Mock chart data
const mockChartData: ChartData = {
  labels: ['一月', '二月', '三月', '四月', '五月', '六月'],
  datasets: [
    {
      label: '用户增长',
      data: [120, 190, 300, 450, 520, 680],
      backgroundColor: 'rgba(59, 130, 246, 0.5)',
      borderColor: 'rgba(59, 130, 246, 1)'
    },
    {
      label: '收入 (万元)',
      data: [30, 45, 62, 85, 95, 120],
      backgroundColor: 'rgba(16, 185, 129, 0.5)',
      borderColor: 'rgba(16, 185, 129, 1)'
    }
  ]
}

// Mock audit logs
const mockAuditLogs: AuditLog[] = [
  {
    id: '1',
    userId: '1',
    userName: 'Admin User',
    action: '登录',
    resource: 'auth',
    details: { ip: '192.168.1.1' },
    ipAddress: '192.168.1.1',
    createdAt: new Date().toISOString()
  },
  {
    id: '2',
    userId: '1',
    userName: 'Admin User',
    action: '创建租户',
    resource: 'tenants',
    resourceId: '2',
    details: { tenantName: 'New Company' },
    createdAt: new Date(Date.now() - 3600000).toISOString()
  }
]

/**
 * Dashboard API
 */
export const dashboardApi = {
  /**
   * 获取仪表板统计数据
   */
  async stats(): Promise<DashboardStats> {
    if (isMockEnabled()) {
      return mockDashboardHandlers.stats()
    }
    const res = await api.get<ApiResponse<DashboardStats>>('/dashboard/stats')
    return res.data.data
  },

  /**
   * 获取仪表板图表数据
   */
  async charts(): Promise<ChartData> {
    if (isMockEnabled()) {
      return mockChartData
    }
    const res = await api.get<ApiResponse<ChartData>>('/dashboard/charts')
    return res.data.data
  },

  /**
   * 获取最近活动
   */
  async activity(params?: { limit?: number }): Promise<AuditLog[]> {
    if (isMockEnabled()) {
      return mockAuditLogs.slice(0, params?.limit || 10)
    }
    const res = await api.get<ApiResponse<AuditLog[]>>('/dashboard/activity', { params })
    return res.data.data
  }
}

/**
 * 角色 API
 */
export const rolesApi = {
  async list(params?: ListParams) {
    return api.get('/roles', { params })
  },

  async getById(id: string) {
    return api.get(`/roles/${id}`)
  },

  async create(data: any) {
    return api.post('/roles', data)
  },

  async update(id: string, data: any) {
    return api.put(`/roles/${id}`, data)
  },

  async delete(id: string) {
    return api.delete(`/roles/${id}`)
  }
}

/**
 * 权限 API
 */
export const permissionsApi = {
  async list() {
    return api.get('/permissions')
  },

  async getByRole(roleId: string) {
    return api.get(`/permissions/role/${roleId}`)
  }
}

/**
 * 审计日志 API
 */
export const auditLogsApi = {
  async list(params?: ListParams) {
    return api.get('/audit-logs', { params })
  },

  async getById(id: string) {
    return api.get(`/audit-logs/${id}`)
  }
}

/**
 * 通知 API
 */
export const notificationsApi = {
  async list(params?: ListParams) {
    return api.get('/notifications', { params })
  },

  async markAsRead(id: string) {
    return api.put(`/notifications/${id}/read`)
  },

  async markAllAsRead() {
    return api.put('/notifications/read-all')
  },

  async unreadCount() {
    return api.get('/notifications/unread-count')
  }
}

/**
 * 设置 API
 */
export const settingsApi = {
  async get(key: string) {
    return api.get(`/settings/${key}`)
  },

  async set(key: string, value: any) {
    return api.put(`/settings/${key}`, { value })
  },

  async getAll() {
    return api.get('/settings')
  }
}
