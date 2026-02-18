/**
 * 认证守卫
 * 处理用户认证相关的路由拦截逻辑
 */

import type { Router } from 'vue-router'
import { useAuthStore } from '@/stores/modules/auth'

/**
 * 设置认证守卫
 */
export function setupAuthGuard(router: Router) {
  router.beforeEach((to, from, next) => {
    const authStore = useAuthStore()

    // 检查路由是否需要认证
    const requiresAuth = to.meta.requiresAuth === true

    // 未认证用户访问需要认证的路由
    if (requiresAuth && !authStore.isAuthenticated) {
      // 重定向到登录页，并携带原始目标路径
      return next({
        name: 'login',
        query: { redirect: to.fullPath },
      })
    }

    // 已认证用户访问登录/注册页，重定向到仪表板
    if (
      (to.name === 'login' || to.name === 'register') &&
      authStore.isAuthenticated
    ) {
      return next({ name: 'overview' })
    }

    // 检查权限（如果路由定义了 requiredRoles）
    if (to.meta.requiredRoles && Array.isArray(to.meta.requiredRoles)) {
      const userRole = authStore.user?.role
      const hasRequiredRole = userRole && to.meta.requiredRoles.includes(userRole)

      if (!hasRequiredRole) {
        return next({ name: 'overview' }) // 或者 403 页面
      }
    }

    next()
  })
}
