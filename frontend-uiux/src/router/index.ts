// Router configuration

import { createRouter, createWebHistory } from 'vue-router'
import type { Router } from 'vue-router'
import { publicRoutes, authRoutes, dashboardRoutes } from './routes'
import { useAuthStore } from '@/stores/modules/auth'
import { useTenantsStore } from '@/stores/modules/tenants'

// Create router instance
const router: Router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [...publicRoutes, ...authRoutes, ...dashboardRoutes],
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    } else {
      return { top: 0 }
    }
  }
})

// Navigation guard
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  const tenantsStore = useTenantsStore()

  // Initialize tenant store
  tenantsStore.initialize()

  // Check if route requires authentication
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    // Redirect to login with redirect query
    next({
      name: 'login',
      query: { redirect: to.fullPath }
    })
  }
  // Redirect authenticated users away from login/register pages
  else if (
    (to.name === 'login' || to.name === 'register') &&
    authStore.isAuthenticated
  ) {
    next({ name: 'dashboard-overview' })
  }
  // Proceed to route
  else {
    next()
  }
})

// Update page title
router.afterEach((to) => {
  const title = to.meta.title
    ? `${to.meta.title} - Multi-Tenant SaaS`
    : 'Multi-Tenant SaaS'
  document.title = title
})

export default router
