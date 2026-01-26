/**
 * 路由守卫统一导出和设置
 */

import type { Router } from 'vue-router'
import { setupAuthGuard } from './auth'
import { setupTenantGuard } from './tenant'
import { setupTitleGuard } from './title'

/**
 * 设置所有路由守卫
 */
export function setupRouterGuards(router: Router) {
  // 按顺序设置守卫
  setupAuthGuard(router)
  setupTenantGuard(router)
  setupTitleGuard(router)
}

// 导出单个守卫函数，方便单独使用
export { setupAuthGuard } from './auth'
export { setupTenantGuard } from './tenant'
export { setupTitleGuard } from './title'
