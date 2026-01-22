// Dashboard routes (authentication required)

import type { AppRouteRecordRaw } from '@/types/router'

export const dashboardRoutes: AppRouteRecordRaw[] = [
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
        name: 'dashboard-overview',
        component: () => import('@/views/dashboard/OverviewView.vue'),
        meta: {
          title: 'Overview',
          icon: 'Dashboard'
        }
      },
      // Tenant Management
      {
        path: 'tenants',
        name: 'tenants',
        component: () => import('@/views/tenants/TenantListView.vue'),
        meta: {
          title: 'Tenants',
          icon: 'Building'
        }
      },
      {
        path: 'tenants/create',
        name: 'tenant-create',
        component: () => import('@/views/tenants/TenantDetailView.vue'),
        meta: {
          title: 'Create Tenant',
          hideInMenu: true
        }
      },
      {
        path: 'tenants/:id',
        name: 'tenant-detail',
        component: () => import('@/views/tenants/TenantDetailView.vue'),
        meta: {
          title: 'Tenant Details',
          hideInMenu: true
        }
      },
      // Service Management
      {
        path: 'services',
        name: 'services',
        component: () => import('@/views/services/ServiceListView.vue'),
        meta: {
          title: 'Services',
          icon: 'Grid'
        }
      },
      {
        path: 'services/:id',
        name: 'service-detail',
        component: () => import('@/views/services/ServiceDetailView.vue'),
        meta: {
          title: 'Service Details',
          hideInMenu: true
        }
      },
      // User Management
      {
        path: 'users',
        name: 'users',
        component: () => import('@/views/users/UserListView.vue'),
        meta: {
          title: 'Users',
          icon: 'Users'
        }
      },
      {
        path: 'users/create',
        name: 'user-create',
        component: () => import('@/views/users/UserDetailView.vue'),
        meta: {
          title: 'Create User',
          hideInMenu: true
        }
      },
      {
        path: 'users/:id',
        name: 'user-detail',
        component: () => import('@/views/users/UserDetailView.vue'),
        meta: {
          title: 'User Details',
          hideInMenu: true
        }
      },
      // Business & Analytics
      {
        path: 'business',
        name: 'business',
        component: () => import('@/views/business/BusinessView.vue'),
        meta: {
          title: 'Business',
          icon: 'Briefcase'
        }
      },
      {
        path: 'analytics',
        name: 'analytics',
        component: () => import('@/views/analytics/AnalyticsView.vue'),
        meta: {
          title: 'Analytics',
          icon: 'BarChart3'
        }
      },
      // Settings & Profile
      {
        path: 'settings',
        name: 'settings',
        component: () => import('@/views/settings/SettingsView.vue'),
        meta: {
          title: 'Settings',
          icon: 'Settings'
        }
      },
      {
        path: 'profile',
        name: 'profile',
        component: () => import('@/views/profile/ProfileView.vue'),
        meta: {
          title: 'Profile',
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
          title: 'Notifications',
          icon: 'Bell',
          hideInMenu: true
        }
      }
    ]
  }
]
