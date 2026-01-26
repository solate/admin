/**
 * 租户守卫
 * 处理租户相关的路由拦截逻辑
 */

import type { Router } from 'vue-router'
import { useTenantsStore } from '@/stores/modules/tenants'

/**
 * 设置租户守卫
 */
export function setupTenantGuard(router: Router) {
  router.beforeEach((to, from, next) => {
    const tenantsStore = useTenantsStore()

    // 初始化租户 store
    tenantsStore.initialize()

    // 检查路由是否需要租户上下文
    const requiresTenant = to.meta.requiresTenant === true

    if (requiresTenant && !tenantsStore.currentTenant) {
      // 如果需要租户但没有当前租户，可以：
      // 1. 重定向到租户选择页
      // 2. 使用用户的第一个租户
      // 3. 显示错误提示

      const firstTenant = tenantsStore.tenants[0]
      if (firstTenant) {
        tenantsStore.setCurrentTenant(firstTenant)
      } else {
        // 没有可用租户，重定向到错误页或显示提示
        console.warn('No available tenant')
      }
    }

    next()
  })
}
