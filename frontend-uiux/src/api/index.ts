// API module exports

export { api, request } from '@/utils/request'
export * from './modules/auth'
export * from './modules/tenants'
export * from './modules/users'
export * from './modules/services'
export * from './modules/dashboard'

// Import all API modules
import { authApi } from './modules/auth'
import { tenantsApi } from './modules/tenants'
import { usersApi } from './modules/users'
import { servicesApi } from './modules/services'
import { dashboardApi, rolesApi, permissionsApi, auditLogsApi, notificationsApi, settingsApi } from './modules/dashboard'

// Legacy API service export for backward compatibility
export const apiService = {
  auth: authApi,
  tenants: tenantsApi,
  users: usersApi,
  services: servicesApi,
  roles: rolesApi,
  permissions: permissionsApi,
  auditLogs: auditLogsApi,
  notifications: notificationsApi,
  dashboard: dashboardApi,
  settings: settingsApi
}

export { api as default } from '@/utils/request'
