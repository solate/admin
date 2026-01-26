// Router configuration

import { createRouter, createWebHistory } from 'vue-router'
import type { Router } from 'vue-router'
import { publicRoutes, authRoutes, dashboardRoutes } from './routes'
import { setupRouterGuards } from './guards'

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

// Setup router guards
setupRouterGuards(router)

export default router
