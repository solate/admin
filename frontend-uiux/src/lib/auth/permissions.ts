/**
 * 权限管理工具
 * 提供权限检查相关的业务逻辑
 */

import { computed } from 'vue'
import { useAuthStore } from '@/stores/modules/auth'

/**
 * 权限定义
 */
export const PERMISSIONS = {
  // 租户管理
  TENANT_VIEW: 'tenant:view',
  TENANT_CREATE: 'tenant:create',
  TENANT_UPDATE: 'tenant:update',
  TENANT_DELETE: 'tenant:delete',

  // 用户管理
  USER_VIEW: 'user:view',
  USER_CREATE: 'user:create',
  USER_UPDATE: 'user:update',
  USER_DELETE: 'user-delete',
  USER_ASSIGN_ROLES: 'user:assign_roles',

  // 服务管理
  SERVICE_VIEW: 'service:view',
  SERVICE_CREATE: 'service:create',
  SERVICE_UPDATE: 'service:update',
  SERVICE_DELETE: 'service:delete',
  SERVICE_TOGGLE: 'service:toggle',

  // 系统设置
  SETTINGS_VIEW: 'settings:view',
  SETTINGS_UPDATE: 'settings:update',

  // 数据分析
  ANALYTICS_VIEW: 'analytics:view',

  // 通知管理
  NOTIFICATION_VIEW: 'notification:view',
  NOTIFICATION_MANAGE: 'notification:manage',
} as const

/**
 * 角色权限映射
 */
export const ROLE_PERMISSIONS = {
  admin: Object.values(PERMISSIONS),
  user: [
    PERMISSIONS.TENANT_VIEW,
    PERMISSIONS.USER_VIEW,
    PERMISSIONS.SERVICE_VIEW,
    PERMISSIONS.SETTINGS_VIEW,
    PERMISSIONS.ANALYTICS_VIEW,
    PERMISSIONS.NOTIFICATION_VIEW,
  ],
  guest: [
    PERMISSIONS.TENANT_VIEW,
  ],
} as const

/**
 * 权限类型
 */
export type Permission = (typeof PERMISSIONS)[keyof typeof PERMISSIONS]
export type Role = keyof typeof ROLE_PERMISSIONS

/**
 * 权限检查类
 */
export class PermissionChecker {
  /**
   * 检查用户是否拥有指定权限
   */
  static has(permission: Permission): boolean {
    const authStore = useAuthStore()
    if (!authStore.isAuthenticated) return false

    const userRoles = authStore.user?.roles || []
    const userPermissions = new Set<Permission>()

    // 收集用户所有角色的权限
    userRoles.forEach((role: string) => {
      const rolePermissions = ROLE_PERMISSIONS[role as Role]
      if (rolePermissions) {
        rolePermissions.forEach((p) => userPermissions.add(p))
      }
    })

    return userPermissions.has(permission)
  }

  /**
   * 检查用户是否拥有所有指定权限
   */
  static hasAll(permissions: Permission[]): boolean {
    return permissions.every((p) => this.has(p))
  }

  /**
   * 检查用户是否拥有任一指定权限
   */
  static hasAny(permissions: Permission[]): boolean {
    return permissions.some((p) => this.has(p))
  }

  /**
   * 获取用户的所有权限
   */
  static getUserPermissions(): Permission[] {
    const authStore = useAuthStore()
    if (!authStore.isAuthenticated) return []

    const userRoles = authStore.user?.roles || []
    const permissions = new Set<Permission>()

    userRoles.forEach((role: string) => {
      const rolePermissions = ROLE_PERMISSIONS[role as Role]
      if (rolePermissions) {
        rolePermissions.forEach((p) => permissions.add(p))
      }
    })

    return Array.from(permissions)
  }
}

/**
 * 权限检查 Composable
 * 用于在 Vue 组件中检查权限
 */
export function usePermissions() {
  const authStore = useAuthStore()

  // 用户的所有权限
  const userPermissions = computed(() => PermissionChecker.getUserPermissions())

  /**
   * 检查是否拥有指定权限
   */
  const hasPermission = (permission: Permission): boolean => {
    return PermissionChecker.has(permission)
  }

  /**
   * 检查是否拥有所有指定权限
   */
  const hasAllPermissions = (permissions: Permission[]): boolean => {
    return PermissionChecker.hasAll(permissions)
  }

  /**
   * 检查是否拥有任一指定权限
   */
  const hasAnyPermission = (permissions: Permission[]): boolean => {
    return PermissionChecker.hasAny(permissions)
  }

  /**
   * 检查是否为管理员
   */
  const isAdmin = computed(() => {
    return authStore.user?.roles?.includes('admin') ?? false
  })

  return {
    userPermissions,
    hasPermission,
    hasAllPermissions,
    hasAnyPermission,
    isAdmin,
  }
}
