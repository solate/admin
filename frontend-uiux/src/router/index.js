import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useTenantsStore } from '@/stores/tenants'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    // Landing Page
    {
      path: '/',
      name: 'landing',
      component: () => import('@/views/LandingView.vue'),
      meta: { requiresAuth: false }
    },
    // Authentication
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/auth/LoginView.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/auth/RegisterView.vue'),
      meta: { requiresAuth: false }
    },
    // Dashboard Layout
    {
      path: '/dashboard',
      component: () => import('@/layouts/DashboardLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          redirect: '/dashboard/overview'
        },
        {
          path: 'overview',
          name: 'dashboard-overview',
          component: () => import('@/views/dashboard/OverviewView.vue'),
          meta: { title: 'Overview' }
        },
        // Tenant Management
        {
          path: 'tenants',
          name: 'tenants',
          component: () => import('@/views/tenants/TenantListView.vue'),
          meta: { title: 'Tenants' }
        },
        {
          path: 'tenants/create',
          name: 'tenant-create',
          component: () => import('@/views/tenants/TenantDetailView.vue'),
          meta: { title: 'Create Tenant' }
        },
        {
          path: 'tenants/:id',
          name: 'tenant-detail',
          component: () => import('@/views/tenants/TenantDetailView.vue'),
          meta: { title: 'Tenant Details' }
        },
        // Service Management
        {
          path: 'services',
          name: 'services',
          component: () => import('@/views/services/ServiceListView.vue'),
          meta: { title: 'Services' }
        },
        {
          path: 'services/:id',
          name: 'service-detail',
          component: () => import('@/views/services/ServiceDetailView.vue'),
          meta: { title: 'Service Details' }
        },
        // User Management
        {
          path: 'users',
          name: 'users',
          component: () => import('@/views/users/UserListView.vue'),
          meta: { title: 'Users' }
        },
        {
          path: 'users/create',
          name: 'user-create',
          component: () => import('@/views/users/UserDetailView.vue'),
          meta: { title: 'Create User' }
        },
        {
          path: 'users/:id',
          name: 'user-detail',
          component: () => import('@/views/users/UserDetailView.vue'),
          meta: { title: 'User Details' }
        },
        // Business & Analytics
        {
          path: 'business',
          name: 'business',
          component: () => import('@/views/business/BusinessView.vue'),
          meta: { title: 'Business' }
        },
        {
          path: 'analytics',
          name: 'analytics',
          component: () => import('@/views/analytics/AnalyticsView.vue'),
          meta: { title: 'Analytics' }
        },
        // Settings & Profile
        {
          path: 'settings',
          name: 'settings',
          component: () => import('@/views/settings/SettingsView.vue'),
          meta: { title: 'Settings' }
        },
        {
          path: 'profile',
          name: 'profile',
          component: () => import('@/views/profile/ProfileView.vue'),
          meta: { title: 'Profile' }
        },
        // Notifications
        {
          path: 'notifications',
          name: 'notifications',
          component: () => import('@/views/notifications/NotificationView.vue'),
          meta: { title: 'Notifications' }
        }
      ]
    },
    // 404 Page
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFoundView.vue')
    }
  ]
})

// Navigation guard
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  const tenantsStore = useTenantsStore()

  // Initialize tenant store
  tenantsStore.initialize()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'login', query: { redirect: to.fullPath } })
  } else if ((to.name === 'login' || to.name === 'register') && authStore.isAuthenticated) {
    next({ name: 'dashboard-overview' })
  } else {
    next()
  }
})

export default router
