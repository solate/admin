// Dashboard API

import { api } from '@/utils/request'
import { dashboardMock, roleMock, permissionMock, auditLogMock, notificationMock, settingMock } from '@/mock/handlers'
import { env } from '@/config/env'
import type { ApiResponse, ListParams } from '@/types/api'
import type { DashboardStats, ChartData, AuditLog } from '@/types/models'

/**
 * Dashboard API
 * 根据 env.useMock 决定使用真实 API 还是 Mock 数据
 */
export const dashboardApi = {
  /**
   * 获取仪表板统计数据
   */
  async stats(): Promise<DashboardStats> {
    if (env.useMock) {
      return dashboardMock.stats()
    }
    const res = await api.get<ApiResponse<DashboardStats>>('/dashboard/stats')
    return res.data.data
  },

  /**
   * 获取仪表板图表数据
   */
  async charts(): Promise<ChartData> {
    if (env.useMock) {
      return dashboardMock.charts()
    }
    const res = await api.get<ApiResponse<ChartData>>('/dashboard/charts')
    return res.data.data
  },

  /**
   * 获取最近活动
   */
  async activity(params?: { limit?: number }): Promise<AuditLog[]> {
    if (env.useMock) {
      return dashboardMock.activity(params)
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
    if (env.useMock) {
      return roleMock.list(params)
    }
    return api.get('/roles', { params })
  },

  async getById(id: string) {
    if (env.useMock) {
      return roleMock.detail(id)
    }
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
    if (env.useMock) {
      return permissionMock.list()
    }
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
    if (env.useMock) {
      return auditLogMock.list(params)
    }
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
    if (env.useMock) {
      return notificationMock.list(params)
    }
    return api.get('/notifications', { params })
  },

  async markAsRead(id: string) {
    if (env.useMock) {
      return notificationMock.markAsRead(id)
    }
    return api.put(`/notifications/${id}/read`)
  },

  async markAllAsRead() {
    if (env.useMock) {
      return notificationMock.markAllAsRead()
    }
    return api.put('/notifications/read-all')
  },

  async unreadCount() {
    if (env.useMock) {
      return notificationMock.unreadCount()
    }
    return api.get('/notifications/unread-count')
  }
}

/**
 * 设置 API
 */
export const settingsApi = {
  async get(key: string) {
    if (env.useMock) {
      return settingMock.get(key)
    }
    return api.get(`/settings/${key}`)
  },

  async set(key: string, value: any) {
    if (env.useMock) {
      return settingMock.set(key, value)
    }
    return api.put(`/settings/${key}`, { value })
  },

  async getAll() {
    if (env.useMock) {
      return settingMock.getAll()
    }
    return api.get('/settings')
  }
}
