// Router types

import type { RouteRecordRaw } from 'vue-router'

/**
 * 应用路由元信息（与 router/types.ts 保持一致）
 */
export interface AppRouteMeta {
  title?: string
  requiresAuth?: boolean
  requiresTenant?: boolean
  requiredRoles?: string[]
  requiredPermissions?: string[]
  icon?: string
  hidden?: boolean
  keepAlive?: boolean
  transition?: string
  external?: string
  badge?: number | string
  order?: number
  // 兼容旧属性
  roles?: string[]
  layout?: string
  hideInMenu?: boolean
  /** 固定标签（不可关闭） */
  affix?: boolean
  /** 是否隐藏标签页 */
  hideTag?: boolean
}

declare module 'vue-router' {
  interface RouteMeta extends AppRouteMeta {}
}

// 使用交叉类型避免 children 属性类型冲突
export type AppRouteRecordRaw = RouteRecordRaw & {
  meta?: AppRouteMeta
  children?: AppRouteRecordRaw[]
}

// 兼容旧类型名称
export type MetaState = AppRouteMeta
