// Router types

import type { RouteRecordRaw } from 'vue-router'

export interface MetaState {
  requiresAuth?: boolean
  title?: string
  roles?: string[]
  layout?: string
  hideInMenu?: boolean
}

declare module 'vue-router' {
  interface RouteMeta extends MetaState {}
}

export interface AppRouteRecordRaw extends Omit<RouteRecordRaw, 'meta'> {
  meta?: MetaState
  children?: AppRouteRecordRaw[]
}
