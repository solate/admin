// Authentication routes

import type { AppRouteRecordRaw } from '@/types/router'

export const authRoutes: AppRouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/auth/LoginView.vue'),
    meta: {
      requiresAuth: false,
      title: 'Login',
      hideInMenu: true
    }
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('@/views/auth/RegisterView.vue'),
    meta: {
      requiresAuth: false,
      title: 'Register',
      hideInMenu: true
    }
  }
]
