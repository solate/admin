// Dashboard routes (authentication required)

import type { AppRouteRecordRaw } from '@/types/router'

export const dashboardRoutes: AppRouteRecordRaw[] = [
  {
    path: '/redirect',
    component: () => import('@/layouts/DashboardLayout.vue'),
    meta: {
      requiresAuth: true
    },
    children: [
      {
        path: '/redirect/:path(.*)',
        name: 'redirect',
        component: () => import('@/views/RedirectView.vue'),
        meta: {
          hideTag: true,
          hideInMenu: true
        }
      }
    ]
  },
  {
    path: '/dashboard',
    component: () => import('@/layouts/DashboardLayout.vue'),
    meta: {
      requiresAuth: true
    },
    children: [
      {
        path: '',
        redirect: '/dashboard/overview'
      },
      {
        path: 'overview',
        name: 'overview',
        component: () => import('@/views/dashboard/OverviewView.vue'),
        meta: {
          title: 'nav.overview',
          icon: 'Dashboard'
        }
      },
      // Tenant Management
      {
        path: 'tenants',
        name: 'tenants',
        component: () => import('@/views/tenants/TenantListView.vue'),
        meta: {
          title: 'nav.tenants',
          icon: 'Building'
        }
      },
      {
        path: 'tenants/create',
        name: 'tenant-create',
        component: () => import('@/views/tenants/TenantDetailView.vue'),
        meta: {
          title: 'tenant.create',
          hideInMenu: true
        }
      },
      {
        path: 'tenants/:id',
        name: 'tenant-detail',
        component: () => import('@/views/tenants/TenantDetailView.vue'),
        meta: {
          title: 'tenant.detail',
          hideInMenu: true
        }
      },
      // Service Management
      {
        path: 'services',
        name: 'services',
        component: () => import('@/views/services/ServiceListView.vue'),
        meta: {
          title: 'nav.services',
          icon: 'Grid'
        }
      },
      {
        path: 'services/:id',
        name: 'service-detail',
        component: () => import('@/views/services/ServiceDetailView.vue'),
        meta: {
          title: 'service.detail',
          hideInMenu: true
        }
      },
      // User Management
      {
        path: 'users',
        name: 'users',
        component: () => import('@/views/users/UserListView.vue'),
        meta: {
          title: 'nav.users',
          icon: 'Users'
        }
      },
      {
        path: 'users/create',
        name: 'user-create',
        component: () => import('@/views/users/UserDetailView.vue'),
        meta: {
          title: 'user.detail.create',
          hideInMenu: true
        }
      },
      {
        path: 'users/:id',
        name: 'user-detail',
        component: () => import('@/views/users/UserDetailView.vue'),
        meta: {
          title: 'user.detail.edit',
          hideInMenu: true
        }
      },
      // Business & Analytics
      {
        path: 'business',
        name: 'business',
        component: () => import('@/views/business/BusinessView.vue'),
        meta: {
          title: 'nav.business',
          icon: 'Briefcase'
        }
      },
      {
        path: 'analytics',
        name: 'analytics',
        component: () => import('@/views/analytics/AnalyticsView.vue'),
        meta: {
          title: 'nav.analytics',
          icon: 'BarChart3'
        }
      },
      // Settings & Profile
      {
        path: 'settings',
        name: 'settings',
        component: () => import('@/views/settings/SettingsView.vue'),
        meta: {
          title: 'nav.settings',
          icon: 'Settings'
        }
      },
      {
        path: 'profile',
        name: 'profile',
        component: () => import('@/views/profile/ProfileView.vue'),
        meta: {
          title: 'nav.profile',
          icon: 'User',
          hideInMenu: true
        }
      },
      // Notifications
      {
        path: 'notifications',
        name: 'notifications',
        component: () => import('@/views/notifications/NotificationView.vue'),
        meta: {
          title: 'nav.notifications',
          icon: 'Bell',
          hideInMenu: true
        }
      }
    ]
  }
]
