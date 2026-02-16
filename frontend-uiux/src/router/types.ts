/**
 * 路由相关类型定义
 */

import type { RouteRecordRaw } from 'vue-router'

/**
 * 应用路由元信息
 */
export interface AppRouteMeta {
  // 路由标题
  title?: string

  // 是否需要认证
  requiresAuth?: boolean

  // 是否需要租户上下文
  requiresTenant?: boolean

  // 需要的角色列表（用户至少拥有其中一个角色才能访问）
  requiredRoles?: string[]

  // 需要的权限列表（用户必须拥有所有权限才能访问）
  requiredPermissions?: string[]

  // 图标
  icon?: string

  // 是否在菜单中隐藏
  hidden?: boolean

  // 是否缓存页面
  keepAlive?: boolean

  // 页面过渡动画
  transition?: string

  // 外链
  external?: string

  // 徽章数量
  badge?: number | string

  // 是否排序（菜单排序）
  order?: number
}

/**
 * 扩展的路由记录类型
 */
export type AppRouteRecordRaw = RouteRecordRaw & {
  meta?: AppRouteMeta
  children?: AppRouteRecordRaw[]
}

/**
 * 菜单项类型
 */
export interface MenuItem {
  name: string
  path: string
  title?: string
  icon?: string
  badge?: number | string
  order?: number
  hidden?: boolean
  children?: MenuItem[]
  external?: string
}

/**
 * 面包屑项类型
 */
export interface BreadcrumbItem {
  title: string
  path?: string
  disabled?: boolean
}
